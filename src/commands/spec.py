"""
Spec command - Manage specifications
"""

from typing import Optional

import typer
from git import Repo
from typing_extensions import Annotated

from env_settings import ENV_SETTINGS
from src.commands.sync import (
    git_commit_and_push,
    git_fetch_and_pull,
    git_has_mem_changes,
)
from src.utils import specs, tasks
from src.utils.github.api import (
    close_issue_with_comment,
    create_pull_request,
    update_github_issue,
)
from src.utils.github.client import get_authenticated_user, get_github_client
from src.utils.github.exceptions import GitHubError
from src.utils.github.git_ops import push_branch, smart_switch, switch_to_branch
from src.utils.github.repo import get_repo_from_git
from src.utils.markdown import slugify

app = typer.Typer()


@app.command()
def new(
    title: Annotated[str, typer.Argument(help="Title of the specification")],
):
    """
    Create a new specification.

    This creates a spec directory with spec.md in .mem/specs/{slug}/
    """
    try:
        spec_file = specs.create_spec(title)
        slug = slugify(title)
        relative_path = spec_file.relative_to(ENV_SETTINGS.caller_dir)

        typer.echo(f"✓ Created spec: {relative_path}")
        typer.echo("\n✨ Spec created successfully!")
        typer.echo("\nNext steps:")
        typer.echo(f"  1. Edit the spec file: {relative_path}")
        typer.echo("  2. Run 'mem sync' to create the GitHub issue")
        typer.echo(f"  3. Run 'mem spec assign {slug}' to assign yourself")
        typer.echo(f"  4. Run 'mem spec activate {slug}' to start working")
        typer.echo("")
        typer.echo("Note: Unassigned specs cannot be activated to prevent conflicts.")

    except ValueError as e:
        typer.echo(f"❌ Error: {e}", err=True)
        raise typer.Exit(code=1)
    except Exception as e:
        typer.echo(f"❌ Error: {e}", err=True)
        raise typer.Exit(code=1)


@app.command("list")
def list_specs_cmd(
    status: Annotated[
        Optional[str],
        typer.Option(
            "--status",
            "-s",
            help="Filter by status (todo, merge_ready, completed, abandoned)",
        ),
    ] = None,
):
    """
    List specifications.

    By default, shows 'todo' and 'merge_ready' specs.
    Active spec is determined by current git branch, not status field.
    """
    try:
        # Get active spec based on current branch
        active_spec = specs.get_active_spec()

        if status:
            spec_list = specs.list_specs(status)
            display_status = status.upper()
        else:
            # Show todo and merge_ready by default
            todo_specs = specs.list_specs("todo")
            merge_ready_specs = specs.list_specs("merge_ready")
            spec_list = todo_specs + merge_ready_specs
            display_status = "TODO & MERGE_READY"

        if not spec_list:
            typer.echo(f"No {display_status.lower()} specs found.")
            return

        # Show active spec info
        if active_spec:
            typer.echo(
                f"\nActive spec (on branch {active_spec.get('branch')}): {active_spec['slug']}"
            )

        # Display specs in a formatted table
        typer.echo(f"\n{display_status} SPECS:\n")
        typer.echo(f"{'Slug':<30} {'Title':<30} {'Status':<12} {'Branch':<25}")
        typer.echo("=" * 100)

        for spec in spec_list:
            slug = spec["slug"]
            title = spec["title"]
            status_text = spec["status"]
            branch = spec.get("branch") or "N/A"

            # Mark active spec
            is_active = active_spec and active_spec["slug"] == slug
            active_marker = " *" if is_active else ""

            # Truncate values for display
            display_slug = slug[:27] + "..." if len(slug) > 30 else slug
            display_title = title[:27] + "..." if len(title) > 30 else title
            display_status_text = (
                status_text[:10] + "..." if len(status_text) > 12 else status_text
            )
            display_branch = branch[:22] + "..." if len(branch) > 25 else branch

            typer.echo(
                f"{display_slug:<30} {display_title:<30} {display_status_text:<12} {display_branch:<25}{active_marker}"
            )

        typer.echo(f"\nTotal: {len(spec_list)} spec(s)")
        if active_spec:
            typer.echo("(* = currently active)")
        typer.echo("\nTo view details: mem spec show <slug>")
        typer.echo("To activate: mem spec activate <slug>")

    except Exception as e:
        typer.echo(f"❌ Error: {e}", err=True)
        raise typer.Exit(code=1)


