"""Markdown utilities for parsing and writing files with YAML frontmatter."""

import re
from pathlib import Path

import yaml


def parse_frontmatter(content: str) -> tuple[dict, str]:
    """Parse YAML frontmatter and body from markdown.

    Expects format:
    ---
    key: value
    ---
    body content

    Returns (metadata, body). If no frontmatter, returns ({}, content).
    """
    if not content.startswith("---"):
        return {}, content

    # Find the closing ---
    match = re.match(r"^---\n(.*?)\n---\n?(.*)", content, re.DOTALL)
    if not match:
        return {}, content

    frontmatter_str = match.group(1)
    body = match.group(2)

    try:
        metadata = yaml.safe_load(frontmatter_str) or {}
    except yaml.YAMLError:
        return {}, content

    return metadata, body


def dump_frontmatter(metadata: dict, body: str) -> str:
    """Combine metadata and body into markdown with frontmatter.

    Returns:
    ---
    key: value
    ---
    body content
    """
    if not metadata:
        return body

    frontmatter_str = yaml.dump(
        metadata,
        default_flow_style=False,
        allow_unicode=True,
        sort_keys=False,
    )

    # Ensure body has leading newline for clean separation
    if body and not body.startswith("\n"):
        body = "\n" + body

    return f"---\n{frontmatter_str}---{body}"


def read_md_file(path: Path) -> tuple[dict, str]:
    """Read a markdown file, return (metadata, body).

    Raises FileNotFoundError if file doesn't exist.
    """
    content = path.read_text()
    return parse_frontmatter(content)


def write_md_file(path: Path, metadata: dict, body: str) -> None:
    """Write markdown file with frontmatter.

    Creates parent directories if they don't exist.
    """
    path.parent.mkdir(parents=True, exist_ok=True)
    content = dump_frontmatter(metadata, body)
    path.write_text(content)


def slugify(text: str) -> str:
    """Convert text to filesystem-safe slug.

    - Lowercases
    - Replaces spaces and special chars with underscores
    - Removes consecutive underscores
    - Strips leading/trailing underscores
    """
    # Lowercase
    slug = text.lower()

    # Replace spaces and common separators with underscores
    slug = re.sub(r"[\s\-]+", "_", slug)

    # Remove anything that isn't alphanumeric or underscore
    slug = re.sub(r"[^a-z0-9_]", "", slug)

    # Collapse multiple underscores
    slug = re.sub(r"_+", "_", slug)

    # Strip leading/trailing underscores
    slug = slug.strip("_")

    return slug
