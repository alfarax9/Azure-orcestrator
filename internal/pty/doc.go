// Package pty runs agent CLIs (Claude Code, Codex) in pseudo-terminals that outlive the
// daemon. Each session has its own host process: the host owns the PTY and the agent,
// keeps the recent output as scrollback, and serves both over a Unix socket. A daemon
// that restarts dials the socket again and repaints the terminal from the scrollback,
// while the agent keeps running (spike S1, F3-02, F3-04).
//
// The agent receives exactly the environment in its Spec and nothing from the daemon,
// so daemon secrets never reach it.
package pty
