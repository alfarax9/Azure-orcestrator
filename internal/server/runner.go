package server

import (
	"context"
	"fmt"
	"io"
	"strings"
)

// Result is what a command did. A non-zero Exit is a result, not an error.
type Result struct {
	Stdout []byte
	Stderr []byte
	Exit   int
}

// Runner runs a command on the machine being checked or installed.
type Runner interface {
	Run(ctx context.Context, cmd string, stdin io.Reader) (Result, error)
}

// as runs a script as root: directly when the login is root, otherwise through sudo without a
// password prompt. The script goes in on stdin, so nothing in it needs quoting.
func as(ctx context.Context, r Runner, root bool, script string) (Result, error) {
	cmd := "sh -s"
	if !root {
		cmd = "sudo -n sh -s"
	}
	return r.Run(ctx, cmd, strings.NewReader("set -eu\n"+script))
}

// failed describes a command that did not succeed, with what it said.
func failed(what string, res Result, err error) error {
	if err != nil {
		return fmt.Errorf("%s: %w", what, err)
	}
	if res.Exit == 0 {
		return nil
	}
	detail := strings.TrimSpace(string(res.Stderr))
	if detail == "" {
		detail = strings.TrimSpace(string(res.Stdout))
	}
	return fmt.Errorf("%s failed (exit %d): %s", what, res.Exit, detail)
}
