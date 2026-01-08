---
title: Auto-switch to dev and git merge rules
status: completed
assigned_to: Benjamin-van-Heerden
issue_id: 21
issue_url: https://github.com/Benjamin-van-Heerden/mem/issues/21
branch: dev-benjamin_van_heerden-auto_switch_to_dev_and_git_merge_rules
pr_url: https://github.com/Benjamin-van-Heerden/mem/pull/22
created_at: '2026-01-08T10:19:07.385117'
updated_at: '2026-01-08T11:23:02.371407'
completed_at: '2026-01-08T11:23:02.369314'
last_synced_at: '2026-01-08T10:20:20.982673'
local_content_hash: 904238b1d6d2cb6040ddc6a6ac7f0e44f0d1c4c0792dfecb2a4ec285e354b72e
remote_content_hash: 904238b1d6d2cb6040ddc6a6ac7f0e44f0d1c4c0792dfecb2a4ec285e354b72e
---
## Overview

Two improvements to prevent accidental work in wrong branches:
1. `mem sync` and `mem onboard` should auto-switch to dev branch
2. Enforce git merge rules via branch protection

## Goals

- Prevent accidental commits to main/test branches
- Enforce proper git flow: dev → test → main
- Allow hotfix branches to merge directly to test

## Technical Approach

### Part 1: Auto-switch to dev
- Add check at start of `mem sync` and `mem onboard` commands
- If not on dev (and not on a feature branch), switch to dev
- Log a message when switching

### Part 2: Git merge rules
- Implement branch protection rules:
  - Anything can merge into dev
  - Only dev and hotfix/* branches can merge into test
  - Only test can merge into main

## Success Criteria

- Running `mem sync` or `mem onboard` from main/test switches to dev
- Merge attempts violating rules are blocked with clear error messages

## Notes

Branch protection can be implemented via GitHub branch protection rules or git hooks.
