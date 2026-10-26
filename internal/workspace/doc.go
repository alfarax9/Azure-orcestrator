// Package workspace manages client workspaces, their projects and the git worktree of each
// worker (F2-01, F2-02, F2-03, F3-01).
//
// Every repository command runs as the workspace's agent user through a Runner; the daemon
// never opens a workspace's files itself (docs/adr/0005). Paths handed to the Runner are
// relative to the workspace's work folder:
//
//	<project>/repo                the clone, on the default branch
//	<project>/workers/<worker-id>  one worktree per worker, on its own branch
package workspace
