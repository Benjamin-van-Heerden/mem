"""
Cleanup command for mem.

Removes stale local and remote branches from completed/abandoned specs.
"""

import subprocess
from typing import Annotated

import typer

from env_settings import ENV_SETTINGS
from src.commands.merge import delete_local_branch, prune_remote_refs
from src.utils import specs
from src.utils.github.api import delete_branch
from src.utils.github.client import get_github_client
from src.utils.github.repo import get_repo_from_git


def get_local_branches() -> list[str]:
    """Get list of local branches matching dev-* pattern."""
    cwd = ENV_SETTINGS.caller_dir
    result = subprocess.run(
        ["git", "branch", "--list", "dev-*"],
        cwd=cwd,
        capture_output=True,
        text=True,
    )
    branches = []
    for line in result.stdout.strip().split("\n"):
        branch = line.strip().lstrip("* ")
        if branch:
            branches.append(branch)
    return branches


def get_remote_branches() -> list[str]:
    """Get list of remote branches matching dev-* pattern."""
    cwd = ENV_SETTINGS.caller_dir
    # Fetch first to get latest remote state
    subprocess.run(
        ["git", "fetch", "--prune", "origin"],
        cwd=cwd,
        capture_output=True,
        text=True,
    )
    result = subprocess.run(
        ["git", "branch", "-r", "--list", "origin/dev-*"],
        cwd=cwd,
        capture_output=True,
        text=True,
    )
    branches = []
    for line in result.stdout.strip().split("\n"):
        branch = line.strip()
        if branch and branch.startswith("origin/"):
            # Remove origin/ prefix
            branches.append(branch[7:])
    return branches


def extract_spec_slug_from_branch(branch_name: str) -> str | None:
    """
    Extract spec slug from branch name.

    Branch format: dev-{username}-{spec_slug}
    Returns spec_slug or None if invalid format.
    """
    if not branch_name.startswith("dev-"):
        return None

    parts = branch_name.split("-", 2)
    if len(parts) < 3:
        return None

    return parts[2]


def run_cleanup(dry_run: bool = False, silent: bool = False) -> tuple[int, int]:
    """
    Core cleanup logic - removes stale local and remote branches.

    Returns (deleted_count, skipped_count) tuple.
    """
    local_branches = get_local_branches()
    remote_branches = get_remote_branches()

    # Combine and deduplicate
    all_branches = set(local_branches) | set(remote_branches)

    if not all_branches:
        if not silent:
            typer.echo("No dev-* branches found.")
        return 0, 0

    # Get GitHub repo for remote deletion
    gh_repo = None
    if not dry_run:
        try:
            client = get_github_client()
            repo_owner, repo_name = get_repo_from_git(ENV_SETTINGS.caller_dir)
            gh_repo = client.get_repo(f"{repo_owner}/{repo_name}")
        except Exception:
            pass  # Will skip remote deletion if can't get repo

    deleted_count = 0
    skipped_count = 0

    for branch in sorted(all_branches):
        spec_slug = extract_spec_slug_from_branch(branch)
        if not spec_slug:
            continue

        # Check if spec exists in completed or abandoned
        spec = specs.get_spec(spec_slug)
        if spec is None:
            continue

        status = spec.get("status")
        if status not in ("completed", "abandoned"):
            skipped_count += 1
            continue

        is_local = branch in local_branches
        is_remote = branch in remote_branches

        if dry_run:
            location = []
            if is_local:
                location.append("local")
            if is_remote:
                location.append("remote")
            if not silent:
                typer.echo(
                    f"🔍 Would delete: {branch} ({', '.join(location)}) (spec {status})"
                )
            deleted_count += 1
        else:
            deleted_any = False

            # Delete local branch
            if is_local:
                if delete_local_branch(branch):
                    if not silent:
                        typer.echo(f"🗑️ Deleted local: {branch} (spec {status})")
                    deleted_any = True
                elif not silent:
                    typer.echo(f"❌ Failed to delete local: {branch}")

            # Delete remote branch
            if is_remote and gh_repo:
                if delete_branch(gh_repo, branch):
                    if not silent:
                        typer.echo(f"🗑️ Deleted remote: {branch} (spec {status})")
                    deleted_any = True
                elif not silent:
                    typer.echo(f"❌ Failed to delete remote: {branch}")

            if deleted_any:
                deleted_count += 1

    # Prune remote refs
    if not dry_run and deleted_count > 0:
        prune_remote_refs()
        if not silent:
            typer.echo("\n🧹 Pruned stale remote tracking refs.")

    return deleted_count, skipped_count


def cleanup(
    dry_run: Annotated[
        bool,
        typer.Option(
            "--dry-run", "-n", help="Show what would be deleted without deleting"
        ),
    ] = False,
):
    """
    Remove stale local and remote branches from completed or abandoned specs.

    Scans branches matching 'dev-*' pattern, extracts the spec slug,
    and deletes the branch if the spec is in completed/ or abandoned/ status.
    """
    typer.echo("🔍 Scanning for stale branches...\n")

    deleted_count, skipped_count = run_cleanup(dry_run=dry_run, silent=False)

    if deleted_count == 0 and skipped_count == 0:
        return

    typer.echo(
        f"\n{'🔍 Would delete' if dry_run else '🗑️ Deleted'}: {deleted_count} branch(es)"
    )
    if skipped_count > 0:
        typer.echo(f"⏭️ Skipped: {skipped_count} branch(es) (spec still active)")
