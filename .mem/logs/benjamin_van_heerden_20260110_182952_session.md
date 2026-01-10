---
created_at: '2026-01-10T18:29:52.354192'
username: benjamin_van_heerden
spec_slug: migrate_from_agent_rules_to_mem
---
# Work Log - Fix spec complete workflow and add migration script

## Overarching Goals

1. Create a migration script to convert old `agent_rules/` format to `mem` format
2. Fix the `spec complete` workflow to be safer - commit/push before rebase so work is never lost

## What Was Accomplished

### Migration script `scripts/migrate_agent_rules.py`

- CLI with `target_dir` argument and `--dry-run` flag
- Uses Gemini 3 Flash via OpenRouter for parsing with Pydantic structured output
- Fallback JSON parsing when Agno's structured output parsing fails
- Converts specs to `.mem/specs/completed/` with proper frontmatter and tasks
- Converts work logs to `.mem/logs/` with proper frontmatter
- Creates and closes GitHub issues for migrated specs
- Fixed filename regex to handle usernames with underscores

### Push feature branch on spec assign

Added `git push --set-upstream origin branch_name` after worktree creation so remote branch exists from the start.

### Fixed spec complete workflow

Changed the order of operations in `spec complete`:

**Before:** Try to rebase first (fails with uncommitted changes) → commit → push
**After:** Commit → push (work is safe) → fetch → rebase → force-push-with-lease → create PR

If rebase fails, the user's work is already safely on the remote branch. They can resolve conflicts manually and run `spec complete` again.

## Key Files Affected

- `scripts/migrate_agent_rules.py` - New migration script
- `src/commands/spec.py` - Fixed complete workflow, added push on assign
- `pyproject.toml` - Added `openai` dependency for agno

## What Comes Next

Spec is complete and ready for PR.
