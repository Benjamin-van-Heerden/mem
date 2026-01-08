---
title: Add auto-switch to dev in sync command
status: completed
subtasks: []
created_at: '2026-01-08T10:20:40.777101'
updated_at: '2026-01-08T10:31:16.834615'
completed_at: '2026-01-08T10:31:16.834607'
---
Check current branch at start of mem sync. If on main or test, switch to dev and log a message.

## Completion Notes

Added ensure_on_dev_branch() helper and called it at start of sync command