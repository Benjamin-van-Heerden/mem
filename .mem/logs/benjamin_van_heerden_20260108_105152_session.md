---
created_at: '2026-01-08T10:51:52.402205'
username: benjamin_van_heerden
spec_slug: auto_switch_to_dev_and_git_merge_rules
---
# Work Log - Auto-switch to dev and git merge rules

## Overarching Goals

Implement two safety features to prevent accidental work in wrong branches:
1. Auto-switch to dev branch when running `mem sync` or `mem onboard` from main/test
2. Enforce git merge rules: anything->dev, dev/hotfix->test, test->main

## What Was Accomplished

### Added ensure_on_dev_branch() utility

Added helper function in `src/utils/specs.py` that checks current branch and switches to dev if on main or test.

### Integrated auto-switch into sync and onboard

Both `mem sync` and `mem onboard` now call `ensure_on_dev_branch()` at the start and display a warning if a switch occurred.

### Created pre-push git hook

Added `create_pre_push_hook()` function in `src/commands/init.py` that creates `.git/hooks/pre-push` to enforce:
- Anything can push to dev
- Only dev and hotfix/* can push to test  
- Only test can push to main

Note: Originally tried pre-merge-commit hook but it doesn't catch fast-forward merges.

### Added commit reminder to mem log

When there's an active spec, `mem log` now shows a reminder to commit and push changes to prevent losing work.

## Key Files Affected

- `src/utils/specs.py` - Added `ensure_on_dev_branch()` function
- `src/commands/sync.py` - Import and call ensure_on_dev_branch at start
- `src/commands/onboard.py` - Import and call ensure_on_dev_branch at start
- `src/commands/init.py` - Added `create_pre_push_hook()` function and call it during init
- `src/commands/log.py` - Added commit/push reminder when active spec exists

## Errors and Barriers

Lost all uncommitted changes during a git stash/checkout sequence while testing the merge hook. Had to redo all changes. Lesson: commit early and often, avoid branch switching with uncommitted work.

## What Comes Next

Still remaining on the spec:
1. Add `setup_branch_protection()` to `src/utils/github/api.py` and call from init (GitHub-side protection)
2. Add completion hints after task creation (task 4)
3. Improve task completion stop instruction (task 5)

Current changes are uncommitted - should commit and push now to protect progress.
