"""
Custom exceptions for GitHub and Git operations
"""


class GitHubError(Exception):
    """Base exception for GitHub-related errors"""

    pass


class GitHubAuthenticationError(GitHubError):
    """Raised when GitHub authentication fails or token is missing"""

    pass


class GitHubRepositoryError(GitHubError):
    """Raised when there are issues with the GitHub repository (not found, access denied)"""

    pass


class GitError(Exception):
    """Base exception for local Git operations"""

    pass


class GitRepositoryNotFoundError(GitError):
    """Raised when a directory is not a git repository"""

    pass
