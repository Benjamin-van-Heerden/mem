"""
Markdown-based spec operations.

Specs are stored as:
  .mem/specs/{slug}/spec.md              # todo, merge_ready
  .mem/specs/completed/{slug}/spec.md    # Completed (PR merged)
  .mem/specs/abandoned/{slug}/spec.md    # Manually abandoned

A spec is "active" when the current git branch matches spec.branch.
This is derived at runtime, not stored in the status field.

Each spec.md has YAML frontmatter with metadata and markdown body.
"""

import shutil
from datetime import datetime
from pathlib import Path
from typing import Any

from git import Repo
from git.exc import InvalidGitRepositoryError

from env_settings import ENV_SETTINGS
from src.models import create_spec_frontmatter
from src.utils.markdown import read_md_file, slugify, write_md_file
from src.utils.worktrees import get_spec_slug_from_worktree, is_worktree


def _get_template_dir() -> Path:
    """Get the templates directory path."""
    return Path(__file__).parent.parent / "templates"


def _load_spec_template() -> str:
    """Load the spec template file, preferring global over local."""
    from src.utils.spec_template import load_spec_template

    return load_spec_template()


def _get_specs_dir() -> Path:
    """Get the specs directory path."""
    return ENV_SETTINGS.specs_dir


def _get_completed_dir() -> Path:
    """Get the completed specs subdirectory path."""
    return _get_specs_dir() / "completed"


def _get_abandoned_dir() -> Path:
    """Get the abandoned specs subdirectory path."""
    return _get_specs_dir() / "abandoned"


def _find_spec_dir(slug: str) -> Path | None:
    """Find spec directory across all locations (root, completed, abandoned).

    Returns the path to the spec directory, or None if not found.
    """
    # Check root specs dir first
    root_path = _get_specs_dir() / slug
    if root_path.exists() and (root_path / "spec.md").exists():
        return root_path

    # Check completed
    completed_path = _get_completed_dir() / slug
    if completed_path.exists() and (completed_path / "spec.md").exists():
        return completed_path

    # Check abandoned
    abandoned_path = _get_abandoned_dir() / slug
    if abandoned_path.exists() and (abandoned_path / "spec.md").exists():
        return abandoned_path

    return None


def _get_spec_dir(slug: str) -> Path:
    """Get path to a specific spec's directory.

    For new specs, returns the root location.
    For existing specs, finds the actual location.
    """
    found = _find_spec_dir(slug)
    if found:
        return found
    # Default to root for new specs
    return _get_specs_dir() / slug


def _get_spec_file(slug: str) -> Path:
    """Get path to a spec's spec.md file."""
    return _get_spec_dir(slug) / "spec.md"


def _now_iso() -> str:
    """Return current timestamp in ISO format."""
    return datetime.now().isoformat()


def _spec_to_dict(slug: str, metadata: dict, body: str) -> dict[str, Any]:
    """Convert parsed spec file to a dict with slug and body included."""
    return {
        "slug": slug,
        "body": body,
        **metadata,
    }


def create_spec(title: str) -> Path:
    """Create new spec directory and spec.md file.

    Returns the path to the created spec.md file.
    """
    slug = slugify(title)
    spec_file = _get_spec_file(slug)

    if spec_file.exists():
        raise ValueError(f"Spec '{slug}' already exists")

    frontmatter = create_spec_frontmatter(title)
    body = _load_spec_template()

    write_md_file(spec_file, frontmatter.to_dict(), body)
    return spec_file


