---
title: Add hints after task new
status: completed
subtasks: []
created_at: '2026-01-09T15:49:53.195054'
updated_at: '2026-01-09T16:14:07.725527'
completed_at: '2026-01-09T16:14:07.725521'
---
Update mem task new output to show hints about amend and rename behavior

## Completion Notes

Updated the new command output in src/commands/task.py to show a 'Refinement options' section with hints about rename and amend commands, including an explanation that amend resets status to todo for iterative refinement cycles.