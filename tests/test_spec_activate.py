"""
Tests for spec activation workflow.

With the new branch-based activation:
- A spec is "active" when you're on its branch
- No assignment required to activate
- No "active"/"inactive" status - activation is derived from git branch
"""

import pytest
import typer
from git import Repo

from src.commands.spec import activate, deactivate, new
from src.utils import specs
from src.utils.github.client import get_authenticated_user


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


def test_spec_activation_creates_branch(initialized_mem, github_client):
    """
    Test the activation workflow:
    1. Create a spec
    2. Activate the spec (creates and switches to branch)
    3. Verify git branch is created correctly
    4. Verify spec has branch recorded
    """
    repo_path = initialized_mem

    # Get current authenticated user info
    user_info = get_authenticated_user(github_client)
    username = user_info["username"]

    # Create a new spec
    spec_title = "Test Specification"
    try:
        new(title=spec_title)
    except typer.Exit:
        pass  # Expected exit after success

    spec_slug = "test_specification"

    # Verify creation
    spec = specs.get_spec(spec_slug)
    assert spec is not None
    assert spec["title"] == spec_title
    assert spec["status"] == "todo"

    # Ensure 'dev' branch exists as base
    repo = Repo(repo_path)
    if "dev" not in [h.name for h in repo.heads]:
        repo.create_head("dev")
    repo.git.checkout("dev")

    # Activate the spec
    try:
        activate(spec_slug=spec_slug)
    except typer.Exit:
        pass

    # Verify Branch Creation
    user_slug = username.lower().replace(" ", "_").replace("-", "_")
    expected_branch = f"dev-{user_slug}-{spec_slug}"

    assert repo.active_branch.name == expected_branch

    # Verify Spec has branch recorded
    spec = specs.get_spec(spec_slug)
    assert spec is not None
    assert spec["branch"] == expected_branch
    # Status should still be "todo" - active is now derived from branch
    assert spec["status"] == "todo"

    # Verify get_active_spec returns this spec
    active = specs.get_active_spec()
    assert active is not None
    assert active["slug"] == spec_slug


def test_deactivate_switches_to_dev(initialized_mem, github_client):
    """
    Verify that deactivation switches back to dev branch.
    """
    repo_path = initialized_mem

    # Create and activate a spec
    try:
        new(title="Deactivate Test")
    except typer.Exit:
        pass

    repo = Repo(repo_path)
    if "dev" not in [h.name for h in repo.heads]:
        repo.create_head("dev")
    repo.git.checkout("dev")

    try:
        activate(spec_slug="deactivate_test")
    except typer.Exit:
        pass

    # Verify we're on the spec branch
    assert repo.active_branch.name != "dev"

    # Deactivate
    try:
        deactivate()
    except typer.Exit:
        pass

    # Verify we're back on dev
    assert repo.active_branch.name == "dev"

    # Verify no active spec on dev branch
    active = specs.get_active_spec()
    assert active is None


def test_activate_multiple_specs(initialized_mem, github_client):
    """
    Verify that activating a second spec just switches to its branch.
    (No longer errors - we just switch branches)
    """
    repo_path = initialized_mem

    # Create two specs
    try:
        new(title="Spec One")
        new(title="Spec Two")
    except typer.Exit:
        pass

    # Ensure 'dev' exists
    repo = Repo(repo_path)
    if "dev" not in [h.name for h in repo.heads]:
        repo.create_head("dev")
    repo.git.checkout("dev")

    # Activate first spec
    try:
        activate(spec_slug="spec_one")
    except typer.Exit:
        pass

    spec_one = specs.get_spec("spec_one")
    assert spec_one is not None
    branch_one = spec_one["branch"]
    assert repo.active_branch.name == branch_one

    # Activate second spec - should switch to its branch
    try:
        activate(spec_slug="spec_two")
    except typer.Exit:
        pass

    spec_two = specs.get_spec("spec_two")
    assert spec_two is not None
    branch_two = spec_two["branch"]
    assert repo.active_branch.name == branch_two

    # Verify active spec is now spec_two
    active = specs.get_active_spec()
    assert active is not None
    assert active["slug"] == "spec_two"


def test_activate_nonexistent_spec(initialized_mem):
    """
    Verify that activation fails for non-existent spec.
    """
    # Attempt to activate non-existent spec
    with pytest.raises(typer.Exit) as excinfo:
        activate(spec_slug="does_not_exist")

    assert excinfo.value.exit_code == 1


def test_branch_based_active_detection(initialized_mem, github_client):
    """
    Verify that get_active_spec correctly detects spec from current branch.
    """
    repo_path = initialized_mem

    # Create a spec and activate it
    try:
        new(title="Branch Detection Test")
    except typer.Exit:
        pass

    repo = Repo(repo_path)
    if "dev" not in [h.name for h in repo.heads]:
        repo.create_head("dev")
    repo.git.checkout("dev")

    # On dev branch, no active spec
    assert specs.get_active_spec() is None

    # Activate spec
    try:
        activate(spec_slug="branch_detection_test")
    except typer.Exit:
        pass

    # Now active spec should be detected
    active = specs.get_active_spec()
    assert active is not None
    assert active["slug"] == "branch_detection_test"

    # Switch back to dev manually
    repo.git.checkout("dev")

    # No active spec again
    assert specs.get_active_spec() is None
