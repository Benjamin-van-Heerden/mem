---
title: Clean up existing stale branches
status: todo
subtasks: []
created_at: '2026-01-06T15:44:05.610209'
updated_at: '2026-01-06T15:44:05.610209'
completed_at: null
---
One-time cleanup of current stale branches after implementing the above.

Branches to delete locally:
- dev-benjamin_van_heerden-close_github_issues_when_pr_is_merged_to_dev
- dev-benjamin_van_heerden-global_config_and_template_alignment
- dev-benjamin_van_heerden-mem_merge_queries_github_prs
- dev-benjamin_van_heerden-sync_git_operations_and_workflow_improvements
- dev-benjamin_van_heerden-test_spec

Steps:
1. Run mem cleanup (after task 3 is done)
2. Verify with 'git branch -a' that branches are gone
3. Check GitHub to confirm remote branches were deleted
4. If any remote branches remain, delete them via GitHub API or manually

This task validates the cleanup command works correctly.