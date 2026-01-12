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
from github.AuthenticatedUser import AuthenticatedUser

load_dotenv()

TEST_REPO_NAME = "mem-test"


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

    Nukes and recreates the test repo at the start of each test session.
    """
    auth = Auth.Token(github_token)
    client = Github(auth=auth)
    user = client.get_user()
    repo_full_name = f"{user.login}/{TEST_REPO_NAME}"

    root_tmp_dir = tmp_path_factory.getbasetemp().parent
    lock_file = root_tmp_dir / "mem_test_repo.lock"
    marker_file = root_tmp_dir / "mem_test_repo_created"

    with FileLock(str(lock_file)):
        if not marker_file.exists():
            try:
                old_repo = client.get_repo(repo_full_name)
                old_repo.delete()
                time.sleep(3)
            except GithubException:
                pass

            assert isinstance(user, AuthenticatedUser)

            user.create_repo(
                TEST_REPO_NAME,
                description="Temporary repo for mem integration tests",
                private=False,
                auto_init=True,
            )
            time.sleep(3)

            marker_file.touch()

    yield client


@pytest.fixture(scope="session")
def cloned_test_repo(github_client, github_token, tmp_path_factory):
    """
    Session-scoped fixture that clones the test repo once.

    This is the "master" clone that gets copied for each test.
    """
    base_dir = tmp_path_factory.mktemp("mem_test_master")

    user = github_client.get_user()
    auth_url = (
        f"https://oauth2:{github_token}@github.com/{user.login}/{TEST_REPO_NAME}.git"
    )

    repo = Repo.clone_from(auth_url, base_dir)

    with repo.config_writer() as config:
        config.set_value("user", "name", "Test User")
        config.set_value("user", "email", "test@example.com")

    # Create dev branch if it doesn't exist
    if "dev" not in [h.name for h in repo.heads]:
        repo.create_head("dev")
    repo.heads["dev"].checkout()

    try:
        repo.git.push("origin", "dev", set_upstream=True, force=True)
    except Exception:
        pass

    repo.close()

    yield base_dir


@pytest.fixture(scope="function")
def setup_test_env(cloned_test_repo, github_token, monkeypatch):
    """
    Sets up a test environment by copying the session-scoped clone.

    This avoids cloning from GitHub for each test, significantly speeding up tests.

    Yields:
        Path: The path to the local test repository

    Teardown:
        - Deletes the local temp directory
        - Cleans up remote branches (best effort)
    """
    base_dir = Path(tempfile.mkdtemp(prefix="mem_test_"))

    # Copy the master clone instead of cloning from GitHub
    shutil.copytree(cloned_test_repo, base_dir, dirs_exist_ok=True)

    repo = Repo(base_dir)

    # Ensure we're on dev branch
    if "dev" not in [h.name for h in repo.heads]:
        repo.create_head("dev")
    repo.heads["dev"].checkout()

    # Fetch latest and reset to origin/dev to ensure clean state
    try:
        repo.git.fetch("origin")
        repo.git.reset("--hard", "origin/dev")
    except Exception:
        pass

    try:
        yield base_dir
    finally:
        repo.close()
        if base_dir.exists():
            shutil.rmtree(base_dir)
