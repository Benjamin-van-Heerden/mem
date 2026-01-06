---
title: Prevent activating unassigned/unsynced specs
status: completed
subtasks: []
created_at: '2026-01-06T14:19:11.306195'
updated_at: '2026-01-06T14:20:05.856603'
completed_at: '2026-01-06T14:20:05.856594'
---
mem spec activate should fail if spec has no assignee or no GitHub issue

## Completion Notes

Added checks in activate command for issue_id and assigned_to. Users must sync and assign before activating.