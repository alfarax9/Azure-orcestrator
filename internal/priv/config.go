package priv

import (
	"path/filepath"

	"github.com/alfarax9/astra/internal/domain"
)

// Config is everything the helper trusts. It is fixed when the helper is built or installed;
// nothing in it comes from the caller.
type Config struct {
	StateDir   string              // holds agents/<slug>/{sessions,work,home}
	DaemonUser string              // the daemon's user and group; the group may reach session sockets
	Executable string              // the astrad binary that runs a session host
	SudoersDir string              // where the per-workspace sudo rule goes
	Agents     map[string][]string // agent kind -> the command a session runs
}

// DefaultConfig is the layout the installer creates.
func DefaultConfig() Config {
	return Config{
		StateDir:   "/var/lib/astra",
		DaemonUser: "astrad",
		Executable: "/usr/local/bin/astrad",
		SudoersDir: "/etc/sudoers.d",
		Agents: map[string][]string{
			"claude": {"/usr/local/bin/claude"},
			"codex":  {"/usr/local/bin/codex"},
			"shell":  {"/bin/bash", "-l"},
		},
	}
}

// Bounds of what a caller may ask for; anything outside is refused, not clamped.
const (
	minMemoryMB, maxMemoryMB, defaultMemoryMB = 256, 8192, 1024
	minCPUPercent, maxCPUPercent, defaultCPU  = 10, 400, 100
	minTerminal, maxTerminal                  = 10, 500
	defaultRows, defaultCols                  = 40, 120
	tasksMax                                  = 512
)

// base is the workspace's folder: sessions, work and home live under it.
func (c Config) base(slug string) string { return filepath.Join(c.StateDir, "agents", slug) }

func (c Config) sessionsDir(slug string) string { return filepath.Join(c.base(slug), "sessions") }
func (c Config) workDir(slug string) string     { return filepath.Join(c.base(slug), "work") }
func (c Config) homeDir(slug string) string     { return filepath.Join(c.base(slug), "home") }

// agentUser is the Linux user of a workspace's agents.
func agentUser(slug string) string { return domain.Workspace{Slug: slug}.AgentUser() }
