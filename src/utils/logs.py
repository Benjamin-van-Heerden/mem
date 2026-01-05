"""
Markdown-based work log operations.

Work logs are stored as:
  .mem/logs/{username}_{date}_session.md

Example: .mem/logs/benjamin_van_heerden_20251229_session.md

Each log has YAML frontmatter with metadata and markdown body.
"""

import tomllib
from datetime import date, datetime
from pathlib import Path
from typing import Any

from env_settings import ENV_SETTINGS
from src.models import create_log_frontmatter
from src.utils.markdown import read_md_file, slugify, write_md_file


def _get_template_dir() -> Path:
    """Get the templates directory path."""
    return Path(__file__).parent.parent / "templates"


def _load_log_template() -> str:
    """Load the log template file."""
    template_file = _get_template_dir() / "log.md"
    return template_file.read_text()


def _get_logs_dir() -> Path:
    """Get the logs directory path."""
    return ENV_SETTINGS.logs_dir


def _get_current_github_username() -> str:
    """Get the GitHub username for the current git user.

    Reads git config user.name and looks up the corresponding GitHub username
    in user_mappings.toml. Returns the slugified GitHub username.
    """
    from src.utils.github.repo import get_git_user_info

    # Get git user info
    try:
        git_user = get_git_user_info(ENV_SETTINGS.caller_dir)
        git_name = git_user["name"]
    except Exception:
        return "unknown"

    # Read user_mappings.toml and do reverse lookup (name -> github username)
    mappings_file = ENV_SETTINGS.mem_dir / "user_mappings.toml"
    if mappings_file.exists():
        try:
            with open(mappings_file, "rb") as f:
                mappings = tomllib.load(f)

            # Find the GitHub username that has this git name
            for github_username, user_info in mappings.items():
                if user_info.get("name") == git_name:
                    return slugify(github_username)
        except Exception:
            pass

    # Fallback to slugified git name
    return slugify(git_name)


def _get_log_filename(log_date: date, username: str | None = None) -> str:
    """Get log filename for a date and user."""
    if username is None:
        username = _get_current_github_username()
    return f"{username}_{log_date.strftime('%Y%m%d')}_session.md"


def _get_log_file(log_date: date, username: str | None = None) -> Path:
    """Get path to a log file."""
    return _get_logs_dir() / _get_log_filename(log_date, username)


def _parse_log_filename(filename: str) -> tuple[str, date] | None:
    """Parse username and date from log filename.

    Filename format: {username}_{YYYYMMDD}_session.md
    Returns (username, date) or None if invalid.
    """
    if not filename.endswith("_session.md"):
        return None

    # Remove suffix
    base = filename.replace("_session.md", "")

    # Split on underscore to find date (last 8 chars should be YYYYMMDD)
    parts = base.rsplit("_", 1)
    if len(parts) != 2:
        return None

    username = parts[0]
    date_str = parts[1]

    try:
        log_date = datetime.strptime(date_str, "%Y%m%d").date()
        return (username, log_date)
    except ValueError:
        return None


def _log_to_dict(
    username: str, log_date: date, metadata: dict, body: str
) -> dict[str, Any]:
    """Convert parsed log file to a dict."""
    return {
        "username": username,
        "date": log_date.isoformat(),
        "filename": _get_log_filename(log_date, username),
        "body": body,
        **metadata,
    }


def create_log(spec_slug: str | None = None) -> Path:
    """Create work log for today for the current user.

    Creates a log file with Pydantic-generated frontmatter and template body.
    If a log for today (for this user) exists, raises ValueError.
    Returns path to the created log file.
    """
    logs_dir = _get_logs_dir()
    logs_dir.mkdir(parents=True, exist_ok=True)

    today = date.today()
    username = _get_current_github_username()
    log_file = _get_log_file(today, username)

    if log_file.exists():
        raise ValueError(f"Log for {today} (user: {username}) already exists")

    frontmatter = create_log_frontmatter(today, username, spec_slug)
    body = _load_log_template()

    write_md_file(log_file, frontmatter.to_dict(), body)
    return log_file


