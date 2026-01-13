"""
Merge command for mem.

Provides subcommands for:
- Merging PRs ready to merge (with [Complete]: prefix)
- Merging between main branches (dev -> test -> main) with back-merging
"""

import subprocess
from typing import Annotated

import typer

from env_settings import ENV_SETTINGS
from src.commands.sync import git_fetch_and_pull
from src.utils import worktrees
from src.utils.github.api import (
    delete_branch,
    list_merge_ready_prs,
    merge_pull_request,
)
from src.utils.github.client import get_github_client
from src.utils.github.repo import get_repo_from_git

app = typer.Typer()


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


@app.callback(invoke_without_command=True)
def merge(
    ctx: typer.Context,
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
    if ctx.invoked_subcommand is not None:
        return

    try:
        # Pre-flight checks
        is_clean, message = check_working_directory_clean()
        if not is_clean:
            typer.echo(f"Error: {message}", err=True)
            raise typer.Exit(code=1)

        typer.echo("🔄 Fetching latest changes...")
        success, message = git_fetch_and_pull()
        if not success:
            typer.echo(f"❌ Error: {message}", err=True)
            raise typer.Exit(code=1)
        typer.echo("✅ Local branch is up to date.\n")

        # Get GitHub client and repo
        client = get_github_client()
        repo_owner, repo_name = get_repo_from_git(ENV_SETTINGS.caller_dir)
        gh_repo = client.get_repo(f"{repo_owner}/{repo_name}")

        typer.echo("🐙 Querying GitHub for merge-ready PRs...\n")

        # Query GitHub for PRs with [Complete]: in title
        prs = list_merge_ready_prs(gh_repo)

        if not prs:
            typer.echo("📭 No PRs ready to merge.")
            typer.echo("\n💡 PRs must have '[Complete]:' in the title to appear here.")
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
            typer.echo("✅ Ready to merge:")
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
            typer.echo("⚠️ Checks failing (use --force to merge anyway):")
            for pr in checks_failing:
                issue_info = (
                    f" (issue #{pr['issue_number']})" if pr["issue_number"] else ""
                )
                typer.echo(f"  - #{pr['number']} {pr['title']}{issue_info}")
            typer.echo()

        if has_conflicts:
            typer.echo("❌ Has conflicts (resolve on GitHub first):")
            for pr in has_conflicts:
                issue_info = (
                    f" (issue #{pr['issue_number']})" if pr["issue_number"] else ""
                )
                typer.echo(f"  - #{pr['number']} {pr['title']}{issue_info}")
            typer.echo()

        # Include failing checks if --force
        if force and checks_failing:
            typer.echo("⚠️ --force: Including PRs with failing checks")
            ready_to_merge.extend(checks_failing)

        if not ready_to_merge:
            typer.echo("📭 No PRs ready to merge.")
            raise typer.Exit(code=0)

        # Dry run - just show what would happen
        if dry_run:
            typer.echo(f"🔍 Dry run: Would merge {len(ready_to_merge)} PR(s)")
            raise typer.Exit(code=0)

        # Select PRs to merge
        if all_ready or len(ready_to_merge) == 1:
            selected = ready_to_merge
            if len(ready_to_merge) == 1:
                typer.echo(f"🔀 Merging PR #{ready_to_merge[0]['number']}...")
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
            typer.echo("📭 No PRs selected.")
            raise typer.Exit(code=0)

        # Merge selected PRs
        typer.echo(f"\n🔀 Merging {len(selected)} PR(s)...\n")

        success_count = 0
        for pr_info in selected:
            typer.echo(f"🔀 Merging #{pr_info['number']}: {pr_info['title']}...")

            # Get full PR object for merge
            pr = gh_repo.get_pull(pr_info["number"])
            result = merge_pull_request(pr, merge_method="rebase")

            if result["success"]:
                typer.echo(f"  ✅ Merged (SHA: {result['sha'][:7]})")
                success_count += 1

                # Delete remote branch
                if delete_branches:
                    branch = pr_info["head_branch"]
                    if delete_branch(gh_repo, branch):
                        typer.echo(f"  🗑️ Deleted remote branch: {branch}")
                    else:
                        typer.echo(
                            f"  ⚠️ Warning: Could not delete remote branch: {branch}"
                        )

                    # Remove worktree first (must happen before branch deletion)
                    spec_slug = extract_spec_slug_from_branch(branch)
                    if spec_slug:
                        try:
                            if worktrees.remove_worktree(
                                ENV_SETTINGS.caller_dir, spec_slug, force=True
                            ):
                                typer.echo(f"  📂 Removed worktree: {spec_slug}")
                        except Exception as e:
                            typer.echo(f"  ⚠️ Warning: Could not remove worktree: {e}")

                    # Delete local branch
                    if delete_local_branch(branch):
                        typer.echo(f"  🗑️ Deleted local branch: {branch}")
                    else:
                        typer.echo(
                            f"  ⚠️ Warning: Could not delete local branch: {branch}"
                        )
            else:
                typer.echo(f"  ❌ Failed: {result['message']}", err=True)

        typer.echo(f"\n✅ Merged {success_count}/{len(selected)} PR(s).")

        # Prune stale remote tracking refs
        if delete_branches and success_count > 0:
            prune_remote_refs()
            typer.echo("🧹 Pruned stale remote tracking refs.")

        # Run sync to update local state
        if not no_sync and success_count > 0:
            typer.echo("\n🔄 Running sync to update local state...")
            from src.commands.sync import sync as run_sync

            run_sync(dry_run=False, no_git=False)

        # Print next step hint
        if success_count > 0:
            typer.echo("\n💡 Next step: mem merge into test")

    except typer.Exit:
        raise
    except Exception as e:
        typer.echo(f"Error: {e}", err=True)
        raise typer.Exit(code=1)


def _get_current_branch() -> str:
    """Get the current git branch name."""
    cwd = ENV_SETTINGS.caller_dir
    result = subprocess.run(
        ["git", "rev-parse", "--abbrev-ref", "HEAD"],
        cwd=cwd,
        capture_output=True,
        text=True,
    )
    return result.stdout.strip()


def _switch_branch(branch: str) -> tuple[bool, str]:
    """Switch to a branch. Returns (success, error_message)."""
    cwd = ENV_SETTINGS.caller_dir
    result = subprocess.run(
        ["git", "checkout", branch],
        cwd=cwd,
        capture_output=True,
        text=True,
    )
    if result.returncode != 0:
        return False, result.stderr.strip()
    return True, ""


def _pull_branch(branch: str) -> tuple[bool, str]:
    """Pull a specific branch from origin. Returns (success, error_message)."""
    cwd = ENV_SETTINGS.caller_dir
    result = subprocess.run(
        ["git", "pull", "--ff-only", "origin", branch],
        cwd=cwd,
        capture_output=True,
        text=True,
    )
    if result.returncode != 0:
        return False, result.stderr.strip()
    return True, ""


def _merge_branch(source: str) -> tuple[bool, str]:
    """Merge source branch into current with ff-only. Returns (success, error_message)."""
    cwd = ENV_SETTINGS.caller_dir
    result = subprocess.run(
        ["git", "merge", "--ff-only", source],
        cwd=cwd,
        capture_output=True,
        text=True,
    )
    if result.returncode != 0:
        return False, result.stderr.strip()
    return True, ""


def _push_branch(branch: str) -> tuple[bool, str]:
    """Push a branch to origin. Returns (success, error_message)."""
    cwd = ENV_SETTINGS.caller_dir
    result = subprocess.run(
        ["git", "push", "origin", branch],
        cwd=cwd,
        capture_output=True,
        text=True,
    )
    if result.returncode != 0:
        return False, result.stderr.strip()
    return True, ""


def _fetch_origin() -> tuple[bool, str]:
    """Fetch from origin. Returns (success, error_message)."""
    cwd = ENV_SETTINGS.caller_dir
    result = subprocess.run(
        ["git", "fetch", "origin"],
        cwd=cwd,
        capture_output=True,
        text=True,
    )
    if result.returncode != 0:
        return False, result.stderr.strip()
    return True, ""


def _print_error_state(
    current_branch: str, original_branch: str, steps_done: list[str]
):
    """Print error state for manual recovery."""
    typer.echo("\n" + "=" * 60, err=True)
    typer.echo("🚨 MERGE FAILED - MANUAL INTERVENTION REQUIRED", err=True)
    typer.echo("=" * 60, err=True)
    typer.echo(f"\nCurrent branch: {current_branch}", err=True)
    typer.echo(f"Original branch: {original_branch}", err=True)
    if steps_done:
        typer.echo("\nCompleted steps:", err=True)
        for step in steps_done:
            typer.echo(f"  ✅ {step}", err=True)
    typer.echo("\nTo recover:", err=True)
    typer.echo(f"  git checkout {original_branch}", err=True)
    typer.echo("=" * 60, err=True)


def _merge_into_test(dry_run: bool = False) -> bool:
    """
    Merge dev into test with fast-forward only.

    Flow:
    1. Check working directory is clean
    2. Fetch latest from remote
    3. Switch to test, pull latest
    4. Merge dev into test (ff-only)
    5. Push test
    6. Switch back to dev

    Returns True on success, False on failure.
    """
    original_branch = _get_current_branch()
    steps_done: list[str] = []

    if dry_run:
        typer.echo("🔍 Dry run: mem merge into test")
        typer.echo("\nWould perform the following steps:")
        typer.echo("  1. Fetch latest from origin")
        typer.echo("  2. Switch to test branch and pull")
        typer.echo("  3. Merge dev into test (fast-forward only)")
        typer.echo("  4. Push test to origin")
        typer.echo("  5. Switch back to dev branch")
        typer.echo("\n💡 Run without --dry-run to execute.")
        return True

    try:
        # 1. Check working directory is clean
        is_clean, message = check_working_directory_clean()
        if not is_clean:
            typer.echo(f"❌ Error: {message}", err=True)
            return False

        # 2. Fetch latest
        typer.echo("🔄 Fetching latest from origin...")
        success, error = _fetch_origin()
        if not success:
            typer.echo(f"❌ Failed to fetch: {error}", err=True)
            return False
        steps_done.append("Fetched from origin")

        # 3. Switch to test and pull
        typer.echo("🌿 Switching to test branch...")
        success, error = _switch_branch("test")
        if not success:
            typer.echo(f"❌ Failed to switch to test: {error}", err=True)
            _print_error_state(_get_current_branch(), original_branch, steps_done)
            return False
        steps_done.append("Switched to test")

        typer.echo("📥 Pulling latest test...")
        success, error = _pull_branch("test")
        if not success:
            typer.echo(f"❌ Failed to pull test: {error}", err=True)
            _print_error_state(_get_current_branch(), original_branch, steps_done)
            return False
        steps_done.append("Pulled test")

        # 4. Merge dev into test (ff-only)
        typer.echo("🔀 Merging dev into test (fast-forward only)...")
        success, error = _merge_branch("dev")
        if not success:
            typer.echo(f"❌ Failed to merge dev into test: {error}", err=True)
            _print_error_state(_get_current_branch(), original_branch, steps_done)
            return False
        steps_done.append("Merged dev into test")

        # 5. Push test
        typer.echo("📤 Pushing test to origin...")
        success, error = _push_branch("test")
        if not success:
            typer.echo(f"❌ Failed to push test: {error}", err=True)
            _print_error_state(_get_current_branch(), original_branch, steps_done)
            return False
        steps_done.append("Pushed test")

        # 6. Switch back to dev
        typer.echo("🌿 Switching back to dev branch...")
        success, error = _switch_branch("dev")
        if not success:
            typer.echo(f"❌ Failed to switch to dev: {error}", err=True)
            _print_error_state(_get_current_branch(), original_branch, steps_done)
            return False
        steps_done.append("Switched to dev")

        typer.echo("\n✅ Successfully merged dev into test!")
        typer.echo("   test is now at the same commit as dev.")
        typer.echo("\n💡 Next step: mem merge into main")
        return True

    except Exception as e:
        typer.echo(f"❌ Unexpected error: {e}", err=True)
        _print_error_state(_get_current_branch(), original_branch, steps_done)
        return False


def _merge_into_main(dry_run: bool = False, force: bool = False) -> bool:
    """
    Merge test into main with fast-forward only.

    By default runs in dry-run mode. Use --force to execute.

    Flow:
    1. Check working directory is clean
    2. Fetch latest from remote
    3. Switch to main, pull latest
    4. Merge test into main (ff-only)
    5. Push main
    6. Switch back to dev

    Returns True on success, False on failure.
    """
    original_branch = _get_current_branch()
    steps_done: list[str] = []

    # Default to dry-run unless --force is specified
    if not force:
        typer.echo("🔍 Dry run: mem merge into main")
        typer.echo("\nWould perform the following steps:")
        typer.echo("  1. Fetch latest from origin")
        typer.echo("  2. Switch to main branch and pull")
        typer.echo("  3. Merge test into main (fast-forward only)")
        typer.echo("  4. Push main to origin")
        typer.echo("  5. Switch back to dev branch")
        typer.echo("\n⚠️  This is a dry run. To execute, run:")
        typer.echo("    mem merge into main --force")
        return True

    try:
        # 1. Check working directory is clean
        is_clean, message = check_working_directory_clean()
        if not is_clean:
            typer.echo(f"❌ Error: {message}", err=True)
            return False

        # 2. Fetch latest
        typer.echo("🔄 Fetching latest from origin...")
        success, error = _fetch_origin()
        if not success:
            typer.echo(f"❌ Failed to fetch: {error}", err=True)
            return False
        steps_done.append("Fetched from origin")

        # 3. Switch to main and pull
        typer.echo("🌿 Switching to main branch...")
        success, error = _switch_branch("main")
        if not success:
            typer.echo(f"❌ Failed to switch to main: {error}", err=True)
            _print_error_state(_get_current_branch(), original_branch, steps_done)
            return False
        steps_done.append("Switched to main")

        typer.echo("📥 Pulling latest main...")
        success, error = _pull_branch("main")
        if not success:
            typer.echo(f"❌ Failed to pull main: {error}", err=True)
            _print_error_state(_get_current_branch(), original_branch, steps_done)
            return False
        steps_done.append("Pulled main")

        # 4. Merge test into main (ff-only)
        typer.echo("🔀 Merging test into main (fast-forward only)...")
        success, error = _merge_branch("test")
        if not success:
            typer.echo(f"❌ Failed to merge test into main: {error}", err=True)
            _print_error_state(_get_current_branch(), original_branch, steps_done)
            return False
        steps_done.append("Merged test into main")

        # 5. Push main
        typer.echo("📤 Pushing main to origin...")
        success, error = _push_branch("main")
        if not success:
            typer.echo(f"❌ Failed to push main: {error}", err=True)
            _print_error_state(_get_current_branch(), original_branch, steps_done)
            return False
        steps_done.append("Pushed main")

        # 6. Switch back to dev
        typer.echo("🌿 Switching back to dev branch...")
        success, error = _switch_branch("dev")
        if not success:
            typer.echo(f"❌ Failed to switch to dev: {error}", err=True)
            _print_error_state(_get_current_branch(), original_branch, steps_done)
            return False
        steps_done.append("Switched to dev")

        typer.echo("\n✅ Successfully merged test into main!")
        typer.echo("   main is now at the same commit as test.")
        return True

    except Exception as e:
        typer.echo(f"❌ Unexpected error: {e}", err=True)
        _print_error_state(_get_current_branch(), original_branch, steps_done)
        return False


@app.command("into")
def into(
    target: Annotated[
        str,
        typer.Argument(help="Target branch to merge into (test or main)"),
    ],
    dry_run: Annotated[
        bool,
        typer.Option(
            "--dry-run", "-n", help="Show what would happen without executing"
        ),
    ] = False,
    force: Annotated[
        bool,
        typer.Option("--force", "-f", help="Execute the merge (required for 'main')"),
    ] = False,
):
    """
    Merge between main branches using fast-forward only.

    This command merges changes forward through the branch hierarchy
    (dev -> test -> main) using fast-forward only merges. This keeps
    branches at the same commit SHA, eliminating GitHub's "X commits
    behind" warnings.

    Usage:
        mem merge into test     Merge dev into test (from dev branch)
        mem merge into main     Merge test into main (dry-run by default)
        mem merge into main --force   Actually execute the merge to main

    The command must be run from the dev branch.
    """
    # Validate target
    target = target.lower()
    if target not in ("test", "main"):
        typer.echo(
            f"❌ Error: Invalid target '{target}'. Must be 'test' or 'main'.", err=True
        )
        raise typer.Exit(code=1)

    # Check we're on dev branch
    current = _get_current_branch()
    if current != "dev":
        typer.echo(
            f"❌ Error: Must be on 'dev' branch to merge. Currently on '{current}'.",
            err=True,
        )
        typer.echo("\n💡 Run: git checkout dev", err=True)
        raise typer.Exit(code=1)

    if target == "test":
        success = _merge_into_test(dry_run=dry_run)
    else:  # main
        success = _merge_into_main(dry_run=dry_run, force=force)

    if not success:
        raise typer.Exit(code=1)
