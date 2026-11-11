---
title: Hook support for Grok
created_at: "2026-11-11T08:49:23+02:00"
---

Codex hooks shipped separately (.codex/hooks.json, see docs/design.md); this is the Grok half of the old hook_support_codex_grok todo. Research from 2026-10-25 is in .mem/specs/archive/hooks_codex_cursor_grok. Grok Build (grok 1.0.67) runs .claude/settings.json and .cursor/hooks.json hooks but discards SessionStart/UserPromptSubmit output and has no compaction injection; it reads AGENTS.md in full and restores it after compaction; it sets GROK_HOOK_EVENT, so mem hook prompt and compact can return early there. Its Stop hook can block with feedback, the same lever as Cursor's stop follow-up.
