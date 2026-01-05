"""
Markdown-based todo operations.

Todos are stored as:
  .mem/todos/{slug}.md

Each todo has YAML frontmatter with metadata and optional markdown body.
"""

from datetime import datetime
from pathlib import Path
from typing import Any

from env_settings import ENV_SETTINGS
from src.utils.markdown import read_md_file, slugify, write_md_file


def _get_todos_dir() -> Path:
    """Get the todos directory path."""
    return ENV_SETTINGS.todos_dir


def _get_todo_file(slug: str) -> Path:
    """Get path to a todo file."""
    return _get_todos_dir() / f"{slug}.md"


def _now_iso() -> str:
    """Return current timestamp in ISO format."""
    return datetime.now().isoformat()


def _todo_to_dict(slug: str, metadata: dict, body: str) -> dict[str, Any]:
    """Convert parsed todo file to a dict."""
    return {
        "slug": slug,
        "body": body,
        **metadata,
    }


def create_todo(title: str, description: str = "") -> Path:
    """Create todo file.

    Returns path to the created todo file.
    """
    todos_dir = _get_todos_dir()
    todos_dir.mkdir(parents=True, exist_ok=True)

    slug = slugify(title)
    todo_file = _get_todo_file(slug)

    if todo_file.exists():
        raise ValueError(f"Todo '{slug}' already exists")

    now = _now_iso()
    metadata = {
        "title": title,
        "status": "open",
        "issue_id": None,
        "issue_url": None,
        "created_at": now,
        "completed_at": None,
    }

    write_md_file(todo_file, metadata, description)
    return todo_file


def get_todo(slug: str) -> dict[str, Any] | None:
    """Read todo metadata + body."""
    todo_file = _get_todo_file(slug)

    if not todo_file.exists():
        return None

    metadata, body = read_md_file(todo_file)
    return _todo_to_dict(slug, metadata, body)


def get_todo_by_issue_id(issue_id: int) -> dict[str, Any] | None:
    """Find todo with matching issue_id."""
    for todo in list_todos():
        if todo.get("issue_id") == issue_id:
            return todo
    return None


def list_todos(status: str | None = None) -> list[dict[str, Any]]:
    """List all todos, optionally filtered by status.

    Returns list of todo dicts sorted by created_at (newest first).
    """
    todos_dir = _get_todos_dir()

    if not todos_dir.exists():
        return []

    todos = []
    for todo_file in todos_dir.iterdir():
        if not todo_file.is_file() or todo_file.suffix != ".md":
            continue

        slug = todo_file.stem
        metadata, body = read_md_file(todo_file)

        if status is not None and metadata.get("status") != status:
            continue

        todos.append(_todo_to_dict(slug, metadata, body))

    # Sort by created_at, newest first
    todos.sort(key=lambda t: t.get("created_at", ""), reverse=True)
    return todos


def update_todo(slug: str, **updates) -> None:
    """Update todo frontmatter fields."""
    todo_file = _get_todo_file(slug)

    if not todo_file.exists():
        raise ValueError(f"Todo '{slug}' not found")

    metadata, body = read_md_file(todo_file)

    for key, value in updates.items():
        metadata[key] = value

    write_md_file(todo_file, metadata, body)


def complete_todo(slug: str) -> None:
    """Mark todo as completed."""
    update_todo(slug, status="completed", completed_at=_now_iso())


def delete_todo(slug: str) -> None:
    """Delete todo file."""
    todo_file = _get_todo_file(slug)

    if not todo_file.exists():
        raise ValueError(f"Todo '{slug}' not found")

    todo_file.unlink()


def get_all_todos() -> list[dict[str, Any]]:
    """Get all todos regardless of status."""
    return list_todos()


def get_unlinked_todos() -> list[dict[str, Any]]:
    """Get todos without GitHub issue linked."""
    return [t for t in list_todos() if t.get("issue_id") is None]


def get_todos_with_issues() -> list[dict[str, Any]]:
    """Get todos that have linked GitHub issues."""
    return [t for t in list_todos() if t.get("issue_id") is not None]


def update_todo_issue_info(slug: str, issue_id: int, issue_url: str) -> None:
    """Update GitHub issue ID and URL for a todo."""
    update_todo(slug, issue_id=issue_id, issue_url=issue_url)
