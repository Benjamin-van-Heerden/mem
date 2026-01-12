"""
Tests for the merge into subcommand.

The merge into command:
1. Merges dev -> test with back-merge
2. Merges test -> main with cascade back-merges
3. Requires being on dev branch
4. Validates target branch (test or main)
5. main target is dry-run by default, requires --force
"""

import uuid
from dataclasses import dataclass
from pathlib import Path

import pytest
import typer
from git import Repo

from src.commands import merge as merge_module
from src.commands.merge import into
from tests.conftest import get_worker_branch_suffix


def unique_filename(prefix: str) -> str:
    """Generate a unique filename to avoid conflicts between tests."""
    return f"{prefix}_{uuid.uuid4().hex[:8]}.txt"


@dataclass
class RepoTestInfo:
    """Information about a test repository with worker-isolated branches."""

    path: Path
    branch_names: dict

    def __truediv__(self, other):
        """Allow path operations like repo_info / 'filename'."""
        return self.path / other


@pytest.fixture
def repo_with_branches(request, setup_test_env, monkeypatch):
    """
    Set up a repo with dev, test, and main branches all pushed to origin.

    Uses worker-specific branch names for xdist compatibility.
    Patches the merge module to use these branch names.

    Returns the repo path with working directory set to it.
    """
    repo_path = setup_test_env
    monkeypatch.chdir(repo_path)
    repo = Repo(repo_path)

    # Get worker-specific suffix for branch isolation
    suffix = get_worker_branch_suffix(request)
    branch_names = {
        "dev": f"dev{suffix}",
        "test": f"test{suffix}",
        "main": f"main{suffix}",
    }

    # Create branches from current HEAD
    current_head = repo.head.commit.hexsha

    for branch_name in branch_names.values():
        if branch_name in [h.name for h in repo.heads]:
            repo.delete_head(branch_name, force=True)
        repo.create_head(branch_name, current_head)

    # Push all branches to origin
    for branch_name in branch_names.values():
        repo.heads[branch_name].checkout()
        try:
            repo.git.push("origin", branch_name, set_upstream=True, force=True)
        except Exception:
            pass

    # Checkout the dev branch
    repo.heads[branch_names["dev"]].checkout()

    # Patch the merge module functions to use our branch names
    original_get_current_branch = merge_module._get_current_branch
    original_switch_branch = merge_module._switch_branch
    original_pull_branch = merge_module._pull_branch
    original_merge_branch = merge_module._merge_branch
    original_push_branch = merge_module._push_branch

    def map_branch(name: str) -> str:
        """Map logical branch name to worker-specific name."""
        return branch_names.get(name, name)

    def patched_get_current_branch() -> str:
        actual = original_get_current_branch()
        # Reverse map: return "dev" if we're on "dev-gw0"
        for logical, actual_name in branch_names.items():
            if actual == actual_name:
                return logical
        return actual

    def patched_switch_branch(branch: str):
        return original_switch_branch(map_branch(branch))

    def patched_pull_branch(branch: str):
        return original_pull_branch(map_branch(branch))

    def patched_merge_branch(source: str, ff_only: bool = False):
        return original_merge_branch(map_branch(source), ff_only)

    def patched_push_branch(branch: str):
        return original_push_branch(map_branch(branch))

    monkeypatch.setattr(merge_module, "_get_current_branch", patched_get_current_branch)
    monkeypatch.setattr(merge_module, "_switch_branch", patched_switch_branch)
    monkeypatch.setattr(merge_module, "_pull_branch", patched_pull_branch)
    monkeypatch.setattr(merge_module, "_merge_branch", patched_merge_branch)
    monkeypatch.setattr(merge_module, "_push_branch", patched_push_branch)

    # Return a RepoTestInfo object
    info = RepoTestInfo(path=repo_path, branch_names=branch_names)

    yield info

    # Cleanup: delete remote branches (best effort)
    for branch_name in branch_names.values():
        try:
            repo.git.push("origin", "--delete", branch_name)
        except Exception:
            pass


class TestMergeIntoValidation:
    """Tests for input validation of merge into command."""

    def test_into_rejects_invalid_target(self, repo_with_branches, capsys):
        """Test that invalid target branch is rejected."""
        with pytest.raises(typer.Exit) as exc_info:
            into(target="invalid")

        assert exc_info.value.exit_code == 1
        captured = capsys.readouterr()
        assert "Invalid target" in captured.err

    def test_into_rejects_when_not_on_dev(self, repo_with_branches, capsys):
        """Test that command fails when not on dev branch."""
        repo = Repo(repo_with_branches.path)
        repo.git.checkout(repo_with_branches.branch_names["test"])

        with pytest.raises(typer.Exit) as exc_info:
            into(target="test")

        assert exc_info.value.exit_code == 1
        captured = capsys.readouterr()
        assert "Must be on 'dev' branch" in captured.err

    def test_into_accepts_test_target(self, repo_with_branches, capsys):
        """Test that 'test' is a valid target."""
        # With dry-run, should succeed
        into(target="test", dry_run=True)

        captured = capsys.readouterr()
        assert "Dry run" in captured.out

    def test_into_accepts_main_target(self, repo_with_branches, capsys):
        """Test that 'main' is a valid target (dry-run by default)."""
        into(target="main")

        captured = capsys.readouterr()
        assert "Dry run" in captured.out
        assert "mem merge into main --force" in captured.out


