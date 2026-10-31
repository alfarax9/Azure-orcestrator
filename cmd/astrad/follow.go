package main

import (
	"context"
	"encoding/json/v2"
	"fmt"
	"os"
	"time"

	"github.com/alfarax9/astra/internal/agent"
	"github.com/alfarax9/astra/internal/pty"
)

const (
	transcriptPoll = 100 * time.Millisecond
	// transcriptSlack covers the gap between the session starting and the follower looking.
	transcriptSlack = 5 * time.Second
)

// followers are the event sources a session host may start with --follow. Each reads the
// transcript of one agent from the host's own home folder, so the daemon, which does not run
// as the agent's user, never opens it (docs/adr/0005).
func followers() map[string]pty.Follower {
	return map[string]pty.Follower{
		"claude": transcriptFollower(agent.ParseClaude, agent.FindClaudeTranscript),
		"codex":  codexFollower(),
	}
}

// codexFollower reads one rollout with a parser of its own, because the parser remembers the
// model across lines.
func codexFollower() pty.Follower {
	return func(ctx context.Context, dir string, emit func([]byte)) error {
		return transcriptFollower(agent.NewCodexParser(), agent.FindCodexRollout)(ctx, dir, emit)
	}
}

// transcriptFollower waits for the agent to start a transcript for the session's folder, then
// emits every structured event it holds as a JSON record.
func transcriptFollower(parse agent.Parser, find finder) pty.Follower {
	return func(ctx context.Context, dir string, emit func([]byte)) error {
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("find home directory: %w", err)
		}
		since := time.Now().Add(-transcriptSlack)
		path, err := waitForTranscript(ctx, func() (string, error) { return find(home, dir, since) }, transcriptPoll)
		if err != nil {
			return err
		}
		return agent.Follow(ctx, path, parse, transcriptPoll, func(e agent.Event) {
			if b, err := json.Marshal(e); err == nil {
				emit(b)
			}
		})
	}
}
