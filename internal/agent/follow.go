package agent

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"time"
)

// Parser turns one transcript line into events.
type Parser func(line []byte) ([]Event, error)

// Follow reads the lines appended to path, from the start, and passes the events of each
// complete line to emit until ctx ends. It checks for new lines every poll, so an event
// reaches emit at most one poll after the agent writes it. A line that does not parse is
// skipped: one odd record must not stop the structured view.
func Follow(ctx context.Context, path string, parse Parser, poll time.Duration, emit func(Event)) error {
	f, err := os.Open(path) //nolint:gosec // the transcript of the agent this daemon runs
	if err != nil {
		return fmt.Errorf("open transcript: %w", err)
	}
	defer func() { _ = f.Close() }()

	tick := time.NewTicker(poll)
	defer tick.Stop()
	var pending []byte
	buf := make([]byte, 64<<10)
	for {
		for {
			n, err := f.Read(buf)
			pending = emitLines(append(pending, buf[:n]...), parse, emit)
			if errors.Is(err, io.EOF) || n == 0 {
				break
			}
			if err != nil {
				return fmt.Errorf("read transcript: %w", err)
			}
		}
		select {
		case <-ctx.Done():
			return nil
		case <-tick.C:
		}
	}
}

// emitLines handles every complete line in data and returns the incomplete rest.
func emitLines(data []byte, parse Parser, emit func(Event)) []byte {
	for {
		i := bytes.IndexByte(data, '\n')
		if i < 0 {
			return data
		}
		if line := bytes.TrimSpace(data[:i]); len(line) > 0 {
			if events, err := parse(line); err == nil {
				for _, e := range events {
					emit(e)
				}
			}
		}
		data = data[i+1:]
	}
}