def resolve_spec_slug_prefix(prefix: str) -> tuple[str | None, list[str]]:
    """
    Resolve a spec slug by unique prefix (git-hash style).

    Returns:
        (resolved_slug_or_none, matches)

    - If `prefix` exactly matches a spec slug, it wins immediately.
    - Otherwise, we search across all spec locations (root, completed, abandoned).
    - If there is exactly one match, we return it.
    - If there are 0 or >1 matches, resolved_slug is None and matches contains
      the candidate slugs (sorted).
    """
    prefix = (prefix or "").strip()
    if not prefix:
        return (None, [])

    # Exact match wins
    exact_dir = _find_spec_dir(prefix)
    if exact_dir is not None:
        return (prefix, [prefix])

    matches: list[str] = []

    # Search all locations
    for base_dir in (_get_specs_dir(), _get_completed_dir(), _get_abandoned_dir()):
        if not base_dir.exists():
            continue

        for spec_dir in base_dir.iterdir():
            if not spec_dir.is_dir():
                continue
            # Skip the completed and abandoned subdirectories if scanning root
            if spec_dir.name in ("completed", "abandoned"):
                continue

            slug = spec_dir.name
            if not slug.startswith(prefix):
                continue

            if (spec_dir / "spec.md").exists():
                matches.append(slug)

    matches = sorted(set(matches))
    if len(matches) == 1:
        return (matches[0], matches)

    return (None, matches)


def get_spec(slug: str) -> dict[str, Any] | None:
    """Read spec metadata + body by slug or unique slug prefix.

    Returns dict with all metadata fields plus 'slug' and 'body',
    or None if spec doesn't exist or slug prefix is ambiguous.
    """
    resolved, matches = resolve_spec_slug_prefix(slug)
    if resolved is None:
        return None

    spec_file = _get_spec_file(resolved)

    if not spec_file.exists():
        return None

    metadata, body = read_md_file(spec_file)
    return _spec_to_dict(resolved, metadata, body)


def get_spec_by_issue_id(issue_id: int) -> dict[str, Any] | None:
    """Find spec with matching issue_id in frontmatter."""
    for spec in list_specs():
        if spec.get("issue_id") == issue_id:
            return spec
    return None


def _list_specs_in_dir(
    directory: Path, status_filter: str | None = None
) -> list[dict[str, Any]]:
    """List specs in a specific directory, optionally filtered by status."""
    if not directory.exists():
        return []

    specs = []
    for spec_dir in directory.iterdir():
        if not spec_dir.is_dir():
            continue
        # Skip the completed and abandoned subdirectories
        if spec_dir.name in ("completed", "abandoned"):
            continue

        spec_file = spec_dir / "spec.md"
        if not spec_file.exists():
            continue

        metadata, body = read_md_file(spec_file)
        slug = spec_dir.name

        if status_filter is not None and metadata.get("status") != status_filter:
            continue

        specs.append(_spec_to_dict(slug, metadata, body))

    return specs


def list_specs(status: str | None = None) -> list[dict[str, Any]]:
    """List specs, optionally filtered by status.

    By default (status=None), lists only active specs from the root directory
    (excludes completed and abandoned).

    Use status="completed" to list completed specs.
    Use status="abandoned" to list abandoned specs.

    Returns list of spec dicts sorted by updated_at (newest first).
    """
    if status == "completed":
        specs = _list_specs_in_dir(_get_completed_dir())
    elif status == "abandoned":
        specs = _list_specs_in_dir(_get_abandoned_dir())
    else:
        # List from root directory, optionally filter by status
        specs = _list_specs_in_dir(_get_specs_dir(), status_filter=status)

    # Sort by updated_at, newest first
    specs.sort(key=lambda s: s.get("updated_at", ""), reverse=True)
    return specs


def get_current_branch() -> str | None:
    """Get the current git branch name."""
    try:
        repo = Repo(ENV_SETTINGS.caller_dir)
        return repo.active_branch.name
    except (InvalidGitRepositoryError, TypeError):
        return None


