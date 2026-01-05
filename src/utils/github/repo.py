"""
GitHub repository discovery and parsing utilities.
"""

import re
from pathlib import Path
from typing import Optional, Tuple

from git import Repo
from git.exc import InvalidGitRepositoryError

from src.utils.github.exceptions import (
    GitHubRepositoryError,
    GitRepositoryNotFoundError,
)


def parse_github_repo_url(url: str) -> Optional[Tuple[str, str]]:
    """
    Parse GitHub repository URL to extract owner and repo name.

    Supports formats:
    - https://github.com/owner/repo.git
    - https://github.com/owner/repo
    - git@github.com:owner/repo.git
    - git@github.com:owner/repo
    - https://oauth2:TOKEN@github.com/owner/repo.git

    Args:
        url: Git remote URL

    Returns:
        Tuple of (owner, repo) or None if not a GitHub URL
    """
    # HTTPS format (including authenticated URLs)
    https_pattern = r"https://(?:[^@]+@)?github\.com/([^/]+)/([^/\.]+)"
    match = re.match(https_pattern, url)
    if match:
        return match.group(1), match.group(2)

    # SSH format
    ssh_pattern = r"git@github\.com:([^/]+)/([^/\.]+)"
    match = re.match(ssh_pattern, url)
    if match:
        return match.group(1), match.group(2)

    return None


def get_repo_from_git(repo_path: Path) -> Tuple[str, str]:
    """
    Discover GitHub repository from git remote.

    Args:
        repo_path: Path to git repository (default: current directory)

    Returns:
        Tuple of (owner, repo_name)

    Raises:
        GitRepositoryNotFoundError: If not a git repo
        GitHubRepositoryError: If no GitHub remote found
    """
    try:
        repo = Repo(repo_path)
    except InvalidGitRepositoryError:
        raise GitRepositoryNotFoundError(f"Not a git repository: {repo_path}")

    if not repo.remotes:
        raise GitHubRepositoryError("No git remotes configured")

    # Try 'origin' first, then any remote
    remote = None
    if "origin" in [r.name for r in repo.remotes]:
        remote = repo.remote("origin")
    else:
        remote = repo.remotes[0]

    url = remote.url
    parsed = parse_github_repo_url(url)

    if not parsed:
        raise GitHubRepositoryError(
            f"Remote URL is not a GitHub repository: {url}\n"
            f"mem requires a GitHub repository."
        )

    return parsed


def get_git_user_info(repo_path: Path) -> dict:
    """
    Get local git user configuration.

    Args:
        repo_path: Path to git repository (default: current directory)

    Returns:
        Dict with git user info: {
            'name': str,
            'email': str,
        }

    Raises:
        GitRepositoryNotFoundError: If not a git repo
        GitHubRepositoryError: If git user not configured
    """
    try:
        repo = Repo(repo_path)
        config = repo.config_reader()

        try:
            name = config.get_value("user", "name")
            email = config.get_value("user", "email")
        except Exception:
            raise GitHubRepositoryError(
                "Git user not configured.\n"
                "Please configure git:\n"
                "  git config user.name 'Your Name'\n"
                "  git config user.email 'your.email@example.com'"
            )

        return {
            "name": name,
            "email": email,
        }
    except InvalidGitRepositoryError:
        raise GitRepositoryNotFoundError(f"Not a git repository: {repo_path}")
