---
title: Remove subtask from models
status: todo
subtasks: []
created_at: '2026-01-10T15:54:59.403718'
updated_at: '2026-01-10T15:54:59.403718'
completed_at: null
---
In src/models.py: (1) Delete the SubtaskItem class, (2) Remove the 'subtasks' field from TaskFrontmatter, (3) Delete the create_subtask_frontmatter function and SubtaskFrontmatter class if they exist.