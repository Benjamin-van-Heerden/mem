---
title: Add mem hook compact with a compact digest
status: completed
created_at: "2026-09-29T14:05:30+02:00"
updated_at: "2026-09-29T14:12:41+02:00"
completed_at: "2026-09-29T14:12:41+02:00"
---

Hidden command 'mem hook compact' (internal/cli/hook.go). Tolerates the SessionStart JSON on stdin. Runs the same catch-up as mem sync (factor the shared steps so both call one function), then prints a plain-text digest under ~3 KB: branch line with ahead/behind and what the sync did; active spec assigned to the user and its next pending task; todos claimed by the user; incoming changes; changed memories in full and changed skills by name/path; nudges; a short instruction (continue with the current work, follow changed memories, raise ⚠️ items). No structure doc, docs or logs. On any error (e.g. not a mem project) print one line and exit 0. Tests: digest content and size in the teammate scenario; exit 0 outside a mem project.

## Completion Notes

Hidden 'mem hook compact' (internal/cli/compact.go) runs the shared catchUp and prints a digest: branch with ahead/behind and what the sync did, the user's active spec with progress and next task, their claimed todos, up to 5 incoming commits and 10 record changes, template lines, changed memories and skills, nudges, and a one-paragraph instruction. It prints one line and exits 0 on any error. Verified end to end: the teammate scenario digest contains all parts, stays under 3000 bytes and has no structure doc or logs; outside a mem project it exits 0 with a one-line note.
