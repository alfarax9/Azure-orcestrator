package agent

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json/jsontext"
	"encoding/json/v2"
	"errors"
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/synctest"
	"time"
)

var update = flag.Bool("update", false, "rewrite the golden files in testdata/")

// parseFile runs a parser over every line of a fixture, skipping lines that do not parse,
// exactly as Follow does.
func parseFile(t *testing.T, path string, parse Parser) []Event {
	t.Helper()
	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = f.Close() }()
	var events []Event
	s := bufio.NewScanner(f)
	for s.Scan() {
		if evs, err := parse(s.Bytes()); err == nil {
			events = append(events, evs...)
		}
	}
	return events
}

// The fixtures copy the structure of real transcripts (Claude Code 2.1.310, Codex
// 0.164.0) with neutral content. The golden files hold the events they must produce.
func TestAdaptersMatchGoldenFiles(t *testing.T) {
	for _, tt := range []struct {
		name  string
		parse Parser
	}{
		{"claude", ParseClaude},
		{"codex", ParseCodex},
	} {
		t.Run(tt.name, func(t *testing.T) {
			events := parseFile(t, filepath.Join("testdata", tt.name+".jsonl"), tt.parse)
			checkGolden(t, filepath.Join("testdata", tt.name+".golden.json"), events)
		})
	}
}

// checkGolden compares events with a golden file, or rewrites it under -update.
func checkGolden(t *testing.T, golden string, events []Event) {
	t.Helper()
	got, err := json.Marshal(events, jsontext.Multiline(true), jsontext.WithIndent("  "))
	if err != nil {
		t.Fatal(err)
	}
	got = append(got, '\n')
	if *update {
		if err := os.WriteFile(golden, got, 0o600); err != nil {
			t.Fatal(err)
		}
	}
	want, err := os.ReadFile(golden)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, want) {
		t.Fatalf("events differ from %s (run go test -update after checking):\n%s", golden, got)
	}
}

func TestParsersRejectLinesThatAreNotJSON(t *testing.T) {
	for _, parse := range []Parser{ParseClaude, ParseCodex} {
		if _, err := parse([]byte("not json")); err == nil {
			t.Error("expected an error")
		}
	}
}

func TestFollowEmitsEachLineOnceItIsComplete(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "session.jsonl")
		f, err := os.Create(path)
		if err != nil {
			t.Fatal(err)
		}
		defer func() { _ = f.Close() }()

		ctx, cancel := context.WithCancel(t.Context())
		var got []Event
		done := make(chan error, 1)
		go func() {
			done <- Follow(ctx, path, ParseClaude, 100*time.Millisecond, func(e Event) { got = append(got, e) })
		}()

		write := func(s string) {
			if _, err := f.WriteString(s); err != nil {
				t.Fatal(err)
			}
			time.Sleep(150 * time.Millisecond) // one poll, on the bubble's fake clock
			synctest.Wait()
		}
		prompt := `{"type":"user","timestamp":"2026-10-26T04:00:00Z","message":{"content":"hello"}}`
		write(prompt[:20])
		if len(got) != 0 {
			t.Fatalf("a half-written line must not be parsed, got %v", got)
		}
		write(prompt[20:] + "\n")
		if len(got) != 1 || got[0].Kind != KindPrompt || got[0].Text != "hello" {
			t.Fatalf("expected the prompt one poll after the line completed, got %v", got)
		}
		write("garbage\n" + strings.Replace(prompt, "hello", "again", 1) + "\n")
		if len(got) != 2 || got[1].Text != "again" {
			t.Fatalf("a bad line must be skipped, not stop the follower: %v", got)
		}
		cancel()
		if err := <-done; err != nil {
			t.Fatal(err)
		}
	})
}

func TestFollowReportsAMissingFile(t *testing.T) {
	err := Follow(t.Context(), filepath.Join(t.TempDir(), "none.jsonl"), ParseClaude, time.Millisecond, func(Event) {})
	if err == nil {
		t.Fatal("expected an error for a missing transcript")
	}
}

func touch(t *testing.T, path, content string, mod time.Time) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Chtimes(path, mod, mod); err != nil {
		t.Fatal(err)
	}
}

func TestFindClaudeTranscriptPicksTheNewestForTheFolder(t *testing.T) {
	home, work := t.TempDir(), t.TempDir()
	dir := ClaudeProjectDir(home, work)
	if strings.ContainsAny(filepath.Base(dir), "/._") {
		t.Fatalf("slug keeps characters Claude Code replaces: %s", dir)
	}
	start := time.Date(2026, 10, 26, 10, 0, 0, 0, time.UTC)
	touch(t, filepath.Join(dir, "old.jsonl"), "{}\n", start.Add(-time.Hour))
	touch(t, filepath.Join(dir, "new.jsonl"), "{}\n", start.Add(time.Minute))
	touch(t, filepath.Join(dir, "newer.txt"), "{}\n", start.Add(2*time.Minute))

	got, err := FindClaudeTranscript(home, work, start)
	if err != nil || filepath.Base(got) != "new.jsonl" {
		t.Fatalf("got %q, %v", got, err)
	}
	if _, err := FindClaudeTranscript(home, work, start.Add(time.Hour)); !errors.Is(err, ErrNoTranscript) {
		t.Fatalf("a transcript older than the session must be ignored: %v", err)
	}
}

func TestFindCodexRolloutMatchesTheFolder(t *testing.T) {
	home, work, other := t.TempDir(), t.TempDir(), t.TempDir()
	start := time.Date(2026, 10, 26, 10, 0, 0, 0, time.UTC)
	day := filepath.Join(home, ".codex", "sessions", "2026", "10", "26")
	meta := func(cwd string) string {
		return `{"type":"session_meta","payload":{"cwd":"` + cwd + `"}}` + "\n"
	}
	touch(t, filepath.Join(day, "rollout-a.jsonl"), meta(work), start.Add(time.Minute))
	touch(t, filepath.Join(day, "rollout-b.jsonl"), meta(other), start.Add(2*time.Minute))
	touch(t, filepath.Join(day, "rollout-c.jsonl"), "not json\n", start.Add(3*time.Minute))

	got, err := FindCodexRollout(home, work, start)
	if err != nil || filepath.Base(got) != "rollout-a.jsonl" {
		t.Fatalf("got %q, %v", got, err)
	}
	if _, err := FindCodexRollout(home, t.TempDir(), start); !errors.Is(err, ErrNoTranscript) {
		t.Fatalf("a folder without a rollout: %v", err)
	}
}

func TestSummaries(t *testing.T) {
	code := 2
	change := &FileChange{Path: "a.go", Op: "update", Hunks: parseUnifiedDiff("@@ -3 +3,2 @@\n-old\n+new\n+more\n\\ No newline at end of file\n")}
	tests := []struct {
		event Event
		want  string
	}{
		{Event{Kind: KindFile, File: change}, "update a.go +2 -1"},
		{Event{Kind: KindTool, Tool: "Read", Target: "a.go"}, "Read a.go"},
		{Event{Kind: KindCommand, Text: "go test ./...\nsecond line", ExitCode: &code}, "go test ./... (exit 2)"},
		{Event{Kind: KindMessage, Text: strings.Repeat("x", 130)}, strings.Repeat("x", 120) + "…"},
	}
	for _, tt := range tests {
		if got := tt.event.Summary(); got != tt.want {
			t.Errorf("got %q, want %q", got, tt.want)
		}
	}
	if h := change.Hunks[0]; h.OldStart != 3 || h.OldLines != 1 || h.NewStart != 3 || h.NewLines != 2 {
		t.Errorf("hunk header parsed as %+v", h)
	}
}
