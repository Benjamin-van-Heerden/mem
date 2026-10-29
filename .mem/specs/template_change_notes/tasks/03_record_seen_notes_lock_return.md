---
title: Record seen notes in the lock and return new ones from Sync
status: completed
created_at: "2026-10-29T08:27:22+02:00"
updated_at: "2026-10-29T08:45:28+02:00"
completed_at: "2026-10-29T08:45:28+02:00"
---

Extend .mem/templates.lock with [[change]] tables (template, name) sorted by key: readLock/writeLock in internal/templates/sync.go, the header comment, state tracking with lockChanged. Give Sync the set of templates whose notes are recorded silently (init: all installed templates; template use: the new one; syncTemplates: none) and update the three call sites (internal/cli/init.go, template.go, onboard_updates.go). For each unseen note of a used template: record it; if its template is not in the adopted set, append it to new Result.Changes. Drop records of notes no longer provided or of templates no longer used, silently. Tests in internal/templates: unseen note returned once and recorded; second Sync returns nothing; adopted template records without returning; dropped records; a lock without [[change]] still reads.

## Completion Notes

Moved the lock code out of sync.go (503 lines) into internal/templates/lock.go and added [[change]] records (LockChange: template, name; omitted when empty, sorted by key). Sync takes adopted templates variadically (init passes all its templates, template use the new one, syncTemplates none); reconcileChanges in changes.go records every unseen note, returns it in Result.Changes unless its template is adopted, and drops records of notes no longer given. TestSyncReturnsEachChangeNoteOnceExceptForAdoptedTemplates covers adoption, a second sync returning nothing without touching the lock, a new note, and a deleted note's record dropped. go test ./internal/templates/, go test ./internal/cli/ -run 'Template|Convention|Init' and go vet ./internal/templates/ pass; a lock without notes keeps its old layout.
