---
title: Fix spec abandon command for worktree workflow
status: completed
subtasks: []
created_at: '2026-01-10T11:48:12.567901'
updated_at: '2026-01-10T12:05:20.702201'
completed_at: '2026-01-10T12:05:20.702193'
---
The abandon command needs to be updated for the worktree-based workflow:

1. Can only abandon from the main repo directory (no active spec)
2. Should NOT be possible while inside a worktree (spec is active)
3. Should NOT allow cross-spec abandonment
4. Abandonment process:
   - Mark spec as abandoned
   - Move to abandoned directory
   - Git add, commit, push with appropriate message
5. Remove the branch switching logic (not needed when abandoning from main repo)
6. Should also clean up the worktree if one exists for the spec being abandoned

## Amendments

Additional requirement: If there are any open PRs or GitHub issues associated with the spec, abandonment should also close them appropriately with a comment indicating the spec was abandoned.

## Completion Notes

Rewrote abandon command for worktree workflow: (1) Must run from main repo, not worktree; (2) Checks no active spec; (3) Removes worktree if exists; (4) Closes GitHub PR with comment; (5) Closes GitHub issue with comment; (6) Moves spec to abandoned; (7) Commits and pushes changes. Also added close_pull_request function to src/utils/github/api.py.