---
title: mem merge robustness and cleanup
status: merge_ready
assigned_to: Benjamin-van-Heerden
issue_id: 11
issue_url: https://github.com/Benjamin-van-Heerden/mem/issues/11
branch: dev-benjamin_van_heerden-mem_merge_robustness_and_cleanup
pr_url: null
created_at: '2026-01-06T15:40:55.184385'
updated_at: '2026-01-07T11:12:26.697460'
completed_at: null
last_synced_at: '2026-01-06T15:42:23.362732'
local_content_hash: e61b2d20afb906dce45e0b588975e9d1fe11b00d0e13bae25a0bee4152db8006
remote_content_hash: e61b2d20afb906dce45e0b588975e9d1fe11b00d0e13bae25a0bee4152db8006
---
## Overview

The `mem merge` command has robustness issues that can leave the repository in an inconsistent state. Additionally, branch cleanup is not working correctly - both local and remote branches are left behind after merges.

## Goals

- `mem merge` should fail early if it can't complete cleanly
- All merged branches should be deleted (both local and remote)
- Stale local branches from completed specs should be cleaned up
- User should never be left in a broken state after running mem commands

## Success Criteria

- Running `mem merge` with uncommitted changes shows clear error and exits
- After successful merge, both local and remote branches are deleted
- `git branch -a` shows no stale branches from completed specs
- No manual git commands needed for normal workflow

## Notes

Current stale branches to clean up:
- dev-benjamin_van_heerden-close_github_issues_when_pr_is_merged_to_dev
- dev-benjamin_van_heerden-global_config_and_template_alignment  
- dev-benjamin_van_heerden-mem_merge_queries_github_prs
- dev-benjamin_van_heerden-sync_git_operations_and_workflow_improvements
- dev-benjamin_van_heerden-test_spec
