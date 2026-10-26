package priv

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
)

// ErrUsage means the request is malformed or asks for something the helper does not do.
var ErrUsage = errors.New("usage")

const usage = `usage: astra-priv <verb> ...
  workspace-add <slug>
  session-start <slug> <id> --agent claude|codex|shell [--dir REL] [--rows N] [--cols N]
                [--memory-max MB] [--cpu-quota PERCENT]
  session-stop  <slug> <id>`

// Dir is a folder the helper creates or repairs.
type Dir struct {
	Path  string
	Owner string
	Group string
	Mode  fs.FileMode // permission bits, plus fs.ModeSetgid where needed
}

// Agent describes an existing agent user.
type Agent struct {
	ExtraGroups []string // groups other than the user's own
}

// Host is the machine the helper changes. The real one is SystemHost; tests use a fake.
type Host interface {
	// Exec runs a program as argv and returns its combined output.
	Exec(ctx context.Context, argv ...string) ([]byte, error)
	// LookupAgent describes user name, or returns an error wrapping fs.ErrNotExist.
	LookupAgent(name string) (Agent, error)
	GroupExists(name string) bool
	// Install creates d if needed and sets its owner and mode; it never touches the contents.
	Install(d Dir) error
	// WriteSudoers checks content with visudo and only then makes it a live sudoers file.
	WriteSudoers(ctx context.Context, dir, name, content string) error
	// ResolveDir returns the real path of root/rel, which must be a folder inside root.
	ResolveDir(root, rel string) (string, error)
}

// Helper runs the three verbs.
type Helper struct {
	Config Config
	Host   Host
}

// Run executes one request. args[0] is the verb. Whatever a verb prints for the caller
// goes to out.
func (h Helper) Run(ctx context.Context, args []string, out io.Writer) error {
	if len(args) == 0 {
		return fmt.Errorf("%w\n%s", ErrUsage, usage)
	}
	switch args[0] {
	case "workspace-add":
		return h.workspaceAdd(ctx, args[1:])
	case "session-start":
		return h.sessionStart(ctx, args[1:], out)
	case "session-stop":
		return h.sessionStop(ctx, args[1:])
	}
	return fmt.Errorf("%w: unknown verb %q\n%s", ErrUsage, args[0], usage)
}