class TestMergeIntoTest:
    """Tests for merge into test functionality."""

    def test_into_test_dry_run_shows_steps(self, repo_with_branches, capsys):
        """Test that dry-run shows what would happen."""
        into(target="test", dry_run=True)

        captured = capsys.readouterr()
        assert "Fetch latest from origin" in captured.out
        assert "Switch to test branch" in captured.out
        assert "Merge dev into test" in captured.out
        assert "fast-forward" in captured.out.lower()

    def test_into_test_executes_merge(self, repo_with_branches, capsys):
        """Test that merge into test actually merges."""
        repo_path = repo_with_branches.path
        branch_names = repo_with_branches.branch_names
        repo = Repo(repo_path)

        # Pull latest first (shared test repo may have been modified)
        repo.git.pull("origin", branch_names["dev"], rebase="true")

        # Create a commit on dev that doesn't exist on test
        filename = unique_filename("dev_change")
        test_file = repo_path / filename
        test_file.write_text("change from dev")
        repo.git.add(filename)
        repo.git.commit("-m", f"Add {filename}")
        repo.git.push("origin", branch_names["dev"])

        # Run the merge
        into(target="test")

        captured = capsys.readouterr()
        assert "Successfully merged dev into test" in captured.out

        # Verify we're back on dev (actual branch name)
        assert repo.active_branch.name == branch_names["dev"]

        # Verify the file exists (back-merge brought test's state to dev)
        assert test_file.exists()

    def test_into_test_branches_at_same_commit(self, repo_with_branches):
        """Test that after merge, dev and test are at the same commit."""
        repo_path = repo_with_branches.path
        branch_names = repo_with_branches.branch_names
        repo = Repo(repo_path)

        # Pull latest first (shared test repo may have been modified)
        repo.git.pull("origin", branch_names["dev"], rebase="true")

        # Create a commit on dev
        filename = unique_filename("sync_test")
        test_file = repo_path / filename
        test_file.write_text("sync test")
        repo.git.add(filename)
        repo.git.commit("-m", f"Add {filename}")
        repo.git.push("origin", branch_names["dev"])

        # Run the merge
        into(target="test")

        # Get commit SHAs using actual branch names
        dev_sha = repo.heads[branch_names["dev"]].commit.hexsha
        test_sha = repo.heads[branch_names["test"]].commit.hexsha

        assert dev_sha == test_sha, "dev and test should be at the same commit"


class TestMergeIntoMain:
    """Tests for merge into main functionality."""

    def test_into_main_dry_run_by_default(self, repo_with_branches, capsys):
        """Test that main target is dry-run by default."""
        into(target="main")

        captured = capsys.readouterr()
        assert "Dry run" in captured.out
        assert "--force" in captured.out

    def test_into_main_dry_run_shows_all_steps(self, repo_with_branches, capsys):
        """Test that dry-run shows all the steps including back-merges."""
        into(target="main")

        captured = capsys.readouterr()
        assert "Merge test into main" in captured.out
        assert "Back-merge main into test" in captured.out
        assert "Back-merge test into dev" in captured.out

    def test_into_main_with_force_executes(self, repo_with_branches, capsys):
        """Test that --force actually executes the merge."""
        repo_path = repo_with_branches.path
        branch_names = repo_with_branches.branch_names
        repo = Repo(repo_path)

        # Pull latest first (shared test repo may have been modified)
        repo.git.pull("origin", branch_names["dev"], rebase="true")

        # First merge dev into test so test has something to merge
        filename = unique_filename("main_test")
        test_file = repo_path / filename
        test_file.write_text("for main")
        repo.git.add(filename)
        repo.git.commit("-m", f"Add {filename}")
        repo.git.push("origin", branch_names["dev"])

        # Merge to test first
        into(target="test")

        # Now merge to main with --force
        into(target="main", force=True)

        captured = capsys.readouterr()
        assert "Successfully merged test into main" in captured.out

        # Verify we're back on dev (actual branch name)
        assert repo.active_branch.name == branch_names["dev"]

    def test_into_main_all_branches_at_same_commit(self, repo_with_branches):
        """Test that after merge to main, all branches are at same commit."""
        repo_path = repo_with_branches.path
        branch_names = repo_with_branches.branch_names
        repo = Repo(repo_path)

        # Pull latest first (shared test repo may have been modified)
        repo.git.pull("origin", branch_names["dev"], rebase="true")

        # Create a commit and propagate through
        filename = unique_filename("all_sync")
        test_file = repo_path / filename
        test_file.write_text("all sync")
        repo.git.add(filename)
        repo.git.commit("-m", f"Add {filename}")
        repo.git.push("origin", branch_names["dev"])

        # Merge to test
        into(target="test")

        # Merge to main with force
        into(target="main", force=True)

        # Get commit SHAs using actual branch names
        dev_sha = repo.heads[branch_names["dev"]].commit.hexsha
        test_sha = repo.heads[branch_names["test"]].commit.hexsha
        main_sha = repo.heads[branch_names["main"]].commit.hexsha

        assert dev_sha == test_sha == main_sha, (
            "All branches should be at the same commit"
        )


class TestMergeIntoErrorHandling:
    """Tests for error handling in merge into command."""

    def test_into_fails_with_uncommitted_changes(self, repo_with_branches, capsys):
        """Test that merge fails if there are uncommitted changes."""
        repo_path = repo_with_branches.path

        # Create uncommitted change
        test_file = repo_path / "uncommitted.txt"
        test_file.write_text("uncommitted")

        with pytest.raises(typer.Exit) as exc_info:
            into(target="test")

        assert exc_info.value.exit_code == 1
        captured = capsys.readouterr()
        assert "Uncommitted changes" in captured.err
