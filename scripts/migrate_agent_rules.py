"""
Migration script: Convert agent_rules/ format to .mem/ format.

This is a one-off migration script for converting projects using the old
agent_rules/ context system to the mem format.

Usage:
    uv run python scripts/migrate_agent_rules.py <target_project_dir> [--dry-run]
"""

import argparse
import re
import sys
from datetime import datetime
from pathlib import Path

from agno.agent import Agent
from agno.models.openrouter import OpenRouter
from pydantic import BaseModel, Field

sys.path.insert(0, str(Path(__file__).parent.parent.resolve()))

from src.models import (
    LogFrontmatter,
    SpecFrontmatter,
    TaskFrontmatter,
)
from src.utils.github.api import (
    close_issue_with_comment,
    create_github_issue,
    ensure_label,
)
from src.utils.github.client import get_github_client
from src.utils.github.repo import get_repo_from_git
from src.utils.markdown import slugify, write_md_file


# Pydantic models for structured output
class ParsedTask(BaseModel):
    title: str = Field(
        ..., description="The task title (e.g. 'Create Database Migration')"
    )
    description: str = Field(
        "",
        description="The full task description including checkbox items and implementation details",
    )


class ParsedSpec(BaseModel):
    title: str = Field(..., description="The spec title from the first heading")
    body: str = Field(
        ...,
        description="Cleaned markdown body with ## Overview, ## Goals, ## Technical Approach sections",
    )
    tasks: list[ParsedTask] = Field(
        default_factory=list, description="List of tasks extracted from the spec"
    )


class ParsedLog(BaseModel):
    title: str = Field(..., description="Short descriptive title for this work session")
    spec_file: str | None = Field(None, description="The spec file path if mentioned")
    body: str = Field(..., description="Cleaned work log content")


# Agno agent setup
model = OpenRouter("google/gemini-3-flash-preview", max_tokens=8192)

spec_parser_agent = Agent(
    model=model,
    name="Spec Parser Agent",
    instructions=[
        """You are an expert at parsing semi-structured markdown spec files and converting them to a structured format.

Given an old spec file, extract:
1. title: The spec title (from the first heading)
2. body: A cleaned markdown body with ## Overview, ## Goals, ## Technical Approach sections
3. tasks: A list of tasks with their titles, and descriptions

When converting the body:
- Use the Description section as the Overview
- Extract goals from the content if present
- Keep technical details in Technical Approach
- Remove any status markers like "%% Status: ... %%"
- Remove task sections from the body (they go in the tasks array)

When extracting tasks:
- Look for "### Task" or "### Task N:" sections (e.g. "### Task 1: Create Database Migration")
- The task title is the text after "Task:" or "Task N:" (e.g. "Create Database Migration")
- Include everything under the task heading in the description (checkbox items, implementation details, etc.)
"""
    ],
    output_schema=ParsedSpec,
)

log_parser_agent = Agent(
    model=model,
    name="Log Parser Agent",
    instructions=[
        """You are an expert at parsing work log files and cleaning them up.

Given an old work log file, extract and clean the content.

When cleaning the body:
- Keep the main sections: Overarching Goals, What Was Accomplished, Key Files Affected, What Comes Next
- Remove the "## Spec File:" section (we extract it separately)
- Clean up any formatting issues
- Keep the content concise and well-structured
"""
    ],
    output_schema=ParsedLog,
)


def parse_old_spec_filename(filename: str) -> tuple[str, str, str]:
    """Parse old spec filename pattern: s_YYYYMMDD_username__feature_name.md

    Returns (date_str, username, feature_name).
    """
    # Username can contain underscores, so we split on double underscore
    match = re.match(r"s_(\d{8})_(.+?)__(.+)\.md", filename)
    if match:
        return match.group(1), match.group(2), match.group(3)
    return "", "", ""


def parse_old_log_filename(filename: str) -> tuple[str, str]:
    """Parse old work log filename pattern: w_YYYYMMDDHHmm_username.md

    Returns (datetime_str, username).
    """
    match = re.match(r"w_(\d{12})_(.+)\.md", filename)
    if match:
        return match.group(1), match.group(2)
    return "", ""


def discover_files(agent_rules_dir: Path) -> tuple[list[Path], list[Path]]:
    """Discover spec and work log files in agent_rules directory.

    Returns (spec_files, log_files).
    """
    spec_dir = agent_rules_dir / "spec"
    log_dir = agent_rules_dir / "work_log"

    spec_files = sorted(spec_dir.glob("s_*_*.md")) if spec_dir.exists() else []
    log_files = sorted(log_dir.glob("w_*_*.md")) if log_dir.exists() else []

    return spec_files, log_files


