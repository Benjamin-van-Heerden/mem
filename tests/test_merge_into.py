"""
Tests for the merge into subcommand.

The merge into command:
1. Merges dev -> test with fast-forward only
2. Merges test -> main with fast-forward only
3. Requires being on dev branch
4. Validates target branch (test or main)
5. main target is dry-run by default, requires --force
"""

import uuid

import pytest
import typer
from git import Repo

from src.commands.merge import into


def unique_filename(prefix: str) -> str:
    """Generate a unique filename to avoid conflicts between tests."""
    return f"{prefix}_{uuid.uuid4().hex[:8]}.txt"


@pytest.fixture
def repo_with_branches(setup_test_env, monkeypatch):
    """
    Set up a repo with dev, test, and main branches all pushed to origin.

    Returns the repo path with working directory set to it.
    """
    repo_path = setup_test_env
    monkeypatch.chdir(repo_path)
    repo = Repo(repo_path)

    current_head = repo.head.commit.hexsha

    # Create test and main branches from current HEAD
    for branch_name in ["test", "main"]:
        if branch_name in [h.name for h in repo.heads]:
            repo.delete_head(branch_name, force=True)
        repo.create_head(branch_name, current_head)

    # Push all branches to origin
    for branch_name in ["dev", "test", "main"]:
        repo.heads[branch_name].checkout()
        try:
            repo.git.push("origin", branch_name, set_upstream=True, force=True)
        except Exception:
            pass

    # Checkout dev branch
    repo.heads["dev"].checkout()

    yield repo_path

    # Cleanup: delete remote branches (best effort)
    for branch_name in ["test", "main"]:
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
        repo = Repo(repo_with_branches)
        repo.git.checkout("test")

        with pytest.raises(typer.Exit) as exc_info:
            into(target="test")

        assert exc_info.value.exit_code == 1
        captured = capsys.readouterr()
        assert "Must be on 'dev' branch" in captured.err

    def test_into_accepts_test_target(self, repo_with_branches, capsys):
        """Test that 'test' is a valid target."""
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
        repo_path = repo_with_branches
        repo = Repo(repo_path)

        # Create a commit on dev that doesn't exist on test
        filename = unique_filename("dev_change")
        test_file = repo_path / filename
        test_file.write_text("change from dev")
        repo.git.add(filename)
        repo.git.commit("-m", f"Add {filename}")
        repo.git.push("origin", "dev")

        # Run the merge
        into(target="test")

        captured = capsys.readouterr()
        assert "Successfully merged dev into test" in captured.out

        # Verify we're back on dev
        assert repo.active_branch.name == "dev"

        # Verify the file exists
        assert test_file.exists()

    def test_into_test_branches_at_same_commit(self, repo_with_branches):
        """Test that after merge, dev and test are at the same commit."""
        repo_path = repo_with_branches
        repo = Repo(repo_path)

        # Create a commit on dev
        filename = unique_filename("sync_test")
        test_file = repo_path / filename
        test_file.write_text("sync test")
        repo.git.add(filename)
        repo.git.commit("-m", f"Add {filename}")
        repo.git.push("origin", "dev")

        # Run the merge
        into(target="test")

        # Get commit SHAs
        dev_sha = repo.heads["dev"].commit.hexsha
        test_sha = repo.heads["test"].commit.hexsha

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
        """Test that dry-run shows all the steps."""
        into(target="main")

        captured = capsys.readouterr()
        assert "Merge test into main" in captured.out
        assert "fast-forward" in captured.out.lower()
        assert "Push main to origin" in captured.out
        assert "Switch back to dev" in captured.out

    def test_into_main_with_force_executes(self, repo_with_branches, capsys):
        """Test that --force actually executes the merge."""
        repo_path = repo_with_branches
        repo = Repo(repo_path)

        # First merge dev into test so test has something to merge
        filename = unique_filename("main_test")
        test_file = repo_path / filename
        test_file.write_text("for main")
        repo.git.add(filename)
        repo.git.commit("-m", f"Add {filename}")
        repo.git.push("origin", "dev")

        # Merge to test first
        into(target="test")

        # Now merge to main with --force
        into(target="main", force=True)

        captured = capsys.readouterr()
        assert "Successfully merged test into main" in captured.out

        # Verify we're back on dev
        assert repo.active_branch.name == "dev"

    def test_into_main_test_and_main_at_same_commit(self, repo_with_branches):
        """Test that after merge to main, test and main are at same commit."""
        repo_path = repo_with_branches
        repo = Repo(repo_path)

        # Create a commit and propagate through
        filename = unique_filename("all_sync")
        test_file = repo_path / filename
        test_file.write_text("all sync")
        repo.git.add(filename)
        repo.git.commit("-m", f"Add {filename}")
        repo.git.push("origin", "dev")

        # Merge to test
        into(target="test")

        # Merge to main with force
        into(target="main", force=True)

        # Get commit SHAs
        test_sha = repo.heads["test"].commit.hexsha
        main_sha = repo.heads["main"].commit.hexsha

        assert test_sha == main_sha, "test and main should be at the same commit"


class TestMergeIntoErrorHandling:
    """Tests for error handling in merge into command."""

    def test_into_fails_with_uncommitted_changes(self, repo_with_branches, capsys):
        """Test that merge fails if there are uncommitted changes."""
        repo_path = repo_with_branches

        # Create uncommitted change
        test_file = repo_path / "uncommitted.txt"
        test_file.write_text("uncommitted")

        with pytest.raises(typer.Exit) as exc_info:
            into(target="test")

        assert exc_info.value.exit_code == 1
        captured = capsys.readouterr()
        assert "Uncommitted changes" in captured.err
