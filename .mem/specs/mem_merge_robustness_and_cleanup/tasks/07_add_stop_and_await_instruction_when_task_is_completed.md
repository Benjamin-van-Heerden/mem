---
title: Add stop-and-await instruction when task is completed
status: completed
subtasks: []
created_at: '2026-01-07T10:28:11.907980'
updated_at: '2026-01-07T10:31:15.532679'
completed_at: '2026-01-07T10:31:15.532673'
---
When a task is marked complete via 'mem task complete', output an explicit instruction for the AI agent to:

1. Stop and take inventory of what was accomplished
2. Await further instructions from the human
3. NOT automatically proceed to the next task

Example output after completing a task:
  Task 'Add feature X' marked complete.
  
  AGENT INSTRUCTION: Stop here. Review what was accomplished and await 
  further instructions before proceeding to the next task.

This prevents 'runaway agents' that attempt to complete everything without human oversight.

## Completion Notes

Added stop instruction to task complete output, shows remaining task count