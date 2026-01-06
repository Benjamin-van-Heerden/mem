---
title: Integrate sync after merge
status: completed
subtasks: []
created_at: '2026-01-06T15:13:25.771515'
updated_at: '2026-01-06T15:16:17.866422'
completed_at: '2026-01-06T15:16:17.866414'
---
After merging PRs, automatically run sync to update local state.

In merge.py after successful merges:
1. Call the sync logic (import from sync.py or shell out)
2. This will:
   - Pull the merged changes to local
   - Detect merged PRs and move specs to completed/
   - Close linked GitHub issues
   - Commit and push the state changes

Options:
- Add --no-sync flag to skip automatic sync
- If sync fails, warn but don't fail the merge operation

This ensures a single 'mem merge' command fully completes the workflow:
PR merged -> branch deleted -> changes pulled -> spec completed -> issue closed

## Completion Notes

Added sync call at end of merge command (imports and calls sync function from sync.py). Added --no-sync flag to skip. Sync runs automatically after successful merges to pull changes, move specs to completed, and close issues.