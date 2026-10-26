package agent

import (
	"bufio"
	"encoding/json/v2"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// ErrNoTranscript means the agent has not written a transcript for that folder yet.
// Claude Code creates it on the first prompt, Codex on its first turn.
var ErrNoTranscript = errors.New("no transcript yet")

// ClaudeProjectDir is where Claude Code keeps the transcripts of sessions started in cwd:
// the absolute path with every character other than a letter or digit turned into '-'.
func ClaudeProjectDir(home, cwd string) string {
	slug := []byte(resolve(cwd))
	for i, c := range slug {
		if !isAlnum(c) {
			slug[i] = '-'
		}
	}
	return filepath.Join(home, ".claude", "projects", string(slug))
}

func isAlnum(c byte) bool {
	return c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9'
}

// FindClaudeTranscript returns the newest transcript for cwd changed at or after since.
func FindClaudeTranscript(home, cwd string, since time.Time) (string, error) {
	files, err := filepath.Glob(filepath.Join(ClaudeProjectDir(home, cwd), "*.jsonl"))
	if err != nil {
		return "", fmt.Errorf("list claude transcripts: %w", err)
	}
	return newest(files, since, func(string) bool { return true })
}

// FindCodexRollout returns the newest Codex rollout for cwd changed at or after since.
// Rollouts are grouped by date, so the cwd is read from each file's first record.
func FindCodexRollout(home, cwd string, since time.Time) (string, error) {
	files, err := filepath.Glob(filepath.Join(home, ".codex", "sessions", "*", "*", "*", "rollout-*.jsonl"))
	if err != nil {
		return "", fmt.Errorf("list codex rollouts: %w", err)
	}
	want := resolve(cwd)
	return newest(files, since, func(f string) bool { return codexCwd(f) == want })
}

func newest(files []string, since time.Time, match func(string) bool) (string, error) {
	var best string
	var bestTime time.Time
	for _, f := range files {
		info, err := os.Stat(f)
		if err != nil || info.ModTime().Before(since) || !info.ModTime().After(bestTime) || !match(f) {
			continue
		}
		best, bestTime = f, info.ModTime()
	}
	if best == "" {
		return "", ErrNoTranscript
	}
	return best, nil
}

// codexCwd reads the working directory from a rollout's session_meta record.
func codexCwd(path string) string {
	f, err := os.Open(path) //nolint:gosec // a Codex rollout under the user's home
	if err != nil {
		return ""
	}
	defer func() { _ = f.Close() }()
	r := bufio.NewReaderSize(f, 64<<10)
	line, _ := r.ReadBytes('\n')
	var meta struct {
		Payload struct {
			Cwd string `json:"cwd"`
		} `json:"payload"`
	}
	if json.Unmarshal(line, &meta) != nil {
		return ""
	}
	return resolve(meta.Payload.Cwd)
}

// resolve gives the physical path, as agents record it: /tmp is /private/tmp on macOS.
func resolve(path string) string {
	if p, err := filepath.EvalSymlinks(path); err == nil {
		return p
	}
	return filepath.Clean(path)
}
