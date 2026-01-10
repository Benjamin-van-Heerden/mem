"""
Markdown-based task operations.

Tasks are stored as:
  .mem/specs/{spec_slug}/tasks/01_{task_slug}.md
  .mem/specs/{spec_slug}/tasks/02_{task_slug}.md

Each task has YAML frontmatter with metadata and markdown body.
"""

import re
from datetime import datetime
from pathlib import Path
from typing import Any

from src.models import create_task_frontmatter
from src.utils.markdown import read_md_file, slugify, write_md_file
from src.utils.specs import get_spec_path


def _get_tasks_dir(spec_slug: str) -> Path:
    """Get the tasks directory for a spec.

    Uses get_spec_path to find the spec in root, completed, or abandoned.
    """
    return get_spec_path(spec_slug) / "tasks"


def _now_iso() -> str:
    """Return current timestamp in ISO format."""
    return datetime.now().isoformat()


def _parse_task_filename(filename: str) -> tuple[int, str] | None:
    """Parse a task filename like '01_my_task.md' into (order, slug).

    Returns None if filename doesn't match expected pattern.
    """
    match = re.match(r"^(\d+)_(.+)\.md$", filename)
    if not match:
        return None
    return int(match.group(1)), match.group(2)


def _make_task_filename(order: int, slug: str) -> str:
    """Create a task filename like '01_my_task.md'."""
    return f"{order:02d}_{slug}.md"


def _task_to_dict(
    spec_slug: str, filename: str, metadata: dict, body: str
) -> dict[str, Any]:
    """Convert parsed task file to a dict."""
    parsed = _parse_task_filename(filename)
    if not parsed:
        return {}

    order, task_slug = parsed
    return {
        "spec_slug": spec_slug,
        "slug": task_slug,
        "filename": filename,
        "order": order,
        "body": body,
        **metadata,
    }


def get_next_task_number(spec_slug: str) -> int:
    """Get next available task number prefix."""
    tasks_dir = _get_tasks_dir(spec_slug)

    if not tasks_dir.exists():
        return 1

    max_num = 0
    for item in tasks_dir.iterdir():
        if item.is_file() and item.suffix == ".md":
            parsed = _parse_task_filename(item.name)
            if parsed:
                max_num = max(max_num, parsed[0])

    return max_num + 1


def create_task(
    spec_slug: str, title: str, description: str, order: int | None = None
) -> Path:
    """Create task file with description as body.

    Returns path to the created task file.
    """
    tasks_dir = _get_tasks_dir(spec_slug)
    tasks_dir.mkdir(parents=True, exist_ok=True)

    if order is None:
        order = get_next_task_number(spec_slug)

    task_slug = slugify(title)
    filename = _make_task_filename(order, task_slug)
    task_file = tasks_dir / filename

    if task_file.exists():
        raise ValueError(f"Task '{filename}' already exists")

    frontmatter = create_task_frontmatter(title)
    body = description

    write_md_file(task_file, frontmatter.to_dict(), body)
    return task_file


def get_task(spec_slug: str, task_filename: str) -> dict[str, Any] | None:
    """Read task metadata + body.

    task_filename should be like '01_my_task.md' or just '01_my_task'.
    """
    if not task_filename.endswith(".md"):
        task_filename = task_filename + ".md"

    task_file = _get_tasks_dir(spec_slug) / task_filename

    if not task_file.exists():
        return None

    metadata, body = read_md_file(task_file)
    return _task_to_dict(spec_slug, task_filename, metadata, body)


def list_tasks(spec_slug: str) -> list[dict[str, Any]]:
    """List all tasks for a spec (sorted by order number)."""
    tasks_dir = _get_tasks_dir(spec_slug)

    if not tasks_dir.exists():
        return []

    tasks = []
    for item in tasks_dir.iterdir():
        if item.is_file() and item.suffix == ".md":
            parsed = _parse_task_filename(item.name)
            if parsed:
                metadata, body = read_md_file(item)
                tasks.append(_task_to_dict(spec_slug, item.name, metadata, body))

    # Sort by order number
    tasks.sort(key=lambda t: t.get("order", 0))
    return tasks


