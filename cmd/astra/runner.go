package main

import (
	"bytes"
	"context"
	"io"
	"os/exec"

	"github.com/alfarax9/astra/internal/remote"
	"github.com/alfarax9/astra/internal/server"
)

// sshRunner runs commands on a server over SSH.
type sshRunner struct{ c *remote.Client }

func (s sshRunner) Run(ctx context.Context, cmd string, stdin io.Reader) (server.Result, error) {
	res, err := s.c.Run(ctx, cmd, stdin)
	return server.Result{Stdout: res.Stdout, Stderr: res.Stderr, Exit: res.Exit}, err
}

// localRunner runs commands on this machine, for installing on the server itself.
type localRunner struct{}

func (localRunner) Run(ctx context.Context, cmd string, stdin io.Reader) (server.Result, error) {
	c := exec.CommandContext(ctx, "sh", "-c", cmd) //nolint:gosec // the commands come from internal/server, never from input
	var out, errOut bytes.Buffer
	c.Stdin, c.Stdout, c.Stderr = stdin, &out, &errOut
	err := c.Run()
	res := server.Result{Stdout: out.Bytes(), Stderr: errOut.Bytes()}
	var exit *exec.ExitError
	switch {
	case err == nil:
		return res, nil
	case asExit(err, &exit):
		res.Exit = exit.ExitCode()
		return res, nil
	}
	return res, err
}

func asExit(err error, target **exec.ExitError) bool {
	e, ok := err.(*exec.ExitError) //nolint:errorlint // exec.Run returns the *ExitError itself
	if ok {
		*target = e
	}
	return ok
}
