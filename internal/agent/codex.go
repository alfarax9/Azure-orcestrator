package agent

import (
	"encoding/json/jsontext"
	"encoding/json/v2"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/alfarax9/astra/internal/domain"
)

// Codex writes one JSON object per line to ~/.codex/sessions/YYYY/MM/DD/rollout-*.jsonl.
// The adapter reads the completed items (event_msg/item_completed), not the raw tool
// calls, which are JavaScript snippets calling Codex's internal tools.
type codexRecord struct {
	Timestamp time.Time `json:"timestamp"`
	Type      string    `json:"type"`
	Payload   struct {
		Type  string         `json:"type"`
		Item  jsontext.Value `json:"item"`
		Model string         `json:"model"` // turn_context
		Info  *struct {
			Total struct {
				Input       int `json:"input_tokens"`
				CachedInput int `json:"cached_input_tokens"`
				Output      int `json:"output_tokens"`
			} `json:"total_token_usage"`
		} `json:"info"` // token_count; null when the record only carries rate limits
	} `json:"payload"`
}

// codexUsageKey is the key of a Codex session's usage. token_count reports the running total of
// the whole session, so there is one usage record per session and each report replaces it.
const codexUsageKey = "session"

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

// codexParser remembers the model a rollout names in its turn_context records, because the
// token_count records that follow do not repeat it.
type codexParser struct{ model string }

// ParseCodex turns one line of a Codex rollout into events. Records it does not use
// (settings, reasoning) give no events. It does not know the model of the usage it reports;
// NewCodexParser does.
func ParseCodex(line []byte) ([]Event, error) {
	return (&codexParser{}).parse(line)
}

// NewCodexParser returns a Parser for one rollout that also names the model on usage events.
func NewCodexParser() Parser {
	return (&codexParser{}).parse
}

func (p *codexParser) parse(line []byte) ([]Event, error) {
	var r codexRecord
	if err := json.Unmarshal(line, &r); err != nil {
		return nil, fmt.Errorf("codex rollout line: %w", err)
	}
	if r.Type == "turn_context" && r.Payload.Model != "" {
		p.model = r.Payload.Model
	}
	if r.Type != "event_msg" {
		return nil, nil
	}
	switch r.Payload.Type {
	case "token_count":
		return p.usage(r), nil
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

// usage reports the session total. Codex counts cached tokens inside input_tokens, so they are
// taken out to keep Input free of the cache, as it is for Claude Code.
func (p *codexParser) usage(r codexRecord) []Event {
	if r.Payload.Info == nil {
		return nil
	}
	t := r.Payload.Info.Total
	u := domain.Usage{
		Key: codexUsageKey, Model: p.model, InputTokens: max(t.Input-t.CachedInput, 0),
		OutputTokens: t.Output, CacheReadTokens: t.CachedInput,
	}
	if u.Empty() {
		return nil
	}
	return []Event{{At: r.Timestamp, Kind: KindUsage, Usage: &u}}
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
