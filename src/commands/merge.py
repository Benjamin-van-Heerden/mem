"""
Merge command for mem.

Lists PRs ready to merge, allows selection, and performs rebase merges.
"""

from typing import Annotated, Optional

import typer

from env_settings import ENV_SETTINGS
from src.utils import specs
from src.utils.github.api import (
    delete_branch,
    get_pr_mergeable_status,
    merge_pull_request,
)
from src.utils.github.client import get_github_client
from src.utils.github.repo import get_repo_from_git


def merge(
    spec_slug: Annotated[
        Optional[str],
        typer.Argument(help="Specific spec slug to merge (optional)"),
    ] = None,
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

    Lists all PRs from specs with 'merge_ready' status, checks their
    mergeability, and allows you to select which ones to merge.

    Uses rebase merge strategy for clean linear history.
    """
    try:
        # Get GitHub client and repo
        client = get_github_client()
        repo_owner, repo_name = get_repo_from_git(ENV_SETTINGS.caller_dir)
        gh_repo = client.get_repo(f"{repo_owner}/{repo_name}")

        # Get all merge_ready specs
        merge_ready_specs = specs.list_specs(status="merge_ready")

        if not merge_ready_specs:
            typer.echo("No specs with 'merge_ready' status found.")
            raise typer.Exit(code=0)

        # Filter to specific spec if provided
        if spec_slug:
            merge_ready_specs = [s for s in merge_ready_specs if s["slug"] == spec_slug]
            if not merge_ready_specs:
                typer.echo(
                    f"Spec '{spec_slug}' not found or not in 'merge_ready' status."
                )
                raise typer.Exit(code=1)

        # Check PR status for each spec
        ready_to_merge = []
        has_conflicts = []
        already_merged = []
        no_pr = []

        typer.echo("Checking PR status...\n")

        for spec in merge_ready_specs:
            pr_url = spec.get("pr_url")
            if not pr_url:
                no_pr.append(spec)
                continue

            status = get_pr_mergeable_status(gh_repo, pr_url)

            if not status["exists"]:
                no_pr.append(spec)
            elif status["merged"]:
                already_merged.append({"spec": spec, "status": status})
            elif status["mergeable"] is True and status["mergeable_state"] == "clean":
                ready_to_merge.append({"spec": spec, "status": status})
            elif status["mergeable"] is False or status["mergeable_state"] in (
                "dirty",
                "blocked",
            ):
                has_conflicts.append({"spec": spec, "status": status})
            else:
                # mergeable is None or state is 'behind' - might be okay
                # GitHub is still computing or needs update
                if status["mergeable_state"] == "behind":
                    # Behind but no conflicts - can still merge
                    ready_to_merge.append({"spec": spec, "status": status})
                else:
                    # Unknown state, treat as not ready
                    has_conflicts.append({"spec": spec, "status": status})

        # Display status
        if ready_to_merge:
            typer.echo("Ready to merge:")
            for i, item in enumerate(ready_to_merge, 1):
                spec = item["spec"]
                pr_url = spec.get("pr_url", "")
                pr_num = pr_url.split("/")[-1] if pr_url else "?"
                state = item["status"]["mergeable_state"]
                state_info = f" (behind)" if state == "behind" else ""
                typer.echo(
                    f'  {i}. [PR #{pr_num}] {spec["slug"]} - "{spec["title"]}"{state_info}'
                )
            typer.echo()

        if has_conflicts:
            typer.echo("Has conflicts (resolve on GitHub):")
            for item in has_conflicts:
                spec = item["spec"]
                pr_url = spec.get("pr_url", "")
                pr_num = pr_url.split("/")[-1] if pr_url else "?"
                state = item["status"]["mergeable_state"]
                typer.echo(f"  - [PR #{pr_num}] {spec['slug']} - {state}")
            typer.echo()

        if already_merged:
            typer.echo("Already merged (will update local status):")
            for item in already_merged:
                spec = item["spec"]
                typer.echo(f"  - {spec['slug']}")
            typer.echo()

        if no_pr:
            typer.echo("No PR found:")
            for spec in no_pr:
                typer.echo(f"  - {spec['slug']}")
            typer.echo()

        # Handle already merged specs
        for item in already_merged:
            spec = item["spec"]
            typer.echo(f"Moving '{spec['slug']}' to completed (already merged)...")
            if not dry_run:
                specs.move_spec_to_completed(spec["slug"])
                if delete_branches and spec.get("branch"):
                    if delete_branch(gh_repo, spec["branch"]):
                        typer.echo(f"  Deleted remote branch: {spec['branch']}")

        if not ready_to_merge:
            if not already_merged:
                typer.echo("No PRs ready to merge.")
            raise typer.Exit(code=0)

        # Select PRs to merge
        if dry_run:
            typer.echo("Dry run - no changes will be made.")
            raise typer.Exit(code=0)

        if all_ready:
            selected = ready_to_merge
        else:
            typer.echo(
                "Select PRs to merge (comma-separated numbers, 'all', or 'q' to quit):"
            )
            selection = typer.prompt("Selection", default="q")

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
        for item in selected:
            spec = item["spec"]
            pr = item["status"]["pr"]
            typer.echo(f"Merging '{spec['slug']}'...")

            result = merge_pull_request(pr, merge_method="rebase")

            if result["success"]:
                typer.echo(f"  Merged successfully (SHA: {result['sha'][:7]})")
                success_count += 1

                # Move spec to completed
                specs.move_spec_to_completed(spec["slug"])
                typer.echo(f"  Moved to completed/")

                # Delete remote branch
                if delete_branches and spec.get("branch"):
                    if delete_branch(gh_repo, spec["branch"]):
                        typer.echo(f"  Deleted remote branch: {spec['branch']}")
                    else:
                        typer.echo(
                            f"  Warning: Could not delete branch: {spec['branch']}"
                        )
            else:
                typer.echo(f"  Failed: {result['message']}", err=True)

        typer.echo(f"\nDone. {success_count}/{len(selected)} PRs merged successfully.")

    except typer.Exit:
        raise
    except Exception as e:
        typer.echo(f"Error: {e}", err=True)
        raise typer.Exit(code=1)
