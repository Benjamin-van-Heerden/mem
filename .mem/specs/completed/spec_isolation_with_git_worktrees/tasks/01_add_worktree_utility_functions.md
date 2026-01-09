---
title: Add worktree utility functions
status: completed
subtasks: []
created_at: '2026-01-09T13:39:13.897420'
updated_at: '2026-01-09T13:41:59.813971'
completed_at: '2026-01-09T13:41:59.813966'
---
Create src/utils/worktrees.py with functions to: detect if in worktree, get main repo path from worktree, create worktree with branch, remove worktree, list worktrees

## Completion Notes

Created src/utils/worktrees.py with: is_worktree(), get_main_repo_path(), create_worktree(), remove_worktree(), list_worktrees(), get_worktree_for_spec(), resolve_repo_and_spec()