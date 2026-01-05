from git import Repo

from src.utils.github.repo import parse_github_repo_url


def test_sandbox_creation(setup_test_env, github_client):
    """
    Verifies that the sandbox environment is correctly set up:
    1. Directory exists
    2. Git repo initialized
    3. Remote 'origin' configured
    4. Remote points to correct GitHub repo (mem)
    5. Local git user configured
    """
    repo_path = setup_test_env

    # 1. Check directory
    assert repo_path.exists()
    assert repo_path.is_dir()

    # 2. Check git repo
    repo = Repo(repo_path)
    assert not repo.bare

    # 3. Check remote
    assert "origin" in [r.name for r in repo.remotes]

    # 4. Check remote URL points to the user's mem repo
    remote_url = repo.remote("origin").url
    parsed = parse_github_repo_url(remote_url)
    assert parsed is not None, f"Failed to parse GitHub URL: {remote_url}"
    owner, repo_name = parsed

    # Get current user to verify owner
    current_user = github_client.get_user().login

    assert owner == current_user
    assert repo_name == "mem-test"

    # 5. Check git config
    reader = repo.config_reader()
    assert reader.get_value("user", "name") == "Test User"
    assert reader.get_value("user", "email") == "test@example.com"
