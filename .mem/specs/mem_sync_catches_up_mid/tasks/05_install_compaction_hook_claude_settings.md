---
title: Install the compaction hook in .claude/settings.json
status: completed
created_at: "2026-09-29T14:05:30+02:00"
updated_at: "2026-09-29T14:15:13+02:00"
completed_at: "2026-09-29T14:15:13+02:00"
---

Merge mem's entry into .claude/settings.json: hooks.SessionStart item {matcher: compact, hooks: [{type: command, command: 'command -v mem >/dev/null 2>&1 && mem hook compact || true', timeout: 60}]}. Identify mem's entry by 'mem hook compact' in the command; add once, update in place, remove when [claude] compact_hook = false in .mem/config.toml (new optional ClaudeConfig in internal/project; absent means on; no schema bump). Preserve all other keys and hooks, and preserve key order where practical; write only when mem's entry changes. Called from mem init, mem import and onboard applyUpdates; publish .claude/settings.json with the other mem files, with the 'already had uncommitted edits' handling used for .gitignore. Tests: fresh install, existing settings preserved, idempotent, opt-out removes it.

## Completion Notes

New package internal/claude: SyncCompactHook merges mem's SessionStart/compact entry (command -v mem ... && mem hook compact || true, timeout 60) into .claude/settings.json with an order-preserving JSON object, identifies mem's entry by 'mem hook compact', writes only on change, removes it when disabled and never creates the file just to opt out. project.ClaudeConfig adds [claude] compact_hook (unset means on). init, import and onboard's applyUpdates install it; onboard publishes it with the other mem files or asks to commit it with existing edits, and reports JSON errors as a warning. Verified with unit tests (fresh install is exact and idempotent, other settings keep their order, opt-out removes only mem's entry) and an end-to-end onboard test for install, publish and opt-out; init in a scratch repo writes no empty [claude] table.
