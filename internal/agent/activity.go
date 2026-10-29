package agent

import (
	"regexp"
	"strings"
	"sync"

	"github.com/alfarax9/astra/internal/domain"
)

// ActivityAfter is what a worker's agent is doing once e has happened. Events that say
// nothing about it leave the activity as it was.
func ActivityAfter(cur domain.Activity, e Event) domain.Activity {
	switch e.Kind {
	case KindTurnEnd:
		return domain.ActivityIdle
	case KindPrompt, KindMessage, KindCommand, KindTool, KindFile:
		return domain.ActivityWorking
	}
	return cur
}

// ansi matches terminal escape sequences, so text can be searched without its formatting.
var ansi = regexp.MustCompile(`\x1b\[[0-9;?<>=]*[ -/]*[@-~]|\x1b\][^\x07\x1b]*(\x07|\x1b\\)|\x1b[()][0-9A-Za-z]|\x1b[=>78]`)

// StripANSI removes terminal escape sequences from terminal output.
func StripANSI(b []byte) []byte { return ansi.ReplaceAll(b, nil) }

// needsYouPrompts are questions an agent's terminal asks that only a person may answer: the
// first-run dialog about trusting a folder, and the approval of a command or an edit. They
// are written without spaces and in lower case, since a full-screen terminal moves the
// cursor instead of printing spaces (see needsDetector).
var needsYouPrompts = []string{
	"trustthefilesinthisfolder",       // Claude Code
	"trustthecontentsofthisdirectory", // Codex
	"doyouwanttoproceed",              // Claude Code: approve a tool call
	"wouldyouliketorunthefollowing",   // Codex: approve a command
}

// tailLimit bounds the text a detector keeps, so an old question scrolls away.
const tailLimit = 4000

// needsDetector watches an agent's terminal output for a question waiting for a person.
// Neither agent's transcript records one until it is answered, so the screen is the only place
// it shows (F3-06, the "Needs you" status). It is used by the goroutine that watches the
// session and by the callers that type into it.
type needsDetector struct {
	mu   sync.Mutex
	tail string
}

// Feed adds output and reports whether a question is on the screen now.
func (d *needsDetector) Feed(out []byte) bool {
	text := strings.Map(func(r rune) rune {
		if r == ' ' || r == '\n' || r == '\r' || r == '\t' || r == ' ' {
			return -1
		}
		return r
	}, strings.ToLower(string(StripANSI(out))))
	d.mu.Lock()
	defer d.mu.Unlock()
	d.tail += text
	if len(d.tail) > tailLimit {
		d.tail = d.tail[len(d.tail)-tailLimit:]
	}
	for _, p := range needsYouPrompts {
		if strings.Contains(d.tail, p) {
			return true
		}
	}
	return false
}

// Reset forgets what was on the screen: the person answered, or the agent moved on.
func (d *needsDetector) Reset() {
	d.mu.Lock()
	d.tail = ""
	d.mu.Unlock()
}
