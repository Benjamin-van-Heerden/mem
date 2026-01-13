---
created_at: '2026-01-13T14:15:33.374178'
username: benjamin_van_heerden
spec_slug: simplify_merge_into_commands_with_fast_forward_only
---
# Work Log - Simplify merge commands to use fast-forward only

## Overarching Goals

Simplify `mem merge into test` and `mem merge into main` commands to use fast-forward only merges instead of merge commits with back-merges. This eliminates the complexity of back-merge logic and prevents "X commits behind" warnings on GitHub since branches end up at the exact same commit SHA.

## What Was Accomplished

### Updated init.py merge settings
Changed `configure_merge_settings()` to set `merge.ff = only` instead of `merge.ff = false`. Updated docstring and echo message to reflect the new behavior.

### Simplified _merge_branch helper
Removed the `ff_only` parameter from `_merge_branch()` function. Now always uses `--ff-only` flag for all merges.

### Simplified _merge_into_test
Removed back-merge steps. New flow:
1. Check working directory is clean
2. Fetch latest from origin
3. Switch to test, pull latest
4. Merge dev into test (ff-only)
5. Push test
6. Switch back to dev

### Simplified _merge_into_main
Removed all back-merge steps. New flow:
1. Check working directory is clean
2. Fetch latest from origin
3. Switch to main, pull latest
4. Merge test into main (ff-only)
5. Push main
6. Switch back to dev

### Updated command docstring
Changed the `into` command docstring from "automatic back-merging" to "fast-forward only".

### Updated tests
- Changed module docstring from back-merge to fast-forward only
- Updated `test_into_main_dry_run_shows_all_steps` to check for ff-only output
- Renamed `test_into_main_all_branches_at_same_commit` to `test_into_main_test_and_main_at_same_commit` since dev is no longer back-merged

## Key Files Affected

- `src/commands/init.py` - Changed merge.ff from false to only
- `src/commands/merge.py` - Simplified _merge_branch, _merge_into_test, _merge_into_main, and updated into command docstring
- `tests/test_merge_into.py` - Updated docstring and test assertions for new behavior

## What Comes Next

All spec tasks completed. Ready to complete the spec and create PR.
