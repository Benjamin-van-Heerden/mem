---
created_at: '2026-01-10T18:23:58.249758'
username: benjamin_van_heerden
spec_slug: migrate_from_agent_rules_to_mem
---
# Work Log - Migration script for agent_rules to mem format

## Overarching Goals

Create a one-off migration script that converts projects using the old `agent_rules/` context system to the `mem` format. The script uses Agno agents with OpenRouter to intelligently parse semi-structured markdown files and convert them to mem's standardized format.

## What Was Accomplished

### Created migration script `scripts/migrate_agent_rules.py`

- CLI with `target_dir` argument and `--dry-run` flag
- Uses Gemini 3 Flash via OpenRouter for parsing
- Pydantic models for structured output: `ParsedSpec`, `ParsedTask`, `ParsedLog`
- Fallback JSON parsing when Agno's structured output fails (known issue with OpenRouter)

### Spec conversion
- Parses old spec files (pattern: `s_YYYYMMDD_username__feature_name.md`)
- Extracts title, body (reformatted to Overview/Goals/Technical Approach), and tasks
- Creates spec in `.mem/specs/completed/<slug>/spec.md` with proper frontmatter
- Creates task files in `.mem/specs/completed/<slug>/tasks/`

### Work log conversion
- Parses old work log files (pattern: `w_YYYYMMDDHHmm_username.md`)
- Extracts title, spec reference, and cleaned body
- Creates log in `.mem/logs/<username>_<date>_<time>_session.md`

### GitHub issue creation
- Creates issue with `mem-spec` and `mem-status:completed` labels
- Immediately closes with migration comment
- Updates spec file with `issue_id` and `issue_url`

### Fixed filename regex
- Username can contain underscores, so changed `([^_]+)` to `(.+?)` with double-underscore delimiter

### Push feature branch on spec assign
- Added `git push --set-upstream origin branch_name` after worktree creation
- Ensures remote branch exists and is tracked from the start

## Key Files Affected

- `scripts/migrate_agent_rules.py` - New migration script (complete implementation)
- `src/commands/spec.py` - Added push to origin after worktree creation in assign command
- `pyproject.toml` - Added `openai` dependency (required by agno for OpenRouter)

## What Comes Next

The spec is complete. To use the migration script on another project:

```bash
uv run python scripts/migrate_agent_rules.py /path/to/project --dry-run  # preview
uv run python scripts/migrate_agent_rules.py /path/to/project            # migrate
```
