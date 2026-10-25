package agenthooks

import (
	"encoding/json"
	"strings"
)

// CursorHooksPath is the project's Cursor hooks file, relative to the repository root.
const CursorHooksPath = ".cursor/hooks.json"

// StopCommand sends the work-log follow-up when a Cursor agent ends its turn. It has no POSIX guard like the Claude
// commands because Cursor runs hook commands through PowerShell on Windows; where mem is missing the hook fails and
// Cursor carries on.
const StopCommand = "mem hook stop"

// cursorEntry is an entry in a Cursor hooks list, which holds commands directly rather than matcher groups.
type cursorEntry struct {
	Command string `json:"command"`
	Timeout int    `json:"timeout,omitempty"`
}

var stopEntry = cursorEntry{Command: StopCommand, Timeout: 30}

// SyncCursor adds, updates or removes mem's stop hook in .cursor/hooks.json. It keeps every other key, event and
// entry, writes only when mem's entry changes and reports whether it did.
func SyncCursor(root string, stop bool) (bool, error) {
	return syncFile(root, CursorHooksPath, `{"version": 1}`, []memEntry{{"stop", stopEntry, isCursorEntry, stop}})
}

func isCursorEntry(raw json.RawMessage) bool {
	var e cursorEntry
	return json.Unmarshal(raw, &e) == nil && strings.Contains(e.Command, StopCommand)
}
