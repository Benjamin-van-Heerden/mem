---
title: Add auto-switch to dev in onboard command
status: completed
subtasks: []
created_at: '2026-01-08T10:20:48.867509'
updated_at: '2026-01-08T10:36:02.425993'
completed_at: '2026-01-08T10:36:02.425986'
---
Check current branch at start of mem onboard. If on main or test, switch to dev and log a message.

## Completion Notes

Added ensure_on_dev_branch() call at start of onboard(), prints warning to stderr when switching