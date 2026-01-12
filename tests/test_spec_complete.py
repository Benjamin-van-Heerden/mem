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
import uuid

import pytest
import typer
from git import Repo

from src.commands.spec import assign, complete, new
from src.commands.sync import sync
from src.commands.task import complete as task_complete
from src.commands.task import new as task_new
from src.utils import specs, worktrees
from tests.conftest import get_worker_id


def unique_slug(base: str, request) -> str:
    """Generate a unique spec slug using worker ID and UUID."""
    worker_id = get_worker_id(request)
    short_uuid = uuid.uuid4().hex[:6]
    return f"{base}_{worker_id}_{short_uuid}"


@pytest.fixture
def initialized_mem(request, setup_test_env, monkeypatch):
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


def test_spec_complete_updates_status(request, initialized_mem, github_client):
    """
    Test that spec complete updates status to merge_ready.
    """
    repo_path = initialized_mem
    # setup_test_env already creates and checks out the worker-specific dev branch

    # Create a spec with unique slug
    spec_slug = unique_slug("complete_test", request)
    spec_title = spec_slug.replace("_", " ").title()
    try:
        new(title=spec_title)
    except typer.Exit:
        pass

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


def test_spec_complete_with_tasks(request, initialized_mem, github_client):
    """
    Test completing a spec that has tasks (all completed).
    """
    repo_path = initialized_mem
    # setup_test_env already creates and checks out the worker-specific dev branch

    # Create a spec with unique slug
    spec_slug = unique_slug("complete_with_tasks", request)
    spec_title = spec_slug.replace("_", " ").title()
    try:
        new(title=spec_title)
    except typer.Exit:
        pass

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


def test_spec_complete_fails_with_incomplete_tasks(
    request, initialized_mem, github_client
):
    """
    Test that completing a spec with incomplete tasks fails.
    """
    repo_path = initialized_mem
    # setup_test_env already creates and checks out the worker-specific dev branch

    # Create a spec with unique slug
    spec_slug = unique_slug("incomplete_tasks", request)
    spec_title = spec_slug.replace("_", " ").title()
    try:
        new(title=spec_title)
    except typer.Exit:
        pass

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


def test_spec_complete_fails_when_spec_not_active(
    request, initialized_mem, github_client
):
    """
    Test that completing a spec fails if the spec is not active.
    """
    repo_path = initialized_mem
    # setup_test_env already creates and checks out the worker-specific dev branch

    # Create a spec but don't assign it
    spec_slug = unique_slug("not_active", request)
    spec_title = spec_slug.replace("_", " ").title()
    try:
        new(title=spec_title)
    except typer.Exit:
        pass

    # Try to complete without assigning - should fail
    with pytest.raises(typer.Exit) as excinfo:
        complete(spec_slug=spec_slug, message="Should fail", no_log=True)

    assert excinfo.value.exit_code == 1

    # Spec should still be todo
    spec = specs.get_spec(spec_slug)
    assert spec is not None
    assert spec["status"] == "todo"


def test_spec_complete_creates_pr_with_github_issue(
    request, initialized_mem, github_client
):
    """
    Test that completing a spec with a linked GitHub issue creates a PR.
    """
    repo_path = initialized_mem
    repo = Repo(repo_path)
    # setup_test_env already creates and checks out the worker-specific dev branch
    # and pushes it to origin

    # Create a spec with unique slug
    spec_slug = unique_slug("pr_test", request)
    spec_title = spec_slug.replace("_", " ").title()
    try:
        new(title=spec_title)
    except typer.Exit:
        pass

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

    # Make a change so there's something to PR
    worktree_repo = Repo(worktree_info.path)
    test_file = worktree_info.path / "test_change.txt"
    test_file.write_text("Test change for PR")
    worktree_repo.git.add("test_change.txt")
    worktree_repo.git.commit("-m", "Add test change")

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
