// Package api is the daemon's HTTP and WebSocket interface for every client: the desktop app,
// the web UI and the CLI (PRD section 7, ADR 0006). Handlers decode and validate a request,
// call a capability, and encode the answer; they hold no logic of their own.
//
// Every route is under /api/v1 and, except the version and pairing, needs the token of a
// paired device. The daemon listens on loopback only; encryption is the SSH tunnel's or the
// proxy's job (N-12).
package api
