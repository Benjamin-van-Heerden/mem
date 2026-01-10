"""
GitHub API utilities for labels, issues, and pull requests.
"""

import re
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
    assignees: Optional[List[str]] = None,
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
        assignees: New assignees to set (replaces existing assignees) (optional)

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
        if assignees is not None:
            edit_kwargs["assignees"] = assignees

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


def list_merge_ready_prs(
    repo: Repository.Repository,
    base_branch: str = "dev",
) -> List[Dict[str, Any]]:
    """
    List open PRs targeting base_branch that are ready to merge.

    Looks for PRs with "[Complete]:" in the title (our convention from mem spec complete).

    Args:
        repo: PyGithub Repository instance
        base_branch: The branch PRs should target (default: dev)

    Returns:
        List of dicts with PR info:
            - number: PR number
            - title: PR title (with [Complete]: prefix stripped)
            - author: GitHub username
            - issue_number: Linked issue number (from "Closes #X" in body) or None
            - checks_passing: bool or None if no checks
            - mergeable: bool or None if unknown
            - html_url: Link to PR
            - head_branch: Branch name to delete after merge
    """
    try:
        pulls = repo.get_pulls(state="open", base=base_branch)
        result = []

        for pr in pulls:
            # Only include PRs with [Complete]: in title
            if "[Complete]:" not in pr.title:
                continue

            # Extract issue number from body (look for "Closes #X")
            issue_number = None
            if pr.body:
                match = re.search(r"Closes\s+#(\d+)", pr.body, re.IGNORECASE)
                if match:
                    issue_number = int(match.group(1))

            # Check if checks are passing
            checks_passing = None
            try:
                commit = repo.get_commit(pr.head.sha)
                combined_status = commit.get_combined_status()
                if combined_status.total_count > 0:
                    checks_passing = combined_status.state == "success"
            except GithubException:
                pass

            # Clean title for display
            display_title = (
                pr.title.replace("[Complete]:", "").replace("[Complete]: ", "").strip()
            )

            result.append(
                {
                    "number": pr.number,
                    "title": display_title,
                    "author": pr.user.login,
                    "issue_number": issue_number,
                    "checks_passing": checks_passing,
                    "mergeable": pr.mergeable,
                    "html_url": pr.html_url,
                    "head_branch": pr.head.ref,
                }
            )

        return result
    except GithubException as e:
        raise GitHubError(f"Failed to list pull requests: {e}")


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


def get_pr_mergeable_status(
    repo: Repository.Repository,
    pr_url: str,
) -> Dict[str, Any]:
    """
    Get the mergeable status of a Pull Request.

    Args:
        repo: PyGithub Repository instance
        pr_url: The full PR URL

    Returns:
        Dict with keys:
            - exists: bool - whether the PR exists
            - merged: bool - whether already merged
            - mergeable: bool | None - whether it can be merged (None if unknown/checking)
            - mergeable_state: str - 'clean', 'dirty', 'blocked', 'behind', 'unknown'
            - pr: PullRequest | None - the PR object if it exists
    """
    pr = get_pull_request_by_url(repo, pr_url)
    if pr is None:
        return {
            "exists": False,
            "merged": False,
            "mergeable": False,
            "mergeable_state": "unknown",
            "pr": None,
        }

    return {
        "exists": True,
        "merged": pr.merged,
        "mergeable": pr.mergeable,
        "mergeable_state": pr.mergeable_state or "unknown",
        "pr": pr,
    }


def merge_pull_request(
    pr: PullRequest.PullRequest,
    merge_method: str = "rebase",
    commit_message: Optional[str] = None,
) -> Dict[str, Any]:
    """
    Merge a Pull Request.

    Args:
        pr: PyGithub PullRequest instance
        merge_method: 'merge', 'squash', or 'rebase'
        commit_message: Optional commit message (used for merge/squash)

    Returns:
        Dict with keys:
            - success: bool
            - sha: str | None - merge commit SHA if successful
            - message: str - success/error message
    """
    try:
        if commit_message:
            result = pr.merge(merge_method=merge_method, commit_message=commit_message)
        else:
            result = pr.merge(merge_method=merge_method)

        return {
            "success": result.merged,
            "sha": result.sha if result.merged else None,
            "message": result.message or "Merged successfully",
        }
    except GithubException as e:
        return {
            "success": False,
            "sha": None,
            "message": f"Failed to merge: {e.data.get('message', str(e))}",
        }


def delete_branch(
    repo: Repository.Repository,
    branch_name: str,
) -> bool:
    """
    Delete a branch from the remote repository.

    Args:
        repo: PyGithub Repository instance
        branch_name: Name of the branch to delete

    Returns:
        True if deleted successfully, False otherwise
    """
    try:
        ref = repo.get_git_ref(f"heads/{branch_name}")
        ref.delete()
        return True
    except GithubException:
        return False


def close_pull_request(
    repo: Repository.Repository,
    pr_url: str,
    comment: Optional[str] = None,
) -> bool:
    """
    Close a Pull Request without merging.

    Args:
        repo: PyGithub Repository instance
        pr_url: The full PR URL
        comment: Optional comment to add before closing

    Returns:
        True if closed successfully, False otherwise
    """
    try:
        pr = get_pull_request_by_url(repo, pr_url)
        if pr is None:
            return False

        if pr.state == "closed":
            return True

        if comment:
            pr.create_issue_comment(comment)

        pr.edit(state="closed")
        return True
    except GithubException:
        return False
