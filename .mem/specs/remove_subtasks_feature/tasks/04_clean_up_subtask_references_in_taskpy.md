---
title: Clean up subtask references in task.py
status: todo
subtasks: []
created_at: '2026-01-10T15:55:08.259484'
updated_at: '2026-01-10T15:55:08.259484'
completed_at: null
---
In src/commands/task.py: (1) Remove the subtask hint from the 'new' command output (the line about 'mem subtask new'), (2) Remove subtask summary display from list_tasks_cmd (the 'Subtasks: X/Y complete' section), (3) Remove the incomplete subtasks check from the 'complete' command (the has_incomplete_subtasks check and error message).