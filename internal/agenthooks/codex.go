package agenthooks

// CodexHooksPath is the project's Codex hooks file, relative to the repository root. Codex loads it only in a project
// the user trusts, and runs each entry only after the user approved it in /hooks.
const CodexHooksPath = ".codex/hooks.json"

// codexEntry is a matcher group in Codex's hooks file, which has the shape of Claude Code's hooks object.
type codexEntry struct {
	Matcher string         `json:"matcher,omitempty"`
	Hooks   []codexCommand `json:"hooks"`
}

// codexCommand gives Windows its own command because the POSIX guard in the shared commands does not run in its
// shells; where mem is missing there, the hook fails and Codex carries on.
type codexCommand struct {
	Type           string `json:"type"`
	Command        string `json:"command"`
	CommandWindows string `json:"commandWindows,omitempty"`
	Timeout        int    `json:"timeout,omitempty"`
}

// Codex records each approval in /hooks under the hooks file's absolute path and the entry's place in it (event, group,
// handler), with a hash of its definition, and asks again for an entry whose record does not match. So syncFile updates
// mem's entries where they are, and once released they must stay exactly as they are.
var (
	codexCompactEntry = codexEntry{Matcher: "compact", Hooks: []codexCommand{{Type: "command", Command: CompactCommand, CommandWindows: "mem hook compact", Timeout: 60}}}
	codexPromptEntry  = codexEntry{Hooks: []codexCommand{{Type: "command", Command: PromptCommand, CommandWindows: "mem hook prompt", Timeout: 10}}}
)

// CodexHooks says which of mem's Codex hooks a project wants.
type CodexHooks struct {
	Compact, Prompt bool
}

// SyncCodex adds, updates or removes mem's hooks in .codex/hooks.json: the SessionStart "compact" catch-up and the
// UserPromptSubmit work-log nudge, whose plain stdout Codex adds to the model's context as it does Claude Code's. It
// keeps every other key, event and entry, writes only when mem's entries change and reports whether it did.
func SyncCodex(root string, want CodexHooks) (bool, error) {
	return syncFile(root, CodexHooksPath, "{}", []memEntry{
		{"SessionStart", codexCompactEntry, commandEntry("mem hook compact"), want.Compact},
		{"UserPromptSubmit", codexPromptEntry, commandEntry("mem hook prompt"), want.Prompt},
	})
}
