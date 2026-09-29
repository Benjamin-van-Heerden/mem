---
title: Sync templates and report changed memories and skills in mem sync
status: completed
created_at: "2026-09-29T14:05:30+02:00"
updated_at: "2026-09-29T14:10:46+02:00"
completed_at: "2026-09-29T14:10:46+02:00"
---

mem sync snapshots readKnowledge before convergence, runs syncTemplates (internal/cli/onboard_updates.go) after it, and prints renderKnowledgeChanges. Replace the 're-read AGENTS.md if it changed' instruction with instructions grounded in what happened: review incoming changes relevant to current work, follow changed memories, tell the user about ⚠️ items. Tests: a teammate's memory set is shown with its text; a newly promoted template memory is installed and shown; output never contains 're-read AGENTS.md'.

## Completion Notes

mem sync now runs a shared catchUp (internal/cli/sync.go): snapshot memories and skills, converge.Sync, reload the project, syncTemplates, afterUpdates, read incoming changes, snapshot again. It prints TEMPLATES, INCOMING and the CHANGED MEMORIES/SKILLS sections, with instructions grounded in each (no re-read AGENTS.md). The hook compact command in task 4 reuses catchUp. Verified end to end: a teammate's memory set shows with its text; a memory promoted from another project is installed by mem sync and shown; the CLI package tests pass.
