"""
Tests for spec assignment and worktree workflow.

With the worktree-based workflow:
- `mem spec assign` creates a worktree and branch
- Being in a worktree means the spec is active
- No activate/deactivate commands
"""

import pytest
import typer
from git import Repo

from src.commands.spec import assign, new
from src.commands.sync import sync
from src.utils import specs, worktrees


@pytest.fixture
def initialized_mem(setup_test_env, monkeypatch):
    """Initialize mem directory structure and return the repo path."""
    repo_path = setup_test_env
    monkeypatch.chdir(repo_path)

    # Create .mem directory structure
    (repo_path / ".mem").mkdir(exist_ok=True)
    (repo_path / ".mem" / "specs").mkdir(exist_ok=True)
    (repo_path / ".mem" / "todos").mkdir(exist_ok=True)
    (repo_path / ".mem" / "logs").mkdir(exist_ok=True)

    return repo_path


def test_assign_creates_worktree_and_branch(initialized_mem, github_client):
    """
    Test that assign creates a worktree and branch:
    1. Create a spec
    2. Sync to GitHub (required for assign)
    3. Assign the spec
    4. Verify worktree created
    5. Verify branch created
    """
    repo_path = initialized_mem

    # Ensure dev branch exists
    repo = Repo(repo_path)
    if "dev" not in [h.name for h in repo.heads]:
        repo.create_head("dev")
    repo.git.checkout("dev")

    # Create a new spec
    spec_title = "Worktree Test Spec"
    try:
        new(title=spec_title)
    except typer.Exit:
        pass

    spec_slug = "worktree_test_spec"

    # Sync to create GitHub issue (required for assign)
    try:
        sync(dry_run=False)
    except typer.Exit:
        pass

    # Assign the spec
    try:
        assign(spec_slug=spec_slug)
    except typer.Exit:
        pass

    # Verify worktree created
    wt = worktrees.get_worktree_for_spec(repo_path, spec_slug)
    assert wt is not None
    assert wt.path.exists()

    # Verify branch exists
    spec = specs.get_spec(spec_slug)
    assert spec is not None
    assert spec.get("branch") is not None
    assert spec["branch"] in [h.name for h in repo.heads]


def test_assign_already_assigned_shows_worktree_path(
    initialized_mem, github_client, capsys
):
    """
    Test that assigning an already-assigned spec shows the worktree path.
    """
    repo_path = initialized_mem

    repo = Repo(repo_path)
    if "dev" not in [h.name for h in repo.heads]:
        repo.create_head("dev")
    repo.git.checkout("dev")

    # Create and sync spec
    try:
        new(title="Double Assign Test")
    except typer.Exit:
        pass

    try:
        sync(dry_run=False)
    except typer.Exit:
        pass

    # First assign
    try:
        assign(spec_slug="double_assign_test")
    except typer.Exit:
        pass

    # Second assign should show existing worktree
    try:
        assign(spec_slug="double_assign_test")
    except typer.Exit:
        pass

    captured = capsys.readouterr()
    assert "already has a worktree" in captured.out


def test_assign_nonexistent_spec_fails(initialized_mem):
    """
    Test that assigning a non-existent spec fails.
    """
    with pytest.raises(typer.Exit) as excinfo:
        assign(spec_slug="does_not_exist")

    assert excinfo.value.exit_code == 1


def test_worktree_detection(initialized_mem, github_client):
    """
    Test that worktree detection works correctly.
    """
    repo_path = initialized_mem

    repo = Repo(repo_path)
    if "dev" not in [h.name for h in repo.heads]:
        repo.create_head("dev")
    repo.git.checkout("dev")

    # Main repo is not a worktree
    assert not worktrees.is_worktree(repo_path)

    # Create spec and assign
    try:
        new(title="Detection Test")
    except typer.Exit:
        pass

    try:
        sync(dry_run=False)
    except typer.Exit:
        pass

    try:
        assign(spec_slug="detection_test")
    except typer.Exit:
        pass

    # Worktree should be detected
    wt_path = worktrees.get_worktree_path(repo_path, "detection_test")
    assert worktrees.is_worktree(wt_path)

    # Main repo path can be resolved from worktree
    main_path = worktrees.get_main_repo_path(wt_path)
    assert main_path is not None
    assert main_path.resolve() == repo_path.resolve()


def test_list_worktrees(initialized_mem, github_client):
    """
    Test listing worktrees.
    """
    repo_path = initialized_mem

    repo = Repo(repo_path)
    if "dev" not in [h.name for h in repo.heads]:
        repo.create_head("dev")
    repo.git.checkout("dev")

    # Initially just the main repo
    wts = worktrees.list_worktrees(repo_path)
    assert len(wts) == 1
    assert wts[0].is_main

    # Create two specs and assign them
    for title in ["List Test One", "List Test Two"]:
        try:
            new(title=title)
        except typer.Exit:
            pass

    try:
        sync(dry_run=False)
    except typer.Exit:
        pass

    for slug in ["list_test_one", "list_test_two"]:
        try:
            assign(spec_slug=slug)
        except typer.Exit:
            pass

    # Should now have 3 worktrees (main + 2 specs)
    wts = worktrees.list_worktrees(repo_path)
    assert len(wts) == 3

    spec_wts = [wt for wt in wts if not wt.is_main]
    assert len(spec_wts) == 2