def ensure_on_dev_branch() -> tuple[bool, str | None]:
    """Ensure we're on the dev branch, switching if on main or test.

    Returns:
        (switched, message) - switched is True if we changed branches,
        message describes what happened or None if already on dev/feature branch.
    """
    current = get_current_branch()
    if current is None:
        return False, None

    if current in ("main", "test"):
        try:
            repo = Repo(ENV_SETTINGS.caller_dir)
            repo.git.checkout("dev")
            return True, f"Switched from '{current}' to 'dev' branch"
        except Exception as e:
            return False, f"Failed to switch to dev: {e}"

    return False, None


def get_branch_diff_stat(branch_name: str | None = None) -> str | None:
    """Get git diff --stat for current branch against dev.

    Returns the diff stat output as a string, or None if not available.
    """
    if branch_name is None:
        branch_name = get_current_branch()

    if branch_name is None or branch_name in ("dev", "main", "master", "test"):
        return None

    try:
        repo = Repo(ENV_SETTINGS.caller_dir)
        diff_stat = repo.git.diff("dev", "--stat")
        if diff_stat.strip():
            return diff_stat
        return None
    except Exception:
        return None


def get_active_spec() -> dict[str, Any] | None:
    """Get the currently active spec based on worktree or git branch.

    Detection order:
    1. If in a worktree, the spec slug is derived from the worktree directory name
    2. Otherwise, returns the spec whose `branch` field matches the current git branch

    Returns None if on 'dev', 'main', 'master', or 'test' branch in the main repo,
    or if no matching spec is found.
    """
    current_dir = ENV_SETTINGS.caller_dir

    # Check if we're in a worktree first
    if is_worktree(current_dir):
        slug = get_spec_slug_from_worktree(current_dir)
        if slug:
            spec = get_spec(slug)
            if spec:
                return spec

    # Fall back to branch-based detection
    current_branch = get_current_branch()

    if current_branch is None:
        return None

    # No active spec on main branches
    if current_branch in ("dev", "main", "master", "test"):
        return None

    # Find spec with matching branch
    all_specs = list_specs()  # Gets todo + other non-archived specs
    for spec in all_specs:
        if spec.get("branch") == current_branch:
            return spec

    return None


def get_branch_status() -> tuple[str, dict | None, str | None]:
    """Get current branch status for onboard display.

    Returns:
        (branch_name, active_spec_or_none, warning_message_or_none)
    """
    current_dir = ENV_SETTINGS.caller_dir

    # Check if we're in a worktree
    if is_worktree(current_dir):
        slug = get_spec_slug_from_worktree(current_dir)
        if slug:
            spec = get_spec(slug)
            if spec:
                branch = spec.get("branch", "unknown")
                return (branch, spec, None)
            return ("worktree", None, f"In worktree '{slug}' but spec not found")
        return ("worktree", None, "In a worktree but could not determine spec")

    current_branch = get_current_branch()

    if current_branch is None:
        return ("unknown", None, "Not in a git repository")

    if current_branch in ("dev", "main", "master", "test"):
        return (current_branch, None, None)

    active_spec = get_active_spec()
    if active_spec:
        return (current_branch, active_spec, None)

    # On a branch but no matching spec
    return (
        current_branch,
        None,
        f"On branch '{current_branch}' but no spec is associated with this branch",
    )


def update_spec(slug: str, **updates) -> None:
    """Update spec frontmatter fields.

    Automatically sets updated_at to current time.
    """
    spec_file = _get_spec_file(slug)

    if not spec_file.exists():
        raise ValueError(f"Spec '{slug}' not found")

    metadata, body = read_md_file(spec_file)

    for key, value in updates.items():
        metadata[key] = value

    metadata["updated_at"] = _now_iso()

    write_md_file(spec_file, metadata, body)


def update_spec_body(slug: str, body: str) -> None:
    """Update spec body content (for sync)."""
    spec_file = _get_spec_file(slug)

    if not spec_file.exists():
        raise ValueError(f"Spec '{slug}' not found")

    metadata, _ = read_md_file(spec_file)
    metadata["updated_at"] = _now_iso()

    write_md_file(spec_file, metadata, body)