def update_task(spec_slug: str, task_filename: str, **updates) -> None:
    """Update task frontmatter fields."""
    if not task_filename.endswith(".md"):
        task_filename = task_filename + ".md"

    task_file = _get_tasks_dir(spec_slug) / task_filename

    if not task_file.exists():
        raise ValueError(f"Task '{task_filename}' not found")

    metadata, body = read_md_file(task_file)

    for key, value in updates.items():
        metadata[key] = value

    metadata["updated_at"] = _now_iso()

    write_md_file(task_file, metadata, body)


def update_task_body(spec_slug: str, task_filename: str, body: str) -> None:
    """Update task body content."""
    if not task_filename.endswith(".md"):
        task_filename = task_filename + ".md"

    task_file = _get_tasks_dir(spec_slug) / task_filename

    if not task_file.exists():
        raise ValueError(f"Task '{task_filename}' not found")

    metadata, _ = read_md_file(task_file)
    metadata["updated_at"] = _now_iso()

    write_md_file(task_file, metadata, body)


def find_task_by_title(spec_slug: str, title: str) -> str | None:
    """Find task filename by title (case-insensitive partial match).

    Returns the task filename or None if not found.
    """
    task_list = list_tasks(spec_slug)
    for task in task_list:
        if title.lower() in task["title"].lower():
            return task["filename"]
    return None


# --- Task completion ---


def complete_task(spec_slug: str, task_filename: str) -> None:
    """Mark task completed."""
    update_task(
        spec_slug,
        task_filename,
        status="completed",
        completed_at=_now_iso(),
    )


def complete_task_with_notes(spec_slug: str, task_filename: str, notes: str) -> None:
    """Complete a task and append completion notes to body."""
    if not task_filename.endswith(".md"):
        task_filename = task_filename + ".md"

    task_file = _get_tasks_dir(spec_slug) / task_filename

    if not task_file.exists():
        raise ValueError(f"Task '{task_filename}' not found")

    metadata, body = read_md_file(task_file)

    # Append completion notes
    completion_section = f"\n\n## Completion Notes\n\n{notes}"
    body = body.rstrip() + completion_section

    # Update metadata
    metadata["status"] = "completed"
    metadata["completed_at"] = _now_iso()
    metadata["updated_at"] = _now_iso()

    write_md_file(task_file, metadata, body)


def delete_task(spec_slug: str, task_filename: str) -> None:
    """Delete task file."""
    if not task_filename.endswith(".md"):
        task_filename = task_filename + ".md"

    task_file = _get_tasks_dir(spec_slug) / task_filename

    if not task_file.exists():
        raise ValueError(f"Task '{task_filename}' not found")

    task_file.unlink()


def amend_task(spec_slug: str, task_filename: str, notes: str) -> None:
    """Amend a task by appending notes and resetting status to todo.

    This enables iterative refinement: Amendments -> Completion -> Amendments -> ...
    """
    if not task_filename.endswith(".md"):
        task_filename = task_filename + ".md"

    task_file = _get_tasks_dir(spec_slug) / task_filename

    if not task_file.exists():
        raise ValueError(f"Task '{task_filename}' not found")

    metadata, body = read_md_file(task_file)

    amendment_section = f"\n\n## Amendments\n\n{notes}"
    body = body.rstrip() + amendment_section

    metadata["status"] = "todo"
    metadata["updated_at"] = _now_iso()
    if "completed_at" in metadata:
        del metadata["completed_at"]

    write_md_file(task_file, metadata, body)


def rename_task(spec_slug: str, task_filename: str, new_title: str) -> None:
    """Rename a task by updating its title in frontmatter.

    The filename/slug remains unchanged for stability.
    """
    if not task_filename.endswith(".md"):
        task_filename = task_filename + ".md"

    task_file = _get_tasks_dir(spec_slug) / task_filename

    if not task_file.exists():
        raise ValueError(f"Task '{task_filename}' not found")

    update_task(spec_slug, task_filename, title=new_title)
