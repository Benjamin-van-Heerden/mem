---
title: Modify spec new to create worktree
status: completed
subtasks: []
created_at: '2026-01-09T13:39:27.959430'
updated_at: '2026-01-09T13:47:28.282414'
completed_at: '2026-01-09T13:47:28.282406'
---
Update src/commands/spec.py new() to create worktree in ../<project>-worktrees/<slug>/ with feature branch. Output the worktree path for user to cd into or start new agent session.

## Completion Notes

Updated spec new command to: get GitHub user, create spec file, create worktree at ../<project>-worktrees/<slug>/, create feature branch, auto-assign spec to current user. Updated help text to reflect new workflow.

## Completion Notes

Changed approach: spec new stays simple (creates spec file only), spec assign now creates the worktree + branch. This makes more sense logically - assignment is when someone claims the work and the worktree is created.