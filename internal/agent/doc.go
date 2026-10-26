// Package agent turns what an agent CLI records about its own work into structured
// events: prompts, messages, commands, tool calls and file changes with their diffs
// (spike S2, F3-05). It reads the transcript files Claude Code and Codex write anyway,
// so it needs no headless mode and no change to the agent.
//
// Each agent has its own adapter, because their formats are undocumented and change
// between releases. An adapter skips what it does not understand; the terminal keeps
// working even when the structured view cannot keep up.
package agent
