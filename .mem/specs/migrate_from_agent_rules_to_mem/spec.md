---
title: Migrate from agent_rules to mem
status: merge_ready
assigned_to: Benjamin-van-Heerden
issue_id: 33
issue_url: https://github.com/Benjamin-van-Heerden/mem/issues/33
branch: dev-benjamin_van_heerden-migrate_from_agent_rules_to_mem
pr_url: null
created_at: '2026-01-10T15:00:16.744004'
updated_at: '2026-01-10T18:24:50.419153'
completed_at: null
last_synced_at: '2026-01-10T15:13:30.169960'
local_content_hash: d392e8b93147a8f54ca6019660204da9eb8e686c11b471b29c66cd161fb92f9f
remote_content_hash: d392e8b93147a8f54ca6019660204da9eb8e686c11b471b29c66cd161fb92f9f
---
## Overview

Create a one-off migration script `scripts/migrate_agent_rules.py` that converts projects using the old `agent_rules/` context system to the `mem` format. The old system stored specs and work logs as markdown files in `agent_rules/spec/` and `agent_rules/work_log/` directories with a semi-structured format.

Since the old format is not rigidly structured, the migration uses Agno agents with OpenRouter to parse and interpret each file, extracting the relevant information and converting it to mem's standardized format.

## Goals

- Create `scripts/migrate_agent_rules.py` that converts `agent_rules/` projects to `.mem/` format
- Use Agno agents to intelligently parse old spec and work log files
- Convert old specs to mem spec format (with tasks extracted from spec body)
- Convert old work logs to mem log format
- Create GitHub issues for migrated specs (as closed issues for historical record)
- Mark all migrated specs as `completed` status

## Technical Approach

### Command Line Interface

```bash
uv run python scripts/migrate_agent_rules.py <target_project_dir> [--dry-run]
```

- `target_project_dir`: Path to project directory containing `agent_rules/` folder
- `--dry-run`: Show what would be migrated without making changes
- Script fails if `agent_rules/` doesn't exist in the target directory

### Script Structure

```python
import sys
from pathlib import Path
from argparse import ArgumentParser

from agno.agent import Agent
from agno.models.openrouter import OpenRouter

# Import mem utilities directly
sys.path.insert(0, str(Path(__file__).parent.parent.resolve()))
from src.utils.specs import ...
from src.utils.tasks import ...
from src.utils.logs import ...
from src.utils.markdown import slugify, write_md_file
from src.utils.github.api import create_github_issue, close_issue_with_comment
from src.utils.github.client import get_github_client
from src.utils.github.repo import get_repo_from_git
from src.models import create_spec_frontmatter, create_task_frontmatter, create_log_frontmatter
```

### Agno Agent Setup

Use OpenRouter with a capable model. No API key needs to be passed - it uses the environment variable automatically:

```python
model = OpenRouter("anthropic/claude-sonnet-4")

spec_parser_agent = Agent(
    model=model,
    name="Spec Parser Agent",
    instructions=[
        """You are an expert at parsing semi-structured markdown spec files and converting them to a structured format.
        
        Given an old spec file, you will extract:
        1. title: The spec title (from the heading or inferred from content)
        2. description: A clean overview/description section
        3. tasks: A list of tasks with their titles and descriptions
        
        Output your response as valid JSON with this structure:
        {
            "title": "string",
            "body": "markdown string with Overview, Goals, Technical Approach sections",
            "tasks": [
                {"title": "string", "description": "string"},
                ...
            ]
        }
        
        When converting the body:
        - Reformat to use ## Overview, ## Goals, ## Technical Approach, ## Success Criteria sections
        - Keep the content concise and actionable
        - Remove any status markers like "%% Status: ... %%"
        
        When extracting tasks:
        - Look for "### Task:" sections
        - Extract the task title and combine the description/details into a single description
        - Convert checkbox items into the description
        """
    ],
    markdown=False,  # We want JSON output
)
```

### Migration Process

1. **Discovery Phase**
   - Scan `agent_rules/spec/` for spec files (pattern: `s_*_*.md`)
   - Scan `agent_rules/work_log/` for work log files (pattern: `w_*_*.md`)
   - Report what will be migrated

