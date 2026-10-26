// Package priv is the code behind astra-priv, the one program Astra runs as root
// (docs/adr/0005). The daemon runs as an unprivileged user and asks this helper for the
// three things that need root: create a workspace's agent user and folders (workspace-add),
// start an agent session as a transient systemd unit (session-start) and stop it
// (session-stop).
//
// The helper trusts nothing from its caller. Every argument is checked against a strict
// pattern before it becomes part of a command, commands run as argv and never through a
// shell, and the caller cannot pick the program, the environment or the user a session runs
// as: those come from Config and from the workspace slug.
package priv
