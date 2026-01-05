"""
Tests for sync PR merge detection functionality.
"""

import os
import time

import pytest
from git import Repo

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


def test_sync_plan_detects_merged_prs(initialized_mem, github_client):
    """
    Test that build_sync_plan correctly identifies merge_ready specs with merged PRs.
    """
    from src.commands.sync import build_sync_plan
    from src.utils.github.repo import get_repo_from_git

    repo_path = initialized_mem

    # Get GitHub repo
    owner, name = get_repo_from_git(repo_path)
    gh_repo = github_client.get_repo(f"{owner}/{name}")

    # Create a local repo for making branches
    local_repo = Repo(repo_path)

    # Create a feature branch and push it with unique names
    branch_name = f"test-feature-branch-{os.getpid()}"
    test_branch = local_repo.create_head(branch_name)
    test_branch.checkout()

    # Make a change with unique filename
    test_file = repo_path / f"feature_{os.getpid()}.txt"
    test_file.write_text(f"Feature content {os.getpid()}")
    local_repo.git.add(A=True)
    local_repo.git.commit("-m", f"Add feature {os.getpid()}")
    local_repo.git.push("origin", branch_name)

    time.sleep(2)

    # Create and merge a PR
    pr = gh_repo.create_pull(
        title=f"Test Feature PR {os.getpid()}",
        body="Test body",
        head=branch_name,
        base="main",
    )
    time.sleep(2)
    pr.merge(merge_method="squash")
    time.sleep(2)

    # Create a spec that simulates being merge_ready with this PR
    specs.create_spec(f"Merged Feature {os.getpid()}")
    spec_slug = f"merged_feature_{os.getpid()}"
    specs.update_spec(
        spec_slug,
        status="merge_ready",
        pr_url=pr.html_url,
        issue_id=None,  # No issue linked for this test
    )

    # Build sync plan
    local_specs = specs.get_all_specs()
    github_issues = []  # No issues for this test

    plan = build_sync_plan(gh_repo, local_specs, github_issues)

    # Verify the spec is in specs_to_complete
    assert len(plan.specs_to_complete) >= 1
    assert any(s["slug"] == spec_slug for s in plan.specs_to_complete)


def test_sync_plan_ignores_unmerged_prs(initialized_mem, github_client):
    """
    Test that build_sync_plan does NOT include specs with unmerged PRs.
    """
    from src.commands.sync import build_sync_plan
    from src.utils.github.repo import get_repo_from_git

    repo_path = initialized_mem

    # Get GitHub repo
    owner, name = get_repo_from_git(repo_path)
    gh_repo = github_client.get_repo(f"{owner}/{name}")

    # Create a local repo for making branches
    local_repo = Repo(repo_path)

    # Create a feature branch and push it with unique names
    branch_name = f"unmerged-feature-{os.getpid()}"
    test_branch = local_repo.create_head(branch_name)
    test_branch.checkout()

    # Make a change with unique filename
    test_file = repo_path / f"unmerged_feature_{os.getpid()}.txt"
    test_file.write_text(f"Unmerged content {os.getpid()}")
    local_repo.git.add(A=True)
    local_repo.git.commit("-m", f"Add unmerged feature {os.getpid()}")
    local_repo.git.push("origin", branch_name)

    time.sleep(2)

    # Create a PR but DON'T merge it
    pr = gh_repo.create_pull(
        title=f"Unmerged Feature PR {os.getpid()}",
        body="Test body",
        head=branch_name,
        base="main",
    )
    time.sleep(2)

    # Create a spec that simulates being merge_ready with this PR
    specs.create_spec(f"Unmerged Feature {os.getpid()}")
    spec_slug = f"unmerged_feature_{os.getpid()}"
    specs.update_spec(
        spec_slug, status="merge_ready", pr_url=pr.html_url, issue_id=None
    )

    # Build sync plan
    local_specs = specs.get_all_specs()
    github_issues = []

    plan = build_sync_plan(gh_repo, local_specs, github_issues)

    # Verify the spec is NOT in specs_to_complete
    assert not any(s["slug"] == spec_slug for s in plan.specs_to_complete)


def test_sync_plan_ignores_non_merge_ready_specs(initialized_mem, github_client):
    """
    Test that build_sync_plan ignores specs that aren't merge_ready.
    """
    from src.commands.sync import build_sync_plan
    from src.utils.github.repo import get_repo_from_git

    repo_path = initialized_mem

    owner, name = get_repo_from_git(repo_path)
    gh_repo = github_client.get_repo(f"{owner}/{name}")

    # Create specs with todo status (the default, not merge_ready)
    specs.create_spec(f"Todo Spec One {os.getpid()}")
    specs.create_spec(f"Todo Spec Two {os.getpid()}")

    # Build sync plan
    local_specs = specs.get_all_specs()
    github_issues = []

    plan = build_sync_plan(gh_repo, local_specs, github_issues)

    # Verify no todo specs are in specs_to_complete
    todo_slugs = [f"todo_spec_one_{os.getpid()}", f"todo_spec_two_{os.getpid()}"]
    for slug in todo_slugs:
        assert not any(s["slug"] == slug for s in plan.specs_to_complete)


def test_sync_execution_moves_merged_spec_to_completed(initialized_mem, github_client):
    """
    Test that execute_sync_plan actually moves specs to completed/.
    """
    from src.commands.sync import build_sync_plan, execute_sync_plan
    from src.utils.github.repo import get_repo_from_git

    repo_path = initialized_mem

    owner, name = get_repo_from_git(repo_path)
    gh_repo = github_client.get_repo(f"{owner}/{name}")

    local_repo = Repo(repo_path)

    # Create and merge a PR with unique names
    branch_name = f"completed-feature-{os.getpid()}"
    test_branch = local_repo.create_head(branch_name)
    test_branch.checkout()

    test_file = repo_path / f"completed_{os.getpid()}.txt"
    test_file.write_text(f"Completed content {os.getpid()}")
    local_repo.git.add(A=True)
    local_repo.git.commit("-m", f"Add completed feature {os.getpid()}")
    local_repo.git.push("origin", branch_name)

    time.sleep(2)

    pr = gh_repo.create_pull(
        title=f"Completed Feature {os.getpid()}",
        body="Test",
        head=branch_name,
        base="main",
    )
    time.sleep(2)
    pr.merge(merge_method="squash")
    time.sleep(2)

    # Create a merge_ready spec
    specs.create_spec(f"Completed Feature {os.getpid()}")
    spec_slug = f"completed_feature_{os.getpid()}"
    specs.update_spec(
        spec_slug, status="merge_ready", pr_url=pr.html_url, issue_id=None
    )

    # Verify spec is in root initially
    all_specs = specs.list_specs()
    assert any(s["slug"] == spec_slug for s in all_specs)

    # Build and execute sync plan
    local_specs = specs.get_all_specs()
    plan = build_sync_plan(gh_repo, local_specs, [])

    assert any(s["slug"] == spec_slug for s in plan.specs_to_complete)

    execute_sync_plan(plan, gh_repo)

    # Verify spec is now in completed
    all_specs = specs.list_specs()  # Default excludes completed
    assert not any(s["slug"] == spec_slug for s in all_specs)

    completed_specs = specs.list_specs(status="completed")
    assert any(s["slug"] == spec_slug for s in completed_specs)

    # Verify spec status is completed
    spec = specs.get_spec(spec_slug)
    assert spec is not None
    assert spec["status"] == "completed"
