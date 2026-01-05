"""
GitHub client and authentication utilities.
"""

import os

from github import Auth, Github, GithubException

from src.utils.github.exceptions import GitHubAuthenticationError


def get_github_token() -> str:
    """
    Get GitHub token from environment variable.

    Returns:
        GitHub personal access token

    Raises:
        GitHubAuthenticationError: If GITHUB_TOKEN environment variable is not set
    """
    token = os.getenv("GITHUB_TOKEN")
    if not token:
        raise GitHubAuthenticationError(
            "GITHUB_TOKEN environment variable not set.\n"
            "Please create a GitHub Personal Access Token with 'repo' and 'user' scopes:\n"
            "  1. Go to https://github.com/settings/tokens\n"
            "  2. Generate new token (classic)\n"
            "  3. Select scopes: 'repo', 'read:user'\n"
            "  4. Export it: export GITHUB_TOKEN='your_token_here'"
        )
    return token


def get_github_client() -> Github:
    """
    Create authenticated GitHub client.

    Returns:
        Authenticated Github client instance

    Raises:
        GitHubAuthenticationError: If authentication fails
    """
    try:
        token = get_github_token()
        auth = Auth.Token(token)
        g = Github(auth=auth)
        # Test authentication by getting user login
        g.get_user().login
        return g
    except GithubException as e:
        raise GitHubAuthenticationError(f"GitHub authentication failed: {e}")
    except Exception as e:
        raise GitHubAuthenticationError(
            f"Unexpected error during GitHub authentication: {e}"
        )


def get_authenticated_user(client: Github) -> dict:
    """
    Get authenticated user information from GitHub.

    Args:
        client: Authenticated Github client

    Returns:
        Dict with user info: {
            'username': str,
            'name': str,
            'email': str,
        }

    Raises:
        GitHubAuthenticationError: If user info retrieval fails
    """
    try:
        user = client.get_user()
        return {
            "username": user.login,
            "name": user.name or user.login,
            "email": user.email or "",
        }
    except GithubException as e:
        raise GitHubAuthenticationError(f"Failed to get GitHub user info: {e}")
