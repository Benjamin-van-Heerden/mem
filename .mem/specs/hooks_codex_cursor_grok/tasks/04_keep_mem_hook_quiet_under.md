---
title: Keep mem hook quiet under Cursor and Grok
status: todo
created_at: "2026-10-25T15:33:39+02:00"
updated_at: "2026-10-25T15:33:39+02:00"
---

Cursor (CURSOR_VERSION) and Grok Build (GROK_HOOK_EVENT) run the hooks in .claude/settings.json but cannot use their output. Cursor rejects plain-text stdout and may fire the SessionStart compact entry at every session start; Grok discards SessionStart and UserPromptSubmit output, and runs mem hook prompt on every message. Add one helper in internal/cli/hook.go (for example hookCannotInject) that checks those two variables, with a comment giving that reason. mem hook compact and mem hook prompt return nil at once without output, sync or Git access when it is true. Do not use CLAUDE_PROJECT_DIR to detect Claude Code, since Cursor and Grok both set it. mem writes no .cursor/hooks.json and no .grok/hooks files. Test: with each variable set, both commands print nothing against a project whose state would otherwise produce output.
