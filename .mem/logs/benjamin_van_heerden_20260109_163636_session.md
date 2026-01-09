---
created_at: '2026-01-09T16:36:36.512891'
username: benjamin_van_heerden
spec_slug: task_amendments_and_work_log_improvements
---
# Work Log - Complete work log and worktree workflow improvements

## Overarching Goals

Implement work log display improvements and add explicit worktree workflow guidance to prevent agents from continuing work in the main repo after spec assignment.

## What Was Accomplished

### Work Log Visual Improvements (src/commands/onboard.py)

1. **Visual separation**: Added box borders using Unicode box-drawing characters around each work log entry:
   ```
   ┌────────────────────────────────────────────────────────────────────┐
     2026-01-09 (username) - spec: spec_slug
   ├────────────────────────────────────────────────────────────────────┤
   [log body]
   └────────────────────────────────────────────────────────────────────┘
   ```

2. **Chronological order**: Reversed work log display order so oldest appears first, newest last (using `reversed(recent_logs)`)

3. **Smarter log selection**: When not on active spec, if all 3 recent logs are from same spec, keeps 2 from that spec and finds 1 from a different spec for diversity

4. **All logs for active spec**: Changed limit from 3 to 100 when fetching logs for active spec

5. **Completed spec bodies**: Updated "Recently completed specs" section to show spec body content (truncated to 500 chars)

### Worktree Workflow Guidance

1. **--spec flag protection (src/commands/task.py)**: Added check in `_resolve_spec_slug` to detect when `--spec` is used from main repo for a spec that has a worktree. Shows error with worktree path.

2. **spec new hints (src/commands/spec.py)**: Added "IMPORTANT: Worktree Workflow" section warning that after assign, a NEW agent session must start in the worktree.

3. **spec assign output (src/commands/spec.py)**: Enhanced with:
   - "WORKTREE READY - START NEW SESSION" header
   - "THIS SESSION MUST END HERE" stop signal
   - "WHY A NEW SESSION?" section explaining worktree isolation

## Key Files Affected

- `src/commands/onboard.py` - Work log formatting, ordering, selection logic, completed spec body display
- `src/commands/task.py` - Added worktree check for --spec flag
- `src/commands/spec.py` - Enhanced new and assign command output with worktree workflow guidance

## What Comes Next

All 11 tasks for the spec are complete. Ready to run `mem spec complete` to create PR.
