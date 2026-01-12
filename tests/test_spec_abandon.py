"""
Tests for the spec abandon command.
"""

import time

import pytest
import typer
from git import Repo

from src.commands.spec import abandon, assign, new
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
    (repo_path / ".mem" / "specs" / "completed").mkdir(exist_ok=True)
    (repo_path / ".mem" / "specs" / "abandoned").mkdir(exist_ok=True)
    (repo_path / ".mem" / "todos").mkdir(exist_ok=True)
    (repo_path / ".mem" / "logs").mkdir(exist_ok=True)

    return repo_path


def test_abandon_spec_moves_to_abandoned(initialized_mem):
    """Test that abandoning a spec moves it to the abandoned directory."""
    # Create a spec
    try:
        new(title="Abandon Test Spec")
    except typer.Exit:
        pass

    spec_slug = "abandon_test_spec"

    # Verify it exists in root
    spec = specs.get_spec(spec_slug)
    assert spec is not None
    assert spec["status"] == "todo"

    # Abandon the spec
    try:
        abandon(spec_slug=spec_slug, reason="No longer needed")
    except typer.Exit:
        pass

    # Verify it's now in abandoned
    spec = specs.get_spec(spec_slug)
    assert spec is not None
    assert spec["status"] == "abandoned"

    # Verify it's not in the default list
    all_specs = specs.list_specs()
    assert not any(s["slug"] == spec_slug for s in all_specs)

    # Verify it's in the abandoned list
    abandoned_specs = specs.list_specs(status="abandoned")
    assert any(s["slug"] == spec_slug for s in abandoned_specs)


def test_abandon_assigned_spec(initialized_mem, github_client):
    """Test that abandoning an assigned spec with worktree works."""
    repo_path = initialized_mem

    # Create a spec
    try:
        new(title="Assigned Abandon Test")
    except typer.Exit:
        pass

    # Ensure dev branch exists
    repo = Repo(repo_path)
    if "dev" not in [h.name for h in repo.heads]:
        repo.create_head("dev")
    repo.git.checkout("dev")

    # Sync to GitHub (required for assign)
    try:
        sync(dry_run=False)
    except typer.Exit:
        pass

    # Assign the spec (creates worktree)
    try:
        assign(spec_slug="assigned_abandon_test")
    except typer.Exit:
        pass

    # Verify worktree was created
    wt = worktrees.get_worktree_for_spec(repo_path, "assigned_abandon_test")
    assert wt is not None

    # Abandon the spec
    try:
        abandon(spec_slug="assigned_abandon_test", reason="Changed direction")
    except typer.Exit:
        pass

    # Verify spec is abandoned
    spec = specs.get_spec("assigned_abandon_test")
    assert spec is not None
    assert spec["status"] == "abandoned"


def test_abandon_spec_with_github_issue(initialized_mem, github_client):
    """Test that abandoning a spec with a linked GitHub issue closes the issue."""
    repo_path = initialized_mem

    from src.utils.github.repo import get_repo_from_git

    # Create a spec
    try:
        new(title="GitHub Abandon Test")
    except typer.Exit:
        pass

    spec_slug = "github_abandon_test"

    # Sync to create GitHub issue
    try:
        sync(dry_run=False)
    except typer.Exit:
        pass

    # Verify issue was created
    spec = specs.get_spec(spec_slug)
    assert spec is not None
    assert spec.get("issue_id") is not None
    issue_id = spec["issue_id"]

    # Abandon the spec
    try:
        abandon(spec_slug=spec_slug, reason="Testing abandon with GitHub")
    except typer.Exit:
        pass

    # Wait for GitHub API
    time.sleep(2)

    # Verify issue is closed
    owner, name = get_repo_from_git(repo_path)
    repo = github_client.get_repo(f"{owner}/{name}")
    issue = repo.get_issue(issue_id)
    assert issue.state == "closed"


def test_abandon_nonexistent_spec_fails(initialized_mem):
    """Test that abandoning a non-existent spec fails."""
    with pytest.raises(typer.Exit) as excinfo:
        abandon(spec_slug="nonexistent_spec", reason="Test")
    assert excinfo.value.exit_code == 1
