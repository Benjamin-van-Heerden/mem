---
created_at: '2026-01-14T15:11:28.495664'
username: benjamin_van_heerden
spec_slug: handle_oversized_onboard_output_via_temp_file
---
# Work Log - Onboard temp-file output + --stdout + agent wording cleanup

## Overarching Goals

Reduce `mem onboard` output overflow in environments that truncate stdout by writing large onboard context to a temp file with a clear directive to read it in full, while also providing an explicit `--stdout` escape hatch and removing user-facing references to a specific agent.

## What Was Accomplished

### Temp-file fallback for oversized onboard output
- Refactored `mem onboard` output assembly so that:
  - A small “About mem” + “Project info” summary remains on stdout.
  - Larger, detailed sections (guidelines/templates, important files, docs, spec details, work logs, todos, workflow hints, next steps, etc.) are assembled into a single onboard context payload.
- When the payload is large, it is written to:
  - `/tmp/mem/mem_onboard_{YYYYMMDD_HHMMSS}.md`
- Added startup cleanup to keep `/tmp/mem` tidy:
  - On each `mem onboard` run, delete `/tmp/mem/mem_onboard_*.md` files older than 1 hour.

### --stdout support
- Added `mem onboard --stdout` to force printing the full onboard context to stdout regardless of size.
- Default behavior remains unchanged (i.e., the temp-file fallback is still used when output is large and `--stdout` is not provided).

### Remove “claude” references in user-facing CLI
- Updated `mem init` behavior so it no longer creates `CLAUDE.md` (or a symlink); it only creates `AGENTS.md`.
- Updated spec assignment UX messaging to be agent-agnostic (no longer telling the user to run a specific agent command).

## Key Files Affected

- `src/commands/onboard.py`
  - Temp-file output path under `/tmp/mem/`
  - On-run cleanup of old `mem_onboard_*.md` files (> 1 hour)
  - Added `--stdout` option to force printing full context to stdout
- `src/commands/init.py`
  - Removed creation of `CLAUDE.md`; only create `AGENTS.md`
- `src/commands/spec.py`
  - Replaced `claude` instruction with agent-agnostic wording

## Errors and Barriers

- During manual runs, `mem onboard` sync step can report “Cannot rebase with uncommitted changes” when the working tree is dirty. This is expected given sync behavior; it did not block implementing the onboard output changes.

## What Comes Next

- Spec tasks are complete.
- Next steps:
  1. Run `mem spec complete handle_oversized_onboard_output_via_temp_file "..."` to open the PR and mark the spec merge-ready.
