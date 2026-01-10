---
title: Clean up subtask references in task.py
status: completed
subtasks: []
created_at: '2026-01-10T15:55:08.259484'
updated_at: '2026-01-10T16:19:03.741897'
completed_at: '2026-01-10T16:19:03.741891'
---
In src/commands/task.py: (1) Remove the subtask hint from the 'new' command output (the line about 'mem subtask new'), (2) Remove subtask summary display from list_tasks_cmd (the 'Subtasks: X/Y complete' section), (3) Remove the incomplete subtasks check from the 'complete' command (the has_incomplete_subtasks check and error message).

## Completion Notes

Removed subtask hint from new command, subtask summary from list_tasks_cmd, and incomplete subtasks check from complete command