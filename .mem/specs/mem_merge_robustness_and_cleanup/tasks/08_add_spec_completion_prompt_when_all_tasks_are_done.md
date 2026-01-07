---
title: Add spec completion prompt when all tasks are done
status: completed
subtasks: []
created_at: '2026-01-07T10:28:18.292851'
updated_at: '2026-01-07T10:31:20.382113'
completed_at: '2026-01-07T10:31:20.382107'
---
When all tasks for a spec are completed, add a clear prompt indicating the spec is ready for completion:

1. In 'mem task complete' - after marking task complete, check if all tasks are now done
2. In 'mem task list' - if all tasks completed, show completion message
3. In 'mem spec show' - if all tasks completed, highlight this

Example message:
  All spec tasks are complete!
  Spec ready for completion via:
    mem spec complete <slug> "commit message"

This ensures the agent knows when to finalize the spec rather than looking for more work.

## Completion Notes

Added check after task complete - shows spec ready message when all tasks done