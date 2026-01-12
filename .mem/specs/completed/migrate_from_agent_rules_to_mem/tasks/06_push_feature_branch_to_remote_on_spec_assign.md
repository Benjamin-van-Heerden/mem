---
title: Push feature branch to remote on spec assign
status: completed
subtasks: []
created_at: '2026-01-10T15:36:33.268772'
updated_at: '2026-01-10T18:22:25.054975'
completed_at: '2026-01-10T18:22:25.054968'
---
When running 'mem spec assign', after creating the local feature branch, also push it to origin with --set-upstream so the remote branch exists and is tracked from the start.

## Completion Notes

Added git push --set-upstream origin branch_name after worktree creation in spec assign command