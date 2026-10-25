---
title: Keep the Claude hooks quiet under Cursor
status: todo
created_at: "2026-10-25T15:40:06+02:00"
updated_at: "2026-10-25T15:40:06+02:00"
---

Cursor runs .claude/settings.json hooks, sets CURSOR_VERSION in every hook's environment, rejects plain-text output and may fire the SessionStart compact entry at every session start. Add one helper in internal/cli/hook.go, with a comment giving that reason, that reports whether CURSOR_VERSION is set. mem hook compact and mem hook prompt return nil at once when it is, before loading the project, with no output, sync or Git access. mem hook stop does not check it. Do not use CLAUDE_PROJECT_DIR, which Cursor also sets. Test: with CURSOR_VERSION set (t.Setenv), mem hook prompt prints nothing in a project where it would otherwise print the due nudge, and mem hook compact prints nothing (no digest).
