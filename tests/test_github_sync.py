"""
Tests for GitHub sync functionality.
"""

import time

import pytest
import typer

from src.commands.spec import new
from src.commands.sync import sync
from src.utils import specs, todos


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


def test_spec_outbound_sync(initialized_mem, github_client, monkeypatch):
    """
    Test the outbound sync workflow:
    1. Create a local spec
    2. Run sync
    3. Verify GitHub issue is created
    4. Verify spec is updated with issue info
    """

    # 1. Create local spec
    spec_title = "Outbound Sync Test"
    try:
        new(title=spec_title)
    except typer.Exit:
        pass

    spec_slug = "outbound_sync_test"
    spec = specs.get_spec(spec_slug)
    assert spec is not None
    assert spec.get("issue_id") is None

    # 2. Run sync
    try:
        sync(dry_run=False)
    except typer.Exit:
        pass

    # 3. Verify spec is updated with issue info
    updated_spec = specs.get_spec(spec_slug)
    assert updated_spec is not None
    assert updated_spec.get("issue_id") is not None
    assert updated_spec.get("issue_url") is not None


def test_github_sync_inbound(initialized_mem, github_client, monkeypatch):
    """
    Test the inbound sync workflow:
    1. Create an issue on GitHub with mem-spec label
    2. Create a normal issue on GitHub
    3. Run mem sync
    4. Verify spec and todo are created locally
    """
    repo_path = initialized_mem

    # Setup: Create issues on the test repo
    from src.utils.github.repo import get_repo_from_git

    owner, name = get_repo_from_git(repo_path)
    repo = github_client.get_repo(f"{owner}/{name}")

    # Create a spec issue
    repo.create_issue(
        title="[Spec]: Remote Spec", body="Remote body content", labels=["mem-spec"]
    )

    # Create a normal issue
    repo.create_issue(title="Normal Todo", body="Todo body content")

    # Wait for GitHub to index the issues
    time.sleep(2)

    # Run sync
    try:
        sync(dry_run=False)
    except typer.Exit:
        pass

    # Verify Spec creation
    spec_list = specs.list_specs()
    assert any(s["title"] == "Remote Spec" for s in spec_list)

    # Verify Todo creation
    todo_list = todos.list_todos()
    assert any(t["title"] == "Normal Todo" for t in todo_list)

    # Verify file content for synced spec
    synced_spec = next(s for s in spec_list if s["title"] == "Remote Spec")
    assert "Remote body content" in synced_spec.get("body", "")


def test_sync_dry_run(initialized_mem, github_client, monkeypatch):
    """
    Test that dry-run doesn't modify anything.
    """

    # Create local spec
    try:
        new(title="Dry Run Test")
    except typer.Exit:
        pass

    spec_slug = "dry_run_test"
    spec = specs.get_spec(spec_slug)
    assert spec is not None
    assert spec.get("issue_id") is None

    # Run sync with dry-run
    try:
        sync(dry_run=True)
    except typer.Exit:
        pass

    # Verify spec is NOT updated
    spec = specs.get_spec(spec_slug)
    assert spec is not None
    assert spec.get("issue_id") is None
