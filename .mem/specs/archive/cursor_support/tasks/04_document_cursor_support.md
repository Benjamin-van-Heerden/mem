---
title: Document Cursor support
status: completed
created_at: "2026-10-25T15:40:06+02:00"
updated_at: "2026-10-25T15:47:59+02:00"
completed_at: "2026-10-25T15:47:59+02:00"
---

README: a short agent support section. Claude Code gets AGENTS.md, the compaction digest and the per-message nudge. Cursor gets AGENTS.md, the post-commit nudge and the stop follow-up when a log is due ([cursor] stop_hook = false turns it off; the workspace must be trusted). Cursor can't take context after compaction or per message. Mention .cursor/hooks.json in the init description. docs/design.md: the hook paragraphs gain a Cursor paragraph (the stop follow-up, its conditions and dedup, the quiet Claude hooks, and why there are no sessionStart, preCompact or beforeSubmitPrompt hooks); the config example gains [cursor]. docs/status.md: implemented, and verified by simulated stdin, not in a real Cursor; the unverified PATH question for Cursor launched from the Dock. .mem/structure.md: internal/agenthooks replaces internal/claude in the layout, key modules, data flow and external interfaces. Then ask the user whether a collaborator can check it in a real Cursor; record that as a todo if it cannot be done now.

## Completion Notes

Documented Cursor support. README: init adds .cursor/hooks.json; a new Agents item says what Claude Code and Cursor each get and covers [cursor] stop_hook and workspace trust. docs/design.md: an In Cursor section (why there are no sessionStart, preCompact or beforeSubmitPrompt hooks; the stop entry, its conditions and dedup, the wording, the quiet Claude hooks); the config example gains [cursor]. docs/status.md: implemented and verified with simulated input only, plus the unverified PATH and matcher questions. .mem/structure.md: internal/agenthooks, hook stop and underCursor, CursorConfig, the external interfaces. Checked end to end with dist/mem-dev in a scratch repository: init wrote .cursor/hooks.json, hook stop returned the follow-up with a log due, hook prompt under CURSOR_VERSION printed nothing. A real Cursor check is recorded as the todo check_cursor_hooks_real_cursor.
