"""
Spec command - Manage specifications
"""

from datetime import datetime, timedelta
from typing import Optional

import typer
from git import Repo
from typing_extensions import Annotated

from env_settings import ENV_SETTINGS
from src.commands.sync import (
    git_fetch_and_pull,
)
from src.utils import logs, specs, tasks, worktrees
from src.utils.github.api import (
    close_issue_with_comment,
    close_pull_request,
    create_pull_request,
    sync_status_labels,
    update_github_issue,
)
from src.utils.github.client import get_authenticated_user, get_github_client
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
    Use 'mem sync' to create the GitHub issue, then 'mem spec assign' to
    claim the spec and create a worktree for working on it.
    """
    try:
        spec_file = specs.create_spec(title)
        slug = slugify(title)
        relative_path = spec_file.relative_to(ENV_SETTINGS.caller_dir)

        typer.echo(f"✅ Created spec: {relative_path}")
        typer.echo("\n✨ Spec created successfully!")
        typer.echo("\n💡 Next steps:")
        typer.echo(f"  1. Edit the spec file: {relative_path}")
        typer.echo("  2. Run 'mem sync' to create the GitHub issue")
        typer.echo(f'  3. Add tasks: mem task new "title" "description" --spec {slug}')
        typer.echo(
            f"  4. Run 'mem spec assign {slug}' to claim it and create a worktree"
        )
        typer.echo("")
        typer.echo("─" * 60)
        typer.echo("🛑 IMPORTANT: Worktree Workflow")
        typer.echo("─" * 60)
        typer.echo("")
        typer.echo("Create tasks BEFORE running 'mem spec assign'.")
        typer.echo("After assignment, start a NEW agent session in the worktree")
        typer.echo("to do the implementation work.")

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
                f"\n📋 Active spec (on branch {active_spec.get('branch')}): {active_spec['slug']}"
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

        typer.echo(f"\n📊 Total: {len(spec_list)} spec(s)")
        if active_spec:
            typer.echo("(* = currently active)")
        typer.echo("\n💡 To view details: mem spec show <slug>")
        typer.echo("💡 To activate: mem spec activate <slug>")

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

        typer.echo(f"\n📋 SPECIFICATION: {spec['title']}")
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
            typer.echo("\n✏️ TASKS:")
            typer.echo("-" * 60)
            for task in task_list:
                status_display = (
                    "[completed]" if task["status"] == "completed" else "[todo]"
                )
                typer.echo(f"{status_display} {task['title']}")

                if verbose:
                    # Show full description
                    body = task.get("body", "").strip()
                    if body:
                        typer.echo(f"       {body}")

                    # Show subtasks in detail
                    subtask_list = task.get("subtasks", [])
                    if subtask_list:
                        for sub in subtask_list:
                            sub_icon = "[x]" if sub["status"] == "completed" else "[ ]"
                            typer.echo(f"       {sub_icon} {sub['title']}")

                    # Show created date
                    created = task.get("created_at", "")
                    if created:
                        typer.echo(f"       Created: {created[:10]}")
                    typer.echo()
                else:
                    # Show subtasks inline (simple view)
                    subtask_list = task.get("subtasks", [])
                    for subtask in subtask_list:
                        sub_icon = "x" if subtask["status"] == "completed" else " "
                        typer.echo(f"    [{sub_icon}] {subtask['title']}")
        else:
            typer.echo("\nNo tasks associated with this spec.")

        typer.echo("\n💡 Commands:")
        typer.echo(
            f'  mem task new "title" "detailed description with implementation notes if necessary" --spec {spec_slug}'
        )
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
    Assign a specification to a GitHub user and create a worktree.

    If no username is provided, assigns to the current authenticated user.
    Creates a git worktree at ../<project>-worktrees/<slug>/ with a feature branch.
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

        # Check if worktree already exists
        main_repo_path = ENV_SETTINGS.caller_dir
        existing_worktree = worktrees.get_worktree_for_spec(main_repo_path, spec_slug)
        if existing_worktree:
            typer.echo(f"📂 Spec '{spec_slug}' already has a worktree at:")
            typer.echo(f"  {existing_worktree.path}")
            typer.echo("\n💡 To work on this spec, open a terminal there.")
            return

        # Create branch name
        user_slug = slugify(username)
        branch_name = spec.get("branch") or f"dev-{user_slug}-{spec_slug}"

        # Update local assignment and branch BEFORE creating worktree
        # so the worktree has the correct metadata
        specs.assign_spec(spec_slug, username)
        if not spec.get("branch"):
            specs.update_spec(spec_slug, branch=branch_name)

        # Commit and push any uncommitted .mem changes before creating worktree
        # This ensures the spec file, tasks, etc. are available in the worktree
        repo = Repo(main_repo_path)
        mem_dir = main_repo_path / ".mem"
        if mem_dir.exists():
            repo.git.add(str(mem_dir))
            if repo.is_dirty(index=True):
                typer.echo("📦 Committing .mem/ changes...")
                repo.git.commit("-m", f"mem: prepare spec {spec_slug} for assignment")
                typer.echo("🔄 Pushing to remote...")
                repo.git.push("origin", repo.active_branch.name)

        try:
            worktree_path = worktrees.create_worktree(
                main_repo_path, spec_slug, branch_name
            )
            typer.echo(f"📂 Created worktree: {worktree_path}")
            typer.echo(f"🌿 Created branch: {branch_name}")
        except Exception as e:
            typer.echo(f"❌ Error creating worktree: {e}", err=True)
            raise typer.Exit(code=1)

        # Sync assignment to GitHub
        try:
            client = get_github_client()
            repo_owner, repo_name = get_repo_from_git(ENV_SETTINGS.caller_dir)
            repo = client.get_repo(f"{repo_owner}/{repo_name}")

            update_github_issue(repo, spec["issue_id"], assignees=[username])
            typer.echo(f"✅ Spec '{spec['title']}' assigned to {username}")
            typer.echo("🐙 Assignment synced to GitHub")
        except Exception as e:
            typer.echo(
                f"⚠️  Warning: Could not sync assignment to GitHub: {e}", err=True
            )
            typer.echo("  Local assignment saved. Run 'mem sync' to retry.")

        typer.echo("\n" + "=" * 60)
        typer.echo("📂 WORKTREE READY - START NEW SESSION")
        typer.echo("=" * 60)
        typer.echo("")
        typer.echo("🛑 THIS SESSION MUST END HERE")
        typer.echo("")
        typer.echo(
            "💡 To work on this spec, start a NEW agent session in the worktree:"
        )
        typer.echo(f"  cd {worktree_path}")
        typer.echo("  claude  # or your preferred agent")
        typer.echo("")
        typer.echo("─" * 60)
        typer.echo("❓ WHY A NEW SESSION?")
        typer.echo("─" * 60)
        typer.echo("")
        typer.echo("• The worktree is an isolated directory with its own branch")
        typer.echo("• This main repo session cannot access the worktree's files")
        typer.echo("• Continuing here would create tasks in the wrong location")

    except typer.Exit:
        raise
    except Exception as e:
        typer.echo(f"❌ Error: {e}", err=True)
        raise typer.Exit(code=1)


@app.command()
def complete(
    spec_slug: Annotated[
        str, typer.Argument(help="Slug of the specification to complete")
    ],
    message: Annotated[str, typer.Argument(help="Commit message for the final push")],
    no_log: Annotated[
        bool,
        typer.Option(
            "--no-log",
            help="Skip the recent work log timing check (not recommended)",
        ),
    ] = False,
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
        typer.echo("🔄 Pulling latest changes...")
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

        # 2. Check if this spec is active (in worktree or on branch)
        active_spec = specs.get_active_spec()
        if not active_spec or active_spec["slug"] != spec_slug:
            typer.echo(
                f"Error: Spec '{spec_slug}' is not currently active.",
                err=True,
            )
            if active_spec:
                typer.echo(f"Active spec: {active_spec['slug']}", err=True)
            else:
                typer.echo("No spec is currently active.", err=True)
            typer.echo("\nTo complete this spec, work from its worktree.", err=True)
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

        # 4. Validate work logs exist (unless --no-log)
        spec_logs = logs.list_logs(limit=100, spec_slug=spec_slug)
        if not no_log and not spec_logs:
            typer.echo(
                f"Error: Cannot complete spec '{spec_slug}'. No work logs found.",
                err=True,
            )
            typer.echo(
                "\nAt least one work log is required before completing a spec.",
                err=True,
            )
            typer.echo("Create a work log with: mem log", err=True)
            raise typer.Exit(code=1)

        # 5. Validate recent work log (within last 3 minutes)
        if not no_log:
            now = datetime.now()
            recent_threshold = now - timedelta(minutes=3)
            recent_log_found = False

            for log in spec_logs:
                log_created_at = log.get("created_at")
                if log_created_at:
                    try:
                        log_datetime = datetime.fromisoformat(log_created_at)
                        if log_datetime >= recent_threshold:
                            recent_log_found = True
                            break
                    except (ValueError, TypeError):
                        continue

            if not recent_log_found:
                typer.echo(
                    f"Error: Cannot complete spec '{spec_slug}'. No recent work log found.",
                    err=True,
                )
                typer.echo("", err=True)
                typer.echo(
                    "A work log must be created within the last 3 minutes before completing a spec.",
                    err=True,
                )
                typer.echo(
                    "This ensures your work is documented while it's fresh.",
                    err=True,
                )
                typer.echo("", err=True)
                typer.echo("To fix this:", err=True)
                typer.echo("  1. Run 'mem log' to create a work log", err=True)
                typer.echo("  2. Document what you accomplished", err=True)
                typer.echo(
                    f"  3. Run 'mem spec complete {spec_slug} \"{message}\"' again",
                    err=True,
                )
                typer.echo("", err=True)
                typer.echo(
                    "To skip this check (not recommended): --no-log",
                    err=True,
                )
                raise typer.Exit(code=1)

        # 6. Mark spec as merge_ready before committing
        typer.echo(f"📋 Completing spec: {spec['title']}...")
        specs.update_spec_status(spec_slug, "merge_ready")

        # 6b. Sync merge_ready status to GitHub immediately
        if spec.get("issue_id"):
            try:
                client = get_github_client()
                repo_owner, repo_name = get_repo_from_git(ENV_SETTINGS.caller_dir)
                gh_repo = client.get_repo(f"{repo_owner}/{repo_name}")
                sync_status_labels(gh_repo, spec["issue_id"], "merge_ready")
                typer.echo("🐙 Updated GitHub issue label to 'merge_ready'")
            except Exception as e:
                typer.echo(f"Warning: Could not update GitHub label: {e}", err=True)

        # 7. Git operations
        repo = Repo(ENV_SETTINGS.caller_dir)
        branch_name = spec.get("branch")

        if not branch_name:
            typer.echo("Error: No branch associated with this spec.", err=True)
            raise typer.Exit(code=1)

        typer.echo("🌿 Committing and pushing changes...")
        repo.git.add(A=True)
        try:
            repo.git.commit("-m", message)
        except Exception as e:
            # If nothing to commit, continue
            if "nothing to commit" not in str(e).lower():
                raise e

        repo.git.push("origin", branch_name)

        # 8. GitHub PR
        pr_url = None
        if spec.get("issue_id"):
            typer.echo("🐙 Creating Pull Request...")
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
                typer.echo(f"✅ Created Pull Request: {pr_url}")

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

        typer.echo(f"\n✅ Spec '{spec_slug}' marked as MERGE READY.")
        if pr_url:
            typer.echo(f"🔗 PR: {pr_url}")
        typer.echo("\n💡 Next steps:")
        typer.echo("  1. Merge the PR on GitHub")
        typer.echo("  2. Run 'mem merge' from the main repo to clean up")

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

    Must be run from the main repository (not from a worktree).

    This command:
    1. Validates we're in the main repo with no active spec
    2. Removes the worktree if one exists for this spec
    3. Closes the linked GitHub PR and issue with comments
    4. Moves the spec to .mem/specs/abandoned/
    5. Commits and pushes the changes
    """
    try:
        # 1. Check we're in the main repo, not a worktree
        if worktrees.is_worktree(ENV_SETTINGS.caller_dir):
            typer.echo("❌ Error: Cannot abandon specs from a worktree.", err=True)
            typer.echo(
                "\nRun this command from the main repository directory.", err=True
            )
            raise typer.Exit(code=1)

        # 2. Check no spec is currently active
        active_spec = specs.get_active_spec()
        if active_spec:
            typer.echo(
                f"❌ Error: Cannot abandon while spec '{active_spec['slug']}' is active.",
                err=True,
            )
            typer.echo("\nSwitch to 'dev' branch first: git checkout dev", err=True)
            raise typer.Exit(code=1)

        # 3. Get spec info
        spec = specs.get_spec(spec_slug)
        if not spec:
            typer.echo(f"❌ Error: Spec '{spec_slug}' not found.", err=True)
            raise typer.Exit(code=1)

        typer.echo(f"🗑️ Abandoning spec: {spec['title']} ({spec_slug})")

        # 4. Remove worktree if it exists
        existing_worktree = worktrees.get_worktree_for_spec(
            ENV_SETTINGS.caller_dir, spec_slug
        )
        if existing_worktree:
            typer.echo(f"📂 Removing worktree: {existing_worktree.path}")
            try:
                worktrees.remove_worktree(
                    ENV_SETTINGS.caller_dir, spec_slug, force=True
                )
                typer.echo("✅ Worktree removed")
            except Exception as e:
                typer.echo(f"⚠️ Warning: Could not remove worktree: {e}", err=True)

        # 5. Close GitHub PR if one exists
        if spec.get("pr_url"):
            typer.echo("🐙 Closing GitHub PR...")
            try:
                client = get_github_client()
                repo_owner, repo_name = get_repo_from_git(ENV_SETTINGS.caller_dir)
                gh_repo = client.get_repo(f"{repo_owner}/{repo_name}")

                pr_comment = f"**Spec Abandoned**\n\n{reason}"
                if close_pull_request(gh_repo, spec["pr_url"], pr_comment):
                    typer.echo(f"✅ Closed PR: {spec['pr_url']}")
                else:
                    typer.echo("⚠️ Warning: Could not close PR (may already be closed)")
            except Exception as e:
                typer.echo(f"⚠️ Warning: Could not close GitHub PR: {e}", err=True)

        # 6. Close GitHub issue if linked
        if spec.get("issue_id"):
            typer.echo("🐙 Closing GitHub issue...")
            try:
                client = get_github_client()
                repo_owner, repo_name = get_repo_from_git(ENV_SETTINGS.caller_dir)
                gh_repo = client.get_repo(f"{repo_owner}/{repo_name}")

                comment = f"**Spec Abandoned**\n\n{reason}"
                close_issue_with_comment(gh_repo, spec["issue_id"], comment)
                typer.echo(f"✅ Closed issue #{spec['issue_id']}")
            except Exception as e:
                typer.echo(f"⚠️ Warning: Could not close GitHub issue: {e}", err=True)

        # 7. Move spec to abandoned directory
        new_path = specs.move_spec_to_abandoned(spec_slug)
        typer.echo(f"📂 Moved to {new_path.relative_to(ENV_SETTINGS.caller_dir)}")

        # 8. Commit and push the changes
        typer.echo("📦 Committing changes...")
        repo = Repo(ENV_SETTINGS.caller_dir)
        repo.git.add(A=True)
        try:
            repo.git.commit("-m", f"mem: abandon spec {spec_slug}")
            typer.echo("🔄 Pushing to remote...")
            repo.git.push("origin", repo.active_branch.name)
        except Exception as e:
            if "nothing to commit" not in str(e).lower():
                typer.echo(f"⚠️ Warning: Could not commit/push: {e}", err=True)

        typer.echo(f"\n✅ Spec '{spec_slug}' abandoned.")
        typer.echo("💡 To view abandoned specs: mem spec list --status abandoned")

    except typer.Exit:
        raise
    except Exception as e:
        typer.echo(f"❌ Error: {e}", err=True)
        raise typer.Exit(code=1)


if __name__ == "__main__":
    app()