@app.command()
def show(
    spec_slug: Annotated[
        Optional[str], typer.Argument(help="Slug of the specification to show")
    ] = None,
    verbose: Annotated[
        bool, typer.Option("--verbose", "-v", help="Show markdown content")
    ] = False,
):
    """
    Show detailed information about a specification.

    If no slug is provided, shows the currently active specification.
    """
    try:
        if spec_slug is None:
            spec = specs.get_active_spec()
            if not spec:
                typer.echo(
                    "❌ Error: No spec slug provided and no spec is currently active.",
                    err=True,
                )
                raise typer.Exit(code=1)
            spec_slug = spec["slug"]
        else:
            spec = specs.get_spec(spec_slug)

        if not spec:
            typer.echo(f"❌ Error: Spec '{spec_slug}' not found.", err=True)
            raise typer.Exit(code=1)

        typer.echo(f"\nSPECIFICATION: {spec['title']}")
        typer.echo("=" * 60)
        typer.echo(f"Slug:         {spec['slug']}")
        typer.echo(f"Status:       {spec['status']}")
        typer.echo(f"Assignee:     {spec.get('assigned_to') or 'Unassigned'}")
        typer.echo(f"Branch:       {spec.get('branch') or 'N/A'}")
        typer.echo(f"GitHub Issue: {spec.get('issue_id') or 'N/A'}")
        if spec.get("issue_url"):
            typer.echo(f"Issue URL:    {spec['issue_url']}")
        if spec.get("pr_url"):
            typer.echo(f"PR URL:       {spec['pr_url']}")
        typer.echo(f"File Path:    .mem/specs/{spec['slug']}/spec.md")
        typer.echo(f"Created:      {spec.get('created_at', 'N/A')}")
        typer.echo(f"Updated:      {spec.get('updated_at', 'N/A')}")

        if verbose:
            typer.echo("\nCONTENT:")
            typer.echo("-" * 60)
            typer.echo(spec.get("body", "").strip())
            typer.echo("-" * 60)

        # Show tasks
        task_list = tasks.list_tasks(spec_slug)

        if task_list:
            typer.echo("\nTASKS:")
            typer.echo("-" * 60)
            for task in task_list:
                status_icon = "x" if task["status"] == "completed" else " "
                typer.echo(f"[{status_icon}] {task['title']}")

                # Show subtasks (now embedded in task frontmatter)
                subtask_list = task.get("subtasks", [])
                for subtask in subtask_list:
                    sub_icon = "x" if subtask["status"] == "completed" else " "
                    typer.echo(f"    [{sub_icon}] {subtask['title']}")
        else:
            typer.echo("\nNo tasks associated with this spec.")

        typer.echo("\nCommands:")
        typer.echo(f'  mem task new "title" "description" --spec {spec_slug}')
        typer.echo(f"  mem spec activate {spec_slug}")

    except Exception as e:
        typer.echo(f"❌ Error: {e}", err=True)
        raise typer.Exit(code=1)


@app.command()
def assign(
    spec_slug: Annotated[str, typer.Argument(help="Slug of the specification")],
    username: Annotated[
        Optional[str],
        typer.Argument(help="GitHub username to assign to (defaults to current user)"),
    ] = None,
):
    """
    Assign a specification to a GitHub user.

    If no username is provided, assigns to the current authenticated user.
    The assignment is synced to GitHub to prevent multiple people working
    on the same spec simultaneously.
    """
    try:
        spec = specs.get_spec(spec_slug)
        if not spec:
            typer.echo(f"❌ Error: Spec '{spec_slug}' not found.", err=True)
            raise typer.Exit(code=1)

        # Get current user if no username provided
        if not username:
            try:
                client = get_github_client()
                username = get_authenticated_user(client)["username"]
            except Exception as e:
                typer.echo(f"❌ Error: Could not get current user: {e}", err=True)
                raise typer.Exit(code=1)

        # Check if spec is synced to GitHub
        if not spec.get("issue_id"):
            typer.echo(
                f"❌ Error: Spec '{spec_slug}' is not synced to GitHub.", err=True
            )
            typer.echo("\nRun 'mem sync' first to create the GitHub issue.")
            raise typer.Exit(code=1)

        # Check if already assigned to someone else
        current_assignee = spec.get("assigned_to")
        if current_assignee and current_assignee != username:
            typer.echo(
                f"❌ Error: Spec is already assigned to '{current_assignee}'.", err=True
            )
            typer.echo(
                "\nSpecs can only be reassigned by the current assignee or repo admin."
            )
            raise typer.Exit(code=1)

        # Update local assignment
        specs.assign_spec(spec_slug, username)

        # Sync assignment to GitHub
        try:
            client = get_github_client()
            repo_owner, repo_name = get_repo_from_git(ENV_SETTINGS.caller_dir)
            repo = client.get_repo(f"{repo_owner}/{repo_name}")

            update_github_issue(repo, spec["issue_id"], assignees=[username])
            typer.echo(f"✓ Spec '{spec['title']}' assigned to {username}")
            typer.echo("✓ Assignment synced to GitHub")
        except Exception as e:
            typer.echo(
                f"⚠️  Warning: Could not sync assignment to GitHub: {e}", err=True
            )
            typer.echo("  Local assignment saved. Run 'mem sync' to retry.")

    except typer.Exit:
        raise
    except Exception as e:
        typer.echo(f"❌ Error: {e}", err=True)
        raise typer.Exit(code=1)


