package agent

import (
	"encoding/json/jsontext"
	"encoding/json/v2"
	"fmt"
	"strings"
	"time"

	"github.com/alfarax9/astra/internal/domain"
)

// Claude Code writes one JSON object per line to ~/.claude/projects/<cwd>/<session>.jsonl.
// Only the fields this adapter reads are declared; json/v2 ignores the rest.
type claudeEntry struct {
	Type        string    `json:"type"`
	Timestamp   time.Time `json:"timestamp"`
	IsMeta      bool      `json:"isMeta"`
	IsSidechain bool      `json:"isSidechain"`
	Message     struct {
		ID         string         `json:"id"`
		Model      string         `json:"model"`
		Content    jsontext.Value `json:"content"`
		StopReason string         `json:"stop_reason"`
		Usage      struct {
			Input      int `json:"input_tokens"`
			Output     int `json:"output_tokens"`
			CacheRead  int `json:"cache_read_input_tokens"`
			CacheWrite int `json:"cache_creation_input_tokens"`
		} `json:"usage"`
	} `json:"message"`
	ToolUseResult jsontext.Value `json:"toolUseResult"`
}

type claudeBlock struct {
	Type  string `json:"type"`
	Text  string `json:"text"`
	Name  string `json:"name"`
	Input struct {
		Command  string `json:"command"`
		FilePath string `json:"file_path"`
		Path     string `json:"path"`
		Pattern  string `json:"pattern"`
		URL      string `json:"url"`
	} `json:"input"`
}

// claudeResult covers the tool results that change files: Edit and MultiEdit carry
// structuredPatch, Write carries type and content, Bash carries bashEditDiff.
type claudeResult struct {
	Type            string `json:"type"`
	FilePath        string `json:"filePath"`
	Content         string `json:"content"`
	StructuredPatch []Hunk `json:"structuredPatch"`
	BashEditDiff    *struct {
		Files []struct {
			FilePath string `json:"filePath"`
			Hunks    []Hunk `json:"hunks"`
			Created  bool   `json:"created"`
		} `json:"files"`
	} `json:"bashEditDiff"`
}

// fileTools report their change in the tool result, which carries the diff.
var fileTools = map[string]bool{"Edit": true, "MultiEdit": true, "Write": true, "NotebookEdit": true}

// ParseClaude turns one line of a Claude Code transcript into events. Lines it does not
// use (settings, attachments, internal notes) give no events, and subagent work gives only usage.
func ParseClaude(line []byte) ([]Event, error) {
	var e claudeEntry
	if err := json.Unmarshal(line, &e); err != nil {
		return nil, fmt.Errorf("claude transcript line: %w", err)
	}
	if e.IsMeta {
		return nil, nil
	}
	if e.IsSidechain {
		// A subagent's work is not shown, but it spends tokens like any other.
		if e.Type == "assistant" {
			return claudeUsage(e), nil
		}
		return nil, nil
	}
	switch e.Type {
	case "user":
		return claudeUser(e), nil
	case "assistant":
		return claudeAssistant(e), nil
	}
	return nil, nil
}

func claudeUser(e claudeEntry) []Event {
	if text, ok := plainText(e.Message.Content); ok {
		return claudePrompt(e.Timestamp, text)
	}
	if len(e.ToolUseResult) == 0 || e.ToolUseResult.Kind() != '{' {
		return nil // a result without a file change, or an error message
	}
	var r claudeResult
	if err := json.Unmarshal(e.ToolUseResult, &r); err != nil {
		return nil
	}
	var events []Event
	for _, f := range claudeFileChanges(r) {
		events = append(events, Event{At: e.Timestamp, Kind: KindFile, File: f})
	}
	return events
}

// claudePrompt reads a typed prompt. Commands typed in bash mode ("!") are recorded as
// <bash-input>; slash commands and their output are internal and skipped.
func claudePrompt(at time.Time, text string) []Event {
	if cmd, ok := between(text, "<bash-input>", "</bash-input>"); ok {
		return []Event{{At: at, Kind: KindCommand, Text: cmd}}
	}
	if strings.HasPrefix(text, "<") {
		return nil
	}
	return []Event{{At: at, Kind: KindPrompt, Text: text}}
}

func claudeFileChanges(r claudeResult) []*FileChange {
	switch {
	case r.BashEditDiff != nil:
		changes := make([]*FileChange, 0, len(r.BashEditDiff.Files))
		for _, f := range r.BashEditDiff.Files {
			op := "update"
			if f.Created {
				op = "add"
			}
			changes = append(changes, &FileChange{Path: f.FilePath, Op: op, Hunks: f.Hunks})
		}
		return changes
	case r.FilePath == "":
		return nil
	case r.Type == "create":
		return []*FileChange{addedFile(r.FilePath, r.Content)}
	case len(r.StructuredPatch) > 0:
		return []*FileChange{{Path: r.FilePath, Op: "update", Hunks: r.StructuredPatch}}
	}
	return nil
}

func claudeAssistant(e claudeEntry) []Event {
	var blocks []claudeBlock
	if e.Message.Content.Kind() == '[' {
		_ = json.Unmarshal(e.Message.Content, &blocks)
	}
	var events []Event
	for _, b := range blocks {
		if ev, ok := claudeBlockEvent(e.Timestamp, b); ok {
			events = append(events, ev)
		}
	}
	// One message is written as several entries (thinking, then text), each carrying the
	// stop reason; only an entry that showed something ends the turn, so it ends once.
	if e.Message.StopReason == "end_turn" && len(events) > 0 {
		events = append(events, Event{At: e.Timestamp, Kind: KindTurnEnd})
	}
	return append(events, claudeUsage(e)...)
}

// claudeUsage reports the tokens of an assistant message. A message is written as several
// entries that each repeat its usage with the same id, and the last one is complete, so the
// id is the key and each report replaces the one before it.
func claudeUsage(e claudeEntry) []Event {
	m := e.Message
	u := domain.Usage{
		Key: m.ID, Model: m.Model, InputTokens: m.Usage.Input, OutputTokens: m.Usage.Output,
		CacheReadTokens: m.Usage.CacheRead, CacheWriteTokens: m.Usage.CacheWrite,
	}
	if m.ID == "" || u.Empty() {
		return nil
	}
	return []Event{{At: e.Timestamp, Kind: KindUsage, Usage: &u}}
}

func claudeBlockEvent(at time.Time, b claudeBlock) (Event, bool) {
	switch {
	case b.Type == "text" && strings.TrimSpace(b.Text) != "":
		return Event{At: at, Kind: KindMessage, Text: b.Text}, true
	case b.Type != "tool_use" || fileTools[b.Name]:
		return Event{}, false
	case b.Name == "Bash":
		return Event{At: at, Kind: KindCommand, Text: b.Input.Command}, true
	}
	// A search shows its pattern rather than the folder it searched.
	target := b.Input.FilePath
	for _, t := range []string{b.Input.Pattern, b.Input.Path, b.Input.URL} {
		if target == "" {
			target = t
		}
	}
	return Event{At: at, Kind: KindTool, Tool: b.Name, Target: target}, true
}

// plainText reports whether v is a JSON string, and its value.
func plainText(v jsontext.Value) (string, bool) {
	if v.Kind() != '"' {
		return "", false
	}
	var s string
	if err := json.Unmarshal(v, &s); err != nil {
		return "", false
	}
	return s, true
}

func between(s, open, closing string) (string, bool) {
	_, rest, ok := strings.Cut(s, open)
	if !ok {
		return "", false
	}
	inner, _, ok := strings.Cut(rest, closing)
	return inner, ok
}
