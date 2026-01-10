---
title: Remove subtasks feature
status: completed
assigned_to: Benjamin-van-Heerden
issue_id: 34
issue_url: https://github.com/Benjamin-van-Heerden/mem/issues/34
branch: dev-benjamin_van_heerden-remove_subtasks_feature
pr_url: https://github.com/Benjamin-van-Heerden/mem/pull/35
created_at: '2026-01-10T15:54:30.433486'
updated_at: '2026-01-10T16:26:26.535279'
completed_at: '2026-01-10T16:26:26.534698'
last_synced_at: '2026-01-10T16:03:54.024526'
local_content_hash: 63ed17ccee87034b5d9f0a0aee6ddbb2993203c7d3268932faef73e0bcc0f3bc
remote_content_hash: 63ed17ccee87034b5d9f0a0aee6ddbb2993203c7d3268932faef73e0bcc0f3bc
---
## Overview

Remove the subtasks feature from mem entirely. Subtasks add unnecessary complexity and overhead for agent workflows. Tasks with good completion notes are sufficient for tracking work progress across sessions.

## Rationale

- **Extra commands** - `mem subtask new`, `mem subtask complete` add friction without proportional benefit
- **Overhead** - Most tasks don't need this level of granularity
- **Agent workflow** - Agents naturally know what they're doing within a task. Subtasks feel like micromanagement.
- **Completion blocking** - Having to complete all subtasks before completing a task can be annoying if priorities shift
- **Completion notes suffice** - The notes added when completing a task already capture what was accomplished

## Goals

- Remove all subtask-related commands (`mem subtask new`, `mem subtask complete`, `mem subtask list`, `mem subtask delete`)
- Remove subtask fields from task frontmatter and models
- Remove subtask-related utility functions
- Clean up any references to subtasks in help text and hints
- Simplify the task completion flow (no more "incomplete subtasks" blocking)

## Technical Approach

### Files to modify:

1. **src/commands/subtask.py** - Delete this entire file
2. **main.py** - Remove the subtask app registration
3. **src/models.py** - Remove `SubtaskItem` model and `subtasks` field from `TaskFrontmatter`
4. **src/utils/tasks.py** - Remove all subtask-related functions:
   - `list_subtasks()`
   - `has_incomplete_subtasks()`
   - `add_subtask()`
   - `complete_subtask()`
   - `delete_subtask()`
5. **src/commands/task.py** - Remove:
   - Subtask hint from `new` command output
   - Subtask summary display from `list_tasks_cmd`
   - Incomplete subtasks check from `complete` command
6. **src/commands/onboard.py** - Check for any subtask references and remove

### Migration consideration:

Any existing tasks with subtasks in their frontmatter will simply have that field ignored after removal. No migration needed - the subtasks list will just be orphaned data that doesn't affect anything.

## Success Criteria

- `mem subtask` command no longer exists
- `mem task new` output has no subtask hint
- `mem task complete` works without checking for subtasks
- `mem task list` doesn't show subtask counts
- All tests pass (if any exist for subtasks, remove them)
- No references to "subtask" remain in codebase (except possibly in git history)

## Notes

- This is a simplification/removal, not a feature addition
- Existing task files with subtask frontmatter will still work - the field will just be ignored
