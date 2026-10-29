package asuser

import (
	"bytes"
	"context"
	"encoding/json/v2"
	"errors"
	"fmt"
	"os"
	"os/exec"
	osuser "os/user"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/alfarax9/astra/internal/domain"
)

// maxOutput bounds what a command may print: repository content is written by agents, and
// the daemon must not be made to hold an arbitrary amount of it.
const maxOutput = 4 << 20

// pathElement is one folder name below a work folder; it also rules out "." and "..".
var pathElement = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]*$`)

// Runner runs commands for workspaces.
type Runner struct {
	Root   string // holds <slug>/{work,home}; /var/lib/astra/agents on a server
	Helper string // the astra-priv binary, called through sudo; unused in local mode
	Local  bool   // run as the caller, without agent users (development only)
}

// Run executes argv as the agent user of workspace slug, in dir below its work folder,
// and returns what it printed on stdout. A failure carries its stderr.
func (r Runner) Run(ctx context.Context, slug, dir string, argv ...string) ([]byte, error) {
	if err := checkTarget(slug, dir); err != nil {
		return nil, err
	}
	if len(argv) == 0 {
		return nil, errors.New("no command")
	}
	cwd := filepath.Join(r.Root, slug, "work", dir)
	env := r.environment(slug)
	var cmd *exec.Cmd
	if r.Local {
		cmd = exec.CommandContext(ctx, argv[0], argv[1:]...) //nolint:gosec // argv comes from the daemon's own code
		cmd.Dir, cmd.Env = cwd, env
	} else {
		// sudo resets the environment; env -i then sets exactly ours, and -C changes into a
		// folder only the agent user can reach.
		args := append([]string{"-n", "-u", domain.Workspace{Slug: slug}.AgentUser(), "--", "/usr/bin/env", "-i", "-C", cwd}, env...)
		cmd = exec.CommandContext(ctx, "sudo", append(args, argv...)...) //nolint:gosec // slug and dir are checked above
		cmd.Dir, cmd.Env = "/", []string{"PATH=/usr/bin:/bin"}
	}
	var stdout, stderr limited
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	err := cmd.Run()
	if stdout.over {
		return nil, fmt.Errorf("%s printed more than %d bytes", argv[0], maxOutput)
	}
	if err != nil {
		return stdout.buf.Bytes(), fmt.Errorf("%s: %w: %s", strings.Join(argv[:min(2, len(argv))], " "), err,
			strings.TrimSpace(stderr.buf.String()))
	}
	return stdout.buf.Bytes(), nil
}

// environment is all a command run for a workspace receives. Git must never stop to ask
// for a password, since no one is at its terminal, and may only use the transports a
// project URL may name (no ext::, which runs a command).
func (r Runner) environment(slug string) []string {
	home := filepath.Join(r.Root, slug, "home")
	user := domain.Workspace{Slug: slug}.AgentUser()
	if me, err := osuser.Current(); r.Local && err == nil {
		user, home = me.Username, me.HomeDir
	}
	return []string{
		"HOME=" + home, "USER=" + user, "LOGNAME=" + user, "LANG=C.UTF-8",
		"PATH=/usr/local/bin:/usr/bin:/bin", "GIT_TERMINAL_PROMPT=0", "GIT_ALLOW_PROTOCOL=https:ssh:file",
		"GIT_SSH_COMMAND=ssh -o BatchMode=yes",
	}
}

// AddWorkspace prepares the machine for a new workspace: its agent user, folders and sudo
// rule, through astra-priv. In local mode it only creates the folders.
func (r Runner) AddWorkspace(ctx context.Context, slug string) error {
	if err := checkTarget(slug, ""); err != nil {
		return err
	}
	if r.Local {
		for _, d := range []string{"work", "home"} {
			if err := os.MkdirAll(filepath.Join(r.Root, slug, d), 0o700); err != nil {
				return fmt.Errorf("create workspace folder: %w", err)
			}
		}
		return nil
	}
	return r.helper(ctx, "workspace-add", slug)
}

// Socket is where the session host of a session listens.
func (r Runner) Socket(slug, id string) string {
	return filepath.Join(r.Root, slug, "sessions", id+".sock")
}

// StartSession runs agent ("claude" or "codex") in dir below the workspace's work folder, as a
// unit of the workspace's agent user. Only a server has units; local mode starts sessions
// itself.
func (r Runner) StartSession(ctx context.Context, slug, id, agent, dir string) error {
	if err := checkTarget(slug, dir); err != nil {
		return err
	}
	args := []string{"session-start", slug, id, "--agent", agent}
	if dir != "" {
		args = append(args, "--dir", dir)
	}
	return r.helper(ctx, args...)
}

// StopSession stops the unit of a session.
func (r Runner) StopSession(ctx context.Context, slug, id string) error {
	return r.helper(ctx, "session-stop", slug, id)
}

// helper calls astra-priv through sudo. It checks nothing itself: the helper refuses what it
// does not accept, and its message comes back as the error.
func (r Runner) helper(ctx context.Context, args ...string) error {
	_, err := r.helperOutput(ctx, args...)
	return err
}

func (r Runner) helperOutput(ctx context.Context, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, "sudo", append([]string{"-n", r.Helper}, args...)...) //nolint:gosec // the helper validates every argument
	cmd.Dir, cmd.Env = "/", []string{"PATH=/usr/bin:/bin"}
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("%s: %w: %s", args[0], err, strings.TrimSpace(stderr.String()))
	}
	return stdout.Bytes(), nil
}

// AgentLogin says which agents are signed in for a workspace. On a server the answer comes from
// astra-priv, since the daemon may not open an agent's home; in local mode the developer's own
// home is looked at, for the existence of the login files only.
func (r Runner) AgentLogin(ctx context.Context, slug string) (map[string]bool, error) {
	if err := checkTarget(slug, ""); err != nil {
		return nil, err
	}
	if r.Local {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("find home directory: %w", err)
		}
		status := map[string]bool{}
		for agent, file := range map[string]string{"claude": ".claude/.credentials.json", "codex": ".codex/auth.json"} {
			_, err := os.Lstat(filepath.Join(home, file))
			status[agent] = err == nil
		}
		return status, nil
	}
	out, err := r.helperOutput(ctx, "agent-status", slug)
	if err != nil {
		return nil, err
	}
	var status map[string]bool
	if err := json.Unmarshal(out, &status); err != nil {
		return nil, fmt.Errorf("read the agent status: %w", err)
	}
	return status, nil
}

func checkTarget(slug, dir string) error {
	if !domain.ValidSlug(slug) {
		return fmt.Errorf("invalid workspace slug %q", slug)
	}
	if dir == "" {
		return nil
	}
	for _, part := range strings.Split(dir, "/") {
		if !pathElement.MatchString(part) {
			return fmt.Errorf("invalid folder %q: use plain folder names below the work folder", dir)
		}
	}
	return nil
}

// limited is a buffer that stops growing at maxOutput and remembers that it did. It does
// not embed bytes.Buffer: io.Copy would use the buffer's ReadFrom and skip the limit.
type limited struct {
	buf  bytes.Buffer
	over bool
}

func (l *limited) Write(p []byte) (int, error) {
	if l.buf.Len()+len(p) > maxOutput {
		l.over = true
		return len(p), nil // keep the command running to completion; Run reports it
	}
	return l.buf.Write(p)
}
