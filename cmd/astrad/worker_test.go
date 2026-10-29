package main

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/alfarax9/astra/internal/domain"
	"github.com/alfarax9/astra/internal/pty"
	"github.com/alfarax9/astra/internal/store"
)

// The test binary doubles as the session host, as astrad does in production.
func TestMain(m *testing.M) {
	if len(os.Args) > 1 && os.Args[1] == pty.HostCommand {
		os.Exit(pty.HostMainWith(os.Args[2:], os.Environ(), followers()))
	}
	os.Exit(m.Run())
}

// fakeClaude behaves like Claude Code as far as Astra can tell: it asks whether to trust the
// folder, waits for a prompt, edits a file and records the turn in a transcript where Claude
// Code keeps it.
const fakeClaude = `#!/bin/sh
dir=$(pwd -P | sed 's/[^A-Za-z0-9]/-/g')
proj="$HOME/.claude/projects/$dir"
mkdir -p "$proj"
printf 'Do you trust the files in this folder?\r\n'
read -r answer
printf '> \r\n'
while read -r line; do
  printf 'a\nb\n' > changed.txt
  now=$(date -u +%Y-%m-%dT%H:%M:%S.000Z)
  printf '{"type":"user","timestamp":"%s","message":{"role":"user","content":"task"}}\n' "$now" >> "$proj/fake.jsonl"
  printf '{"type":"assistant","timestamp":"%s","message":{"role":"assistant","content":[{"type":"text","text":"done"}],"stop_reason":"end_turn"}}\n' "$now" >> "$proj/fake.jsonl"
done
`

// waitFor polls until the worker's facts satisfy ok.
func waitFor(t *testing.T, st *store.Store, id domain.WorkerID, what string, ok func(domain.Worker) bool) domain.Worker {
	t.Helper()
	deadline := time.Now().Add(20 * time.Second)
	for {
		w, err := st.Worker(t.Context(), id)
		if err != nil {
			t.Fatal(err)
		}
		if ok(w) {
			return w
		}
		if time.Now().After(deadline) {
			t.Fatalf("timed out waiting for %s; the worker is %+v", what, w)
		}
		time.Sleep(50 * time.Millisecond)
	}
}

// localEnv points astrad at a throwaway local setup whose "claude" is fakeClaude.
func localEnv(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git is needed")
	}
	// Short paths: a Unix socket path is limited to about 100 bytes.
	root, err := os.MkdirTemp("/tmp", "astra-w-")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(root) })
	bin, home := filepath.Join(root, "bin"), filepath.Join(root, "home")
	for _, d := range []string{bin, home} {
		if err := os.MkdirAll(d, 0o700); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(bin, "claude"), []byte(fakeClaude), 0o700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", bin+":"+os.Getenv("PATH"))
	t.Setenv("HOME", home)
	t.Setenv("ASTRA_LOCAL", "1")
	t.Setenv("ASTRA_DB", filepath.Join(root, "astra.db"))
	t.Setenv("ASTRA_AGENTS_DIR", filepath.Join(root, "agents"))
}

// localApp wires the services to a local setup and returns them with a worker ready to start.
func localApp(t *testing.T) (app, domain.WorkerID) {
	t.Helper()
	localEnv(t)
	a, err := open(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = a.store.Close() })
	if _, err := a.ws.AddWorkspace(t.Context(), "acme", "Acme"); err != nil {
		t.Fatal(err)
	}
	p, err := a.ws.AddProject(t.Context(), "acme", "site", "")
	if err != nil {
		t.Fatal(err)
	}
	w, err := a.ws.CreateWorker(t.Context(), p.ID, "Fix the header", domain.AgentClaude, true)
	if err != nil {
		t.Fatal(err)
	}
	return a, w.ID
}

func TestAWorkerRunsFromTrustDialogToFinishedTurn(t *testing.T) {
	ctx := t.Context()
	a, id := localApp(t)
	if err := a.ag.Start(ctx, id); err != nil {
		t.Fatal(err)
	}
	watchCtx, stopWatching := context.WithCancel(ctx)
	watched := make(chan error, 1)
	go func() { _, err := a.ag.Watch(watchCtx, id); watched <- err }()

	// The agent asks about the folder; nothing is typed until the person answers.
	waitFor(t, a.store, id, "the trust question", func(w domain.Worker) bool { return w.Activity == domain.ActivityNeedsInput })
	if err := a.ag.Press(ctx, id, "\r"); err != nil {
		t.Fatal(err)
	}
	if err := a.ag.Send(ctx, id, "Fix the header"); err != nil {
		t.Fatal(err)
	}

	// The turn shows up as events, and the agent's file change as counted lines.
	done := waitFor(t, a.store, id, "the finished turn", func(w domain.Worker) bool {
		return w.Activity == domain.ActivityIdle && w.Added == 2
	})
	if done.Removed != 0 {
		t.Errorf("removed = %d", done.Removed)
	}
	stopWatching()
	if err := <-watched; err != nil {
		t.Errorf("Watch = %v", err)
	}
	if err := a.ag.Stop(ctx, id); err != nil {
		t.Fatal(err)
	}
}
