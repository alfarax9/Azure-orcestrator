package main

import (
	"context"
	"encoding/json/v2"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/alfarax9/astra/internal/agent"
)

const eventsUsage = `usage: astrad events -agent claude|codex -dir DIR [-since 10m] [-for 0] [-poll 100ms] [-json]`

// runEvents is the spike S2 command line: it follows an agent's transcript for a folder
// and prints each structured event with its lag behind the agent's own timestamp.
func runEvents(args []string) int {
	fs := flag.NewFlagSet("events", flag.ContinueOnError)
	name := fs.String("agent", "", "claude or codex")
	wd, _ := os.Getwd()
	dir := fs.String("dir", wd, "working directory the agent runs in")
	since := fs.Duration("since", 10*time.Minute, "only transcripts changed within this window")
	limit := fs.Duration("for", 0, "stop after this long (0: until Ctrl-C)")
	poll := fs.Duration("poll", 100*time.Millisecond, "how often to check for new lines")
	asJSON := fs.Bool("json", false, "print events as JSON lines")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	parse, find, err := adapter(*name)
	if err != nil {
		fmt.Fprintln(os.Stderr, eventsUsage)
		return 2
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if *limit > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, *limit)
		defer cancel()
	}
	if err := followEvents(ctx, find, parse, time.Now().Add(-*since), *dir, *poll, *asJSON); err != nil {
		fmt.Fprintln(os.Stderr, "astrad:", err)
		return 1
	}
	return 0
}

type finder func(home, cwd string, since time.Time) (string, error)

func adapter(name string) (agent.Parser, finder, error) {
	switch name {
	case "claude":
		return agent.ParseClaude, agent.FindClaudeTranscript, nil
	case "codex":
		return agent.ParseCodex, agent.FindCodexRollout, nil
	}
	return nil, nil, fmt.Errorf("unknown agent %q", name)
}

func followEvents(ctx context.Context, find finder, parse agent.Parser, since time.Time, dir string, poll time.Duration, asJSON bool) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("find home directory: %w", err)
	}
	path, err := waitForTranscript(ctx, func() (string, error) { return find(home, dir, since) }, poll)
	if err != nil {
		return err
	}
	fmt.Fprintln(os.Stderr, "following", path)
	return agent.Follow(ctx, path, parse, poll, func(e agent.Event) {
		if asJSON {
			b, _ := json.Marshal(e)
			fmt.Println(string(b))
			return
		}
		lag := time.Since(e.At).Round(time.Millisecond)
		fmt.Printf("%s  lag %-8s %-8s %s\n", e.At.Local().Format("15:04:05.000"), lag, e.Kind, e.Summary())
	})
}

func waitForTranscript(ctx context.Context, find func() (string, error), poll time.Duration) (string, error) {
	tick := time.NewTicker(poll)
	defer tick.Stop()
	for {
		path, err := find()
		if err == nil {
			return path, nil
		}
		if !errors.Is(err, agent.ErrNoTranscript) {
			return "", err
		}
		select {
		case <-ctx.Done():
			return "", errors.New("no transcript appeared for that folder")
		case <-tick.C:
		}
	}
}