@app.command()
def activate(
    spec_slug: Annotated[
        str, typer.Argument(help="Slug of the specification to activate")
    ],
):
    """
    Activate a specification by switching to its branch.

    If the spec doesn't have a branch yet, creates one and switches to it.
    A spec is considered "active" when you're on its branch.
    """
    try:
        # 1. Get spec info
        spec = specs.get_spec(spec_slug)

        if not spec:
            typer.echo(f"Error: Spec '{spec_slug}' not found.", err=True)
            raise typer.Exit(code=1)

        title = spec["title"]

        # 2. Check if spec is synced to GitHub
        if not spec.get("issue_id"):
            typer.echo(f"Error: Spec '{spec_slug}' is not synced to GitHub.", err=True)
            typer.echo("\nTo activate a spec, it must first be synced:")
            typer.echo("  1. Edit the spec file if needed")
            typer.echo("  2. Run 'mem sync' to create the GitHub issue")
            typer.echo("  3. Run 'mem spec assign <slug>' to assign yourself")
            typer.echo("  4. Then run 'mem spec activate <slug>'")
            raise typer.Exit(code=1)

        # 3. Check if spec is assigned
        if not spec.get("assigned_to"):
            typer.echo(
                f"Error: Spec '{spec_slug}' is not assigned to anyone.", err=True
            )
            typer.echo("\nTo activate a spec, it must be assigned to you:")
            typer.echo("  Run 'mem spec assign <slug>' to assign yourself")
            raise typer.Exit(code=1)

        # 4. Check if assigned to current user
        try:
            client = get_github_client()
            current_gh_user = get_authenticated_user(client)["username"]
            if spec.get("assigned_to") != current_gh_user:
                typer.echo(
                    f"Error: Spec '{spec_slug}' is assigned to '{spec.get('assigned_to')}', not you.",
                    err=True,
                )
                typer.echo("\nYou can only activate specs assigned to you.")
                raise typer.Exit(code=1)
        except Exception as e:
            if "assigned to" in str(e).lower():
                raise
            # If we can't verify, proceed with warning
            typer.echo(f"Warning: Could not verify GitHub user: {e}", err=True)

        # 5. Check if already on this spec's branch
        current_branch = specs.get_current_branch()
        if spec.get("branch") and current_branch == spec.get("branch"):
            typer.echo(f"Already on spec '{spec_slug}' branch: {current_branch}")
            return

        # 6. Prepare branch name
        branch_name = spec.get("branch")
        if not branch_name:
            # Create a new branch name for this spec
            user_slug = slugify(current_gh_user)
            branch_name = f"dev-{user_slug}-{spec_slug}"

        typer.echo(f"Activating spec: {title}")
        typer.echo(f"Branch: {branch_name}")

        # 7. Git operations
        try:
            is_new = smart_switch(ENV_SETTINGS.caller_dir, branch_name)
            if is_new:
                push_branch(ENV_SETTINGS.caller_dir, branch_name)
                typer.echo(f"Created and switched to branch '{branch_name}'")
            else:
                typer.echo(f"Switched to existing branch '{branch_name}'")
        except Exception as e:
            typer.echo(f"\nFailed to switch to branch '{branch_name}': {e}", err=True)
            typer.echo("Please manually resolve the git issue and try again.", err=True)
            raise typer.Exit(code=1)

        # 8. Update spec with branch name (if it was new)
        if not spec.get("branch"):
            specs.update_spec(spec_slug, branch=branch_name)

        typer.echo(f"\nSpec '{spec_slug}' is now active.")

    except GitHubError as e:
        typer.echo(f"GitHub Error: {e}", err=True)
        raise typer.Exit(code=1)
    except Exception as e:
        if not isinstance(e, typer.Exit):
            typer.echo(f"Error: {e}", err=True)
        raise typer.Exit(code=1)


