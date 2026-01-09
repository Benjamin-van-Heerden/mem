---
title: Implement task rename command
status: completed
subtasks: []
created_at: '2026-01-09T15:49:48.377579'
updated_at: '2026-01-09T16:13:18.331720'
completed_at: '2026-01-09T16:13:18.331714'
---
Add mem task rename <old_title> <new_title> that updates the title in task frontmatter without changing the filename

## Completion Notes

Added rename_task function in src/utils/tasks.py that updates the title in frontmatter while keeping the filename stable. Added the rename CLI command in src/commands/task.py with helpful output showing old and new titles.