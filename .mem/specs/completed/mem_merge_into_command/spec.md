---
title: mem merge into command
status: completed
assigned_to: Benjamin-van-Heerden
issue_id: 43
issue_url: https://github.com/Benjamin-van-Heerden/mem/issues/43
branch: dev-benjamin_van_heerden-mem_merge_into_command
pr_url: https://github.com/Benjamin-van-Heerden/mem/pull/44
created_at: '2026-01-12T11:09:06.582678'
updated_at: '2026-01-12T15:49:27.735520'
completed_at: '2026-01-12T15:49:27.734232'
last_synced_at: '2026-01-12T11:21:43.147352'
local_content_hash: 7b9de1cdd778c41633a234e5b33d18133cf1c339835a1a32044f89390ecf97af
remote_content_hash: 7b9de1cdd778c41633a234e5b33d18133cf1c339835a1a32044f89390ecf97af
---
## Overview

Add a `mem merge into <test|main>` subcommand that merges between the main branches with automatic back-merging to eliminate GitHub's "X commits behind" warnings.

The issue: When merging `dev → test → main`, each merge creates a merge commit that only exists on the target branch. GitHub then shows warnings like "main is 1 commit behind dev" even though the code content is identical. This is confusing and clutters the GitHub UI.

The solution: After each forward merge, perform a fast-forward back-merge so both branches point to the same commit.

## Goals

- Add `mem merge into test` command (works from `dev` branch)
- Add `mem merge into main` command (works from `dev` branch, handles intermediate steps automatically)
- Automatic back-merge (ff-only) to keep branches synced at the same commit
- Progressive hints guiding users through the merge workflow
- `mem merge into main` runs in dry-run mode by default, requires `--force` to execute
- Clear error messages with recovery recommendations on failure
- Keep existing `mem merge` behavior for merging PRs unchanged

## Technical Approach

### 1. Convert merge.py to Typer sub-app
Currently `merge.py` exports a single `merge` function. Convert it to use a Typer app (like `spec.py` does) so we can have subcommands.

### 2. Keep existing merge as default command
The existing PR merge logic becomes accessible via `mem merge` (with all its current options like `--all`, `--dry-run`, `--force`, etc.).

### 3. Add `into` subcommand
New command: `mem merge into <target>` where target is `test` or `main`.

**All commands work from `dev` branch.** The commands handle switching to the required branches automatically.

**Flow for `mem merge into test`** (run from `dev`):
1. Check working directory is clean
2. Fetch latest from remote
3. Switch to `test`
4. Pull latest `test`
5. Merge `dev` into `test` (regular merge)
6. Push `test`
7. Switch back to `dev`
8. Merge `test` into `dev` (ff-only back-merge)
9. Push `dev`
10. Confirm success
11. Print hint: "Next step: mem merge into main"

**Flow for `mem merge into main`** (run from `dev`):
- By default, runs in **dry-run mode**: shows what would happen, then prints instructions to run `mem merge into main --force` to actually execute
- With `--force` flag, executes the full merge:
  1. Check working directory is clean
  2. Fetch latest from remote
  3. Switch to `test`, pull latest
  4. Switch to `main`, pull latest
  5. Merge `test` into `main` (regular merge)
  6. Push `main`
  7. Merge `main` back into `test` (ff-only)
  8. Push `test`
  9. Merge `test` back into `dev` (ff-only)
  10. Push `dev`
  11. Switch back to `dev`
  12. Confirm success

### 4. Usage hints (progressive disclosure)
- After `mem merge` (PR merge) completes: print hint "Next step: mem merge into test (requires confirmation)"
- After `mem merge into test` completes: print hint "Next step: mem merge into main"
- `mem merge into main` dry-run output shows `--force` flag - this is the ONLY place where `--force` is mentioned (not in previous hints)

### 5. Error handling
On any failure:
- Print clear error message explaining what went wrong
- Print the current state (which branch we're on, what was done)
- Provide specific recommendations for manual resolution
- Try to leave repo in a recoverable state

### 6. Update main.py
Change from registering single command to registering the Typer app:
```python
# Before
app.command(name="merge", help="...")(merge_command)

# After
app.add_typer(merge_app, name="merge", help="Merge operations")
```

## Success Criteria

- `mem merge` continues to work exactly as before (PR merging), with new hint at end
- `mem merge into test` works from `dev` branch
- `mem merge into main` works from `dev` branch, shows dry-run by default, executes with `--force`
- After successful merge, all involved branches are at the same commit (no "behind" warnings)
- Progressive hints guide users through workflow without exposing `--force` prematurely
- Clear error messages guide users to resolution on failure

## Notes

- Use GitPython (`repo.git.*`) for git operations, consistent with rest of codebase
- Use `--ff-only` for back-merges to ensure they're truly fast-forward
- The back-merge should always succeed since we just merged in that direction
- Reuse existing helpers like `check_working_directory_clean()` and `git_fetch_and_pull()`
