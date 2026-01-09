"""
Tests for the mem init command.
"""

import typer

from src.commands.init import init
from src.utils.github.client import get_authenticated_user


def test_mem_init_success(setup_test_env, github_client, monkeypatch):
    """
    Test the mem init workflow on an already-initialized repo (force=True).

    Since we're using the actual mem repo clone, .mem already exists.
    We test that init with force=True recreates the config properly.
    """
    repo_path = setup_test_env
    monkeypatch.chdir(repo_path)

    # Run the init command with force to reinitialize
    try:
        init(force=True)
    except typer.Exit as e:
        # Exit code 0 is success
        assert e.exit_code == 0 or e.exit_code is None

    # Verify directory structure exists
    assert (repo_path / ".mem").exists()
    assert (repo_path / ".mem" / "specs").exists()
    assert (repo_path / ".mem" / "logs").exists()
    assert (repo_path / ".mem" / "todos").exists()
    assert (repo_path / ".mem" / "config.toml").exists()
    assert (repo_path / ".mem" / "user_mappings.toml").exists()

    # Verify config content - new format
    config_content = (repo_path / ".mem" / "config.toml").read_text()

    # New config structure has [vars], [project], and [[files]]
    assert "[vars]" in config_content
    assert "[project]" in config_content
    assert "[[files]]" in config_content
    assert "github_token_env" in config_content

    # Get current GH user to verify user_mappings
    gh_user = get_authenticated_user(github_client)

    # Verify user_mappings.toml content
    mappings_content = (repo_path / ".mem" / "user_mappings.toml").read_text()
    assert f"[{gh_user['username']}]" in mappings_content


def test_mem_init_already_initialized_no_force(setup_test_env, monkeypatch):
    """
    Test that mem init respects existing initialization when force is False.
    """
    repo_path = setup_test_env
    monkeypatch.chdir(repo_path)

    # The repo already has .mem from the clone, so this tests the "already initialized" path

    # Mock typer.confirm to return False (don't reinitialize)
    monkeypatch.setattr("typer.confirm", lambda msg, default: False)

    try:
        init(force=False)
    except typer.Exit as e:
        # Should exit gracefully
        assert e.exit_code == 0 or e.exit_code is None
