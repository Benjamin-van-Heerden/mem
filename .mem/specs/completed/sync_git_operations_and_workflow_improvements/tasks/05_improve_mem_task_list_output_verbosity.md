---
title: Improve mem task list output verbosity
status: completed
subtasks: []
created_at: '2026-01-06T14:57:42.488975'
updated_at: '2026-01-06T15:02:16.625559'
completed_at: '2026-01-06T15:02:16.625553'
---
The current mem task list output is too minimal - just shows checkboxes and titles. Improve it to show:

1. Task status explicitly (todo/in_progress/completed) not just checkbox
2. Task description (first 2-3 lines or ~150 chars, truncated with ...)
3. Subtask summary if any (e.g., '2/5 subtasks complete')
4. Created date
5. Add --verbose/-v flag for full description output

Current output:
[ ] Add git pull before sync operations

Improved output:
[todo] Add git pull before sync operations
       Before executing any sync operations, mem sync should run git fetch...
       Created: 2026-01-06

With --verbose:
[todo] Add git pull before sync operations
       <full description>
       Subtasks: 0/0
       Created: 2026-01-06

Files to modify:
- src/commands/task.py: Update list command output formatting

Consider also adding --json flag for machine-readable output.

## Completion Notes

Added --verbose/-v flag, description preview (first 150 chars), subtask summary (X/Y complete), created date. Verbose mode shows full description and individual subtasks.