def convert_spec(spec_file: Path, mem_dir: Path, dry_run: bool = False) -> dict | None:
    """Convert an old spec file to mem format using Agno agent.

    Returns spec info dict with slug, title, issue_id for later GitHub operations.
    """
    print(f"   Processing: {spec_file.name}")

    # Parse filename for metadata
    date_str, username, _ = parse_old_spec_filename(spec_file.name)

    # Read the old spec content
    content = spec_file.read_text()

    # Use Agno agent to parse the spec (returns structured ParsedSpec)
    response = spec_parser_agent.run(f"Parse this spec file:\n\n{content}")

    if isinstance(response.content, ParsedSpec):
        parsed = response.content
    elif isinstance(response.content, str):
        # Fallback: try to parse the JSON string manually
        print("      [fallback] Parsing JSON string manually")
        import json

        try:
            data = json.loads(response.content)
            parsed = ParsedSpec(**data)
        except (json.JSONDecodeError, ValueError) as e:
            print(f"      Failed to parse spec: {e}")
            return None
    else:
        print(
            f"      Failed to parse spec: unexpected type {type(response.content).__name__}"
        )
        return None

    title = parsed.title
    body = parsed.body
    tasks = parsed.tasks

    slug = slugify(title)

    # Create timestamps from filename date
    if date_str:
        try:
            created_dt = datetime.strptime(date_str, "%Y%m%d")
        except ValueError:
            created_dt = datetime.now()
    else:
        created_dt = datetime.now()

    created_at = created_dt.isoformat()

    if dry_run:
        print(f"      Would create spec: {slug}")
        print(f"      Title: {title}")
        print(f"      Tasks: {len(tasks)}")
        return {"slug": slug, "title": title, "dry_run": True}

    # Create spec directory
    spec_dir = mem_dir / "specs" / "completed" / slug
    spec_dir.mkdir(parents=True, exist_ok=True)

    # Create spec frontmatter
    frontmatter = SpecFrontmatter(
        title=title,
        status="completed",
        assigned_to=username.replace("_", "-") if username else None,
        created_at=created_at,
        updated_at=created_at,
        completed_at=created_at,
    )

    # Write spec.md
    write_md_file(spec_dir / "spec.md", frontmatter.to_dict(), body)
    print(f"      Created spec: {slug}")

    # Create tasks directory and task files
    if tasks:
        tasks_dir = spec_dir / "tasks"
        tasks_dir.mkdir(exist_ok=True)

        for i, task in enumerate(tasks, 1):
            task_title = task.title
            task_desc = task.description
            task_slug = slugify(task_title)

            task_frontmatter = TaskFrontmatter(
                title=task_title,
                status="completed",
                created_at=created_at,
                updated_at=created_at,
                completed_at=created_at,
            )

            task_filename = f"{i:02d}_{task_slug}.md"
            write_md_file(
                tasks_dir / task_filename, task_frontmatter.to_dict(), task_desc
            )

        print(f"      Created {len(tasks)} task(s)")

    return {"slug": slug, "title": title, "spec_dir": spec_dir}


def convert_log(log_file: Path, mem_dir: Path, dry_run: bool = False) -> bool:
    """Convert an old work log file to mem format using Agno agent.

    Returns True if successful.
    """
    print(f"   Processing: {log_file.name}")

    # Parse filename for metadata
    datetime_str, username = parse_old_log_filename(log_file.name)

    if not datetime_str or not username:
        print("      Skipping: Could not parse filename")
        return False

    # Parse datetime from filename (YYYYMMDDHHmm)
    try:
        created_dt = datetime.strptime(datetime_str, "%Y%m%d%H%M")
    except ValueError:
        created_dt = datetime.now()

    # Read the old log content
    content = log_file.read_text()

    # Use Agno agent to parse the log (returns structured ParsedLog)
    response = log_parser_agent.run(f"Parse this work log file:\n\n{content}")

    if isinstance(response.content, ParsedLog):
        parsed = response.content
    elif isinstance(response.content, str):
        # Fallback: try to parse the JSON string manually
        print("      [fallback] Parsing JSON string manually")
        import json

        try:
            data = json.loads(response.content)
            parsed = ParsedLog(**data)
        except (json.JSONDecodeError, ValueError) as e:
            print(f"      Failed to parse log: {e}")
            return False
    else:
        print(
            f"      Failed to parse log: unexpected type {type(response.content).__name__}"
        )
        return False

    title = parsed.title
    spec_file_path = parsed.spec_file
    body = parsed.body

    # Try to extract spec_slug from spec_file path if present
    spec_slug = None
    if spec_file_path:
        # Try to extract slug from path like "agent_rules/spec/s_...__feature_name.md"
        match = re.search(r"s_\d+_[^_]+__(.+)\.md", spec_file_path)
        if match:
            spec_slug = slugify(match.group(1))

    # Format log filename
    log_filename = f"{username}_{created_dt.strftime('%Y%m%d')}_{created_dt.strftime('%H%M%S')}_session.md"

    if dry_run:
        print(f"      Would create log: {log_filename}")
        print(f"      Title: {title}")
        if spec_slug:
            print(f"      Linked spec: {spec_slug}")
        return True

    # Create logs directory
    logs_dir = mem_dir / "logs"
    logs_dir.mkdir(exist_ok=True)

    # Create log frontmatter
    frontmatter = LogFrontmatter(
        created_at=created_dt.isoformat(),
        username=username,
        spec_slug=spec_slug,
    )

    # Prepend title to body
    full_body = f"# Work Log - {title}\n\n{body}" if not body.startswith("# ") else body

    # Write log file
    write_md_file(logs_dir / log_filename, frontmatter.to_dict(), full_body)
    print(f"      Created log: {log_filename}")

    return True


