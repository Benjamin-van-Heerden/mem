---
title: Integrate cleanup into mem sync
status: completed
subtasks: []
created_at: '2026-01-07T11:17:55.866030'
updated_at: '2026-01-07T11:19:59.217351'
completed_at: '2026-01-07T11:19:59.217345'
---
Call cleanup logic at end of sync command after moving specs to completed. Add --no-cleanup flag to skip if needed. This means onboard gets cleanup for free since it runs sync.

## Completion Notes

Added --no-cleanup flag to sync command and call run_cleanup() at end of sync. Cleanup runs silently and reports count if branches were deleted.