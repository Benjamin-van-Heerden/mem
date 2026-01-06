---
title: Rewrite mem merge command to use GitHub API
status: completed
subtasks: []
created_at: '2026-01-06T15:13:15.794199'
updated_at: '2026-01-06T15:16:10.370021'
completed_at: '2026-01-06T15:16:10.370012'
---
Rewrite the merge command to query GitHub PRs instead of local specs.

Location: src/commands/merge.py

New workflow:
1. Query GitHub for merge-ready PRs using list_merge_ready_prs()
2. Display list to user:
   PR #8: Implement user auth (issue #5) - checks passing
   PR #12: Add dark mode (issue #11) - checks failing
3. If no PRs found, inform user and exit
4. Options:
   - mem merge --all: Merge all PRs with passing checks
   - mem merge: Interactive selection or merge all if only one
   - mem merge --force: Merge even if checks failing (with warning)
5. For each PR to merge:
   - Call merge_pull_request()
   - Call delete_branch() to clean up
   - Report success/failure
6. After all merges, run sync to update local state

Display format for PR list:
  #8  [Complete]: User Authentication
      Issue: #5 | Author: username | Checks: passing
      
  #12 [Complete]: Dark Mode  
      Issue: #11 | Author: username | Checks: failing

Add --dry-run flag to preview what would be merged without merging.

## Completion Notes

Rewrote merge.py to query GitHub PRs with list_merge_ready_prs() instead of local specs. Shows PR number, title, issue, author, branch, checks status. Supports --all, --dry-run, --force, --no-sync, --delete-branches/--keep-branches options.