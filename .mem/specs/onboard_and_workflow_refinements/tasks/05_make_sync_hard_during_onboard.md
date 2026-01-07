---
title: Make sync hard during onboard
status: completed
subtasks: []
created_at: '2026-01-07T11:42:12.788347'
updated_at: '2026-01-07T12:02:51.709694'
completed_at: '2026-01-07T12:02:51.709688'
---
Change sync call in onboard from dry-run to actual sync - need things in synced state before continuing work

## Completion Notes

Explicitly pass dry_run=False to sync call