@app.command()
def complete(
    spec_slug: Annotated[
        str, typer.Argument(help="Slug of the specification to complete")
    ],
    message: Annotated[str, typer.Argument(help="Commit message for the final push")],
):
    """
    Complete a specification.

    This command:
    1. Pulls latest changes from remote.
    2. Validates that all tasks are completed.
    3. Commits and pushes all changes.
    4. Creates a Pull Request on GitHub.
    5. Marks the spec as 'merge_ready'.
    6. Switches back to the 'dev' branch.
    """
    try:
        # 0. Pull latest changes first
        typer.echo("Pulling latest changes...")
        success, pull_msg = git_fetch_and_pull()
        if not success:
            typer.echo(f"Error: {pull_msg}", err=True)
            typer.echo("Please resolve conflicts and try again.", err=True)
            raise typer.Exit(code=1)

        # 1. Get spec info
        spec = specs.get_spec(spec_slug)
        if not spec:
            typer.echo(f"Error: Spec '{spec_slug}' not found.", err=True)
            raise typer.Exit(code=1)

        # 2. Check if on spec's branch (i.e., spec is active)
        current_branch = specs.get_current_branch()
        spec_branch = spec.get("branch")
        if not spec_branch or current_branch != spec_branch:
            typer.echo(
                "Error: You must be on the spec's branch to complete it.",
                err=True,
            )
            typer.echo(f"Current branch: {current_branch}", err=True)
            typer.echo(f"Spec branch: {spec_branch or 'not set'}", err=True)
            typer.echo(f"\nRun: mem spec activate {spec_slug}", err=True)
            raise typer.Exit(code=1)

        # 3. Validate tasks
        task_list = tasks.list_tasks(spec_slug)
        incomplete_tasks = []
        for task in task_list:
            if task["status"] != "completed":
                incomplete_tasks.append(task)
            # Check subtasks (now embedded in frontmatter)
            subtask_list = task.get("subtasks", [])
            incomplete_tasks.extend(
                [
                    {"title": s["title"], "status": s["status"]}
                    for s in subtask_list
                    if s["status"] != "completed"
                ]
            )

        if incomplete_tasks:
            typer.echo(
                f"Error: Cannot complete spec '{spec_slug}'. There are incomplete tasks:",
                err=True,
            )
            for t in incomplete_tasks:
                typer.echo(f"  - {t['title']} ({t['status']})", err=True)
            raise typer.Exit(code=1)

        # 4. Mark spec as merge_ready before committing
        typer.echo(f"Completing spec: {spec['title']}...")
        specs.update_spec_status(spec_slug, "merge_ready")

        # 5. Git operations
        repo = Repo(ENV_SETTINGS.caller_dir)
        branch_name = spec.get("branch")

        if not branch_name:
            typer.echo("Error: No branch associated with this spec.", err=True)
            raise typer.Exit(code=1)

        typer.echo("Committing and pushing changes...")
        repo.git.add(A=True)
        try:
            repo.git.commit("-m", message)
        except Exception as e:
            # If nothing to commit, continue
            if "nothing to commit" not in str(e).lower():
                raise e

        repo.git.push("origin", branch_name)

        # 6. GitHub PR
        pr_url = None
        if spec.get("issue_id"):
            typer.echo("Creating Pull Request...")
            client = get_github_client()
            repo_owner, repo_name = get_repo_from_git(ENV_SETTINGS.caller_dir)
            gh_repo = client.get_repo(f"{repo_owner}/{repo_name}")

            pr_body = f"This PR completes the specification: {spec['title']}\n\nCloses #{spec['issue_id']}"
            try:
                pr = create_pull_request(
                    repo=gh_repo,
                    title=f"[Complete]: {spec['title']}",
                    body=pr_body,
                    head=branch_name,
                    base="dev",
                )
                pr_url = pr.html_url
                specs.update_spec_pr_url(spec_slug, pr_url)
                typer.echo(f"Created Pull Request: {pr_url}")

                # Commit and push the PR URL update
                repo.git.add(A=True)
                try:
                    repo.git.commit("-m", f"Add PR URL for {spec_slug}")
                    repo.git.push("origin", branch_name)
                except Exception as e:
                    if "nothing to commit" not in str(e).lower():
                        raise e
            except Exception as e:
                typer.echo(f"Warning: Could not create Pull Request: {e}", err=True)
        else:
            typer.echo("Warning: Spec has no GitHub issue. Skipping PR creation.")

        # 7. Switch back to dev
        try:
            switch_to_branch(ENV_SETTINGS.caller_dir, "dev")
            typer.echo("Switched back to 'dev' branch")
        except Exception as e:
            typer.echo(f"Warning: Could not switch to 'dev' branch: {e}", err=True)

        typer.echo(f"\nSpec '{spec_slug}' marked as MERGE READY.")
        if pr_url:
            typer.echo(f"PR: {pr_url}")
        typer.echo("\nNext steps: Merge the PR on GitHub.")

    except Exception as e:
        if not isinstance(e, typer.Exit):
            typer.echo(f"Error: {e}", err=True)
        raise typer.Exit(code=1)


