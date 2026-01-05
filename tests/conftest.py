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
def setup_test_env(github_client, github_token):
    """
    Sets up a test environment using a temp directory that clones the test repo.

    Uses a dedicated mem-test repo that gets nuked at session start.

    Yields:
        Path: The path to the local test repository

    Teardown:
        - Deletes the local temp directory
    """
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

    # Ensure we're on a test branch to avoid polluting main
    test_branch = f"test-{os.getpid()}"
    repo.create_head(test_branch)
    repo.heads[test_branch].checkout()

    # Create a local 'dev' branch from test branch so branch switching works
    if "dev" in [h.name for h in repo.heads]:
        repo.delete_head("dev", force=True)
    repo.create_head("dev")

    try:
        yield base_dir
    finally:
        # Cleanup - just delete the local temp directory
        repo.close()
        if base_dir.exists():
            shutil.rmtree(base_dir)
