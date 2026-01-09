---
title: Update spec detection for worktrees
status: completed
subtasks: []
created_at: '2026-01-09T13:39:28.228564'
updated_at: '2026-01-09T13:53:10.703944'
completed_at: '2026-01-09T13:53:10.703938'
---
Modify get_active_spec() and related functions in src/utils/specs.py to detect active spec based on worktree location, not just branch name. When in a worktree, resolve the spec from the worktree directory name.

## Completion Notes

Updated get_active_spec() and get_branch_status() in src/utils/specs.py to check for worktree first before falling back to branch-based detection. Added imports for worktree utilities.