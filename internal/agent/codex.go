package agent

import (
	"encoding/json/jsontext"
	"encoding/json/v2"
	"fmt"
	"slices"
	"strings"
	"time"
)

// Codex writes one JSON object per line to ~/.codex/sessions/YYYY/MM/DD/rollout-*.jsonl.
// The adapter reads the completed items (event_msg/item_completed), not the raw tool
// calls, which are JavaScript snippets calling Codex's internal tools.
type codexRecord struct {
	Timestamp time.Time `json:"timestamp"`
	Type      string    `json:"type"`
	Payload   struct {
		Type string         `json:"type"`
		Item jsontext.Value `json:"item"`
	} `json:"payload"`
}

type codexItem struct {
	Type    string `json:"type"`
	Content []struct {
		Text string `json:"text"`
	} `json:"content"`
	Command  []string               `json:"command"`
	ExitCode *int                   `json:"exit_code"`
	Changes  map[string]codexChange `json:"changes"`
}

type codexChange struct {
	Type        string `json:"type"` // "add", "update" or "delete"
	Content     string `json:"content"`
	UnifiedDiff string `json:"unified_diff"`
}

// ParseCodex turns one line of a Codex rollout into events. Records it does not use
// (settings, reasoning, token counts) give no events.
func ParseCodex(line []byte) ([]Event, error) {
	var r codexRecord
	if err := json.Unmarshal(line, &r); err != nil {
		return nil, fmt.Errorf("codex rollout line: %w", err)
	}
	if r.Type != "event_msg" {
		return nil, nil
	}
	switch r.Payload.Type {
	case "task_complete":
		return []Event{{At: r.Timestamp, Kind: KindTurnEnd}}, nil
	case "item_completed":
		var item codexItem
		if len(r.Payload.Item) == 0 || json.Unmarshal(r.Payload.Item, &item) != nil {
			return nil, nil
		}
		return codexItemEvents(r.Timestamp, item), nil
	}
	return nil, nil
}

func codexItemEvents(at time.Time, item codexItem) []Event {
	switch item.Type {
	case "UserMessage":
		return textEvent(at, KindPrompt, item)
	case "AgentMessage":
		return textEvent(at, KindMessage, item)
	case "CommandExecution":
		return []Event{{At: at, Kind: KindCommand, Text: shellCommand(item.Command), ExitCode: item.ExitCode}}
	case "FileChange":
		return codexFileEvents(at, item.Changes)
	}
	return nil
}

func textEvent(at time.Time, kind Kind, item codexItem) []Event {
	parts := make([]string, 0, len(item.Content))
	for _, c := range item.Content {
		parts = append(parts, c.Text)
	}
	text := strings.Join(parts, "\n")
	if strings.TrimSpace(text) == "" {
		return nil
	}
	return []Event{{At: at, Kind: kind, Text: text}}
}

// shellCommand shows what the agent ran: the script of `sh -lc <script>`, else argv.
func shellCommand(argv []string) string {
	if n := len(argv); n == 3 && strings.HasSuffix(argv[1], "c") && strings.HasPrefix(argv[1], "-") {
		return argv[2]
	}
	return strings.Join(argv, " ")
}

// codexFileEvents gives one event per changed file, in path order so output is stable.
func codexFileEvents(at time.Time, changes map[string]codexChange) []Event {
	paths := make([]string, 0, len(changes))
	for p := range changes {
		paths = append(paths, p)
	}
	slices.Sort(paths)
	events := make([]Event, 0, len(paths))
	for _, p := range paths {
		c := changes[p]
		var f *FileChange
		switch c.Type {
		case "add":
			f = addedFile(p, c.Content)
		case "delete":
			f = addedFile(p, c.Content)
			f.Op = "delete"
			for i := range f.Hunks {
				f.Hunks[i] = removedHunk(f.Hunks[i])
			}
		default:
			f = &FileChange{Path: p, Op: "update", Hunks: parseUnifiedDiff(c.UnifiedDiff)}
		}
		events = append(events, Event{At: at, Kind: KindFile, File: f})
	}
	return events
}

// removedHunk turns the hunk of an added file into the hunk of a deleted one.
func removedHunk(h Hunk) Hunk {
	lines := make([]string, len(h.Lines))
	for i, l := range h.Lines {
		lines[i] = "-" + strings.TrimPrefix(l, "+")
	}
	return Hunk{OldStart: 1, OldLines: len(lines), Lines: lines}
}
