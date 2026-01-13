---
title: Simplify merge into commands with fast-forward only
status: merge_ready
assigned_to: Benjamin-van-Heerden
issue_id: 45
issue_url: https://github.com/Benjamin-van-Heerden/mem/issues/45
branch: dev-benjamin_van_heerden-simplify_merge_into_commands_with_fast_forward_only
pr_url: https://github.com/Benjamin-van-Heerden/mem/pull/46
created_at: '2026-01-13T13:51:00.593412'
updated_at: '2026-01-13T14:16:25.969244'
completed_at: null
last_synced_at: '2026-01-13T13:55:25.916833'
local_content_hash: 4afbe224a8d9a5c8f7cfa66c2cf91abbeba5be629d88d24afc960c6740fb2583
remote_content_hash: 4afbe224a8d9a5c8f7cfa66c2cf91abbeba5be629d88d24afc960c6740fb2583
---
## Overview

Simplify `mem merge into test` and `mem merge into main` commands to use fast-forward only merges instead of merge commits with back-merges.

Currently, the commands create merge commits and then attempt to back-merge to keep branches in sync. This is problematic because:
1. The back-merges use `--ff-only` which fails if the source branch has moved ahead
2. Merge commits create "X commits behind" warnings on GitHub
3. The logic is complex and error-prone

With fast-forward only merges, branches end up at the exact same commit SHA, eliminating "behind" warnings entirely.

## Goals

- Remove merge commits from the merge flow
- Remove back-merge logic (no longer needed with ff-only)
- Update `init.py` to set `merge.ff = only` instead of `merge.ff = false`
- Simplify the merge into commands significantly

## Technical Approach

### 1. Update `src/commands/init.py`

Change `configure_merge_settings()` to set `merge.ff = only` instead of `merge.ff = false`:
```python
subprocess.run(
    ["git", "config", "merge.ff", "only"],
    cwd=project_root,
    check=True,
    capture_output=True,
)
```

Update the echo message accordingly.

### 2. Simplify `_merge_into_test()` in `src/commands/merge.py`

New flow:
1. Check working directory is clean
2. Fetch latest from origin
3. Switch to test, pull latest
4. Merge dev into test with `--ff-only`
5. Push test
6. Switch back to dev

Remove steps 6-8 (back-merge to dev, push dev). No back-merge needed since we're not creating merge commits.

Update success message to: "test is now at the same commit as dev."

### 3. Simplify `_merge_into_main()` in `src/commands/merge.py`

New flow:
1. Check working directory is clean
2. Fetch latest from origin  
3. Switch to main, pull latest
4. Merge test into main with `--ff-only`
5. Push main
6. Switch back to dev

Remove steps 7-10 (back-merge to test, back-merge to dev, push both). 

Update success message to: "main is now at the same commit as test."

### 4. Update `_merge_branch()` helper

Change to always use `--ff-only`:
```python
def _merge_branch(source: str) -> tuple[bool, str]:
    """Merge source branch into current with ff-only. Returns (success, error_message)."""
    cwd = ENV_SETTINGS.caller_dir
    result = subprocess.run(
        ["git", "merge", "--ff-only", source],
        cwd=cwd,
        capture_output=True,
        text=True,
    )
    if result.returncode != 0:
        return False, result.stderr.strip()
    return True, ""
```

Remove the `ff_only` parameter since it's now always true.

### 5. Update dry-run output

Update the dry-run messages in both functions to reflect the simplified flow.

## Success Criteria

- `mem merge into test` performs ff-only merge from dev to test
- `mem merge into main` performs ff-only merge from test to main  
- No merge commits are created
- No back-merge logic exists
- `mem init` sets `merge.ff = only`
- All existing tests pass (update tests as needed)
- Commands fail gracefully if ff-only is not possible (target branch has diverged)

## Notes

- The pre-merge-commit hook in init.py can remain unchanged - it still enforces branch merge rules
- If a user's repo already has `merge.ff = false`, they'll need to run `git config merge.ff only` manually or re-run `mem init`
- Fast-forward only means if test/main have commits not on dev, the merge will fail - this is the desired behavior as all work should flow through dev
