import os
import shutil
import tempfile
import time
from pathlib import Path

import pytest
from dotenv import load_dotenv
from filelock import FileLock
from git import Repo
from github import Auth, Github, GithubException

# Load environment variables
load_dotenv()

# Dedicated test repo - keeps the main mem repo clean
TEST_REPO_NAME = "mem-test"


def get_worker_id(request) -> str:
    """Get the xdist worker id, or 'master' if not running with xdist."""
    if hasattr(request.config, "workerinput"):
        return request.config.workerinput["workerid"]
    return "master"


def get_worker_branch_suffix(request) -> str:
    """Get a suffix for branch names to isolate parallel test workers."""
    worker_id = get_worker_id(request)
    if worker_id == "master":
        return ""
    return f"-{worker_id}"


@pytest.fixture(scope="session")
def github_token():
    """Retrieve GITHUB_TOKEN from environment."""
    token = os.getenv("GITHUB_TOKEN")
    if not token:
        pytest.skip("GITHUB_TOKEN not set in environment or .env file")
    return token


@pytest.fixture(scope="session")
def github_client(github_token, tmp_path_factory):
    """
    Create an authenticated GitHub client.

    Nukes and recreates the test repo at the start of each test session
    for a clean slate. Uses a file lock to ensure only one worker does this
    when running with pytest-xdist.
    """
    auth = Auth.Token(github_token)
    client = Github(auth=auth)
    user = client.get_user()
    repo_full_name = f"{user.login}/{TEST_REPO_NAME}"

    # Use a file lock to ensure only one worker creates the repo
    root_tmp_dir = tmp_path_factory.getbasetemp().parent
    lock_file = root_tmp_dir / "mem_test_repo.lock"
    marker_file = root_tmp_dir / "mem_test_repo_created"

    with FileLock(str(lock_file)):
        if not marker_file.exists():
            # First worker - delete and recreate the repo
            try:
                old_repo = client.get_repo(repo_full_name)
                old_repo.delete()
                time.sleep(3)  # Give GitHub time to process deletion
            except GithubException:
                pass  # Repo doesn't exist, that's fine

            # Create fresh test repo
            user.create_repo(  # type: ignore
                TEST_REPO_NAME,
                description="Temporary repo for mem integration tests",
                private=False,
                auto_init=True,  # Creates initial commit so we can clone
            )
            time.sleep(3)  # Give GitHub time to initialize the repo

            # Mark that repo has been created
            marker_file.touch()

    yield client


