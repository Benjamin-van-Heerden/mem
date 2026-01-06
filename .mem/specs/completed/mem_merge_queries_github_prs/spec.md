---
title: mem merge queries GitHub PRs
status: completed
assigned_to: Benjamin-van-Heerden
issue_id: 9
issue_url: https://github.com/Benjamin-van-Heerden/mem/issues/9
branch: dev-benjamin_van_heerden-mem_merge_queries_github_prs
pr_url: https://github.com/Benjamin-van-Heerden/mem/pull/10
created_at: '2026-01-06T15:11:59.918258'
updated_at: '2026-01-06T15:37:10.237100'
completed_at: '2026-01-06T15:37:10.235944'
last_synced_at: '2026-01-06T15:12:35.150875'
local_content_hash: 1354e11140a49fa79220e89d2e22194d4981bd8f6092853ef5e0176274a69eb7
remote_content_hash: 1354e11140a49fa79220e89d2e22194d4981bd8f6092853ef5e0176274a69eb7
---
## Overview

Currently `mem merge` looks for local specs with `merge_ready` status, but this fails when on dev branch because completed specs only exist on their feature branches until merged. The solution is to have `mem merge` query GitHub directly for open PRs that are ready to merge.

## Goals

- `mem merge` queries GitHub for open PRs targeting dev
- Works regardless of local branch state
- Provides clear list of PRs ready to merge
- Merges selected PRs via GitHub API
- Cleans up merged branches
- Triggers sync after merge to update local state

## Technical Approach

1. Query GitHub API for open PRs where:
   - Base branch is `dev`
   - Title contains "[Complete]:" (our convention from `mem spec complete`)
   - Or PR is linked to a mem-spec issue

2. Display PRs to user with:
   - PR number and title
   - Linked issue number
   - Author
   - Whether checks are passing

3. Allow user to select which PRs to merge (or merge all)

4. For each selected PR:
   - Merge via GitHub API
   - Delete the feature branch
   
5. Run `mem sync` to:
   - Pull the merged changes
   - Move specs to completed
   - Close linked issues

## Success Criteria

- Running `mem merge` on dev shows list of PRs ready to merge
- Can merge PRs without needing local spec files
- After merge, `mem sync` properly completes the workflow
- Branches are cleaned up after merge

## Notes

This decouples `mem merge` from local file state entirely - it becomes a pure GitHub operation. The local state reconciliation happens via `mem sync` afterward.
