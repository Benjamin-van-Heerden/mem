---
title: Remove subtask utilities from tasks.py
status: completed
subtasks: []
created_at: '2026-01-10T15:55:03.246403'
updated_at: '2026-01-10T16:13:30.593559'
completed_at: '2026-01-10T16:13:30.593552'
---
In src/utils/tasks.py, remove these functions: list_subtasks(), has_incomplete_subtasks(), add_subtask(), complete_subtask(), delete_subtask(). These are the subtask operations embedded in task frontmatter.

## Completion Notes

Removed list_subtasks, has_incomplete_subtasks, add_subtask, complete_subtask, delete_subtask functions and removed subtask checks from complete_task and complete_task_with_notes