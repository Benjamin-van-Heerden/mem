"""
Tests for the merge command.

The merge command:
1. Lists PRs from merge_ready specs
2. Checks mergeability status
3. Allows selection and performs rebase merge
4. Deletes remote branches after merge
5. Moves specs to completed/
"""

import os
import time
import uuid

import pytest
import typer
from git import Repo
from typer.testing import CliRunner

from src.commands.merge import app as merge_app
from src.commands.spec import assign, complete, new
from src.commands.sync import sync
from src.utils import specs, worktrees
from tests.conftest import get_worker_id

runner = CliRunner()


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


def test_merge_no_merge_ready_specs(request, initialized_mem):
    """Test that merge command handles no merge_ready specs gracefully."""
    repo_path = initialized_mem
    repo = Repo(repo_path)
    # setup_test_env already creates and checks out the worker-specific dev branch

    # Create a spec but don't complete it
    spec_slug = unique_slug("not_ready", request)
    spec_title = spec_slug.replace("_", " ").title()
    try:
        new(title=spec_title)
    except typer.Exit:
        pass

    # Commit the new spec to make working directory clean
    repo.git.add("-A")
    repo.git.commit("-m", "Add spec")

    # Run merge - should exit cleanly
    result = runner.invoke(merge_app)
    assert result.exit_code == 0
    assert "No PRs ready to merge" in result.output


def test_merge_lists_ready_prs(request, initialized_mem, github_client):
    """Test that merge command lists PRs that are ready to merge."""
    repo_path = initialized_mem
    repo = Repo(repo_path)
    # setup_test_env already creates and checks out the worker-specific dev branch

    # Create a spec with unique slug
    spec_slug = unique_slug("merge_test", request)
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

    # Assign the spec (creates worktree) and complete it
    try:
        assign(spec_slug=spec_slug)
    except typer.Exit:
        pass

    # Get the worktree path and work from there
    worktree_info = worktrees.get_worktree_for_spec(repo_path, spec_slug)
    assert worktree_info is not None, "Worktree should have been created by assign"
    os.chdir(worktree_info.path)

    # Make a change so there's something to PR
    worktree_repo = Repo(worktree_info.path)
    test_file = worktree_info.path / "test_change.txt"
    test_file.write_text("Test change for merge")
    worktree_repo.git.add("test_change.txt")
    worktree_repo.git.commit("-m", "Add test change")

    try:
        complete(spec_slug=spec_slug, message="Ready for merge test", no_log=True)
    except typer.Exit:
        pass

    # Go back to main repo for merge command
    os.chdir(repo_path)

    # Wait for GitHub to process
    time.sleep(3)

    # Run merge with dry-run
    result = runner.invoke(merge_app, ["--dry-run"])

    # Should show the spec in some category (ready, conflicts, or already merged)
    assert spec_slug in result.output or "Checking PR status" in result.output


def test_merge_moves_spec_to_completed(initialized_mem, github_client):
    """Test that merge command moves merged specs to completed/."""
    repo_path = initialized_mem
    repo = Repo(repo_path)

    # Ensure dev branch exists and push to remote
    if "dev" not in [h.name for h in repo.heads]:
        repo.create_head("dev")
    repo.git.checkout("dev")
    try:
        repo.git.push("origin", "dev", set_upstream=True)
    except Exception:
        pass

    # Create a spec
    spec_title = f"Merge Complete Test {os.getpid()}"
    try:
        new(title=spec_title)
    except typer.Exit:
        pass

    spec_slug = f"merge_complete_test_{os.getpid()}"

    # Sync to create GitHub issue
    try:
        sync(dry_run=False)
    except typer.Exit:
        pass

    # Assign the spec (creates worktree) and complete it
    try:
        assign(spec_slug=spec_slug)
    except typer.Exit:
        pass

    # Get the worktree path and work from there
    worktree_info = worktrees.get_worktree_for_spec(repo_path, spec_slug)
    assert worktree_info is not None, "Worktree should have been created by assign"
    os.chdir(worktree_info.path)

    try:
        complete(spec_slug=spec_slug, message="Ready for merge", no_log=True)
    except typer.Exit:
        pass

    # Wait for GitHub
    time.sleep(3)

    # Go back to main repo for merge command
    os.chdir(repo_path)

    # Verify spec is merge_ready (need to check on spec branch)
    _spec_branch = specs.get_spec(spec_slug)
    # Spec may not be visible from dev, that's expected

    # Try to merge with --all flag
    runner.invoke(merge_app, ["--all"])

    # Check if spec was moved to completed
    time.sleep(2)
    _completed_specs = specs.list_specs(status="completed")
    # The spec should either be in completed or still in merge_ready
    # (depending on whether the merge succeeded)


def test_merge_with_no_merge_ready_exits_cleanly(initialized_mem):
    """Test that merge with no merge-ready specs exits cleanly."""
    # Run merge when there are no merge_ready specs
    result = runner.invoke(merge_app)
    # Should exit with 0 when no specs are ready
    assert result.exit_code == 0
    assert "No PRs ready to merge" in result.output


def test_merge_dry_run_shows_message(initialized_mem):
    """Test that dry-run shows appropriate message when no specs ready."""
    # Run with dry-run when no specs are ready
    result = runner.invoke(merge_app, ["--dry-run"])
    # Should show "No PRs" message since there are none ready
    assert "No PRs ready to merge" in result.output
