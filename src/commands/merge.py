"""
Merge command for mem.

Queries GitHub for PRs ready to merge (with [Complete]: prefix),
allows selection, and performs merges via GitHub API.
"""

import subprocess
from typing import Annotated

import typer

from env_settings import ENV_SETTINGS
from src.commands.sync import git_fetch_and_pull
from src.utils.github.api import (
    delete_branch,
    list_merge_ready_prs,
    merge_pull_request,
)
from src.utils.github.client import get_github_client
from src.utils.github.repo import get_repo_from_git


def check_working_directory_clean() -> tuple[bool, str]:
    """
    Check if working directory has uncommitted changes.

    Returns:
        (is_clean, message) tuple
    """
    cwd = ENV_SETTINGS.caller_dir
    result = subprocess.run(
        ["git", "status", "--porcelain"],
        cwd=cwd,
        capture_output=True,
        text=True,
    )
    if result.stdout.strip():
        return False, "Uncommitted changes detected. Commit or stash before merging."
    return True, ""


def delete_local_branch(branch_name: str) -> bool:
    """
    Delete a local git branch.

    Tries -d first (safe delete), falls back to -D (force delete) if needed.
    Returns True if deleted successfully, False otherwise.
    """
    cwd = ENV_SETTINGS.caller_dir

    # Try safe delete first
    result = subprocess.run(
        ["git", "branch", "-d", branch_name],
        cwd=cwd,
        capture_output=True,
        text=True,
    )
    if result.returncode == 0:
        return True

    # Fall back to force delete
    result = subprocess.run(
        ["git", "branch", "-D", branch_name],
        cwd=cwd,
        capture_output=True,
        text=True,
    )
    return result.returncode == 0


def prune_remote_refs() -> None:
    """Prune stale remote tracking references."""
    cwd = ENV_SETTINGS.caller_dir
    subprocess.run(
        ["git", "remote", "prune", "origin"],
        cwd=cwd,
        capture_output=True,
        text=True,
    )


