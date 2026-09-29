---
title: Install the compaction hook in .claude/settings.json
status: todo
created_at: "2026-09-29T14:05:30+02:00"
updated_at: "2026-09-29T14:05:30+02:00"
---

Merge mem's entry into .claude/settings.json: hooks.SessionStart item {matcher: compact, hooks: [{type: command, command: 'command -v mem >/dev/null 2>&1 && mem hook compact || true', timeout: 60}]}. Identify mem's entry by 'mem hook compact' in the command; add once, update in place, remove when [claude] compact_hook = false in .mem/config.toml (new optional ClaudeConfig in internal/project; absent means on; no schema bump). Preserve all other keys and hooks, and preserve key order where practical; write only when mem's entry changes. Called from mem init, mem import and onboard applyUpdates; publish .claude/settings.json with the other mem files, with the 'already had uncommitted edits' handling used for .gitignore. Tests: fresh install, existing settings preserved, idempotent, opt-out removes it.
