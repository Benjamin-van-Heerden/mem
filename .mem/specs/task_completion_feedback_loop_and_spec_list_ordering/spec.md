---
title: Task completion feedback loop and spec list ordering
status: todo
assigned_to: Benjamin-van-Heerden
issue_id: 49
issue_url: https://github.com/Benjamin-van-Heerden/mem/issues/49
branch: dev-benjamin_van_heerden-task_completion_feedback_loop_and_spec_list_ordering
pr_url: null
created_at: '2026-01-14T10:47:32.762192'
updated_at: '2026-01-14T10:53:54.270051'
completed_at: null
last_synced_at: '2026-01-14T10:48:37.567766'
local_content_hash: 604e581c2411df98b2f2857c6c1f6394149f80defc695e81ecb56583918d06ca
remote_content_hash: 604e581c2411df98b2f2857c6c1f6394149f80defc695e81ecb56583918d06ca
---
## Overview

Two improvements to mem CLI:

1. **Two-step task completion flow** - Add a feedback loop to `mem task complete` so agents must get user confirmation before actually completing a task
2. **Spec list ordering** - Make `mem spec list -s completed` display specs ordered by completion date with human-readable dates

## Goals

- Prevent agents from marking tasks as complete when they aren't actually complete
- Create a clear feedback loop: agent summarizes work -> user confirms -> task marked complete
- Improve completed spec list readability with chronological ordering and dates

## Technical Approach

### Task Completion Feedback Loop

Modify `mem task complete` command in `src/commands/task.py`:

**Without `--accept` flag:**
- Do NOT mark the task as complete
- Output agent instructions telling it to:
  1. Summarize what was done for this task
  2. Ask the user if the work is acceptable
  3. If acceptable, run the command again with `--accept`
  4. If not acceptable, iterate on the work

**With `--accept` flag:**
- Actually mark the task as complete (current behavior)
- Show remaining tasks / completion status

### Spec List Ordering

Modify `mem spec list` in `src/commands/spec.py`:

- When listing completed specs (`-s completed`), sort by `completed_at` datetime
- Display the completion date in human-readable format (e.g., "Jan 14, 2026" or relative like "2 days ago")
- Keep existing ordering for other statuses

## Success Criteria

- `mem task complete "title" "notes"` without `--accept` outputs agent instructions and does NOT complete the task
- `mem task complete "title" "notes" --accept` actually completes the task
- `mem spec list -s completed` shows specs ordered by completion date (most recent first or last, TBD)
- Completion dates are displayed in a human-readable format

## Notes

The `notes` argument to `mem task complete` should still be required even without `--accept` - the agent should have done the work and have notes ready, this is just the confirmation step before persisting.