def merge(
    all_ready: Annotated[
        bool,
        typer.Option("--all", "-a", help="Merge all ready PRs without prompting"),
    ] = False,
    dry_run: Annotated[
        bool,
        typer.Option(
            "--dry-run", "-n", help="Show what would be merged without merging"
        ),
    ] = False,
    force: Annotated[
        bool,
        typer.Option(
            "--force", "-f", help="Merge even if checks are failing (with warning)"
        ),
    ] = False,
    no_sync: Annotated[
        bool,
        typer.Option("--no-sync", help="Skip running sync after merge"),
    ] = False,
    delete_branches: Annotated[
        bool,
        typer.Option(
            "--delete-branches/--keep-branches",
            help="Delete remote branches after merge",
        ),
    ] = True,
):
    """
    Merge pull requests for completed specs.

    Queries GitHub for open PRs with '[Complete]:' in the title (created by
    mem spec complete). Shows PR status and allows selection for merge.

    Uses rebase merge strategy for clean linear history.
    After merging, runs 'mem sync' to update local state.
    """
    try:
        # Pre-flight checks
        is_clean, message = check_working_directory_clean()
        if not is_clean:
            typer.echo(f"Error: {message}", err=True)
            raise typer.Exit(code=1)

        typer.echo("Fetching latest changes...")
        success, message = git_fetch_and_pull()
        if not success:
            typer.echo(f"Error: {message}", err=True)
            raise typer.Exit(code=1)
        typer.echo("Local branch is up to date.\n")

        # Get GitHub client and repo
        client = get_github_client()
        repo_owner, repo_name = get_repo_from_git(ENV_SETTINGS.caller_dir)
        gh_repo = client.get_repo(f"{repo_owner}/{repo_name}")

        typer.echo("Querying GitHub for merge-ready PRs...\n")

        # Query GitHub for PRs with [Complete]: in title
        prs = list_merge_ready_prs(gh_repo)

        if not prs:
            typer.echo("No PRs ready to merge.")
            typer.echo("\nPRs must have '[Complete]:' in the title to appear here.")
            typer.echo("Use 'mem spec complete <slug> \"message\"' to create such PRs.")
            raise typer.Exit(code=0)

        # Categorize PRs
        ready_to_merge = []
        checks_failing = []
        has_conflicts = []

        for pr in prs:
            if pr["mergeable"] is False:
                has_conflicts.append(pr)
            elif pr["checks_passing"] is False:
                checks_failing.append(pr)
            else:
                ready_to_merge.append(pr)

        # Display PRs
        if ready_to_merge:
            typer.echo("Ready to merge:")
            for i, pr in enumerate(ready_to_merge, 1):
                issue_info = (
                    f" (issue #{pr['issue_number']})" if pr["issue_number"] else ""
                )
                checks_info = ""
                if pr["checks_passing"] is True:
                    checks_info = " [checks: passing]"
                elif pr["checks_passing"] is None:
                    checks_info = " [checks: none]"
                typer.echo(
                    f"  {i}. #{pr['number']} {pr['title']}{issue_info}{checks_info}"
                )
                typer.echo(
                    f"      Author: {pr['author']} | Branch: {pr['head_branch']}"
                )
            typer.echo()

        if checks_failing:
            typer.echo("Checks failing (use --force to merge anyway):")
            for pr in checks_failing:
                issue_info = (
                    f" (issue #{pr['issue_number']})" if pr["issue_number"] else ""
                )
                typer.echo(f"  - #{pr['number']} {pr['title']}{issue_info}")
            typer.echo()

        if has_conflicts:
            typer.echo("Has conflicts (resolve on GitHub first):")
            for pr in has_conflicts:
                issue_info = (
                    f" (issue #{pr['issue_number']})" if pr["issue_number"] else ""
                )
                typer.echo(f"  - #{pr['number']} {pr['title']}{issue_info}")
            typer.echo()

        # Include failing checks if --force
        if force and checks_failing:
            typer.echo("--force: Including PRs with failing checks")
            ready_to_merge.extend(checks_failing)

        if not ready_to_merge:
            typer.echo("No PRs ready to merge.")
            raise typer.Exit(code=0)

        # Dry run - just show what would happen
        if dry_run:
            typer.echo(f"Dry run: Would merge {len(ready_to_merge)} PR(s)")
            raise typer.Exit(code=0)

        # Select PRs to merge
        if all_ready or len(ready_to_merge) == 1:
            selected = ready_to_merge
            if len(ready_to_merge) == 1:
                typer.echo(f"Merging PR #{ready_to_merge[0]['number']}...")
        else:
            typer.echo(
                "Select PRs to merge (comma-separated numbers, 'all', or 'q' to quit):"
            )
            selection = typer.prompt("Selection", default="all")

            if selection.lower() == "q":
                typer.echo("Cancelled.")
                raise typer.Exit(code=0)

            if selection.lower() == "all":
                selected = ready_to_merge
            else:
                try:
                    indices = [int(x.strip()) - 1 for x in selection.split(",")]
                    selected = [
                        ready_to_merge[i]
                        for i in indices
                        if 0 <= i < len(ready_to_merge)
                    ]
                except (ValueError, IndexError):
                    typer.echo("Invalid selection.", err=True)
                    raise typer.Exit(code=1)

        if not selected:
            typer.echo("No PRs selected.")
            raise typer.Exit(code=0)

        # Merge selected PRs
        typer.echo(f"\nMerging {len(selected)} PR(s)...\n")

        success_count = 0
        for pr_info in selected:
            typer.echo(f"Merging #{pr_info['number']}: {pr_info['title']}...")

            # Get full PR object for merge
            pr = gh_repo.get_pull(pr_info["number"])
            result = merge_pull_request(pr, merge_method="rebase")

            if result["success"]:
                typer.echo(f"  Merged (SHA: {result['sha'][:7]})")
                success_count += 1

                # Delete remote branch
                if delete_branches:
                    branch = pr_info["head_branch"]
                    if delete_branch(gh_repo, branch):
                        typer.echo(f"  Deleted remote branch: {branch}")
                    else:
                        typer.echo(
                            f"  Warning: Could not delete remote branch: {branch}"
                        )

                    # Delete local branch
                    if delete_local_branch(branch):
                        typer.echo(f"  Deleted local branch: {branch}")
                    else:
                        typer.echo(
                            f"  Warning: Could not delete local branch: {branch}"
                        )
            else:
                typer.echo(f"  Failed: {result['message']}", err=True)

        typer.echo(f"\nMerged {success_count}/{len(selected)} PR(s).")

        # Prune stale remote tracking refs
        if delete_branches and success_count > 0:
            prune_remote_refs()
            typer.echo("Pruned stale remote tracking refs.")

        # Run sync to update local state
        if not no_sync and success_count > 0:
            typer.echo("\nRunning sync to update local state...")
            from src.commands.sync import sync as run_sync

            run_sync(dry_run=False, no_git=False)

    except typer.Exit:
        raise
    except Exception as e:
        typer.echo(f"Error: {e}", err=True)
        raise typer.Exit(code=1)
