---
created_at: '2026-01-08T11:20:36'
username: benjamin_van_heerden
spec_slug: auto_switch_to_dev_and_git_merge_rules
---
# Work Log - Task UX improvements and log message clarity

## Overarching Goals

Complete the remaining tasks for the auto_switch_to_dev_and_git_merge_rules spec:
1. Add completion hints after task creation
2. Improve task completion stop instruction

Also fix ambiguity in log.py message and standardize "commit message" wording.

## What Was Accomplished

### Added completion hint to task new command

Modified `src/commands/task.py` to show how to complete a task after creating it:

```
Hints:
  Complete with: mem task complete "task title" "notes"
  For complex tasks, break into subtasks: mem subtask new "subtask title" --task "task title"
```

### Improved task completion stop instruction

Replaced the shouty all-caps instruction with a structured agent instruction format:

```
  ✓ Task completed: {title}

  [AGENT INSTRUCTION]
  Your next response must:
  1. Summarize what was done for this task
  2. Ask the user if they want to continue
  Do NOT call any tools. Do NOT start the next task.
```

### Clarified log.py git instructions

Rewrote the ambiguous message to be clear about when to commit:

```
If this is the LAST log before completing the spec:
  No action needed - `mem spec complete` handles git automatically.

Otherwise, commit and push your changes:
  git add -A && git commit -m '<describe what was done>' && git push
```

### Standardized commit message wording

Changed all instances of `"commit message"` to `"detailed commit message"` in:
- src/commands/onboard.py (4 locations)
- src/commands/task.py (2 locations)

## Key Files Affected

- `src/commands/task.py` - Added completion hint, improved stop instruction, updated commit message wording
- `src/commands/log.py` - Clarified git instructions for last log scenario
- `src/commands/onboard.py` - Updated commit message wording in 4 locations

## What Comes Next

All tasks for this spec are complete. Ready to create PR.
