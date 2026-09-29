---
title: Sync templates and report changed memories and skills in mem sync
status: todo
created_at: "2026-09-29T14:05:30+02:00"
updated_at: "2026-09-29T14:05:30+02:00"
---

mem sync snapshots readKnowledge before convergence, runs syncTemplates (internal/cli/onboard_updates.go) after it, and prints renderKnowledgeChanges. Replace the 're-read AGENTS.md if it changed' instruction with instructions grounded in what happened: review incoming changes relevant to current work, follow changed memories, tell the user about ⚠️ items. Tests: a teammate's memory set is shown with its text; a newly promoted template memory is installed and shown; output never contains 're-read AGENTS.md'.
