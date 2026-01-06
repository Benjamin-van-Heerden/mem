---
title: Add PR merge and branch cleanup functions
status: completed
subtasks: []
created_at: '2026-01-06T15:13:04.046584'
updated_at: '2026-01-06T15:14:53.706232'
completed_at: '2026-01-06T15:14:53.706227'
---
Create functions to merge a PR and delete its branch.

Location: src/utils/github/api.py

Functions needed:

1. merge_pull_request(repo, pr_number: int) -> bool:
   - Merge the PR using GitHub API
   - Use 'squash' merge method
   - Return True on success, False on failure
   - Handle merge conflicts gracefully

2. delete_branch(repo, branch_name: str) -> bool:
   - Delete the remote branch after merge
   - Return True on success, False on failure
   - Should not fail if branch already deleted

Error handling:
- If PR has merge conflicts, inform user
- If checks are failing, warn but allow merge with confirmation
- If branch deletion fails, warn but don't fail the operation

## Completion Notes

Functions merge_pull_request() and delete_branch() already existed in api.py. merge_pull_request takes a PR object and merge method, returns success/sha/message dict. delete_branch takes repo and branch name, returns bool.