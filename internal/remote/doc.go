// Package remote is the SSH side of onboarding (F1-02, F1-05, ADR 0007): it connects to a
// server with a key, runs commands there, forwards a local port to the daemon, and keeps the
// host keys it has seen. It makes no decisions about what to install; internal/server does.
//
// A host key is never trusted silently. An unknown host stops the connection with its
// fingerprint so a person can confirm it; a host whose key changed is refused outright.
package remote
