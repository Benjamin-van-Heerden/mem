---
created_at: '2026-01-15T09:30:34.161138'
username: benjamin_van_heerden
---
# Work Log - Onboard temp output + spec CLI QoL fixes

## Overarching Goals

Move `mem onboard` temp output into the project directory so project-scoped agents can read it, while keeping the output ephemeral and uncommitted. Also clean up a couple spec command UX issues (remove legacy `activate` hints; support partial slug matching in `spec show`).

## What Was Accomplished

### Onboard temp output location + cleanup
- Changed `mem onboard` to write its generated context markdown into the caller project’s `.mem/tmp/` instead of `/tmp/mem`.
- Kept the existing 1-hour cleanup behavior, but applied it to `.mem/tmp/` (removing `mem_onboard_*.md` older than 1 hour).
- Switched path handling to use `ENV_SETTINGS` so the temp dir is based on the caller project root consistently.

### Enforce gitignore for temp artifacts
- Added logic to ensure `.mem/tmp/` is present in the caller project’s `.gitignore`, adding it if absent (and creating `.gitignore` if needed). This ensures onboard output stays readable to agents but never gets committed.

### Spec command UX fixes
- Removed references to legacy `mem spec activate` from `mem spec list -s` output and from the footer of `mem spec show`.
- Implemented git-style slug-prefix resolution for `mem spec show <slug>`:
  - If the provided slug is a unique prefix, it resolves to the full slug and shows the spec.
  - If ambiguous, it prints the number of matches, lists candidates, and suggests adding more characters.

## Key Files Affected

- `mem/src/commands/onboard.py`
  - Temp output moved to `.mem/tmp/` under the caller project and cleanup updated.
  - Gitignore enforcement helper added and invoked at runtime.
- `mem/src/commands/spec.py`
  - Removed `mem spec activate` hints.
  - Added prefix-resolution flow and improved ambiguity messaging for `spec show`.
- `mem/src/utils/specs.py`
  - Added slug prefix resolution helper and integrated it into spec lookup.
- `mem/.gitignore`
  - Added `.mem/tmp/` (for the mem repo itself).

## What Comes Next

- No follow-up required for this set of changes.
- If desired later: consider reusing a shared “ensure gitignore entry” helper across commands (e.g. init/docs/onboard) to avoid duplication and keep behavior consistent.
