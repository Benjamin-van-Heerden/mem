---
title: Update onboard for worktree awareness
status: completed
subtasks: []
created_at: '2026-01-09T13:39:29.029566'
updated_at: '2026-01-09T14:04:15.077313'
completed_at: '2026-01-09T14:04:15.077307'
---
Update src/commands/onboard.py to show worktree paths for active specs when run from main repo. Show which specs have worktrees and where they are located.

## Completion Notes

Updated onboard.py: added worktrees import, updated key commands to show new workflow (spec new, spec assign), added section showing active worktrees when in main repo, simplified next steps to reflect worktree-based workflow.