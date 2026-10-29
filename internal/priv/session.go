package priv

import (
	"context"
	"flag"
	"fmt"
	"io"
	"regexp"
	"slices"
	"strconv"
	"strings"

	"github.com/alfarax9/astra/internal/domain"
)

const (
	systemdRun = "/usr/bin/systemd-run"
	systemctl  = "/usr/bin/systemctl"
)

// pathElement is one folder name of a session's working directory. It must start with a
// letter or digit, which also rules out "." and "..".
var pathElement = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]*$`)

// sessionRequest is a validated session-start.
type sessionRequest struct {
	slug, id, agent, dir string
	rows, cols           int
	memoryMB, cpuPercent int
}

func parseSessionStart(args []string) (sessionRequest, error) {
	r := sessionRequest{}
	if len(args) < 2 || strings.HasPrefix(args[0], "-") || strings.HasPrefix(args[1], "-") {
		return r, fmt.Errorf("%w: session-start takes <slug> <id> and options", ErrUsage)
	}
	r.slug, r.id = args[0], args[1]
	fs := flag.NewFlagSet("session-start", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	fs.StringVar(&r.agent, "agent", "", "claude, codex or shell")
	fs.StringVar(&r.dir, "dir", "", "folder under work/")
	fs.IntVar(&r.rows, "rows", defaultRows, "terminal rows")
	fs.IntVar(&r.cols, "cols", defaultCols, "terminal columns")
	fs.IntVar(&r.memoryMB, "memory-max", defaultMemoryMB, "memory limit in MB")
	fs.IntVar(&r.cpuPercent, "cpu-quota", defaultCPU, "CPU limit in percent of one core")
	if err := fs.Parse(args[2:]); err != nil || fs.NArg() != 0 {
		return r, fmt.Errorf("%w: unexpected option or argument", ErrUsage)
	}
	return r, r.validate()
}

func (r sessionRequest) validate() error {
	switch {
	case !domain.ValidSlug(r.slug):
		return fmt.Errorf("%w: invalid workspace slug", ErrUsage)
	case !domain.ValidSessionID(r.id):
		return fmt.Errorf("%w: invalid session id", ErrUsage)
	case r.rows < minTerminal || r.rows > maxTerminal || r.cols < minTerminal || r.cols > maxTerminal:
		return fmt.Errorf("%w: terminal size must be %d to %d", ErrUsage, minTerminal, maxTerminal)
	case r.memoryMB < minMemoryMB || r.memoryMB > maxMemoryMB:
		return fmt.Errorf("%w: memory-max must be %d to %d MB", ErrUsage, minMemoryMB, maxMemoryMB)
	case r.cpuPercent < minCPUPercent || r.cpuPercent > maxCPUPercent:
		return fmt.Errorf("%w: cpu-quota must be %d to %d percent", ErrUsage, minCPUPercent, maxCPUPercent)
	}
	if r.dir != "" {
		for _, part := range strings.Split(r.dir, "/") {
			if !pathElement.MatchString(part) {
				return fmt.Errorf("%w: dir must be a relative path of plain folder names", ErrUsage)
			}
		}
	}
	return nil
}

func unitName(slug, id string) string { return "astra-s-" + slug + "-" + id }

// sessionStart runs the agent's session host as a transient unit of the workspace's agent
// user, inside the workspace's slice with resource limits. It prints the socket path.
func (h Helper) sessionStart(ctx context.Context, args []string, out io.Writer) error {
	r, err := parseSessionStart(args)
	if err != nil {
		return err
	}
	spec, ok := h.Config.Agents[r.agent]
	if !ok {
		return fmt.Errorf("%w: agent must be one of %s", ErrUsage, strings.Join(h.agentKinds(), ", "))
	}
	if _, err := h.Host.LookupAgent(agentUser(r.slug)); err != nil {
		return fmt.Errorf("workspace %s has no agent user (run workspace-add): %w", r.slug, err)
	}
	dir, err := h.Host.ResolveDir(h.Config.workDir(r.slug), r.dir)
	if err != nil {
		return fmt.Errorf("%w: working directory: %w", ErrUsage, err)
	}
	sock := h.Config.sessionsDir(r.slug) + "/" + r.id + ".sock"
	if output, err := h.Host.Exec(ctx, h.unitCommand(r, dir, sock, spec)...); err != nil {
		return fmt.Errorf("start session %s: %w: %s", r.id, err, strings.TrimSpace(string(output)))
	}
	_, err = fmt.Fprintln(out, sock)
	return err
}

// unitCommand builds the systemd-run command line. The agent's environment is listed here
// in full, so nothing the daemon knows reaches it and no secret appears in a process list.
func (h Helper) unitCommand(r sessionRequest, dir, sock string, spec AgentSpec) []string {
	agent := agentUser(r.slug)
	argv := []string{
		systemdRun, "--quiet", "--collect",
		"--unit=" + unitName(r.slug, r.id),
		"--slice=astra-agents-" + r.slug + ".slice",
		"--uid=" + agent, "--gid=" + agent,
		"--working-directory=" + dir,
		"--property=MemoryMax=" + strconv.Itoa(r.memoryMB) + "M",
		"--property=CPUQuota=" + strconv.Itoa(r.cpuPercent) + "%",
		"--property=TasksMax=" + strconv.Itoa(tasksMax),
		"--property=NoNewPrivileges=yes",
		// The agent sees its own processes only, so it cannot read the command line or the
		// environment of the daemon or of another workspace's agent (ADR 0005, finding 3).
		"--property=ProtectProc=invisible", "--property=ProcSubset=pid",
	}
	for _, kv := range []string{
		"HOME=" + h.Config.homeDir(r.slug), "USER=" + agent, "LOGNAME=" + agent, "SHELL=/bin/bash",
		"LANG=C.UTF-8", "PATH=/usr/local/bin:/usr/bin:/bin", "TERM=xterm-256color", "COLORTERM=truecolor",
	} {
		argv = append(argv, "--setenv="+kv)
	}
	argv = append(argv, h.Config.Executable, "session-host", "--socket", sock, "--socket-mode", "660",
		"--dir", dir, "--rows", strconv.Itoa(r.rows), "--cols", strconv.Itoa(r.cols))
	if spec.Follow != "" {
		argv = append(argv, "--follow", spec.Follow)
	}
	return append(append(argv, "--"), spec.Command...)
}

func (h Helper) agentKinds() []string {
	kinds := make([]string, 0, len(h.Config.Agents))
	for k := range h.Config.Agents {
		kinds = append(kinds, k)
	}
	slices.Sort(kinds)
	return kinds
}

// sessionStop stops the unit of one session, and so everything the agent started in it.
func (h Helper) sessionStop(ctx context.Context, args []string) error {
	if len(args) != 2 || !domain.ValidSlug(args[0]) || !domain.ValidSessionID(args[1]) {
		return fmt.Errorf("%w: session-stop takes <slug> <id>", ErrUsage)
	}
	unit := unitName(args[0], args[1]) + ".service"
	if out, err := h.Host.Exec(ctx, systemctl, "stop", unit); err != nil {
		return fmt.Errorf("stop session %s: %w: %s", args[1], err, strings.TrimSpace(string(out)))
	}
	return nil
}
