---
title: Move the hook files into internal/agenthooks with a Cursor stop entry
status: completed
created_at: "2026-10-25T15:40:06+02:00"
updated_at: "2026-10-25T15:44:41+02:00"
completed_at: "2026-10-25T15:44:41+02:00"
---

Rename internal/claude to internal/agenthooks. Split it into json.go (the ordered object, marshal, indent), claude.go (ClaudeSettingsPath, CompactCommand, PromptCommand, Hooks, SyncClaude: today's SyncHooks unchanged in behavior) and cursor.go (CursorHooksPath .cursor/hooks.json, StopCommand "mem hook stop" with no POSIX guard because Cursor uses PowerShell on Windows, SyncCursor(root, stop bool)). SyncCursor keeps version and every other key, event and entry. mem's entry is {"command":"mem hook stop","timeout":30} under hooks.stop (Cursor entries are flat, matched by the command text). It writes "version": 1 for a new file, removes the event key when mem's entry was the only one, writes only on change, and creates no file when the hook is off. Add project.CursorConfig ([cursor] stop_hook, *bool, unset means on, StopHookEnabled) to Config. init and import call SyncCursor(root, true) and list the file. applyUpdates handles .cursor/hooks.json like .claude/settings.json (ignored: stays local; uncommitted edits: not committed; else published with checkpoint.ProjectFilesCommit), with its own update line naming [cursor] stop_hook = false when off; generalise settingsIgnored to take a path. Tests: agenthooks unit tests for SyncCursor (new file, merge with other hooks and keys, removal, idempotence); the existing claude tests move with the package; a cli test where init writes the file and onboard removes the entry under stop_hook = false.

## Completion Notes

Renamed internal/claude to internal/agenthooks: json.go holds the ordered object and a shared syncFile/syncEntry, claude.go SyncClaude (behavior unchanged), cursor.go SyncCursor maintaining {"version": 1, "hooks": {"stop": [{"command": "mem hook stop", "timeout": 30}]}} while keeping other keys, events and entries. project.CursorConfig adds [cursor] stop_hook. init and import write .cursor/hooks.json and list it; onboard syncs it with the same ignored/uncommitted/publish handling as .claude/settings.json (settingsIgnored takes the path) and its own update line. Verified: agenthooks tests (fresh file, merge with other hooks, removal, opt-out), TestOnboardInstallsTheCursorHookAndHonoursTheOptOut, go test ./internal/cli/ ./internal/project/ ./internal/agenthooks/ and go vet pass.
