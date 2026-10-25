---
title: Document Cursor support
status: todo
created_at: "2026-10-25T15:40:06+02:00"
updated_at: "2026-10-25T15:40:06+02:00"
---

README: a short agent support section. Claude Code gets AGENTS.md, the compaction digest and the per-message nudge. Cursor gets AGENTS.md, the post-commit nudge and the stop follow-up when a log is due ([cursor] stop_hook = false turns it off; the workspace must be trusted). Cursor can't take context after compaction or per message. Mention .cursor/hooks.json in the init description. docs/design.md: the hook paragraphs gain a Cursor paragraph (the stop follow-up, its conditions and dedup, the quiet Claude hooks, and why there are no sessionStart, preCompact or beforeSubmitPrompt hooks); the config example gains [cursor]. docs/status.md: implemented, and verified by simulated stdin, not in a real Cursor; the unverified PATH question for Cursor launched from the Dock. .mem/structure.md: internal/agenthooks replaces internal/claude in the layout, key modules, data flow and external interfaces. Then ask the user whether a collaborator can check it in a real Cursor; record that as a todo if it cannot be done now.
