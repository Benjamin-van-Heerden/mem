---
created_at: '2026-01-14T11:08:57.428482'
username: benjamin_van_heerden
spec_slug: task_completion_feedback_loop_and_spec_list_ordering
---
# Work Log - Task completion feedback loop and spec list ordering

## Overarching Goals

Implement two improvements to the mem CLI:
1. Add a two-step confirmation flow to `mem task complete` so agents must get user confirmation before actually marking a task complete
2. Improve the completed specs list to show specs ordered by completion date with human-readable dates

## What Was Accomplished

### Task Completion Feedback Loop

Modified `src/commands/task.py` to add an `--accept` flag to the `complete` command:

- **Without `--accept`**: The command outputs agent instructions including:
  - The task title and full description (including any amendments)
  - The agent's completion notes
  - Instructions to summarize work, ask user for confirmation, and only run with `--accept` if approved
  - The task is NOT marked as complete

- **With `--accept`**: The command actually marks the task as complete (original behavior)

Also improved the task completion output:
- When tasks remain, now shows the actual task titles instead of just a count
- When all tasks are complete, reminds agent to create a work log first

### Spec List Completion Date Ordering

Modified `src/commands/spec.py` to improve the completed specs listing:

- Added `_format_completed_date()` helper to format ISO dates as "Jan 13, 2026"
- Completed specs are now sorted by `completed_at` (oldest first, most recent at bottom)
- Changed column layout for completed specs to show: Completed date, Title, Slug

### Onboard Signature Fix

Updated `src/commands/onboard.py` to show correct command signature:
- Changed from `mem task complete "title"` to `mem task complete "title" "notes"`
- Did NOT mention `--accept` flag (intentionally hidden until agent attempts completion)

## Key Files Affected

- `src/commands/task.py` - Added `--accept` flag, task description display, remaining tasks list, work log reminder
- `src/commands/spec.py` - Added completion date sorting and formatting for completed specs list
- `src/commands/onboard.py` - Fixed task complete signature in key commands

## What Comes Next

All spec tasks are complete. The spec is ready for PR creation and merge.
