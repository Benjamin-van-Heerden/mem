"""
GitHub API utilities for labels, issues, and pull requests.
"""

from typing import Any, Dict, List, Optional

from github import GithubException, Issue, PullRequest, Repository

from src.utils.github.exceptions import GitHubError


def ensure_label(
    repo: Repository.Repository, name: str, color: str, description: str = ""
) -> None:
    """
    Ensure a label exists in the repository. Creates it if it doesn't.

    Args:
        repo: PyGithub Repository instance
        name: Name of the label
        color: Hex color code (without #)
        description: Optional description
    """
    try:
        repo.get_label(name)
    except GithubException as e:
        if e.status == 404:
            try:
                repo.create_label(name, color, description)
            except GithubException as ce:
                raise GitHubError(f"Failed to create label '{name}': {ce}")
        else:
            raise GitHubError(f"Failed to check for label '{name}': {e}")


def create_github_issue(
    repo: Repository.Repository,
    title: str,
    body: str,
    labels: Optional[List[str]] = None,
    assignees: Optional[List[str]] = None,
) -> Issue.Issue:
    """
    Create a new issue on GitHub.

    Args:
        repo: PyGithub Repository instance
        title: Issue title
        body: Issue body (markdown)
        labels: List of label names to apply
        assignees: List of GitHub usernames to assign

    Returns:
        The created Issue instance
    """
    try:
        return repo.create_issue(
            title=title,
            body=body,
            labels=labels or [],
            assignees=assignees or [],
        )
    except GithubException as e:
        raise GitHubError(f"Failed to create GitHub issue: {e}")


def create_pull_request(
    repo: Repository.Repository, title: str, body: str, head: str, base: str = "dev"
) -> PullRequest.PullRequest:
    """
    Create a new Pull Request on GitHub.

    Args:
        repo: PyGithub Repository instance
        title: PR title
        body: PR body (markdown)
        head: The name of the branch where your changes are implemented
        base: The name of the branch you want the changes pulled into

    Returns:
        The created PullRequest instance
    """
    try:
        return repo.create_pull(title=title, body=body, head=head, base=base)
    except GithubException as e:
        raise GitHubError(f"Failed to create Pull Request: {e}")


def get_issue_by_id(repo: Repository.Repository, issue_id: int) -> Issue.Issue:
    """
    Retrieve a specific issue by its ID.

    Args:
        repo: PyGithub Repository instance
        issue_id: The issue number
    """
    try:
        return repo.get_issue(number=issue_id)
    except GithubException as e:
        raise GitHubError(f"Failed to retrieve issue #{issue_id}: {e}")


def list_repo_issues(
    repo: Repository.Repository, labels: Optional[List[str]] = None, state: str = "open"
) -> List[Issue.Issue]:
    """
    List issues in the repository, optionally filtered by labels and state.
    Excludes pull requests (which GitHub's API returns as issues).

    Args:
        repo: PyGithub Repository instance
        labels: List of label names to filter by
        state: 'open', 'closed', or 'all'
    """
    try:
        if labels:
            issues = repo.get_issues(state=state, labels=labels)
        else:
            issues = repo.get_issues(state=state)
        # Filter out pull requests (they have a pull_request attribute)
        return [issue for issue in issues if issue.pull_request is None]
    except GithubException as e:
        raise GitHubError(f"Failed to list issues: {e}")


def get_comments(issue: Issue.Issue) -> List[Dict[str, Any]]:
    """
    Get all comments for an issue, formatted for local storage.

    Args:
        issue: PyGithub Issue instance

    Returns:
        List of dicts containing comment info
    """
    try:
        comments = issue.get_comments()
        return [
            {
                "user": comment.user.login,
                "body": comment.body,
                "created_at": comment.created_at.isoformat(),
            }
            for comment in comments
        ]
    except GithubException as e:
        raise GitHubError(f"Failed to retrieve comments for issue #{issue.number}: {e}")


# Status label configuration
# Maps spec status to GitHub label name and color
STATUS_LABELS = {
    "todo": {
        "label": "mem-status:todo",
        "color": "6B7280",
        "description": "Spec not yet started",
    },
    "active": {
        "label": "mem-status:active",
        "color": "22C55E",
        "description": "Spec currently being worked on",
    },
    "inactive": {
        "label": "mem-status:inactive",
        "color": "EAB308",
        "description": "Spec paused",
    },
    "completed": {
        "label": "mem-status:completed",
        "color": "3B82F6",
        "description": "Spec completed",
    },
    "merge_ready": {
        "label": "mem-status:merge-ready",
        "color": "8B5CF6",
        "description": "Spec ready to merge",
    },
    "archived": {
        "label": "mem-status:archived",
        "color": "374151",
        "description": "Spec archived",
    },
}


