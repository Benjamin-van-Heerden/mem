---
title: Fix --spec flag for worktree workflow
status: completed
subtasks: []
created_at: '2026-01-09T15:50:23.620142'
updated_at: '2026-01-09T16:26:07.772362'
completed_at: '2026-01-09T16:26:07.772356'
---
Ensure mem task new --spec creates tasks in the worktree (on feature branch) not in main repo (on dev)

## Completion Notes

Added check in _resolve_spec_slug to detect when --spec is used from main repo for a spec that has a worktree. Shows warning with path to worktree and exits with error.