package main

import (
	"context"
	"encoding/json/v2"
	"errors"
	"fmt"
	"net"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"time"

	"github.com/alfarax9/astra/internal/agent"
	"github.com/alfarax9/astra/internal/asuser"
	"github.com/alfarax9/astra/internal/pty"
)

// sessionHost implements agent.Sessions. On a server a session is a unit of the workspace's
// agent user, made by astra-priv, and the daemon reaches it through its socket. Locally the
// developer's own agent runs under a session host the daemon starts itself.
type sessionHost struct {
	run asuser.Runner
}

func (h sessionHost) Start(ctx context.Context, spec agent.SessionSpec) error {
	if !h.run.Local {
		if err := h.run.StartSession(ctx, spec.Slug, spec.ID, string(spec.Agent), spec.Dir); err != nil {
			return err
		}
		// The unit is started, but its host opens the socket a moment later: nothing may
		// attach before that, or the worker would look like it has no session.
		return pty.WaitReady(ctx, h.run.Socket(spec.Slug, spec.ID))
	}
	path, err := exec.LookPath(string(spec.Agent))
	if err != nil {
		return fmt.Errorf("find %s: %w", spec.Agent, err)
	}
	me, err := user.Current()
	if err != nil {
		return fmt.Errorf("look up the current user: %w", err)
	}
	home, err := os.UserHomeDir() // the agent is logged in there
	if err != nil {
		return fmt.Errorf("find home directory: %w", err)
	}
	exe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("find own binary: %w", err)
	}
	sp := pty.Spawner{SocketDir: filepath.Dir(h.run.Socket(spec.Slug, spec.ID)), Executable: exe, Scrollback: 1 << 20, Linger: 10 * time.Minute}
	_, err = sp.Start(ctx, pty.Spec{
		ID: spec.ID, Argv: []string{path}, Follow: string(spec.Agent),
		Dir: filepath.Join(h.run.Root, spec.Slug, "work", spec.Dir), Rows: 40, Cols: 120,
		Env: []string{"HOME=" + home, "USER=" + me.Username, "LOGNAME=" + me.Username, "LANG=C.UTF-8",
			"PATH=/usr/local/bin:/opt/homebrew/bin:/usr/bin:/bin", "TERM=xterm-256color", "COLORTERM=truecolor"},
	})
	return err
}

func (h sessionHost) Attach(ctx context.Context, slug, id string) (agent.Stream, error) {
	c, err := pty.Dial(ctx, h.run.Socket(slug, id))
	if err != nil {
		return nil, err
	}
	return stream{c}, nil
}

func (h sessionHost) Stop(ctx context.Context, slug, id string) error {
	if !h.run.Local {
		return h.run.StopSession(ctx, slug, id)
	}
	c, err := pty.Dial(ctx, h.run.Socket(slug, id))
	if err != nil {
		return err
	}
	defer func() { _ = c.Close() }()
	return c.Stop()
}

// stream turns a session connection into agent signals, decoding the host's event records.
type stream struct{ c *pty.Conn }

func (s stream) Next() (agent.Signal, error) {
	for {
		ev, err := s.c.Next()
		if err != nil {
			return agent.Signal{}, err
		}
		switch {
		case ev.Exited:
			return agent.Signal{Exited: true, ExitCode: ev.ExitCode}, nil
		case ev.Record != nil:
			var e agent.Event
			if err := json.Unmarshal(ev.Record, &e); err != nil {
				continue // a record this daemon does not understand is skipped, like a transcript line
			}
			return agent.Signal{Event: &e}, nil
		case len(ev.Output) > 0:
			return agent.Signal{Output: ev.Output}, nil
		}
	}
}

func (s stream) Write(p []byte) error {
	if _, err := s.c.Write(p); err != nil {
		return fmt.Errorf("write to session: %w", err)
	}
	return nil
}

func (s stream) Resize(rows, cols uint16) error { return s.c.Resize(rows, cols) }

func (s stream) Close() error {
	if err := s.c.Close(); err != nil && !errors.Is(err, net.ErrClosed) {
		return err
	}
	return nil
}
