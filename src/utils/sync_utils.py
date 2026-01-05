"""
Sync utilities for bidirectional synchronization between local specs and GitHub issues.
"""

import hashlib
import re
from pathlib import Path

# Separator between spec body and comments section in markdown files
SEPARATOR = "\n\n===\n***\n===\n\n"


def compute_content_hash(content: str) -> str:
    """
    Compute SHA-256 hash of content for change detection.

    Args:
        content: The text content to hash

    Returns:
        Hex-encoded SHA-256 hash string
    """
    return hashlib.sha256(content.encode("utf-8")).hexdigest()


def extract_body_from_spec_file(file_path: Path) -> str:
    """
    Extract the main body from a spec file, excluding frontmatter and comments.

    The spec file format is:
        ---
        title: ...
        status: ...
        ---
        [Main body content]
        ===
        ***
        ===
        [Comments section]

    This function returns only the main body, suitable for syncing to GitHub issue body.

    Args:
        file_path: Path to the spec markdown file

    Returns:
        The body content (without frontmatter, before comments separator),
        or empty string if file doesn't exist
    """
    if not file_path.exists():
        return ""

    content = file_path.read_text()

    # Remove frontmatter if present
    if content.startswith("---"):
        match = re.match(r"^---\n.*?\n---\n?(.*)", content, re.DOTALL)
        if match:
            content = match.group(1)

    # Remove comments section if present
    if SEPARATOR in content:
        content = content.split(SEPARATOR)[0]

    return content.strip()


def content_differs(hash1: str | None, hash2: str | None) -> bool:
    """
    Check if two content hashes differ.

    None is treated as "unknown/unsynced", which counts as different.

    Args:
        hash1: First hash (or None)
        hash2: Second hash (or None)

    Returns:
        True if hashes differ or either is None
    """
    if hash1 is None or hash2 is None:
        return True
    return hash1 != hash2


def slugify(text: str) -> str:
    """
    Convert text to a slug suitable for filenames.

    Removes [Spec]: prefix if present, lowercases, and replaces
    spaces/special chars with underscores.

    Args:
        text: The text to slugify

    Returns:
        A slug-safe string
    """
    # Remove [Spec]: prefix if present
    clean_text = text.replace("[Spec]:", "").replace("[Spec]: ", "").strip()

    # Lowercase
    slug = clean_text.lower()

    # Replace spaces and common separators with underscores
    slug = re.sub(r"[\s\-]+", "_", slug)

    # Remove anything that isn't alphanumeric or underscore
    slug = re.sub(r"[^a-z0-9_]", "", slug)

    # Collapse multiple underscores
    slug = re.sub(r"_+", "_", slug)

    # Strip leading/trailing underscores
    slug = slug.strip("_")

    return slug
