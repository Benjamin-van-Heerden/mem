---
title: Cleanup command improvements
status: merge_ready
assigned_to: Benjamin-van-Heerden
issue_id: 13
issue_url: https://github.com/Benjamin-van-Heerden/mem/issues/13
branch: dev-benjamin_van_heerden-cleanup_command_improvements
pr_url: null
created_at: '2026-01-07T11:16:55.097013'
updated_at: '2026-01-07T11:20:35.670698'
completed_at: null
last_synced_at: '2026-01-07T11:17:28.862293'
local_content_hash: 175083c8f7886c790eb13397a8d9dab7e76ee79de683e275660ada697fd82354
remote_content_hash: 175083c8f7886c790eb13397a8d9dab7e76ee79de683e275660ada697fd82354
---
## Overview

The `mem cleanup` command currently only deletes local branches for completed/abandoned specs. It should also delete remote branches on GitHub. Additionally, cleanup should run automatically as part of `mem sync` (and by extension `mem onboard`).

## Goals

- `mem cleanup` deletes both local AND remote branches for completed/abandoned specs
- `mem sync` runs cleanup automatically after syncing
- `mem onboard` gets cleanup for free since it runs sync

## Technical Approach

1. Add remote branch deletion to cleanup.py:
   - Use GitHub API to delete remote branches for completed/abandoned specs
   - Reuse `delete_branch` from `src/utils/github/api.py`

2. Integrate cleanup into sync:
   - Call cleanup logic at end of sync (after moving specs to completed)
   - Add `--no-cleanup` flag to skip if needed

## Success Criteria

- After merging a PR on GitHub and running `mem sync`, both local and remote branches are deleted
- `git branch -a` shows no stale branches from completed specs
- No manual cleanup steps needed in normal workflow

## Notes

Current gap: When PRs are merged directly on GitHub (not via `mem merge`), the remote branch isn't deleted. This should be handled by cleanup.
