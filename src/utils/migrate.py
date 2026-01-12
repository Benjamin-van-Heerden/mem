"""
Migration utilities for converting agent_rules/ format to .mem/ format.
"""

import re
from datetime import datetime
from pathlib import Path

from src.models import LogFrontmatter, SpecFrontmatter, TaskFrontmatter
from src.utils.ai.log_parser import parse_log
from src.utils.ai.spec_parser import parse_spec
from src.utils.github.api import (
    close_issue_with_comment,
    create_github_issue,
    ensure_label,
)
from src.utils.github.client import get_github_client
from src.utils.github.repo import get_repo_from_git
from src.utils.markdown import read_md_file, slugify, write_md_file


def parse_old_spec_filename(filename: str) -> tuple[str, str, str]:
    """Parse old spec filename pattern: s_YYYYMMDD_username__feature_name.md

    Returns (date_str, username, feature_name).
    """
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
    """Convert an old spec file to mem format using AI agent.

    Returns spec info dict with slug, title, spec_dir for later GitHub operations.
    """
    print(f"   Processing: {spec_file.name}")

    date_str, username, _ = parse_old_spec_filename(spec_file.name)
    content = spec_file.read_text()

    parsed = parse_spec(content)
    if not parsed:
        print("      Failed to parse spec")
        return None

    title = parsed.title
    body = parsed.body
    tasks = parsed.tasks
    slug = slugify(title)

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

    spec_dir = mem_dir / "specs" / "completed" / slug
    spec_dir.mkdir(parents=True, exist_ok=True)

    frontmatter = SpecFrontmatter(
        title=title,
        status="completed",
        assigned_to=username.replace("_", "-") if username else None,
        created_at=created_at,
        updated_at=created_at,
        completed_at=created_at,
    )

    write_md_file(spec_dir / "spec.md", frontmatter.to_dict(), body)
    print(f"      Created spec: {slug}")

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
    """Convert an old work log file to mem format using AI agent.

    Returns True if successful.
    """
    print(f"   Processing: {log_file.name}")

    datetime_str, username = parse_old_log_filename(log_file.name)

    if not datetime_str or not username:
        print("      Skipping: Could not parse filename")
        return False

    try:
        created_dt = datetime.strptime(datetime_str, "%Y%m%d%H%M")
    except ValueError:
        created_dt = datetime.now()

    content = log_file.read_text()

    parsed = parse_log(content)
    if not parsed:
        print("      Failed to parse log")
        return False

    title = parsed.title
    spec_file_path = parsed.spec_file
    body = parsed.body

    spec_slug = None
    if spec_file_path:
        match = re.search(r"s_\d+_[^_]+__(.+)\.md", spec_file_path)
        if match:
            spec_slug = slugify(match.group(1))

    log_filename = f"{username}_{created_dt.strftime('%Y%m%d')}_{created_dt.strftime('%H%M%S')}_session.md"

    if dry_run:
        print(f"      Would create log: {log_filename}")
        print(f"      Title: {title}")
        if spec_slug:
            print(f"      Linked spec: {spec_slug}")
        return True

    logs_dir = mem_dir / "logs"
    logs_dir.mkdir(exist_ok=True)

    frontmatter = LogFrontmatter(
        created_at=created_dt.isoformat(),
        username=username,
        spec_slug=spec_slug,
    )

    full_body = f"# Work Log - {title}\n\n{body}" if not body.startswith("# ") else body

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

        ensure_label(repo, "mem-spec", "0366d6", "Spec managed by mem")
        ensure_label(repo, "mem-status:completed", "0e8a16", "Completed spec")
    except Exception as e:
        print(f"   Could not connect to GitHub: {e}")
        print("   Skipping GitHub issue creation.")
        return

    for spec in specs:
        if spec.get("dry_run"):
            continue

        try:
            title = f"[Spec]: {spec['title']}"
            body = "Migrated from legacy agent_rules system.\n\nThis spec has been completed."
            labels = ["mem-spec", "mem-status:completed"]

            issue = create_github_issue(repo, title=title, body=body, labels=labels)

            close_issue_with_comment(
                repo,
                issue.number,
                "Migrated from legacy agent_rules system - already completed.",
            )

            spec_file = spec.get("spec_dir")
            if spec_file:
                metadata, body_content = read_md_file(spec_file / "spec.md")
                metadata["issue_id"] = issue.number
                metadata["issue_url"] = issue.html_url
                write_md_file(spec_file / "spec.md", metadata, body_content)

            print(f"   Created and closed issue #{issue.number} for: {spec['title']}")
        except Exception as e:
            print(f"   Failed to create issue for {spec['title']}: {e}")


def run_migration(target_dir: Path, dry_run: bool = False) -> None:
    """Run the full migration from agent_rules/ to .mem/ format."""
    agent_rules_dir = target_dir / "agent_rules"
    mem_dir = target_dir / ".mem"

    print("=" * 60)
    print("Agent Rules Migration Tool")
    print("=" * 60)
    print(f"\nTarget: {target_dir}")
    print(f"Source: {agent_rules_dir}")
    print(f"Destination: {mem_dir}")
    if dry_run:
        print("\n*** DRY RUN - No changes will be made ***")

    print("\n1. Discovering files...")
    spec_files, log_files = discover_files(agent_rules_dir)
    print(f"   Found {len(spec_files)} spec file(s)")
    print(f"   Found {len(log_files)} work log file(s)")

    if not spec_files and not log_files:
        print("\nNo files to migrate.")
        return

    migrated_specs = []
    if spec_files:
        print("\n2. Converting specs...")
        for spec_file in spec_files:
            result = convert_spec(spec_file, mem_dir, dry_run)
            if result:
                migrated_specs.append(result)

    if log_files:
        print("\n3. Converting work logs...")
        for log_file in log_files:
            convert_log(log_file, mem_dir, dry_run)

    create_github_issues_for_specs(migrated_specs, target_dir, dry_run)

    print("\n" + "=" * 60)
    if dry_run:
        print("DRY RUN COMPLETE")
        print("Run without --dry-run to apply changes.")
    else:
        print("MIGRATION COMPLETE")
        print(f"Migrated {len(migrated_specs)} spec(s) and {len(log_files)} log(s)")
    print("=" * 60)
