---
created_at: '2026-01-08T15:07:28.649246'
username: benjamin_van_heerden
---
# Work Log - Fix pre-merge-commit hook and add auto-recreation

## Overarching Goals

Fix the git pre-merge-commit hook which was incorrectly blocking valid merges (dev → test), and ensure the hook is automatically recreated for new clones via `mem onboard`.

## What Was Accomplished

### Fixed pre-merge-commit hook source branch detection

The hook was using `git rev-parse --abbrev-ref MERGE_HEAD` which returns the literal string "MERGE_HEAD" instead of the actual branch name. This caused all merges to fail validation.

Fixed by using `git name-rev --name-only MERGE_HEAD` which correctly returns the branch name (e.g., "dev").

```bash
# Before (broken)
SOURCE_BRANCH=$(git rev-parse --abbrev-ref MERGE_HEAD 2>/dev/null || echo "")

# After (working)
SOURCE_BRANCH=$(git name-rev --name-only MERGE_HEAD 2>/dev/null | sed 's|remotes/origin/||')
```

### Replaced pre-push hook with pre-merge-commit hook in init.py

The init.py was creating a `pre-push` hook but the actual enforcement was via a `pre-merge-commit` hook. Renamed the function to `create_pre_merge_commit_hook()` with the corrected logic.

Added a `quiet` parameter for silent operation when called from onboard.

### Added silent hook check to onboard

`mem onboard` now calls `create_pre_merge_commit_hook(quiet=True)` at startup, ensuring the hook exists for all users who run onboard - even on fresh clones where the hook wouldn't exist.

## Key Files Affected

- `.git/hooks/pre-merge-commit` - Fixed the source branch detection
- `src/commands/init.py` - Replaced `create_pre_push_hook()` with `create_pre_merge_commit_hook()`, added `quiet` parameter
- `src/commands/onboard.py` - Added import and silent call to ensure hook exists

## What Comes Next

Changes are uncommitted on dev branch. Could be committed directly or wrapped into a spec if desired.
