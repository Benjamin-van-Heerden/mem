---
title: Close GitHub issues when PR is merged to dev
status: merge_ready
assigned_to: Benjamin-van-Heerden
issue_id: 5
issue_url: https://github.com/Benjamin-van-Heerden/mem/issues/5
branch: dev-benjamin_van_heerden-close_github_issues_when_pr_is_merged_to_dev
pr_url: https://github.com/Benjamin-van-Heerden/mem/pull/6
created_at: '2026-01-06T14:32:40.669327'
updated_at: '2026-01-06T14:34:51.562404'
completed_at: null
last_synced_at: '2026-01-06T14:33:19.443848'
local_content_hash: c702bc975d8f66716b4b73bb46c6b5c216453e4db74bd88bcba6e1d3bec96f34
remote_content_hash: c702bc975d8f66716b4b73bb46c6b5c216453e4db74bd88bcba6e1d3bec96f34
---
## Overview

GitHub's "Closes #X" syntax in PR descriptions only auto-closes issues when merging to the default branch (usually `main`). Since mem merges to `dev`, linked issues remain open. We need to detect merged PRs during sync and close their linked issues.

## Goals

- Automatically close GitHub issues when their linked PR is merged
- Keep sync as the single source of truth for GitHub state

## Technical Approach

1. In `mem sync`, after moving specs to `completed/`:
   - Check if the spec has an `issue_id` and the issue is still open
   - If the PR was merged (spec moved to completed), close the issue via API
   - Add a comment like "Closed by PR #X"

2. Alternatively, enhance `mem merge` to close the issue after successful merge

## Success Criteria

- When a PR is merged to dev, the linked GitHub issue is automatically closed
- The closure is done via sync or merge command, not relying on GitHub's default behavior

## Notes

This is needed because we use `dev` as the working branch, not `main`.
