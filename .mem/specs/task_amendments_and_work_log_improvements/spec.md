---
title: Task amendments and work log improvements
status: todo
assigned_to: null
issue_id: 27
issue_url: https://github.com/Benjamin-van-Heerden/mem/issues/27
branch: null
pr_url: null
created_at: '2026-01-09T15:38:41.022297'
updated_at: '2026-01-09T15:39:22.886728'
completed_at: null
last_synced_at: '2026-01-09T15:39:22.885709'
local_content_hash: 039a6b034b9a4dbe7fa93afdae92c94ecfb5f69e77685e15b71f66fbc3913a3b
remote_content_hash: 039a6b034b9a4dbe7fa93afdae92c94ecfb5f69e77685e15b71f66fbc3913a3b
---
## Overview

Add task amendment/rename functionality and improve work log display in onboard output.

## Goals

- Enable iterative task refinement with `mem task amend` and `mem task rename`
- Improve visual separation between work logs in onboard output
- Show work logs in chronological order (oldest first, newest last)
- Smarter work log selection when not on an active spec
- Show all related work logs when on an active spec

## Technical Approach

### 1. Task Amendment Commands

**`mem task amend <title> "<notes>"`**
- Appends `\n\n## Amendments\n\n{amendment_notes}` to the task body
- Resets task status to `todo` (allows re-completion cycle)
- This enables a pattern: Amendments -> Completion Notes -> Amendments -> Completion Notes...

**`mem task rename <old_title> "<new_title>"`**
- Updates the `title` field in task frontmatter
- Does not change the filename (slug remains stable)

**Hints after `mem task new`**
- Add output hints explaining amend/rename behavior

### 2. Work Log Display Improvements

**Visual separation**
- Add clear dividers between work logs (e.g., horizontal rules or box borders)

**Chronological order**
- Reverse the order so oldest appears first, newest last
- Applies to both active spec and no-spec contexts

**Smarter log selection (no active spec)**
- Show last 3 work logs
- BUT if all 3 are from the same spec, show only 2 from that spec
- Use the 3rd slot for the most recent log from any other spec
- This ensures variety in context

**All logs for active spec**
- When on an active spec, show ALL work logs related to that spec (not just 3)

## Success Criteria

- `mem task amend` adds amendments section and resets status to todo
- `mem task rename` updates task title in frontmatter
- `mem task new` shows hints about amend/rename
- Work logs have clear visual separation in onboard output
- Work logs appear oldest-first in onboard
- No-active-spec context shows diverse logs (not all from same spec)
- Active-spec context shows complete log history for that spec

## Notes

The amendment pattern supports iterative refinement where a task can go through multiple cycles of amendment and completion.
