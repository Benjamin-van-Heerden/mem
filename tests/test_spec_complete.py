"""
Tests for the spec complete command.

The complete command orchestrates multiple operations:
1. Validates all tasks are completed
2. Updates spec status to merge_ready
3. Commits and pushes all changes
4. Creates a Pull Request on GitHub
5. Switches back to dev branch

These tests ensure the full workflow executes correctly, especially
that the branch switch succeeds (which requires all changes to be committed).
"""

import os
import time

import pytest
import typer
from git import Repo

from src.commands.spec import activate, complete, new
from src.commands.task import complete as task_complete
from src.commands.task import new as task_new
from src.utils import specs


@pytest.fixture
def initialized_mem(setup_test_env, monkeypatch):
    """Initialize mem directory structure and return the repo path."""
    repo_path = setup_test_env
    monkeypatch.chdir(repo_path)

    # Create .mem directory structure
    (repo_path / ".mem").mkdir(exist_ok=True)
    (repo_path / ".mem" / "specs").mkdir(exist_ok=True)
    (repo_path / ".mem" / "specs" / "completed").mkdir(exist_ok=True)
    (repo_path / ".mem" / "specs" / "abandoned").mkdir(exist_ok=True)
    (repo_path / ".mem" / "todos").mkdir(exist_ok=True)
    (repo_path / ".mem" / "logs").mkdir(exist_ok=True)

    return repo_path


def test_spec_complete_switches_to_dev_cleanly(initialized_mem, github_client):
    """
    Test that spec complete can switch to dev branch without uncommitted changes error.

    This was a bug where the status update happened AFTER the push but BEFORE
    the branch switch, leaving uncommitted changes that prevented the switch.
    """
    repo_path = initialized_mem
    repo = Repo(repo_path)

    # Ensure dev branch exists
    if "dev" not in [h.name for h in repo.heads]:
        repo.create_head("dev")
    repo.git.checkout("dev")

    # Create a spec
    spec_title = f"Complete Test {os.getpid()}"
    try:
        new(title=spec_title)
    except typer.Exit:
        pass

    spec_slug = f"complete_test_{os.getpid()}"

    # Activate the spec
    try:
        activate(spec_slug=spec_slug)
    except typer.Exit:
        pass

    # Verify we're on the spec branch
    assert repo.active_branch.name != "dev"
    spec_branch = repo.active_branch.name

    # Complete the spec (no tasks, so it should succeed)
    try:
        complete(spec_slug=spec_slug, message="Test completion")
    except typer.Exit:
        pass

    # Key assertion: we should be back on dev branch
    assert repo.active_branch.name == "dev", (
        "Should have switched to dev branch after completion. "
        "If this fails, there may be uncommitted changes blocking the switch."
    )

    # Verify no uncommitted changes on dev
    assert not repo.is_dirty(), "There should be no uncommitted changes on dev branch"

    # Switch back to spec branch to verify status (spec file only exists on spec branch until merged)
    repo.git.checkout(spec_branch)
    spec = specs.get_spec(spec_slug)
    assert spec is not None
    assert spec["status"] == "merge_ready"


def test_spec_complete_with_tasks(initialized_mem, github_client):
    """
    Test completing a spec that has tasks (all completed).
    """
    repo_path = initialized_mem
    repo = Repo(repo_path)

    # Ensure dev branch exists
    if "dev" not in [h.name for h in repo.heads]:
        repo.create_head("dev")
    repo.git.checkout("dev")

    # Create a spec
    spec_title = f"Complete With Tasks {os.getpid()}"
    try:
        new(title=spec_title)
    except typer.Exit:
        pass

    spec_slug = f"complete_with_tasks_{os.getpid()}"

    # Activate the spec
    try:
        activate(spec_slug=spec_slug)
    except typer.Exit:
        pass

    # Create and complete a task
    try:
        task_new(title="Test Task", description="A test task", spec_slug=spec_slug)
    except typer.Exit:
        pass

    try:
        task_complete(title="Test Task", notes="Done", spec_slug=spec_slug)
    except typer.Exit:
        pass

    # Remember spec branch before completing
    spec_branch = repo.active_branch.name

    # Complete the spec
    try:
        complete(spec_slug=spec_slug, message="Completed with tasks")
    except typer.Exit:
        pass

    # Should be back on dev
    assert repo.active_branch.name == "dev"

    # Switch back to spec branch to verify status (spec file only exists on spec branch until merged)
    repo.git.checkout(spec_branch)
    spec = specs.get_spec(spec_slug)
    assert spec is not None
    assert spec["status"] == "merge_ready"


