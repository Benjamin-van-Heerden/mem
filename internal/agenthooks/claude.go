// Package agenthooks maintains mem's hook entries in the configuration files of the coding agents it supports.
package agenthooks

import (
	"encoding/json"
	"strings"
)

// ClaudeSettingsPath is the project's shared Claude Code settings file, relative to the repository root.
const ClaudeSettingsPath = ".claude/settings.json"

// CompactCommand runs mem's catch-up after compaction; PromptCommand reports the work since the last work log
// when the user sends a message. Both do nothing where mem is not installed.
const (
	CompactCommand = "command -v mem >/dev/null 2>&1 && mem hook compact || true"
	PromptCommand  = "command -v mem >/dev/null 2>&1 && mem hook prompt || true"
)

type hookEntry struct {
	Matcher string        `json:"matcher,omitempty"`
	Hooks   []hookCommand `json:"hooks"`
}

type hookCommand struct {
	Type    string `json:"type"`
	Command string `json:"command"`
	Timeout int    `json:"timeout,omitempty"`
}

var (
	compactEntry = hookEntry{Matcher: "compact", Hooks: []hookCommand{{Type: "command", Command: CompactCommand, Timeout: 60}}}
	promptEntry  = hookEntry{Hooks: []hookCommand{{Type: "command", Command: PromptCommand, Timeout: 10}}}
)

// ClaudeHooks says which of mem's Claude Code hooks a project wants.
type ClaudeHooks struct {
	Compact, Prompt bool
}

// SyncClaude adds, updates or removes mem's hooks in .claude/settings.json: the SessionStart "compact" catch-up and
// the UserPromptSubmit work-log nudge. It keeps every other setting and its order, writes only when mem's entries
// change and reports whether it did.
func SyncClaude(root string, want ClaudeHooks) (bool, error) {
	return syncFile(root, ClaudeSettingsPath, "{}", []memEntry{
		{"SessionStart", compactEntry, claudeEntry("mem hook compact"), want.Compact},
		{"UserPromptSubmit", promptEntry, claudeEntry("mem hook prompt"), want.Prompt},
	})
}

// claudeEntry recognises mem's entry by a command in its nested hooks list.
func claudeEntry(marker string) func(json.RawMessage) bool {
	return func(raw json.RawMessage) bool { return isEntry(raw, marker) }
}

func isEntry(raw json.RawMessage, marker string) bool {
	var e hookEntry
	if json.Unmarshal(raw, &e) != nil {
		return false
	}
	for _, h := range e.Hooks {
		if strings.Contains(h.Command, marker) {
			return true
		}
	}
	return false
}
