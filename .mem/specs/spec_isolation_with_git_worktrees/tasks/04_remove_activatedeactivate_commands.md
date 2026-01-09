---
title: Remove activate/deactivate commands
status: completed
subtasks: []
created_at: '2026-01-09T13:39:28.494008'
updated_at: '2026-01-09T13:59:01.590457'
completed_at: '2026-01-09T13:59:01.590449'
---
Remove spec activate and deactivate commands from src/commands/spec.py. Being in a worktree = spec is active. Update help text and onboard output accordingly.

## Completion Notes

Removed activate and deactivate commands from spec.py. Cleaned up unused imports (push_branch, smart_switch, GitHubError). These commands are no longer needed - being in a worktree means the spec is active.