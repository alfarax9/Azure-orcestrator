package priv

import (
	"path/filepath"

	"github.com/alfarax9/astra/internal/domain"
)

// Config is everything the helper trusts. It is fixed when the helper is built or installed;
// nothing in it comes from the caller.
type Config struct {
	StateDir   string               // holds agents/<slug>/{sessions,work,home}
	DaemonUser string               // the daemon's user and group; the group may reach session sockets
	Executable string               // the astrad binary that runs a session host
	SudoersDir string               // where the per-workspace sudo rule goes
	Agents     map[string]AgentSpec // agent kind -> what a session runs
}

// AgentSpec is what the helper starts for one kind of agent.
type AgentSpec struct {
	Command []string
	Follow  string // the event source the session host starts: the name of the agent's transcript reader
}

// DefaultConfig is the layout the installer creates.
func DefaultConfig() Config {
	return Config{
		StateDir:   "/var/lib/astra",
		DaemonUser: "astrad",
		Executable: "/usr/local/bin/astrad",
		SudoersDir: "/etc/sudoers.d",
		Agents: map[string]AgentSpec{
			"claude": {Command: []string{"/usr/local/bin/claude"}, Follow: "claude"},
			"codex":  {Command: []string{"/usr/local/bin/codex"}, Follow: "codex"},
			"shell":  {Command: []string{"/bin/bash", "-l"}},
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

// credentialFiles are where each agent keeps its login in the agent user's home. Only whether
// they exist is ever looked at, never what is in them (F3-03); where an agent keeps them is an
// assumption about its current version, and a wrong one only reads as "not signed in".
var credentialFiles = map[string]string{
	"claude": ".claude/.credentials.json",
	"codex":  ".codex/auth.json",
}

// base is the workspace's folder: sessions, work and home live under it.
func (c Config) base(slug string) string { return filepath.Join(c.StateDir, "agents", slug) }

func (c Config) sessionsDir(slug string) string { return filepath.Join(c.base(slug), "sessions") }
func (c Config) workDir(slug string) string     { return filepath.Join(c.base(slug), "work") }
func (c Config) homeDir(slug string) string     { return filepath.Join(c.base(slug), "home") }

// agentUser is the Linux user of a workspace's agents.
func agentUser(slug string) string { return domain.Workspace{Slug: slug}.AgentUser() }