def get_spec_path(slug: str) -> Path:
    """Get path to spec directory."""
    return _get_spec_dir(slug)


def get_spec_file_path(slug: str) -> Path:
    """Get path to spec.md file."""
    return _get_spec_file(slug)


# --- Convenience functions matching old DB API ---


def update_spec_status(slug: str, status: str) -> None:
    """Update the status of a specification."""
    update_spec(slug, status=status)


def update_spec_branch(slug: str, branch_name: str) -> None:
    """Update the branch associated with a specification."""
    update_spec(slug, branch=branch_name)


def assign_spec(slug: str, username: str) -> None:
    """Assign a specification to a GitHub user."""
    update_spec(slug, assigned_to=username)


def update_spec_issue_info(slug: str, issue_id: int, issue_url: str) -> None:
    """Update GitHub issue ID and URL for a specification."""
    update_spec(slug, issue_id=issue_id, issue_url=issue_url)


def update_spec_pr_url(slug: str, pr_url: str) -> None:
    """Update the Pull Request URL for a specification."""
    update_spec(slug, pr_url=pr_url)


def mark_spec_synced(
    slug: str,
    local_content_hash: str,
    remote_content_hash: str,
) -> None:
    """Mark a spec as synced with GitHub."""
    update_spec(
        slug,
        last_synced_at=_now_iso(),
        local_content_hash=local_content_hash,
        remote_content_hash=remote_content_hash,
    )


def get_unlinked_specs() -> list[dict[str, Any]]:
    """Get specs that have no GitHub issue linked (need outbound create)."""
    return [s for s in list_specs() if s.get("issue_id") is None]


def get_specs_with_issues() -> list[dict[str, Any]]:
    """Get all specs that have linked GitHub issues."""
    return [s for s in list_specs() if s.get("issue_id") is not None]


def get_all_specs() -> list[dict[str, Any]]:
    """Get all specs regardless of status."""
    return list_specs()


def delete_spec(slug: str) -> None:
    """Delete a spec and its directory.

    This removes the entire spec directory including tasks.
    """
    spec_dir = _get_spec_dir(slug)

    if not spec_dir.exists():
        raise ValueError(f"Spec '{slug}' not found")

    shutil.rmtree(spec_dir)


def move_spec_to_completed(slug: str) -> Path:
    """Move a spec to the completed subdirectory.

    Updates the spec status to 'completed' and moves the directory.
    Returns the new path to the spec directory.
    """
    spec_dir = _find_spec_dir(slug)
    if spec_dir is None:
        raise ValueError(f"Spec '{slug}' not found")

    # Update status before moving
    update_spec(slug, status="completed", completed_at=_now_iso())

    # Ensure completed directory exists
    completed_dir = _get_completed_dir()
    completed_dir.mkdir(parents=True, exist_ok=True)

    # Move the spec directory
    new_path = completed_dir / slug
    if new_path.exists():
        raise ValueError(f"Spec '{slug}' already exists in completed/")

    shutil.move(str(spec_dir), str(new_path))
    return new_path


def move_spec_to_abandoned(slug: str) -> Path:
    """Move a spec to the abandoned subdirectory.

    Updates the spec status to 'abandoned' and moves the directory.
    Returns the new path to the spec directory.
    """
    spec_dir = _find_spec_dir(slug)
    if spec_dir is None:
        raise ValueError(f"Spec '{slug}' not found")

    # Update status before moving
    update_spec(slug, status="abandoned")

    # Ensure abandoned directory exists
    abandoned_dir = _get_abandoned_dir()
    abandoned_dir.mkdir(parents=True, exist_ok=True)

    # Move the spec directory
    new_path = abandoned_dir / slug
    if new_path.exists():
        raise ValueError(f"Spec '{slug}' already exists in abandoned/")

    shutil.move(str(spec_dir), str(new_path))
    return new_path