2. **Spec Conversion** (using Agno agent)
   - For each old spec file:
     - Read the file content
     - Run through `spec_parser_agent` to get structured JSON
     - Parse the JSON response
     - Create spec directory: `.mem/specs/completed/<slug>/`
     - Write `spec.md` with proper `SpecFrontmatter`:
       - `status: "completed"`
       - `completed_at`: derived from filename date or now
       - `created_at`: derived from filename date (format: `s_YYYYMMDD_...`)
     - Write each task to `.mem/specs/completed/<slug>/tasks/`
     - Create GitHub issue via API with labels `["mem-spec", "mem-status:completed"]`
     - Immediately close the issue with comment "Migrated from legacy agent_rules system"
     - Update spec frontmatter with `issue_id` and `issue_url`

3. **Work Log Conversion** (using Agno agent)
   - For each old work log file:
     - Read the file content
     - Use agent to extract/clean the content and identify associated spec if any
     - Parse username from filename (format: `w_YYYYMMDDHHmm_{username}.md`)
     - Parse timestamp from filename
     - Write to `.mem/logs/{username}_{YYYYMMDD}_{HHMMSS}_session.md` with proper `LogFrontmatter`

### Using Existing Utilities

The script should use existing `src/utils/` functions directly rather than reimplementing:

**From `src/utils/markdown.py`:**
- `slugify(text)` - Convert title to filesystem-safe slug
- `write_md_file(path, metadata, body)` - Write markdown with YAML frontmatter

**From `src/utils/github/api.py`:**
- `create_github_issue(repo, title, body, labels, assignees)` - Create issue
- `close_issue_with_comment(repo, issue_number, comment)` - Close with comment
- `ensure_label(repo, name, color, description)` - Ensure labels exist

**From `src/utils/github/client.py`:**
- `get_github_client()` - Get authenticated PyGithub client

**From `src/utils/github/repo.py`:**
- `get_repo_from_git(path)` - Get (owner, repo_name) from git remote

**From `src/models.py`:**
- `create_spec_frontmatter(title, status, ...)` - Create SpecFrontmatter
- `create_task_frontmatter(title, status)` - Create TaskFrontmatter
- `create_log_frontmatter(created_at, username, spec_slug)` - Create LogFrontmatter

### File Format Reference

**Old Spec Format (agent_rules/spec/):**
```markdown
# {Title}

%% Status: {Draft | In Progress | Completed | Archived | Abandoned} %%

## Description
{description}

## Tasks

### Task: {task title}
- [ ] goal 1
- [x] goal 2

#### Implementation Details
{details}
```

**New Spec Format (.mem/specs/completed/):**
```yaml
---
title: Example spec
status: completed
assigned_to: null
issue_id: 123
issue_url: https://github.com/...
created_at: '2026-01-10T10:00:00'
updated_at: '2026-01-10T10:00:00'
completed_at: '2026-01-10T10:00:00'
---
## Overview
{description}

## Goals
- Goal 1

## Technical Approach
{approach}
```

**Old Work Log Format (agent_rules/work_log/):**
```markdown
# Work Log - {short title}

## Spec File: `{path}`

## Overarching Goals
{goals}

## What Was Accomplished
{accomplishments}

## Key Files Affected
{files}

## What Comes Next
{next steps}
```

**New Log Format (.mem/logs/):**
```yaml
---
created_at: '2026-01-10T10:00:00'
username: benjamin_van_heerden
spec_slug: example_spec
---
# Work Log - {title}

## Overarching Goals
{goals}

## What Was Accomplished
{accomplishments}

## Key Files Affected
{files}

## What Comes Next
{next steps}
```

## Example Files for Reference

**Completed spec example:** `.mem/specs/completed/emojify_cli_output/spec.md`
**Completed task example:** `.mem/specs/completed/emojify_cli_output/tasks/01_emojify_command_files.md`
**Work log example:** `.mem/logs/benjamin_van_heerden_20260106_session.md`

**Old spec examples in:** `agent_rules/spec/`
**Old work log examples in:** `agent_rules/work_log/`

## Success Criteria

- Running `uv run python scripts/migrate_agent_rules.py <path>` converts all specs and logs
- All migrated specs have corresponding closed GitHub issues
- All migrated files have valid frontmatter matching Pydantic models
- `mem onboard` shows migrated specs in "Recently completed specs" section
- Work logs are correctly associated with their specs via `spec_slug`
- Dry run mode shows accurate preview without making changes
- Script fails gracefully if `agent_rules/` directory doesn't exist

## Notes

- Old spec filename pattern: `s_YYYYMMDD_username__feature_name.md`
- Old work log filename pattern: `w_YYYYMMDDHHmm_username.md`
- Agno is already installed in the project via `uv`
- OpenRouter API key is available via environment variable (no need to pass explicitly)
- This is a one-off migration script, not a mem command - it goes in `scripts/` not `src/commands/`