def ensure_status_labels(repo: Repository.Repository) -> None:
    """
    Create all mem-status:* labels if they don't exist.

    Args:
        repo: PyGithub Repository instance
    """
    for status, config in STATUS_LABELS.items():
        ensure_label(repo, config["label"], config["color"], config["description"])


def get_status_label_name(status: str) -> Optional[str]:
    """
    Get the GitHub label name for a spec status.

    Args:
        status: The spec status (e.g., 'active', 'todo')

    Returns:
        The label name (e.g., 'mem-status:active') or None if status is unknown
    """
    if status in STATUS_LABELS:
        return STATUS_LABELS[status]["label"]
    return None


def get_status_from_labels(labels: List[str]) -> Optional[str]:
    """
    Extract spec status from a list of GitHub label names.

    Args:
        labels: List of label names from a GitHub issue

    Returns:
        The spec status (e.g., 'active') or None if no status label found
    """
    # Build reverse mapping: label -> status
    label_to_status = {
        config["label"]: status for status, config in STATUS_LABELS.items()
    }

    for label in labels:
        if label in label_to_status:
            return label_to_status[label]

    return None


def update_github_issue(
    repo: Repository.Repository,
    issue_number: int,
    title: Optional[str] = None,
    body: Optional[str] = None,
    state: Optional[str] = None,
    labels: Optional[List[str]] = None,
) -> Issue.Issue:
    """
    Update an existing GitHub issue.

    Args:
        repo: PyGithub Repository instance
        issue_number: The issue number to update
        title: New title (optional)
        body: New body content (optional)
        state: New state - 'open' or 'closed' (optional)
        labels: New labels to set (replaces existing labels) (optional)

    Returns:
        The updated Issue instance
    """
    try:
        issue = repo.get_issue(number=issue_number)

        # Build kwargs for edit() - only include non-None values
        edit_kwargs: Dict[str, Any] = {}
        if title is not None:
            edit_kwargs["title"] = title
        if body is not None:
            edit_kwargs["body"] = body
        if state is not None:
            edit_kwargs["state"] = state
        if labels is not None:
            edit_kwargs["labels"] = labels

        if edit_kwargs:
            issue.edit(**edit_kwargs)

        return issue
    except GithubException as e:
        raise GitHubError(f"Failed to update issue #{issue_number}: {e}")


def sync_status_labels(
    repo: Repository.Repository,
    issue_number: int,
    new_status: str,
) -> None:
    """
    Synchronize status labels on a GitHub issue.

    Removes any existing mem-status:* labels and adds the label for the new status.

    Args:
        repo: PyGithub Repository instance
        issue_number: The issue number to update
        new_status: The new spec status (e.g., 'active', 'completed')
    """
    try:
        issue = repo.get_issue(number=issue_number)

        # Get current labels
        current_labels = [label.name for label in issue.labels]

        # Remove all existing mem-status:* labels
        new_labels = [
            label for label in current_labels if not label.startswith("mem-status:")
        ]

        # Add the new status label
        new_status_label = get_status_label_name(new_status)
        if new_status_label:
            new_labels.append(new_status_label)

        # Update labels on the issue
        issue.edit(labels=new_labels)

    except GithubException as e:
        raise GitHubError(
            f"Failed to sync status labels for issue #{issue_number}: {e}"
        )


def close_issue_with_comment(
    repo: Repository.Repository,
    issue_number: int,
    comment: str,
) -> Issue.Issue:
    """
    Close a GitHub issue with a comment explaining why.

    Args:
        repo: PyGithub Repository instance
        issue_number: The issue number to close
        comment: Comment to add before closing

    Returns:
        The closed Issue instance
    """
    try:
        issue = repo.get_issue(number=issue_number)
        issue.create_comment(comment)
        issue.edit(state="closed")
        return issue
    except GithubException as e:
        raise GitHubError(f"Failed to close issue #{issue_number}: {e}")


def get_pull_request_by_url(
    repo: Repository.Repository,
    pr_url: str,
) -> Optional[PullRequest.PullRequest]:
    """
    Get a Pull Request by its URL.

    Args:
        repo: PyGithub Repository instance
        pr_url: The full PR URL (e.g., https://github.com/owner/repo/pull/123)

    Returns:
        The PullRequest instance or None if not found
    """
    try:
        # Extract PR number from URL
        # URL format: https://github.com/owner/repo/pull/123
        parts = pr_url.rstrip("/").split("/")
        pr_number = int(parts[-1])
        return repo.get_pull(pr_number)
    except (ValueError, IndexError, GithubException):
        return None


def is_pr_merged(
    repo: Repository.Repository,
    pr_url: str,
) -> bool:
    """
    Check if a Pull Request has been merged.

    Args:
        repo: PyGithub Repository instance
        pr_url: The full PR URL

    Returns:
        True if the PR is merged, False otherwise
    """
    pr = get_pull_request_by_url(repo, pr_url)
    if pr is None:
        return False
    return pr.merged
