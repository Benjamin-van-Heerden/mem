---
title: Make mem spec complete run sync first
status: completed
subtasks: []
created_at: '2026-01-06T14:56:30.048989'
updated_at: '2026-01-06T15:00:24.721234'
completed_at: '2026-01-06T15:00:24.721228'
---
The spec complete command should run mem sync before proceeding with completion:

1. At the start of the complete() function, call sync operations
2. If sync fails or has conflicts, abort completion and inform user
3. After sync succeeds, proceed with existing completion logic (create PR, update status, etc.)

This ensures GitHub state is current before creating the PR, and any merged specs are properly moved to completed before we try to complete our own.

Files to modify:
- src/commands/spec.py: Add sync call at start of complete() function

Implementation:
- Import and call the sync logic (or shell out to mem sync)
- Check sync result before proceeding
- Consider: should we sync again after completion? (PR creation updates GitHub)

## Completion Notes

Added git_fetch_and_pull() call at the start of spec complete command. If pull fails (conflicts), command aborts with helpful message.