def test_spec_complete_fails_with_incomplete_tasks(initialized_mem, github_client):
    """
    Test that completing a spec with incomplete tasks fails.
    """
    repo_path = initialized_mem
    repo = Repo(repo_path)

    # Ensure dev branch exists
    if "dev" not in [h.name for h in repo.heads]:
        repo.create_head("dev")
    repo.git.checkout("dev")

    # Create a spec
    spec_title = f"Incomplete Tasks Test {os.getpid()}"
    try:
        new(title=spec_title)
    except typer.Exit:
        pass

    spec_slug = f"incomplete_tasks_test_{os.getpid()}"

    # Activate the spec
    try:
        activate(spec_slug=spec_slug)
    except typer.Exit:
        pass

    spec_branch = repo.active_branch.name

    # Create a task but don't complete it
    try:
        task_new(title="Incomplete Task", description="Not done", spec_slug=spec_slug)
    except typer.Exit:
        pass

    # Try to complete the spec - should fail
    with pytest.raises(typer.Exit) as excinfo:
        complete(spec_slug=spec_slug, message="Should fail")

    assert excinfo.value.exit_code == 1

    # Should still be on spec branch (not switched)
    assert repo.active_branch.name == spec_branch

    # Spec should still be todo
    spec = specs.get_spec(spec_slug)
    assert spec is not None
    assert spec["status"] == "todo"


def test_spec_complete_fails_when_not_on_spec_branch(initialized_mem, github_client):
    """
    Test that completing a spec fails if not on the spec's branch.
    """
    repo_path = initialized_mem
    repo = Repo(repo_path)

    # Ensure dev branch exists
    if "dev" not in [h.name for h in repo.heads]:
        repo.create_head("dev")
    repo.git.checkout("dev")

    # Create a spec but don't activate it
    spec_title = f"Not Active Test {os.getpid()}"
    try:
        new(title=spec_title)
    except typer.Exit:
        pass

    spec_slug = f"not_active_test_{os.getpid()}"

    # Try to complete without activating - should fail
    with pytest.raises(typer.Exit) as excinfo:
        complete(spec_slug=spec_slug, message="Should fail")

    assert excinfo.value.exit_code == 1

    # Spec should still be todo
    spec = specs.get_spec(spec_slug)
    assert spec is not None
    assert spec["status"] == "todo"


def test_spec_complete_creates_pr_with_github_issue(initialized_mem, github_client):
    """
    Test that completing a spec with a linked GitHub issue creates a PR.
    """
    repo_path = initialized_mem
    repo = Repo(repo_path)

    from src.commands.sync import sync

    # Ensure dev branch exists and push it to remote
    if "dev" not in [h.name for h in repo.heads]:
        repo.create_head("dev")
    repo.git.checkout("dev")

    # Push dev branch to remote so PR can target it
    try:
        repo.git.push("origin", "dev", set_upstream=True)
    except Exception:
        pass  # May already exist

    # Create a spec
    spec_title = f"PR Test {os.getpid()}"
    try:
        new(title=spec_title)
    except typer.Exit:
        pass

    spec_slug = f"pr_test_{os.getpid()}"

    # Sync to create GitHub issue
    try:
        sync(dry_run=False)
    except typer.Exit:
        pass

    # Verify issue was created
    spec = specs.get_spec(spec_slug)
    assert spec is not None
    assert spec.get("issue_id") is not None

    # Activate the spec
    try:
        activate(spec_slug=spec_slug)
    except typer.Exit:
        pass

    spec_branch = repo.active_branch.name

    # Complete the spec
    try:
        complete(spec_slug=spec_slug, message="Test PR creation")
    except typer.Exit:
        pass

    # Wait for GitHub API
    time.sleep(2)

    # Should be back on dev
    assert repo.active_branch.name == "dev", (
        "Should have switched to dev branch after completion. "
        "If this fails, there may be uncommitted changes blocking the switch."
    )

    # Switch back to spec branch to verify PR was created (spec file only exists on spec branch until merged)
    repo.git.checkout(spec_branch)
    spec = specs.get_spec(spec_slug)
    assert spec is not None
    assert spec["status"] == "merge_ready"
    assert spec.get("pr_url") is not None


def test_spec_complete_commits_status_change(initialized_mem, github_client):
    """
    Test that the status change to merge_ready is included in the commit.

    This verifies the fix for the bug where status was updated after push,
    leaving uncommitted changes.
    """
    repo_path = initialized_mem
    repo = Repo(repo_path)

    # Ensure dev branch exists
    if "dev" not in [h.name for h in repo.heads]:
        repo.create_head("dev")
    repo.git.checkout("dev")

    # Create a spec
    spec_title = f"Status Commit Test {os.getpid()}"
    try:
        new(title=spec_title)
    except typer.Exit:
        pass

    spec_slug = f"status_commit_test_{os.getpid()}"

    # Activate the spec
    try:
        activate(spec_slug=spec_slug)
    except typer.Exit:
        pass

    spec_branch = repo.active_branch.name

    # Complete the spec
    try:
        complete(spec_slug=spec_slug, message="Status should be committed")
    except typer.Exit:
        pass

    # Switch back to spec branch to check the commit
    repo.git.checkout(spec_branch)

    # Verify no uncommitted changes on spec branch either
    assert not repo.is_dirty(), (
        "There should be no uncommitted changes on spec branch. "
        "The status update should have been committed."
    )

    # Verify the spec file on this branch has merge_ready status
    spec = specs.get_spec(spec_slug)
    assert spec is not None
    assert spec["status"] == "merge_ready"
