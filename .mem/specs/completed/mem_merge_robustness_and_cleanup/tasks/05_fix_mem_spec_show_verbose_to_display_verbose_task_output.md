---
title: Fix mem spec show --verbose to display verbose task output
status: completed
subtasks: []
created_at: '2026-01-07T10:04:09.946565'
updated_at: '2026-01-07T10:28:57.066272'
completed_at: '2026-01-07T10:28:57.066267'
---
When using --verbose flag on mem spec show, the tasks section should display the full verbose task output (with descriptions, subtasks, dates) instead of just the simple checkbox list

## Completion Notes

Updated spec.py show command to display full task descriptions, subtasks with checkboxes, and created dates when --verbose flag is used