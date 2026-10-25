---
title: Check the Cursor hooks in a real Cursor
created_at: "2026-10-25T15:47:59+02:00"
---

The Cursor support (spec cursor_support) is verified only with simulated hook input; Cursor is not installed on Benjamin's machine. In a trusted Cursor workspace of a mem project (a collaborator's, after a release), check: (1) mem's stop hook in .cursor/hooks.json runs, and with a work log due its follow-up is submitted once as the next message, marked as coming from mem, and the agent writes the log or says it is due; (2) Cursor launched from the Dock finds mem on PATH for hooks (see the hook logs in Cursor's Output panel); (3) the Claude hooks Cursor loads from .claude/settings.json cause no errors or delays (mem hook compact/prompt return at once under CURSOR_VERSION); (4) how Cursor treats the Claude SessionStart compact matcher. Record what you find in docs/status.md.
