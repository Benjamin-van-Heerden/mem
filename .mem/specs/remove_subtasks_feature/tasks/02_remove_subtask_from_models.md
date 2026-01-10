---
title: Remove subtask from models
status: completed
subtasks: []
created_at: '2026-01-10T15:54:59.403718'
updated_at: '2026-01-10T16:11:45.087362'
completed_at: '2026-01-10T16:11:45.087352'
---
In src/models.py: (1) Delete the SubtaskItem class, (2) Remove the 'subtasks' field from TaskFrontmatter, (3) Delete the create_subtask_frontmatter function and SubtaskFrontmatter class if they exist.

## Completion Notes

Removed SubtaskItem model, subtasks field from TaskFrontmatter, SubtaskFrontmatter model, and create_subtask_frontmatter function