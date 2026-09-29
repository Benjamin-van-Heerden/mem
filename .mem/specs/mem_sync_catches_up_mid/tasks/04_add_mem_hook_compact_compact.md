---
title: Add mem hook compact with a compact digest
status: todo
created_at: "2026-09-29T14:05:30+02:00"
updated_at: "2026-09-29T14:05:30+02:00"
---

Hidden command 'mem hook compact' (internal/cli/hook.go). Tolerates the SessionStart JSON on stdin. Runs the same catch-up as mem sync (factor the shared steps so both call one function), then prints a plain-text digest under ~3 KB: branch line with ahead/behind and what the sync did; active spec assigned to the user and its next pending task; todos claimed by the user; incoming changes; changed memories in full and changed skills by name/path; nudges; a short instruction (continue with the current work, follow changed memories, raise ⚠️ items). No structure doc, docs or logs. On any error (e.g. not a mem project) print one line and exit 0. Tests: digest content and size in the teammate scenario; exit 0 outside a mem project.
