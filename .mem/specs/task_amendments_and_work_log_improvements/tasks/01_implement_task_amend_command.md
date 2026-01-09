---
title: Implement task amend command
status: completed
subtasks: []
created_at: '2026-01-09T15:39:41.037937'
updated_at: '2026-01-09T16:10:21.783500'
completed_at: '2026-01-09T16:10:21.783494'
---
Add mem task amend <title> <notes> that appends an Amendments section to the task body and resets status to todo

## Completion Notes

Added amend_task function in src/utils/tasks.py that appends an Amendments section to the task body and resets status to todo. Added the amend CLI command in src/commands/task.py with --spec flag support and helpful output messages.