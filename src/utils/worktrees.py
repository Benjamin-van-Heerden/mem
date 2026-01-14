"""
Git worktree utilities for spec isolation.

Worktrees allow each spec to have its own working directory with a dedicated
feature branch, enabling parallel work on multiple specs with separate agent sessions.

Directory structure:
  /path/to/project/                    # Main repo, stays on 'dev'
  /path/to/project-worktrees/          # Sibling directory for worktrees
    ├── user_auth/                     # Worktree on 'dev-user-user_auth' branch
    └── fix_sync/                      # Worktree on 'dev-user-fix_sync' branch
"""

from pathlib import Path
from typing import NamedTuple

from git import Repo


class WorktreeInfo(NamedTuple):
    """Information about a git worktree."""

    path: Path
    branch: str
    is_main: bool


def is_worktree(path: Path) -> bool:
    """Check if the given path is a git worktree (not the main repo).

    In a worktree, .git is a file containing "gitdir: /path/to/main/.git/worktrees/<name>"
    In the main repo, .git is a directory.
    """
    git_path = path / ".git"
    return git_path.is_file()


def get_main_repo_path(worktree_path: Path) -> Path | None:
    """Get the main repository path from a worktree.

    Returns None if not in a worktree.
    """
    git_file = worktree_path / ".git"

    if not git_file.is_file():
        return None

    content = git_file.read_text().strip()
    if not content.startswith("gitdir:"):
        return None

    gitdir_path = content.split("gitdir:", 1)[1].strip()
    gitdir = Path(gitdir_path)

    if "/worktrees/" in str(gitdir):
        main_git_dir = str(gitdir).split("/worktrees/")[0]
        return Path(main_git_dir).parent

    return None


def get_worktrees_base_dir(main_repo_path: Path) -> Path:
    """Get the base directory for worktrees (sibling to main repo).

    Creates: /path/to/project-worktrees/
    Uses resolve() to handle symlinks (e.g., /var -> /private/var on macOS).
    """
    resolved = main_repo_path.resolve()
    return resolved.parent / f"{resolved.name}-worktrees"


def get_worktree_path(main_repo_path: Path, slug: str) -> Path:
    """Get the path for a specific spec's worktree."""
    return get_worktrees_base_dir(main_repo_path) / slug


def create_worktree(main_repo_path: Path, slug: str, branch_name: str) -> Path:
    """Create a new worktree for a spec.

    Creates the worktree at ../<project>-worktrees/<slug>/ with the given branch.
    Creates the branch if it doesn't exist.

    Returns the path to the created worktree.
    """
    repo = Repo(main_repo_path)
    worktree_path = get_worktree_path(main_repo_path, slug)

    worktree_path.parent.mkdir(parents=True, exist_ok=True)

    if branch_name in repo.heads:
        repo.git.worktree("add", str(worktree_path), branch_name)
    else:
        repo.git.worktree("add", "-b", branch_name, str(worktree_path))

    return worktree_path


def remove_worktree(main_repo_path: Path, slug: str, force: bool = False) -> bool:
    """Remove a worktree for a spec.

    Returns True if removed, False if worktree didn't exist.
    """
    repo = Repo(main_repo_path)
    worktree_path = get_worktree_path(main_repo_path, slug)

    if not worktree_path.exists():
        return False

    if force:
        repo.git.worktree("remove", str(worktree_path), "--force")
    else:
        repo.git.worktree("remove", str(worktree_path))

    # git worktree remove sometimes leaves an empty directory shell behind
    if worktree_path.exists() and worktree_path.is_dir():
        try:
            worktree_path.rmdir()  # only succeeds if directory is empty
        except OSError:
            pass  # directory not empty or other issue, leave it alone

    return True


def list_worktrees(main_repo_path: Path) -> list[WorktreeInfo]:
    """List all worktrees for the repository.

    Returns list of WorktreeInfo with path, branch, and whether it's the main repo.
    Uses resolve() to handle symlinks consistently.
    """
    repo = Repo(main_repo_path)
    resolved_main = main_repo_path.resolve()
    worktrees = []

    output = repo.git.worktree("list", "--porcelain")

    current_path = None
    current_branch = None

    for line in output.split("\n"):
        if line.startswith("worktree "):
            current_path = Path(line.split(" ", 1)[1])
        elif line.startswith("branch "):
            current_branch = line.split(" ", 1)[1].replace("refs/heads/", "")
        elif line == "" and current_path is not None:  # type: ignore
            is_main = current_path.resolve() == resolved_main
            worktrees.append(
                WorktreeInfo(
                    path=current_path,
                    branch=current_branch or "",
                    is_main=is_main,
                )
            )
            current_path = None
            current_branch = None

    if current_path is not None:
        is_main = current_path.resolve() == resolved_main
        worktrees.append(
            WorktreeInfo(
                path=current_path,
                branch=current_branch or "",
                is_main=is_main,
            )
        )

    return worktrees


def get_worktree_for_spec(main_repo_path: Path, slug: str) -> WorktreeInfo | None:
    """Get worktree info for a specific spec slug.

    Returns None if no worktree exists for this spec.
    Uses resolve() to handle symlinks consistently.
    """
    expected_path = get_worktree_path(main_repo_path, slug).resolve()

    for wt in list_worktrees(main_repo_path):
        if wt.path.resolve() == expected_path:
            return wt

    return None


def get_spec_slug_from_worktree(worktree_path: Path) -> str | None:
    """Extract the spec slug from a worktree path.

    Assumes worktree is at ../<project>-worktrees/<slug>/
    Returns None if path doesn't match expected pattern.
    """
    if not is_worktree(worktree_path):
        return None

    parent_name = worktree_path.parent.name
    if not parent_name.endswith("-worktrees"):
        return None

    return worktree_path.name


def resolve_repo_and_spec(
    current_path: Path,
) -> tuple[Path, str | None, bool]:
    """Resolve the main repo path and active spec from current location.

    Returns:
        (main_repo_path, active_spec_slug_or_none, is_in_worktree)

    If in main repo: returns (current_path, None, False)
    If in worktree: returns (main_repo_path, spec_slug, True)
    """
    if is_worktree(current_path):
        main_repo = get_main_repo_path(current_path)
        if main_repo:
            slug = get_spec_slug_from_worktree(current_path)
            return (main_repo, slug, True)

    return (current_path, None, False)
