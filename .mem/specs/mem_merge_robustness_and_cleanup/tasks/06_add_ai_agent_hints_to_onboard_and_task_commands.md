---
title: Add AI agent hints to onboard and task commands
status: completed
subtasks: []
created_at: '2026-01-07T10:28:03.047101'
updated_at: '2026-01-07T10:31:02.719177'
completed_at: '2026-01-07T10:31:02.719171'
---
Add contextual hints/instructions for AI agents throughout the mem CLI:

1. In onboard output:
   - Add hints about working with tasks (how to create, complete, add subtasks)
   - Remind to mark tasks complete when done
   - Explain the subtask workflow

2. When creating tasks (mem task new):
   - Hint about adding subtasks for complex tasks
   - Show example: mem subtask new "subtask title" --task "task title"

3. When listing tasks (mem task list):
   - If tasks exist, remind about completing them
   - Show relevant next-step commands

4. General principle: Every mem command output should guide the agent to logical next steps

## Completion Notes

Added AGENT WORKFLOW HINTS section to onboard, hints to task new and task list commands