def get_log(log_date: date, username: str | None = None) -> dict[str, Any] | None:
    """Get log for a specific date and user.

    If username is not provided, uses the current user.
    """
    if username is None:
        username = _get_current_github_username()

    log_file = _get_log_file(log_date, username)

    if not log_file.exists():
        return None

    metadata, body = read_md_file(log_file)
    return _log_to_dict(username, log_date, metadata, body)


def get_today_log() -> dict[str, Any] | None:
    """Get today's work log for the current user if exists."""
    return get_log(date.today())


def list_logs(
    limit: int = 10, spec_slug: str | None = None, username: str | None = None
) -> list[dict[str, Any]]:
    """List recent work logs (newest first).

    Optionally filter by spec_slug and/or username.
    If no filters are provided, lists all logs from all users.
    """
    logs_dir = _get_logs_dir()

    if not logs_dir.exists():
        return []

    logs = []
    for log_file in logs_dir.iterdir():
        if not log_file.is_file() or not log_file.name.endswith("_session.md"):
            continue

        parsed = _parse_log_filename(log_file.name)
        if parsed is None:
            continue

        file_username, log_date = parsed

        metadata, body = read_md_file(log_file)

        if spec_slug is not None and metadata.get("spec_slug") != spec_slug:
            continue

        if username is not None and file_username != username:
            continue

        logs.append(_log_to_dict(file_username, log_date, metadata, body))

    # Sort by date, newest first
    logs.sort(key=lambda log: log.get("date", ""), reverse=True)

    return logs[:limit]


def update_log(log_date: date, username: str | None = None, **updates) -> None:
    """Update log frontmatter fields.

    If username is not provided, uses the current user.
    """
    if username is None:
        username = _get_current_github_username()

    log_file = _get_log_file(log_date, username)

    if not log_file.exists():
        raise ValueError(f"Log for {log_date} (user: {username}) not found")

    metadata, body = read_md_file(log_file)

    for key, value in updates.items():
        metadata[key] = value

    write_md_file(log_file, metadata, body)


def update_log_body(log_date: date, body: str, username: str | None = None) -> None:
    """Update log body content.

    If username is not provided, uses the current user.
    """
    if username is None:
        username = _get_current_github_username()

    log_file = _get_log_file(log_date, username)

    if not log_file.exists():
        raise ValueError(f"Log for {log_date} (user: {username}) not found")

    metadata, _ = read_md_file(log_file)
    write_md_file(log_file, metadata, body)


def append_to_log(section: str, content: str) -> None:
    """Append content to a section of today's log for the current user.

    Creates today's log if it doesn't exist.
    """
    today = date.today()
    username = _get_current_github_username()
    log_file = _get_log_file(today, username)

    if not log_file.exists():
        create_log()

    metadata, body = read_md_file(log_file)

    # Find the section and append to it
    section_header = f"## {section}"
    if section_header in body:
        # Find the section and the next section
        lines = body.split("\n")
        new_lines = []
        in_section = False
        appended = False

        for i, line in enumerate(lines):
            new_lines.append(line)

            if line.strip() == section_header:
                in_section = True
            elif in_section and line.startswith("## "):
                # Next section found, insert content before it
                new_lines.insert(-1, content)
                new_lines.insert(-1, "")
                in_section = False
                appended = True

        if in_section and not appended:
            # Section was at the end
            new_lines.append(content)

        body = "\n".join(new_lines)
    else:
        # Section doesn't exist, add it
        body = body.rstrip() + f"\n\n## {section}\n\n{content}\n"

    write_md_file(log_file, metadata, body)


def delete_log(log_date: date, username: str | None = None) -> None:
    """Delete a log file.

    If username is not provided, uses the current user.
    """
    if username is None:
        username = _get_current_github_username()

    log_file = _get_log_file(log_date, username)

    if not log_file.exists():
        raise ValueError(f"Log for {log_date} (user: {username}) not found")

    log_file.unlink()
