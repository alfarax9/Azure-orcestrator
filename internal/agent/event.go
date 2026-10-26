package agent

import (
	"strconv"
	"strings"
	"time"
)

// Kind says what an Event describes.
type Kind string

// Event kinds.
const (
	KindPrompt  Kind = "prompt"   // the developer, or the daemon on their behalf, asked something
	KindMessage Kind = "message"  // the agent wrote text to the developer
	KindCommand Kind = "command"  // the agent ran a shell command
	KindTool    Kind = "tool"     // the agent used another tool, such as reading or searching
	KindFile    Kind = "file"     // a file was added, changed or deleted
	KindTurnEnd Kind = "turn_end" // the agent finished its turn and waits for input
)

// Event is one structured thing an agent did, as recorded in its transcript.
type Event struct {
	At       time.Time   `json:"at"`
	Kind     Kind        `json:"kind"`
	Text     string      `json:"text,omitempty"`   // prompt, message or command line
	Tool     string      `json:"tool,omitempty"`   // tool name, for KindTool
	Target   string      `json:"target,omitempty"` // what the tool worked on: a file, pattern or URL
	ExitCode *int        `json:"exit_code,omitempty"`
	File     *FileChange `json:"file,omitempty"`
}

// FileChange is one file an agent added, updated or deleted, with its diff.
type FileChange struct {
	Path  string `json:"path"`
	Op    string `json:"op"` // "add", "update" or "delete"
	Hunks []Hunk `json:"hunks,omitempty"`
}

// Hunk is one block of a diff. Lines keep their unified-diff prefix: ' ', '+' or '-'.
type Hunk struct {
	OldStart int      `json:"oldStart"`
	OldLines int      `json:"oldLines"`
	NewStart int      `json:"newStart"`
	NewLines int      `json:"newLines"`
	Lines    []string `json:"lines"`
}

// Summary is a one-line description of e for logs and terminals.
func (e Event) Summary() string {
	switch {
	case e.File != nil:
		added, removed := e.File.Stats()
		return e.File.Op + " " + e.File.Path + " +" + strconv.Itoa(added) + " -" + strconv.Itoa(removed)
	case e.Tool != "":
		return strings.TrimSpace(e.Tool + " " + e.Target)
	case e.ExitCode != nil:
		return firstLine(e.Text) + " (exit " + strconv.Itoa(*e.ExitCode) + ")"
	default:
		return firstLine(e.Text)
	}
}

// Stats counts added and removed lines across the change's hunks.
func (f FileChange) Stats() (added, removed int) {
	for _, h := range f.Hunks {
		for _, l := range h.Lines {
			switch {
			case strings.HasPrefix(l, "+"):
				added++
			case strings.HasPrefix(l, "-"):
				removed++
			}
		}
	}
	return added, removed
}

func firstLine(s string) string {
	line, _, _ := strings.Cut(strings.TrimSpace(s), "\n")
	const limit = 120
	if len(line) > limit {
		return line[:limit] + "…"
	}
	return line
}

// addedFile is the diff of a new file: every line added.
func addedFile(path, content string) *FileChange {
	lines := splitLines(content)
	plus := make([]string, len(lines))
	for i, l := range lines {
		plus[i] = "+" + l
	}
	return &FileChange{Path: path, Op: "add", Hunks: []Hunk{{NewStart: 1, NewLines: len(lines), Lines: plus}}}
}

func splitLines(s string) []string {
	if s == "" {
		return nil
	}
	return strings.Split(strings.TrimSuffix(s, "\n"), "\n")
}

// parseUnifiedDiff reads the hunks of a unified diff; file headers are skipped.
func parseUnifiedDiff(diff string) []Hunk {
	var hunks []Hunk
	for _, line := range splitLines(diff) {
		switch {
		case strings.HasPrefix(line, "@@"):
			hunks = append(hunks, parseHunkHeader(line))
		case strings.HasPrefix(line, "---"), strings.HasPrefix(line, "+++"):
		case len(hunks) > 0 && line != "" && strings.ContainsRune(" +-", rune(line[0])):
			last := &hunks[len(hunks)-1]
			last.Lines = append(last.Lines, line)
		}
	}
	return hunks
}

// parseHunkHeader reads "@@ -a,b +c,d @@"; a missing count means 1.
func parseHunkHeader(line string) Hunk {
	var h Hunk
	for _, field := range strings.Fields(line) {
		if len(field) < 2 || (field[0] != '-' && field[0] != '+') {
			continue
		}
		start, count := parseRange(field[1:])
		if field[0] == '-' {
			h.OldStart, h.OldLines = start, count
		} else {
			h.NewStart, h.NewLines = start, count
		}
	}
	return h
}

func parseRange(r string) (start, count int) {
	s, c, found := strings.Cut(r, ",")
	start, _ = strconv.Atoi(s)
	count = 1
	if found {
		count, _ = strconv.Atoi(c)
	}
	return start, count
}