@pytest.fixture(scope="function")
def setup_test_env(request, github_client, github_token, monkeypatch):
    """
    Sets up a test environment using a temp directory that clones the test repo.

    Uses a dedicated mem-test repo that gets nuked at session start.
    Each worker gets isolated branches to avoid push conflicts in xdist.

    Yields:
        Path: The path to the local test repository

    Teardown:
        - Deletes the local temp directory
        - Cleans up remote branches (best effort)
    """
    # Get worker-specific suffix for branch isolation
    suffix = get_worker_branch_suffix(request)
    dev_branch = f"dev{suffix}"

    # Create a temp directory
    base_dir = Path(tempfile.mkdtemp(prefix="mem_test_"))

    # Clone the test repo
    user = github_client.get_user()
    auth_url = (
        f"https://oauth2:{github_token}@github.com/{user.login}/{TEST_REPO_NAME}.git"
    )

    repo = Repo.clone_from(auth_url, base_dir)

    # Configure git user for the test repo
    with repo.config_writer() as config:
        config.set_value("user", "name", "Test User")
        config.set_value("user", "email", "test@example.com")

    # Create worker-isolated dev branch
    if dev_branch in [h.name for h in repo.heads]:
        repo.delete_head(dev_branch, force=True)
    repo.create_head(dev_branch)
    repo.heads[dev_branch].checkout()

    # Push the dev branch to origin (force to handle conflicts)
    try:
        repo.git.push("origin", dev_branch, set_upstream=True, force=True)
    except Exception:
        pass

    # Patch specs module to recognize our worker-specific dev branch as "dev"
    from src.utils import specs as specs_module

    original_ensure_on_dev = specs_module.ensure_on_dev_branch
    original_get_current = specs_module.get_current_branch

    def patched_ensure_on_dev_branch():
        """Treat worker-specific dev branch as dev."""
        current = specs_module.get_current_branch()
        if current is None:
            return False, None
        # If on our worker's dev branch, we're good
        if current == dev_branch:
            return False, None
        # If on main or test, switch to our dev branch
        if current in ("main", "test"):
            try:
                from git import Repo as GitRepo

                from env_settings import ENV_SETTINGS

                r = GitRepo(ENV_SETTINGS.caller_dir)
                r.git.checkout(dev_branch)
                return True, f"Switched from '{current}' to '{dev_branch}' branch"
            except Exception as e:
                return False, f"Failed to switch to {dev_branch}: {e}"
        return False, None

    monkeypatch.setattr(
        specs_module, "ensure_on_dev_branch", patched_ensure_on_dev_branch
    )

    # Also patch sync module's rebase target
    from src.commands import sync as sync_module

    original_git_fetch_and_pull = sync_module.git_fetch_and_pull

    def patched_git_fetch_and_pull():
        """Use worker-specific dev branch for rebase."""
        import subprocess

        from env_settings import ENV_SETTINGS

        cwd = ENV_SETTINGS.caller_dir

        try:
            subprocess.run(
                ["git", "fetch", "origin"],
                cwd=cwd,
                check=True,
                capture_output=True,
                text=True,
            )
        except subprocess.CalledProcessError as e:
            return False, f"git fetch failed: {e.stderr}"

        current_branch = sync_module.get_current_git_branch()

        if sync_module.is_feature_branch(current_branch):
            # Feature branch: rebase onto our worker's dev branch
            if sync_module.has_uncommitted_changes():
                return False, "UNCOMMITTED_CHANGES"

            try:
                result = subprocess.run(
                    ["git", "rebase", f"origin/{dev_branch}"],
                    cwd=cwd,
                    capture_output=True,
                    text=True,
                )
                if result.returncode != 0:
                    subprocess.run(
                        ["git", "rebase", "--abort"],
                        cwd=cwd,
                        capture_output=True,
                        text=True,
                    )
                    return False, "REBASE_FAILED"
            except subprocess.CalledProcessError:
                subprocess.run(
                    ["git", "rebase", "--abort"],
                    cwd=cwd,
                    capture_output=True,
                    text=True,
                )
                return False, "REBASE_FAILED"
        else:
            # Non-feature branch: pull with fast-forward
            try:
                result = subprocess.run(
                    ["git", "pull", "--ff-only"],
                    cwd=cwd,
                    capture_output=True,
                    text=True,
                )
                if result.returncode != 0:
                    stderr_lower = result.stderr.lower()
                    if "not possible to fast-forward" in stderr_lower:
                        return (
                            False,
                            "Cannot fast-forward. Please resolve conflicts manually.",
                        )
                    if "no tracking information" in stderr_lower:
                        return True, "OK (no upstream to pull from)"
                    return False, f"git pull failed: {result.stderr}"
            except subprocess.CalledProcessError as e:
                return False, f"git pull failed: {e.stderr}"

        return True, "OK"

    monkeypatch.setattr(sync_module, "git_fetch_and_pull", patched_git_fetch_and_pull)

    # Patch git.Git.execute to replace origin/dev with worker-specific branch
    # This is needed for spec complete which calls repo.git.rebase("origin/dev")
    from git import Git

    original_git_execute = Git.execute

    def patched_git_execute(self, command, **kwargs):
        """Intercept git commands and replace origin/dev with worker-specific branch."""
        if isinstance(command, (list, tuple)):
            command = list(command)
            for i, arg in enumerate(command):
                if arg == "origin/dev":
                    command[i] = f"origin/{dev_branch}"
        return original_git_execute(self, command, **kwargs)

    monkeypatch.setattr(Git, "execute", patched_git_execute)

    # Patch create_pull_request to use worker-specific dev branch as base
    # Need to patch it in the spec module where it's imported
    from src.commands import spec as spec_module
    from src.utils.github import api as github_api

    original_create_pull_request = github_api.create_pull_request

    def patched_create_pull_request(repo, title, body, head, base="dev"):
        """Use worker-specific dev branch as PR base."""
        if base == "dev":
            base = dev_branch
        return original_create_pull_request(repo, title, body, head, base)

    monkeypatch.setattr(github_api, "create_pull_request", patched_create_pull_request)
    monkeypatch.setattr(spec_module, "create_pull_request", patched_create_pull_request)

    # Track branches to clean up
    branches_to_cleanup = [dev_branch]

    try:
        yield base_dir
    finally:
        # Cleanup remote branches (best effort)
        for branch in branches_to_cleanup:
            try:
                repo.git.push("origin", "--delete", branch)
            except Exception:
                pass

        # Cleanup local temp directory
        repo.close()
        if base_dir.exists():
            shutil.rmtree(base_dir)


@pytest.fixture(scope="function")
def setup_test_env_isolated(request, github_client, github_token):
    """
    Sets up a test environment with worker-isolated branches for xdist compatibility.

    Each xdist worker gets its own set of branches (dev-gw0, test-gw0, main-gw0, etc.)
    to prevent conflicts when tests run in parallel.

    Yields:
        dict with:
            - path: Path to the local test repository
            - branches: dict mapping logical names to actual branch names
                       e.g., {"dev": "dev-gw0", "test": "test-gw0", "main": "main-gw0"}

    Teardown:
        - Deletes the local temp directory
        - Cleans up remote branches
    """
    suffix = get_worker_branch_suffix(request)
    branch_names = {
        "dev": f"dev{suffix}",
        "test": f"test{suffix}",
        "main": f"main{suffix}",
    }

    # Create a temp directory
    base_dir = Path(tempfile.mkdtemp(prefix="mem_test_"))

    # Clone the test repo
    user = github_client.get_user()
    auth_url = (
        f"https://oauth2:{github_token}@github.com/{user.login}/{TEST_REPO_NAME}.git"
    )

    repo = Repo.clone_from(auth_url, base_dir)

    # Configure git user for the test repo
    with repo.config_writer() as config:
        config.set_value("user", "name", "Test User")
        config.set_value("user", "email", "test@example.com")

    # Create worker-isolated branches from the default branch
    default_branch = repo.active_branch.name

    for logical_name, actual_name in branch_names.items():
        # Delete if exists locally
        if actual_name in [h.name for h in repo.heads]:
            repo.delete_head(actual_name, force=True)
        # Create from default branch
        repo.create_head(actual_name)

    # Checkout the isolated dev branch
    repo.heads[branch_names["dev"]].checkout()

    try:
        yield {
            "path": base_dir,
            "branches": branch_names,
        }
    finally:
        # Cleanup remote branches (best effort)
        for actual_name in branch_names.values():
            try:
                repo.git.push("origin", "--delete", actual_name)
            except Exception:
                pass  # Branch may not have been pushed

        # Cleanup local temp directory
        repo.close()
        if base_dir.exists():
            shutil.rmtree(base_dir)