@app.command()
def deactivate():
    """
    Deactivate the currently active specification.

    This switches back to the 'dev' branch. The spec remains in 'todo' status
    since activation is now branch-based.
    """
    try:
        # 1. Find active spec
        active_spec = specs.get_active_spec()
        if not active_spec:
            typer.echo("No spec is currently active.")
            return

        typer.echo(f"Stopping work on: {active_spec['title']} ({active_spec['slug']})")

        # 2. Git operations - switch back to dev
        try:
            switch_to_branch(ENV_SETTINGS.caller_dir, "dev")
            typer.echo("Switched back to 'dev' branch")
        except Exception as e:
            typer.echo(f"Warning: Could not switch to 'dev' branch: {e}", err=True)
            typer.echo("Please switch branches manually.", err=True)
            raise typer.Exit(code=1)

        typer.echo("Spec deactivated.")

    except Exception as e:
        if not isinstance(e, typer.Exit):
            typer.echo(f"Error: {e}", err=True)
        raise typer.Exit(code=1)


@app.command()
def abandon(
    spec_slug: Annotated[
        str, typer.Argument(help="Slug of the specification to abandon")
    ],
    reason: Annotated[
        str,
        typer.Option("--reason", "-r", help="Reason for abandoning the spec"),
    ] = "Spec abandoned.",
):
    """
    Abandon a specification.

    This command:
    1. Moves the spec to .mem/specs/abandoned/
    2. Closes the linked GitHub issue with a comment
    3. Switches back to 'dev' branch if this was the active spec
    """
    try:
        # 1. Get spec info
        spec = specs.get_spec(spec_slug)
        if not spec:
            typer.echo(f"Error: Spec '{spec_slug}' not found.", err=True)
            raise typer.Exit(code=1)

        typer.echo(f"Abandoning spec: {spec['title']} ({spec_slug})")

        # 2. If this is the active spec, switch to dev first
        active_spec = specs.get_active_spec()
        if active_spec and active_spec["slug"] == spec_slug:
            typer.echo("Switching away from active spec...")
            try:
                switch_to_branch(ENV_SETTINGS.caller_dir, "dev")
                typer.echo("Switched to 'dev' branch")
            except Exception as e:
                typer.echo(f"Warning: Could not switch to 'dev' branch: {e}", err=True)

        # 3. Close GitHub issue if linked
        if spec.get("issue_id"):
            typer.echo("Closing GitHub issue...")
            try:
                client = get_github_client()
                repo_owner, repo_name = get_repo_from_git(ENV_SETTINGS.caller_dir)
                gh_repo = client.get_repo(f"{repo_owner}/{repo_name}")

                comment = f"**Spec Abandoned**\n\n{reason}"
                close_issue_with_comment(gh_repo, spec["issue_id"], comment)
                typer.echo(f"Closed issue #{spec['issue_id']}")
            except Exception as e:
                typer.echo(f"Warning: Could not close GitHub issue: {e}", err=True)

        # 4. Move spec to abandoned directory
        new_path = specs.move_spec_to_abandoned(spec_slug)
        typer.echo(f"Moved to {new_path.relative_to(ENV_SETTINGS.caller_dir)}")

        typer.echo(f"\nSpec abandoned. Moved to .mem/specs/abandoned/{spec_slug}/")
        typer.echo("To view abandoned specs: mem spec list --status abandoned")

    except Exception as e:
        if not isinstance(e, typer.Exit):
            typer.echo(f"Error: {e}", err=True)
        raise typer.Exit(code=1)


if __name__ == "__main__":
    app()
