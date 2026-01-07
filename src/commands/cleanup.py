"""
Cleanup command for mem.

Removes stale local branches from completed/abandoned specs.
"""

import subprocess
from typing import Annotated

import typer

from env_settings import ENV_SETTINGS
from src.commands.merge import delete_local_branch, prune_remote_refs
from src.utils import specs


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


def cleanup(
    dry_run: Annotated[
        bool,
        typer.Option(
            "--dry-run", "-n", help="Show what would be deleted without deleting"
        ),
    ] = False,
):
    """
    Remove stale local branches from completed or abandoned specs.

    Scans local branches matching 'dev-*' pattern, extracts the spec slug,
    and deletes the branch if the spec is in completed/ or abandoned/ status.
    """
    typer.echo("Scanning for stale branches...\n")

    branches = get_local_branches()
    if not branches:
        typer.echo("No dev-* branches found.")
        raise typer.Exit(code=0)

    deleted_count = 0
    skipped_count = 0

    for branch in branches:
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

        if dry_run:
            typer.echo(f"Would delete: {branch} (spec {status})")
            deleted_count += 1
        else:
            if delete_local_branch(branch):
                typer.echo(f"Deleted: {branch} (spec {status})")
                deleted_count += 1
            else:
                typer.echo(f"Failed to delete: {branch}")

    # Prune remote refs
    if not dry_run and deleted_count > 0:
        prune_remote_refs()
        typer.echo("\nPruned stale remote tracking refs.")

    typer.echo(
        f"\n{'Would delete' if dry_run else 'Deleted'}: {deleted_count} branch(es)"
    )
    if skipped_count > 0:
        typer.echo(f"Skipped: {skipped_count} branch(es) (spec still active)")
