"""
Tests for GitHub API utility functions.
"""

import time

import pytest

from src.utils.github.api import (
    close_issue_with_comment,
    get_pull_request_by_url,
    is_pr_merged,
)


@pytest.fixture
def test_repo(github_client, setup_test_env):
    """Get the test repository."""
    from src.utils.github.repo import get_repo_from_git

    repo_path = setup_test_env
    owner, name = get_repo_from_git(repo_path)
    return github_client.get_repo(f"{owner}/{name}")


def test_close_issue_with_comment(test_repo):
    """Test closing an issue with a comment."""
    # Create an issue
    issue = test_repo.create_issue(
        title="Test Issue to Close", body="This issue will be closed with a comment."
    )

    # Wait for GitHub
    time.sleep(1)

    # Close it with a comment
    closed_issue = close_issue_with_comment(
        test_repo, issue.number, "Closing this issue for testing purposes."
    )

    # Verify it's closed
    assert closed_issue.state == "closed"

    # Verify comment was added
    comments = list(closed_issue.get_comments())
    assert len(comments) >= 1
    assert "Closing this issue for testing purposes" in comments[-1].body


def test_get_pull_request_by_url(test_repo, setup_test_env):
    """Test getting a PR by URL."""
    import os

    from git import Repo

    repo_path = setup_test_env

    # Create a branch and PR
    local_repo = Repo(repo_path)

    # Use unique branch and file names based on PID
    branch_name = f"test-pr-branch-{os.getpid()}"
    test_branch = local_repo.create_head(branch_name)
    test_branch.checkout()

    # Make a change with unique filename
    test_file = repo_path / f"test_pr_file_{os.getpid()}.txt"
    test_file.write_text(f"Test content for PR {os.getpid()}")
    local_repo.git.add(A=True)
    local_repo.git.commit("-m", f"Test commit for PR {os.getpid()}")
    local_repo.git.push("origin", branch_name)

    time.sleep(2)

    # Create PR on GitHub (base is main, which exists on remote)
    pr = test_repo.create_pull(
        title="Test PR", body="Test PR body", head=branch_name, base="main"
    )

    time.sleep(2)

    # Get PR by URL
    pr_url = pr.html_url
    fetched_pr = get_pull_request_by_url(test_repo, pr_url)

    assert fetched_pr is not None
    assert fetched_pr.number == pr.number
    assert fetched_pr.title == "Test PR"


def test_get_pull_request_by_url_invalid(test_repo):
    """Test that get_pull_request_by_url returns None for invalid URLs."""
    # Invalid PR number
    result = get_pull_request_by_url(
        test_repo, "https://github.com/owner/repo/pull/99999"
    )
    assert result is None

    # Invalid URL format
    result = get_pull_request_by_url(test_repo, "not-a-url")
    assert result is None


def test_is_pr_merged_unmerged(test_repo, setup_test_env):
    """Test is_pr_merged returns False for unmerged PRs."""
    import os

    from git import Repo

    repo_path = setup_test_env
    local_repo = Repo(repo_path)

    # Create test branch from main with unique name
    branch_name = f"test-unmerged-pr-{os.getpid()}"
    test_branch = local_repo.create_head(branch_name)
    test_branch.checkout()

    # Make a change with unique filename
    test_file = repo_path / f"unmerged_test_{os.getpid()}.txt"
    test_file.write_text(f"Unmerged content {os.getpid()}")
    local_repo.git.add(A=True)
    local_repo.git.commit("-m", f"Unmerged commit {os.getpid()}")
    local_repo.git.push("origin", branch_name)

    time.sleep(2)

    # Create PR (base is main)
    pr = test_repo.create_pull(
        title="Unmerged PR",
        body="This PR will not be merged",
        head=branch_name,
        base="main",
    )

    time.sleep(2)

    # Check if merged (should be False)
    assert is_pr_merged(test_repo, pr.html_url) is False


def test_is_pr_merged_after_merge(test_repo, setup_test_env):
    """Test is_pr_merged returns True for merged PRs."""
    import os

    from git import Repo

    repo_path = setup_test_env
    local_repo = Repo(repo_path)

    # Create test branch from main with unique name
    branch_name = f"test-merged-pr-{os.getpid()}"
    test_branch = local_repo.create_head(branch_name)
    test_branch.checkout()

    # Make a change with unique filename
    test_file = repo_path / f"merged_test_{os.getpid()}.txt"
    test_file.write_text(f"Merged content {os.getpid()}")
    local_repo.git.add(A=True)
    local_repo.git.commit("-m", f"Merged commit {os.getpid()}")
    local_repo.git.push("origin", branch_name)

    time.sleep(2)

    # Create and merge PR (base is main)
    pr = test_repo.create_pull(
        title="Merged PR", body="This PR will be merged", head=branch_name, base="main"
    )

    time.sleep(2)

    # Merge the PR
    pr.merge(merge_method="squash")

    time.sleep(2)

    # Check if merged (should be True)
    assert is_pr_merged(test_repo, pr.html_url) is True


def test_is_pr_merged_invalid_url(test_repo):
    """Test is_pr_merged returns False for invalid URLs."""
    assert is_pr_merged(test_repo, "invalid-url") is False
    assert is_pr_merged(test_repo, "https://github.com/owner/repo/pull/99999") is False
