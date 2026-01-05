"""
Tests for username-prefixed log files.
"""

from datetime import date
from pathlib import Path

import pytest

from src.utils import logs
from src.utils.markdown import slugify


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

    # Create user_mappings.toml
    mappings_content = """# GitHub username to Git user mappings
[test-github-user]
name = "Test User"
email = "test@example.com"
"""
    (repo_path / ".mem" / "user_mappings.toml").write_text(mappings_content)

    return repo_path


def test_log_filename_includes_username(initialized_mem):
    """Test that log filenames include the username prefix."""
    # Create a log
    log_path = logs.create_log()

    # Verify filename format: {username}_{YYYYMMDD}_session.md
    filename = log_path.name
    assert filename.endswith("_session.md")

    # Should have format: username_YYYYMMDD_session.md
    parts = filename.replace("_session.md", "").rsplit("_", 1)
    assert len(parts) == 2

    username_part = parts[0]
    date_part = parts[1]

    # Username should be slugified
    assert username_part == slugify(username_part)

    # Date part should be 8 digits
    assert len(date_part) == 8
    assert date_part.isdigit()


def test_log_metadata_includes_username(initialized_mem):
    """Test that log metadata includes username field."""
    logs.create_log()

    log = logs.get_today_log()
    assert log is not None
    assert "username" in log
    assert log["username"] is not None


def test_get_log_finds_user_log(initialized_mem):
    """Test that get_log finds the correct user's log."""
    # Create a log
    logs.create_log()

    # Get it back
    today = date.today()
    log = logs.get_log(today)

    assert log is not None
    assert log.get("body") is not None


def test_list_logs_returns_all_users(initialized_mem):
    """Test that list_logs returns logs from all users by default."""
    logs_dir = Path(initialized_mem) / ".mem" / "logs"

    # Manually create logs for different users
    from src.utils.markdown import write_md_file

    today = date.today()
    date_str = today.strftime("%Y%m%d")

    # Create log for user1
    user1_file = logs_dir / f"user_one_{date_str}_session.md"
    write_md_file(
        user1_file,
        {"date": today.isoformat(), "username": "user_one"},
        "User one's log",
    )

    # Create log for user2
    user2_file = logs_dir / f"user_two_{date_str}_session.md"
    write_md_file(
        user2_file,
        {"date": today.isoformat(), "username": "user_two"},
        "User two's log",
    )

    # List all logs
    all_logs = logs.list_logs()

    # Should have both users
    usernames = [log["username"] for log in all_logs]
    assert "user_one" in usernames
    assert "user_two" in usernames


def test_list_logs_filter_by_username(initialized_mem):
    """Test that list_logs can filter by username."""
    logs_dir = Path(initialized_mem) / ".mem" / "logs"

    from src.utils.markdown import write_md_file

    today = date.today()
    date_str = today.strftime("%Y%m%d")

    # Create logs for different users
    user1_file = logs_dir / f"alice_{date_str}_session.md"
    write_md_file(
        user1_file, {"date": today.isoformat(), "username": "alice"}, "Alice's log"
    )

    user2_file = logs_dir / f"bob_{date_str}_session.md"
    write_md_file(
        user2_file, {"date": today.isoformat(), "username": "bob"}, "Bob's log"
    )

    # Filter by alice
    alice_logs = logs.list_logs(username="alice")
    assert len(alice_logs) == 1
    assert alice_logs[0]["username"] == "alice"

    # Filter by bob
    bob_logs = logs.list_logs(username="bob")
    assert len(bob_logs) == 1
    assert bob_logs[0]["username"] == "bob"


def test_parse_log_filename_extracts_username_and_date(initialized_mem):
    """Test that _parse_log_filename correctly extracts username and date."""
    # Test valid filename
    result = logs._parse_log_filename("benjamin_van_heerden_20251230_session.md")
    assert result is not None
    username, log_date = result
    assert username == "benjamin_van_heerden"
    assert log_date == date(2025, 12, 30)

    # Test simple username
    result = logs._parse_log_filename("alice_20251225_session.md")
    assert result is not None
    username, log_date = result
    assert username == "alice"
    assert log_date == date(2025, 12, 25)

    # Test invalid filename (no date)
    result = logs._parse_log_filename("invalid_session.md")
    assert result is None

    # Test invalid filename (wrong suffix)
    result = logs._parse_log_filename("alice_20251225.md")
    assert result is None


def test_append_to_log_uses_current_user(initialized_mem):
    """Test that append_to_log appends to the current user's log."""
    # Append to log (creates it if doesn't exist)
    logs.append_to_log("Test Section", "Test content")

    # Get today's log
    log = logs.get_today_log()
    assert log is not None
    assert "Test Section" in log["body"]
    assert "Test content" in log["body"]


def test_update_log_updates_correct_user_log(initialized_mem):
    """Test that update_log updates the correct user's log."""
    # Create a log
    logs.create_log()

    today = date.today()

    # Update it
    logs.update_log(today, spec_slug="test_spec")

    # Verify update
    log = logs.get_log(today)
    assert log is not None
    assert log["spec_slug"] == "test_spec"


def test_delete_log_deletes_correct_user_log(initialized_mem):
    """Test that delete_log deletes the correct user's log."""
    # Create a log
    logs.create_log()

    today = date.today()

    # Verify it exists
    assert logs.get_log(today) is not None

    # Delete it
    logs.delete_log(today)

    # Verify it's gone
    assert logs.get_log(today) is None


def test_github_username_lookup(initialized_mem):
    """Test that _get_current_github_username looks up correctly."""
    # The test fixture sets up user_mappings.toml with:
    # [test-github-user]
    # name = "Test User"
    # email = "test@example.com"
    #
    # And git config has:
    # user.name = "Test User"

    username = logs._get_current_github_username()

    # Should return slugified github username
    assert username == "test_github_user"


def test_multiple_users_same_day(initialized_mem):
    """Test that multiple users can have logs for the same day."""
    logs_dir = Path(initialized_mem) / ".mem" / "logs"

    from src.utils.markdown import write_md_file

    today = date.today()
    date_str = today.strftime("%Y%m%d")

    # Create logs for multiple users on same day
    for user in ["alice", "bob", "charlie"]:
        user_file = logs_dir / f"{user}_{date_str}_session.md"
        write_md_file(
            user_file,
            {"date": today.isoformat(), "username": user},
            f"{user}'s work for today",
        )

    # List all logs
    all_logs = logs.list_logs()
    assert len(all_logs) == 3

    # Each user should have their own log
    usernames = {log["username"] for log in all_logs}
    assert usernames == {"alice", "bob", "charlie"}
