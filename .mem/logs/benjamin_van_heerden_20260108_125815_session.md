---
created_at: '2026-01-08T12:58:15.386089'
username: benjamin_van_heerden
spec_slug: polish_and_small_improvements
---
# Work Log - Polish and improvements for git hooks and onboard

## Overarching Goals

Improve the mem developer experience with better git branch protection and enhanced onboard context for AI agents.

## What Was Accomplished

### Fixed git merge hook bypass via fast-forward

Discovered that the `pre-merge-commit` hook wasn't triggering for fast-forward merges. Solution: set `git config merge.ff false` to force merge commits.

- Added `configure_merge_settings()` to `src/commands/init.py`
- Called during `mem init` after pre-push hook creation

### Show recently completed specs when no active work

When there are no todo or merge_ready specs, onboard now shows the 2 most recently completed specs for context:

```
### Recently completed specs:
These were the last completed specs for context:
  - slug_1: Title 1
  - slug_2: Title 2
```

### Show git diff stat for active spec

Added `get_branch_diff_stat()` to `src/utils/specs.py` that runs `git diff dev --stat`. Now shown in:
- `mem onboard` under the active spec section
- `mem spec activate` after switching to the branch

Output example:
```
Files modified in this spec (vs dev):
 src/commands/init.py  | 22 ++++++++++++++++++++++
 src/utils/specs.py    | 21 +++++++++++++++++++++
```

### Branch merge rules in onboard

Added clear documentation of branch merge rules to the About mem section:

```
**Branch merge rules:**
- anything → dev (feature branches merge here)
- dev or hotfix/* → test
- test → main
```

## Key Files Affected

- `src/commands/init.py` - Added `configure_merge_settings()` function
- `src/commands/onboard.py` - Added diff stat display, recently completed specs, branch rules
- `src/commands/spec.py` - Added diff stat display on activate
- `src/utils/specs.py` - Added `get_branch_diff_stat()` function

## What Comes Next

Spec is complete. Ready for PR.