def create_github_issues_for_specs(
    specs: list[dict], target_dir: Path, dry_run: bool = False
) -> None:
    """Create closed GitHub issues for migrated specs."""
    if not specs:
        return

    print("\n4. Creating GitHub issues for migrated specs...")

    if dry_run:
        for spec in specs:
            print(f"   Would create and close issue for: {spec['title']}")
        return

    try:
        client = get_github_client()
        repo_owner, repo_name = get_repo_from_git(target_dir)
        repo = client.get_repo(f"{repo_owner}/{repo_name}")

        # Ensure labels exist
        ensure_label(repo, "mem-spec", "0366d6", "Spec managed by mem")
        ensure_label(repo, "mem-status:completed", "0e8a16", "Completed spec")
    except Exception as e:
        print(f"   Could not connect to GitHub: {e}")
        return

    for spec in specs:
        if spec.get("dry_run"):
            continue

        try:
            # Create issue
            title = f"[Spec]: {spec['title']}"
            body = "Migrated from legacy agent_rules system.\n\nThis spec has been completed."
            labels = ["mem-spec", "mem-status:completed"]

            issue = create_github_issue(repo, title=title, body=body, labels=labels)

            # Close immediately
            close_issue_with_comment(
                repo,
                issue.number,
                "Migrated from legacy agent_rules system - already completed.",
            )

            # Update spec file with issue info
            spec_file = spec.get("spec_dir")
            if spec_file:
                from src.utils.markdown import read_md_file

                metadata, body_content = read_md_file(spec_file / "spec.md")
                metadata["issue_id"] = issue.number
                metadata["issue_url"] = issue.html_url
                write_md_file(spec_file / "spec.md", metadata, body_content)

            print(f"   Created and closed issue #{issue.number} for: {spec['title']}")
        except Exception as e:
            print(f"   Failed to create issue for {spec['title']}: {e}")


def main():
    parser = argparse.ArgumentParser(
        description="Migrate agent_rules/ format to .mem/ format"
    )
    parser.add_argument(
        "target_dir",
        type=Path,
        help="Path to project directory containing agent_rules/ folder",
    )
    parser.add_argument(
        "--dry-run",
        action="store_true",
        help="Preview migration without making changes",
    )

    args = parser.parse_args()
    target_dir = args.target_dir.resolve()
    dry_run = args.dry_run

    # Validate target directory
    if not target_dir.exists():
        print(f"Error: Directory does not exist: {target_dir}")
        sys.exit(1)

    agent_rules_dir = target_dir / "agent_rules"
    if not agent_rules_dir.exists():
        print(f"Error: No agent_rules/ directory found in: {target_dir}")
        sys.exit(1)

    mem_dir = target_dir / ".mem"

    print("=" * 60)
    print("Agent Rules Migration Tool")
    print("=" * 60)
    print(f"\nTarget: {target_dir}")
    print(f"Source: {agent_rules_dir}")
    print(f"Destination: {mem_dir}")
    if dry_run:
        print("\n*** DRY RUN - No changes will be made ***")

    # Discover files
    print("\n1. Discovering files...")
    spec_files, log_files = discover_files(agent_rules_dir)
    print(f"   Found {len(spec_files)} spec file(s)")
    print(f"   Found {len(log_files)} work log file(s)")

    if not spec_files and not log_files:
        print("\nNo files to migrate.")
        sys.exit(0)

    # Convert specs
    migrated_specs = []
    if spec_files:
        print("\n2. Converting specs...")
        for spec_file in spec_files:
            result = convert_spec(spec_file, mem_dir, dry_run)
            if result:
                migrated_specs.append(result)

    # Convert logs
    if log_files:
        print("\n3. Converting work logs...")
        for log_file in log_files:
            convert_log(log_file, mem_dir, dry_run)

    # Create GitHub issues
    create_github_issues_for_specs(migrated_specs, target_dir, dry_run)

    # Summary
    print("\n" + "=" * 60)
    if dry_run:
        print("DRY RUN COMPLETE")
        print("Run without --dry-run to apply changes.")
    else:
        print("MIGRATION COMPLETE")
        print(f"Migrated {len(migrated_specs)} spec(s) and {len(log_files)} log(s)")
    print("=" * 60)


if __name__ == "__main__":
    main()
