"""
Tests for the spec complete command.

With the worktree-based workflow, complete:
1. Validates all tasks are completed
2. Updates spec status to merge_ready
3. Commits and pushes all changes
4. Creates a Pull Request on GitHub
5. Stays in the worktree (cleanup happens later via mem merge)
"""

import os
import time

import pytest
import typer
from git import Repo

from src.commands.spec import assign, complete, new
from src.commands.sync import sync
from src.commands.task import complete as task_complete
from src.commands.task import new as task_new
from src.utils import specs, worktrees


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


def test_spec_complete_updates_status(initialized_mem, github_client):
    """
    Test that spec complete updates status to merge_ready.
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

    # Sync to GitHub (required for assign)
    try:
        sync(dry_run=False)
    except typer.Exit:
        pass

    # Assign the spec (creates worktree and branch)
    try:
        assign(spec_slug=spec_slug)
    except typer.Exit:
        pass

    # Work in the worktree (where the branch is checked out)
    worktree_info = worktrees.get_worktree_for_spec(repo_path, spec_slug)
    assert worktree_info is not None
    os.chdir(worktree_info.path)

    # Complete the spec (no tasks, so it should succeed with --no-log)
    try:
        complete(spec_slug=spec_slug, message="Test completion", no_log=True)
    except typer.Exit:
        pass

    # Verify status is merge_ready
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

    # Sync and assign
    try:
        sync(dry_run=False)
    except typer.Exit:
        pass

    try:
        assign(spec_slug=spec_slug)
    except typer.Exit:
        pass

    # Work in the worktree
    worktree_info = worktrees.get_worktree_for_spec(repo_path, spec_slug)
    assert worktree_info is not None
    os.chdir(worktree_info.path)

    # Create and complete a task
    try:
        task_new(title="Test Task", description="A test task", spec_slug=spec_slug)
    except typer.Exit:
        pass

    try:
        task_complete(title="Test Task", notes="Done", spec_slug=spec_slug)
    except typer.Exit:
        pass

    # Complete the spec
    try:
        complete(spec_slug=spec_slug, message="Completed with tasks", no_log=True)
    except typer.Exit:
        pass

    # Verify status
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

    # Sync and assign
    try:
        sync(dry_run=False)
    except typer.Exit:
        pass

    try:
        assign(spec_slug=spec_slug)
    except typer.Exit:
        pass

    # Work in the worktree
    worktree_info = worktrees.get_worktree_for_spec(repo_path, spec_slug)
    assert worktree_info is not None
    os.chdir(worktree_info.path)

    # Create a task but don't complete it
    try:
        task_new(title="Incomplete Task", description="Not done", spec_slug=spec_slug)
    except typer.Exit:
        pass

    # Try to complete the spec - should fail
    with pytest.raises(typer.Exit) as excinfo:
        complete(spec_slug=spec_slug, message="Should fail", no_log=True)

    assert excinfo.value.exit_code == 1

    # Spec should still be todo
    spec = specs.get_spec(spec_slug)
    assert spec is not None
    assert spec["status"] == "todo"


def test_spec_complete_fails_when_spec_not_active(initialized_mem, github_client):
    """
    Test that completing a spec fails if the spec is not active.
    """
    repo_path = initialized_mem
    repo = Repo(repo_path)

    # Ensure dev branch exists
    if "dev" not in [h.name for h in repo.heads]:
        repo.create_head("dev")
    repo.git.checkout("dev")

    # Create a spec but don't assign it
    spec_title = f"Not Active Test {os.getpid()}"
    try:
        new(title=spec_title)
    except typer.Exit:
        pass

    spec_slug = f"not_active_test_{os.getpid()}"

    # Try to complete without assigning - should fail
    with pytest.raises(typer.Exit) as excinfo:
        complete(spec_slug=spec_slug, message="Should fail", no_log=True)

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

    # Assign the spec
    try:
        assign(spec_slug=spec_slug)
    except typer.Exit:
        pass

    # Work in the worktree
    worktree_info = worktrees.get_worktree_for_spec(repo_path, spec_slug)
    assert worktree_info is not None
    os.chdir(worktree_info.path)

    # Complete the spec
    try:
        complete(spec_slug=spec_slug, message="Test PR creation", no_log=True)
    except typer.Exit:
        pass

    # Wait for GitHub API
    time.sleep(2)

    # Verify PR was created
    spec = specs.get_spec(spec_slug)
    assert spec is not None
    assert spec["status"] == "merge_ready"
    assert spec.get("pr_url") is not None
