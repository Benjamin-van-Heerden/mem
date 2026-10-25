---
title: Maintain Codex hooks in .codex/hooks.json
status: todo
created_at: "2026-10-25T15:33:39+02:00"
updated_at: "2026-10-25T15:33:39+02:00"
---

Rename internal/claude to internal/agenthooks and generalise SyncHooks to take a target: ClaudeSettings (.claude/settings.json) and CodexHooks (.codex/hooks.json, the same {hooks:{Event:[{matcher,hooks:[{type,command,timeout}]}]}} shape; keep other top-level keys such as description). Codex entries are the same as Claude's: SessionStart with matcher compact (a regex on source) running CompactCommand with a 60 s timeout, and UserPromptSubmit running PromptCommand with a 10 s timeout. Add no additionalContextLimit. Comment that the entries must stay stable: Codex trusts project hooks by hash, so any change makes each user re-approve them in /hooks. init and import sync both targets. applyUpdates loops over both with today's per-file handling (gitignored: stays local; uncommitted edits: not committed; else published with checkpoint.ProjectFilesCommit); generalise settingsIgnored, sharedSettingsWarning and claudeHooksLine per target and agent name. When .codex/hooks.json is created or mem's entries in it change, the line tells Codex users to trust the project and approve the hooks in Codex's /hooks. Never write Codex trust records. Update init's and import's file listings. Tests: the agenthooks unit tests cover both targets; a cli test shows init writing both files and onboard publishing an updated .codex/hooks.json.
