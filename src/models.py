"""
Pydantic models for frontmatter validation.

These models define the structure of YAML frontmatter in markdown files.
Templates contain only the body content; frontmatter is generated from these models.
"""

from datetime import datetime
from typing import Literal

from pydantic import BaseModel

# --- Status types ---

# Note: "active" state is now derived from git branch, not stored in status
SpecStatus = Literal["todo", "merge_ready", "completed", "abandoned"]
TaskStatus = Literal["todo", "completed"]


# --- Frontmatter models ---


class SpecFrontmatter(BaseModel):
    """Frontmatter for spec.md files."""

    title: str
    status: SpecStatus = "todo"
    assigned_to: str | None = None

    # GitHub integration
    issue_id: int | None = None
    issue_url: str | None = None
    branch: str | None = None
    pr_url: str | None = None

    # Timestamps
    created_at: str  # ISO format datetime string
    updated_at: str  # ISO format datetime string
    completed_at: str | None = None

    # Sync state
    last_synced_at: str | None = None
    local_content_hash: str | None = None
    remote_content_hash: str | None = None

    def to_dict(self) -> dict:
        """Convert to dict for YAML serialization, excluding None values."""
        return {k: v for k, v in self.model_dump().items()}


class TaskFrontmatter(BaseModel):
    """Frontmatter for task markdown files."""

    title: str
    status: TaskStatus = "todo"

    # Timestamps
    created_at: str  # ISO format datetime string
    updated_at: str  # ISO format datetime string
    completed_at: str | None = None

    def to_dict(self) -> dict:
        """Convert to dict for YAML serialization."""
        return self.model_dump()


class LogFrontmatter(BaseModel):
    """Frontmatter for work log files."""

    created_at: str  # ISO format datetime string (YYYY-MM-DDTHH:MM:SS)
    username: str
    spec_slug: str | None = None

    def to_dict(self) -> dict:
        """Convert to dict for YAML serialization."""
        return {k: v for k, v in self.model_dump().items() if v is not None}


# --- Factory functions for creating frontmatter with defaults ---


def create_spec_frontmatter(
    title: str,
    status: SpecStatus = "todo",
    assigned_to: str | None = None,
    issue_id: int | None = None,
    issue_url: str | None = None,
    branch: str | None = None,
    pr_url: str | None = None,
) -> SpecFrontmatter:
    """Create SpecFrontmatter with current timestamps."""
    now = datetime.now().isoformat()
    return SpecFrontmatter(
        title=title,
        status=status,
        assigned_to=assigned_to,
        issue_id=issue_id,
        issue_url=issue_url,
        branch=branch,
        pr_url=pr_url,
        created_at=now,
        updated_at=now,
    )


def create_task_frontmatter(title: str, status: TaskStatus = "todo") -> TaskFrontmatter:
    """Create TaskFrontmatter with current timestamps."""
    now = datetime.now().isoformat()
    return TaskFrontmatter(
        title=title,
        status=status,
        created_at=now,
        updated_at=now,
    )


def create_log_frontmatter(
    created_at: datetime, username: str, spec_slug: str | None = None
) -> LogFrontmatter:
    """Create LogFrontmatter for a work log."""
    return LogFrontmatter(
        created_at=created_at.isoformat(),
        username=username,
        spec_slug=spec_slug,
    )
