---
title: Remove subtask utilities from tasks.py
status: todo
subtasks: []
created_at: '2026-01-10T15:55:03.246403'
updated_at: '2026-01-10T15:55:03.246403'
completed_at: null
---
In src/utils/tasks.py, remove these functions: list_subtasks(), has_incomplete_subtasks(), add_subtask(), complete_subtask(), delete_subtask(). These are the subtask operations embedded in task frontmatter.