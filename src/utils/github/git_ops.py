"""
Git operations and branch management utilities.
"""

from pathlib import Path
from typing import List, Optional

from git import Repo

from src.utils.github.exceptions import GitHubError, GitRepositoryNotFoundError


def ensure_branches_exist(
    repo_path: Path, branches: Optional[List[str]] = None
) -> None:
    """
    Ensure specified branches exist both locally and on remote.
    Creates them from current HEAD and pushes to origin if they don't exist.

    Args:
        repo_path: Path to git repository
        branches: List of branch names to ensure exist (default: ['main', 'test', 'dev'])

    Raises:
        GitHubError: If branch creation or remote sync fails
        GitRepositoryNotFoundError: If path is not a git repository
    """
    if branches is None:
        branches = ["main", "test", "dev"]

    try:
        repo = Repo(repo_path)
        if repo.bare:
            raise GitRepositoryNotFoundError(f"Bare repository found at {repo_path}")

        origin = repo.remote("origin")

        # Fetch to get latest remote state
        origin.fetch()

        for branch_name in branches:
            # Check if branch exists locally
            local_exists = branch_name in [h.name for h in repo.heads]

            # Check if branch exists on remote
            remote_exists = f"origin/{branch_name}" in [ref.name for ref in origin.refs]

            if not local_exists and not remote_exists:
                # Create branch from current HEAD
                branch = repo.create_head(branch_name)
                # Push to remote and set upstream
                repo.git.push("origin", branch_name, set_upstream=True)
            elif not local_exists and remote_exists:
                # Create local tracking branch from remote
                remote_ref = origin.refs[branch_name]
                branch = repo.create_head(branch_name, remote_ref)
                branch.set_tracking_branch(remote_ref)
            elif local_exists and not remote_exists:
                # Push existing local branch to remote
                repo.git.push("origin", branch_name, set_upstream=True)
            # else: both exist, nothing to do

    except GitRepositoryNotFoundError:
        raise
    except Exception as e:
        raise GitHubError(f"Failed to ensure branches exist: {e}")


def switch_to_branch(repo_path: Path, branch_name: str = "dev") -> None:
    """
    Switch to the specified branch.

    Args:
        repo_path: Path to git repository
        branch_name: Branch to switch to

    Raises:
        GitHubError: If branch switch fails
        GitRepositoryNotFoundError: If path is not a git repository
    """
    try:
        repo = Repo(repo_path)
        repo.git.switch(branch_name)
    except Exception as e:
        raise GitHubError(f"Failed to switch to branch '{branch_name}': {e}")


def smart_switch(repo_path: Path, branch_name: str, base_branch: str = "dev") -> bool:
    """
    Switch to a branch if it exists (locally or on remote), otherwise create it from base_branch.

    Args:
        repo_path: Path to git repository
        branch_name: Branch to switch to or create
        base_branch: Branch to create from if branch_name doesn't exist

    Returns:
        bool: True if a new branch was created, False if it already existed.

    Raises:
        GitHubError: If git operations fail
    """
    try:
        repo = Repo(repo_path)
        origin = repo.remote("origin")
        origin.fetch()

        # Check local
        local_exists = branch_name in [h.name for h in repo.heads]
        # Check remote
        remote_exists = f"origin/{branch_name}" in [ref.name for ref in origin.refs]

        if local_exists:
            repo.git.switch(branch_name)
            return False
        elif remote_exists:
            # Create local tracking branch from remote
            repo.git.switch("--track", f"origin/{branch_name}")
            return False
        else:
            # Create new from base
            # First ensure base branch is checked out and updated
            repo.git.switch(base_branch)
            try:
                repo.git.pull("origin", base_branch)
            except Exception:
                # Best effort pull, continue if it fails
                pass

            repo.git.switch("-c", branch_name)
            return True
    except Exception as e:
        raise GitHubError(f"Failed to switch to or create branch '{branch_name}': {e}")


def get_current_branch(repo_path: Path) -> str:
    """
    Get the name of the currently active branch.
    """
    try:
        repo = Repo(repo_path)
        return repo.active_branch.name
    except Exception as e:
        raise GitHubError(f"Failed to get current branch: {e}")


def push_branch(repo_path: Path, branch_name: str, set_upstream: bool = True) -> None:
    """
    Push a branch to origin.
    """
    try:
        repo = Repo(repo_path)
        args = ["origin", branch_name]
        if set_upstream:
            args.append("--set-upstream")
        repo.git.push(*args)
    except Exception as e:
        raise GitHubError(f"Failed to push branch '{branch_name}': {e}")
