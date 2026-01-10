"""
Sync command - Bidirectional synchronization between GitHub issues and local specs
"""

import subprocess
from dataclasses import dataclass, field
from enum import Enum
from typing import Any

import typer

from env_settings import ENV_SETTINGS
from src.utils import specs, todos
from src.utils.github.api import (
    close_issue_with_comment,
    create_github_issue,
    get_comments,
    get_status_from_labels,
    get_status_label_name,
    is_pr_merged,
    list_repo_issues,
    sync_status_labels,
    update_github_issue,
)
from src.utils.github.client import get_github_client
from src.utils.github.repo import get_repo_from_git
from src.utils.specs import ensure_on_dev_branch
from src.utils.sync_utils import (
    SEPARATOR,
    compute_content_hash,
    content_differs,
    extract_body_from_spec_file,
    slugify,
)

app = typer.Typer(help="Synchronize with GitHub")


def git_fetch_and_pull() -> tuple[bool, str]:
    """
    Fetch and pull from remote.

    Returns:
        (success, message) tuple
    """
    cwd = ENV_SETTINGS.caller_dir

    try:
        subprocess.run(
            ["git", "fetch", "origin"],
            cwd=cwd,
            check=True,
            capture_output=True,
            text=True,
        )
    except subprocess.CalledProcessError as e:
        return False, f"git fetch failed: {e.stderr}"

    try:
        result = subprocess.run(
            ["git", "pull", "--ff-only"],
            cwd=cwd,
            capture_output=True,
            text=True,
        )
        if result.returncode != 0:
            stderr_lower = result.stderr.lower()
            if "not possible to fast-forward" in stderr_lower:
                return (
                    False,
                    "Cannot fast-forward. Please resolve conflicts manually and try again.",
                )
            # New branches with no upstream are OK - nothing to pull
            if "no tracking information" in stderr_lower:
                return True, "OK (no upstream to pull from)"
            return False, f"git pull failed: {result.stderr}"
    except subprocess.CalledProcessError as e:
        return False, f"git pull failed: {e.stderr}"

    return True, "OK"


def git_has_mem_changes() -> bool:
    """Check if there are uncommitted changes in .mem/ directory."""
    cwd = ENV_SETTINGS.caller_dir
    result = subprocess.run(
        ["git", "status", "--porcelain", ".mem/"],
        cwd=cwd,
        capture_output=True,
        text=True,
    )
    return bool(result.stdout.strip())


def git_commit_and_push(message: str) -> tuple[bool, str]:
    """
    Add .mem/ changes, commit, and push.

    Returns:
        (success, message) tuple
    """
    cwd = ENV_SETTINGS.caller_dir

    try:
        subprocess.run(
            ["git", "add", ".mem/"],
            cwd=cwd,
            check=True,
            capture_output=True,
            text=True,
        )
    except subprocess.CalledProcessError as e:
        return False, f"git add failed: {e.stderr}"

    try:
        subprocess.run(
            ["git", "commit", "-m", message],
            cwd=cwd,
            check=True,
            capture_output=True,
            text=True,
        )
    except subprocess.CalledProcessError as e:
        if "nothing to commit" in e.stdout.lower():
            return True, "Nothing to commit"
        return False, f"git commit failed: {e.stderr}"

    try:
        subprocess.run(
            ["git", "push"],
            cwd=cwd,
            check=True,
            capture_output=True,
            text=True,
        )
    except subprocess.CalledProcessError as e:
        return False, f"git push failed: {e.stderr}. You may need to pull first."

    return True, "OK"


def build_sync_commit_message(plan: "SyncPlan", actions_executed: int) -> str:
    """Build a descriptive commit message for sync changes."""
    parts = []

    if plan.outbound_creates:
        parts.append(f"created {len(plan.outbound_creates)} issue(s)")
    if plan.outbound_updates:
        parts.append(f"updated {len(plan.outbound_updates)} issue(s)")
    if plan.inbound_creates:
        parts.append(f"synced {len(plan.inbound_creates)} spec(s) from GitHub")
    if plan.inbound_updates:
        parts.append(f"updated {len(plan.inbound_updates)} spec(s) from GitHub")
    if plan.specs_to_complete:
        parts.append(f"completed {len(plan.specs_to_complete)} spec(s)")
    if plan.todos_to_create:
        parts.append(f"created {len(plan.todos_to_create)} todo(s)")

    if parts:
        return f"mem sync: {', '.join(parts)}"
    return f"mem sync: {actions_executed} action(s)"


class SyncDirection(Enum):
    """Direction of sync action."""

    INBOUND = "inbound"  # GitHub -> Local
    OUTBOUND = "outbound"  # Local -> GitHub
    CONFLICT = "conflict"  # Both changed


@dataclass
class SyncAction:
    """Represents a single sync action to be performed."""

    direction: SyncDirection
    action_type: str  # 'create', 'update', 'close', 'status'
    spec_slug: str | None
    issue_number: int | None
    title: str
    description: str


@dataclass
class SyncPlan:
    """Complete plan for sync operations."""

    inbound_creates: list[SyncAction] = field(default_factory=list)
    inbound_updates: list[SyncAction] = field(default_factory=list)
    outbound_creates: list[SyncAction] = field(default_factory=list)
    outbound_updates: list[SyncAction] = field(default_factory=list)
    status_syncs: list[SyncAction] = field(default_factory=list)
    conflicts: list[SyncAction] = field(default_factory=list)
    todos_to_create: list[dict[str, Any]] = field(default_factory=list)
    specs_to_complete: list[dict[str, Any]] = field(default_factory=list)

    @property
    def has_changes(self) -> bool:
        return any(
            [
                self.inbound_creates,
                self.inbound_updates,
                self.outbound_creates,
                self.outbound_updates,
                self.status_syncs,
                self.conflicts,
                self.todos_to_create,
                self.specs_to_complete,
            ]
        )

    @property
    def total_actions(self) -> int:
        return (
            len(self.inbound_creates)
            + len(self.inbound_updates)
            + len(self.outbound_creates)
            + len(self.outbound_updates)
            + len(self.status_syncs)
            + len(self.todos_to_create)
            + len(self.specs_to_complete)
        )


def build_sync_plan(
    repo: Any,
    local_specs: list[dict[str, Any]],
    github_issues: list[Any],
) -> SyncPlan:
    """
    Compare local specs and GitHub issues to determine sync actions.

    Logic:
    1. For each GitHub issue with mem-spec label:
       - If no local spec exists with that issue_id -> INBOUND CREATE
       - If local spec exists -> compare hashes for changes
    2. For each local spec without issue_id -> OUTBOUND CREATE
    3. For status/label mismatches -> STATUS SYNC
    4. Other issues -> create as todos
    5. For merge_ready specs with merged PRs -> move to completed
    """
    plan = SyncPlan()

    # Check for merge_ready specs with merged PRs
    for spec in local_specs:
        if spec.get("status") == "merge_ready" and spec.get("pr_url"):
            if is_pr_merged(repo, spec["pr_url"]):
                plan.specs_to_complete.append(spec)

    # Build lookup maps
    specs_by_issue_id: dict[int, dict[str, Any]] = {
        spec["issue_id"]: spec for spec in local_specs if spec.get("issue_id")
    }
    issue_numbers_seen: set = set()

    # Process GitHub issues
    for issue in github_issues:
        labels = [label.name for label in issue.labels]
        issue_title = issue.title.replace("[Spec]:", "").replace("[Spec]: ", "").strip()

        if "mem-spec" in labels:
            issue_numbers_seen.add(issue.number)
            spec = specs_by_issue_id.get(issue.number)

            if spec is None:
                # New issue from GitHub -> INBOUND CREATE
                plan.inbound_creates.append(
                    SyncAction(
                        direction=SyncDirection.INBOUND,
                        action_type="create",
                        spec_slug=None,
                        issue_number=issue.number,
                        title=issue_title,
                        description=f"Create local spec from GitHub issue #{issue.number}",
                    )
                )
            else:
                # Existing spec - check for changes
                remote_body = issue.body or ""
                remote_hash = compute_content_hash(remote_body)

                # Get local content
                spec_file = specs.get_spec_file_path(spec["slug"])
                local_body = extract_body_from_spec_file(spec_file)
                local_hash = compute_content_hash(local_body)

                stored_local_hash = spec.get("local_content_hash")
                stored_remote_hash = spec.get("remote_content_hash")

                local_changed = content_differs(local_hash, stored_local_hash)
                remote_changed = content_differs(remote_hash, stored_remote_hash)

                if local_changed and remote_changed:
                    # Both sides changed -> CONFLICT
                    plan.conflicts.append(
                        SyncAction(
                            direction=SyncDirection.CONFLICT,
                            action_type="conflict",
                            spec_slug=spec["slug"],
                            issue_number=issue.number,
                            title=issue_title,
                            description=f"Both local and remote changed for spec '{spec['slug']}'",
                        )
                    )
                elif remote_changed:
                    # Only remote changed -> INBOUND UPDATE
                    plan.inbound_updates.append(
                        SyncAction(
                            direction=SyncDirection.INBOUND,
                            action_type="update",
                            spec_slug=spec["slug"],
                            issue_number=issue.number,
                            title=issue_title,
                            description=f"Update local spec '{spec['slug']}' from GitHub",
                        )
                    )
                elif local_changed:
                    # Only local changed -> OUTBOUND UPDATE
                    plan.outbound_updates.append(
                        SyncAction(
                            direction=SyncDirection.OUTBOUND,
                            action_type="update",
                            spec_slug=spec["slug"],
                            issue_number=issue.number,
                            title=issue_title,
                            description=f"Update GitHub issue #{issue.number} from local",
                        )
                    )

                # Check status label sync (only if no content conflict)
                if not (local_changed and remote_changed):
                    github_status = get_status_from_labels(labels)
                    local_status = spec.get("status")

                    if local_status and local_status != github_status:
                        # Local status differs from GitHub - sync outbound (local is source of truth)
                        plan.status_syncs.append(
                            SyncAction(
                                direction=SyncDirection.OUTBOUND,
                                action_type="status",
                                spec_slug=spec["slug"],
                                issue_number=issue.number,
                                title=issue_title,
                                description=f"Update GitHub labels to '{local_status}' from local",
                            )
                        )
                    elif github_status and not local_status:
                        # GitHub has status but local doesn't - sync inbound
                        plan.status_syncs.append(
                            SyncAction(
                                direction=SyncDirection.INBOUND,
                                action_type="status",
                                spec_slug=spec["slug"],
                                issue_number=issue.number,
                                title=issue_title,
                                description=f"Update local status to '{github_status}' from GitHub labels",
                            )
                        )
        else:
            # Not a spec issue - check if we should create a todo
            existing_todo = todos.get_todo_by_issue_id(issue.number)
            if not existing_todo:
                # Also check by title (for backwards compat)
                existing_by_title = todos.get_todo(slugify(issue.title))
                if not existing_by_title:
                    plan.todos_to_create.append(
                        {
                            "title": issue.title,
                            "body": issue.body or "",
                            "issue_number": issue.number,
                            "issue_url": issue.html_url,
                        }
                    )

    # Find local specs without issue_id -> OUTBOUND CREATE
    for spec in local_specs:
        if spec.get("issue_id") is None:
            plan.outbound_creates.append(
                SyncAction(
                    direction=SyncDirection.OUTBOUND,
                    action_type="create",
                    spec_slug=spec["slug"],
                    issue_number=None,
                    title=spec["title"],
                    description=f"Create GitHub issue from local spec '{spec['slug']}'",
                )
            )

    return plan


def print_sync_plan(plan: SyncPlan) -> None:
    """Print a human-readable preview of what sync would do."""
    typer.echo("\n" + "=" * 60)
    typer.echo("🔍 SYNC PREVIEW (dry-run)")
    typer.echo("=" * 60)

    if not plan.has_changes:
        typer.echo("\n✅ Everything is in sync. No changes needed.")
        return

    if plan.outbound_creates:
        typer.echo("\n📤 WILL CREATE on GitHub:")
        for action in plan.outbound_creates:
            typer.echo(f'   + Spec "{action.spec_slug}": "{action.title}"')

    if plan.outbound_updates:
        typer.echo("\n📤 WILL UPDATE on GitHub:")
        for action in plan.outbound_updates:
            typer.echo(f'   ~ Issue #{action.issue_number}: "{action.title}"')

    if plan.inbound_creates:
        typer.echo("\n📥 WILL CREATE locally:")
        for action in plan.inbound_creates:
            typer.echo(f'   + Issue #{action.issue_number}: "{action.title}"')

    if plan.inbound_updates:
        typer.echo("\n📥 WILL UPDATE locally:")
        for action in plan.inbound_updates:
            typer.echo(f'   ~ Spec "{action.spec_slug}": "{action.title}"')

    if plan.status_syncs:
        typer.echo("\n🏷️  WILL SYNC status labels:")
        for action in plan.status_syncs:
            direction = (
                "GitHub → Local"
                if action.direction == SyncDirection.INBOUND
                else "Local → GitHub"
            )
            typer.echo(f"   ~ {direction}: {action.description}")

    if plan.todos_to_create:
        typer.echo("\n📋 WILL CREATE todos:")
        for todo in plan.todos_to_create:
            typer.echo(f'   + "{todo["title"]}"')

    if plan.specs_to_complete:
        typer.echo("\n✅ WILL MOVE TO COMPLETED (PR merged):")
        for spec in plan.specs_to_complete:
            typer.echo(f'   → "{spec["slug"]}": "{spec["title"]}"')

    if plan.conflicts:
        typer.echo("\n⚠️  CONFLICTS (require manual resolution):")
        for action in plan.conflicts:
            typer.echo(
                f'   ! Spec "{action.spec_slug}" / Issue #{action.issue_number}: "{action.title}"'
            )
            typer.echo(
                "     Both local file and GitHub issue have changed since last sync."
            )
            typer.echo("     Resolve by editing the local file, then run sync again.")

    typer.echo("\n" + "-" * 60)
    typer.echo(f"📊 Total actions: {plan.total_actions}")
    if plan.conflicts:
        typer.echo(f"⚠️ Conflicts: {len(plan.conflicts)} (will be skipped)")
    typer.echo("\n💡 Run without --dry-run to apply these changes.")


def execute_outbound_create(repo: Any, spec: dict[str, Any]) -> None:
    """Create a GitHub issue from a local spec."""
    # Read spec body
    spec_file = specs.get_spec_file_path(spec["slug"])
    body = extract_body_from_spec_file(spec_file)

    # Determine labels
    labels = ["mem-spec"]
    status_label = get_status_label_name(spec.get("status", "todo"))
    if status_label:
        labels.append(status_label)

    # Create the issue
    title = f"[Spec]: {spec['title']}"
    issue = create_github_issue(repo, title=title, body=body, labels=labels)

    # Update spec with issue info
    specs.update_spec_issue_info(spec["slug"], issue.number, issue.html_url)

    # Store content hashes
    local_hash = compute_content_hash(body)
    remote_hash = compute_content_hash(issue.body or "")
    specs.mark_spec_synced(spec["slug"], local_hash, remote_hash)

    typer.echo(f'   ✓ Created issue #{issue.number} for spec "{spec["slug"]}"')


def execute_outbound_update(repo: Any, spec: dict[str, Any]) -> None:
    """Update a GitHub issue from local spec changes."""
    # Read spec body
    spec_file = specs.get_spec_file_path(spec["slug"])
    body = extract_body_from_spec_file(spec_file)

    # Update the issue
    issue = update_github_issue(repo, spec["issue_id"], body=body)

    # Store content hashes
    local_hash = compute_content_hash(body)
    remote_hash = compute_content_hash(issue.body or "")
    specs.mark_spec_synced(spec["slug"], local_hash, remote_hash)

    typer.echo(f'   ✓ Updated issue #{spec["issue_id"]} from spec "{spec["slug"]}"')


def execute_inbound_create(repo: Any, issue: Any) -> None:
    """Create a local spec from a GitHub issue."""
    issue_title = issue.title.replace("[Spec]:", "").replace("[Spec]: ", "").strip()
    assignee = issue.assignee.login if issue.assignee else None
    labels = [label.name for label in issue.labels]

    # Create the spec (creates directory and spec.md)
    slug = slugify(issue_title)

    # Check if spec already exists
    existing = specs.get_spec(slug)
    if existing:
        typer.echo(f'   ⚠ Spec "{slug}" already exists, skipping create')
        return

    # Create spec file
    specs.create_spec(issue_title)

    # Fetch and format comments
    comments = get_comments(issue)
    formatted_comments = []
    for c in comments:
        formatted_comments.append(
            f"### Comment by @{c['user']} on {c['created_at']}\n\n{c['body']}"
        )

    # Build body with comments
    body = issue.body or ""
    if formatted_comments:
        body += SEPARATOR + "\n\n".join(formatted_comments)

    # Update body
    specs.update_spec_body(slug, body)

    # Update metadata
    specs.update_spec_issue_info(slug, issue.number, issue.html_url)

    if assignee:
        specs.assign_spec(slug, assignee)

    # Set status from labels if present
    github_status = get_status_from_labels(labels)
    if github_status:
        specs.update_spec_status(slug, github_status)

    # Store content hashes
    body_hash = compute_content_hash(issue.body or "")
    specs.mark_spec_synced(slug, body_hash, body_hash)

    typer.echo(f'   ✓ Created spec "{slug}" from issue #{issue.number}')


def execute_inbound_update(repo: Any, issue: Any, spec: dict[str, Any]) -> None:
    """Update a local spec from GitHub issue changes."""
    issue_title = issue.title.replace("[Spec]:", "").replace("[Spec]: ", "").strip()
    assignee = issue.assignee.login if issue.assignee else None

    # Fetch and format comments
    comments = get_comments(issue)
    formatted_comments = []
    for c in comments:
        formatted_comments.append(
            f"### Comment by @{c['user']} on {c['created_at']}\n\n{c['body']}"
        )

    # Build body with comments
    body = issue.body or ""
    if formatted_comments:
        body += SEPARATOR + "\n\n".join(formatted_comments)

    # Update body
    specs.update_spec_body(spec["slug"], body)

    # Update metadata
    if assignee:
        specs.assign_spec(spec["slug"], assignee)

    # Update title if it changed
    if spec["title"] != issue_title:
        specs.update_spec(spec["slug"], title=issue_title)

    # Store content hashes
    body_hash = compute_content_hash(issue.body or "")
    specs.mark_spec_synced(spec["slug"], body_hash, body_hash)

    typer.echo(f'   ✓ Updated spec "{spec["slug"]}" from issue #{issue.number}')


def execute_status_sync(
    repo: Any, action: SyncAction, spec: dict[str, Any], issue: Any
) -> None:
    """Sync status between local spec and GitHub labels."""
    labels = [label.name for label in issue.labels]

    if action.direction == SyncDirection.INBOUND:
        # GitHub -> Local
        github_status = get_status_from_labels(labels)
        if github_status:
            specs.update_spec_status(spec["slug"], github_status)
            typer.echo(
                f"   ✓ Updated spec \"{spec['slug']}\" status to '{github_status}'"
            )
    else:
        # Local -> GitHub
        local_status = spec.get("status")
        if local_status:
            sync_status_labels(repo, issue.number, local_status)
            typer.echo(f"   ✓ Updated issue #{issue.number} labels to '{local_status}'")


def execute_sync_plan(plan: SyncPlan, repo: Any) -> int:
    """
    Execute the sync plan.

    Returns:
        Number of actions executed
    """
    actions_executed = 0

    # Get fresh data for execution
    all_specs = specs.get_all_specs()
    specs_by_slug = {spec["slug"]: spec for spec in all_specs}

    # Fetch issues for execution
    issues = list_repo_issues(repo, state="all")
    issues_by_number = {issue.number: issue for issue in issues}

    # Execute outbound creates
    if plan.outbound_creates:
        typer.echo("\n📤 Creating on GitHub...")
        for action in plan.outbound_creates:
            spec = specs_by_slug.get(action.spec_slug)
            if spec:
                execute_outbound_create(repo, spec)
                actions_executed += 1

    # Execute outbound updates
    if plan.outbound_updates:
        typer.echo("\n📤 Updating on GitHub...")
        for action in plan.outbound_updates:
            spec = specs_by_slug.get(action.spec_slug)
            if spec:
                execute_outbound_update(repo, spec)
                actions_executed += 1

    # Execute inbound creates
    if plan.inbound_creates:
        typer.echo("\n📥 Creating locally...")
        for action in plan.inbound_creates:
            assert action.issue_number is not None
            issue = issues_by_number.get(action.issue_number)
            if issue:
                execute_inbound_create(repo, issue)
                actions_executed += 1

    # Execute inbound updates
    if plan.inbound_updates:
        typer.echo("\n📥 Updating locally...")
        for action in plan.inbound_updates:
            spec = specs_by_slug.get(action.spec_slug)
            assert action.issue_number is not None
            issue = issues_by_number.get(action.issue_number)
            if spec and issue:
                execute_inbound_update(repo, issue, spec)
                actions_executed += 1

    # Execute status syncs
    if plan.status_syncs:
        typer.echo("\n🏷️  Syncing status labels...")
        for action in plan.status_syncs:
            spec = specs_by_slug.get(action.spec_slug)
            assert action.issue_number is not None
            issue = issues_by_number.get(action.issue_number)
            if spec and issue:
                execute_status_sync(repo, action, spec, issue)
                actions_executed += 1

    # Create todos
    if plan.todos_to_create:
        typer.echo("\n📋 Creating todos...")
        for todo_data in plan.todos_to_create:
            try:
                todos.create_todo(todo_data["title"], todo_data.get("body", ""))
                slug = slugify(todo_data["title"])
                # Link to GitHub issue
                if todo_data.get("issue_number"):
                    todos.update_todo_issue_info(
                        slug, todo_data["issue_number"], todo_data.get("issue_url", "")
                    )
                title_display = todo_data["title"]
                if len(title_display) > 50:
                    title_display = title_display[:50] + "..."
                typer.echo(f'   ✓ Created todo: "{title_display}"')
                actions_executed += 1
            except ValueError:
                # Todo already exists
                typer.echo(f'   ⚠ Todo "{todo_data["title"][:30]}..." already exists')

    # Move completed specs and close their GitHub issues
    if plan.specs_to_complete:
        typer.echo("\n✅ Moving merged specs to completed...")
        for spec in plan.specs_to_complete:
            try:
                specs.move_spec_to_completed(spec["slug"])
                typer.echo(f'   ✓ Moved "{spec["slug"]}" to completed/')
                actions_executed += 1

                # Close the GitHub issue if it exists
                if spec.get("issue_id"):
                    try:
                        pr_url = spec.get("pr_url", "")
                        if pr_url:
                            comment = f"Completed via PR: {pr_url}"
                        else:
                            comment = "Completed and merged."
                        close_issue_with_comment(repo, spec["issue_id"], comment)
                        typer.echo(f"   ✓ Closed issue #{spec['issue_id']}")
                    except Exception as e:
                        typer.echo(
                            f"   ⚠ Could not close issue #{spec['issue_id']}: {e}"
                        )
            except Exception as e:
                typer.echo(f'   ⚠ Could not move "{spec["slug"]}": {e}')

    # Report conflicts
    if plan.conflicts:
        typer.echo("\n⚠️  Skipped conflicts:")
        for action in plan.conflicts:
            typer.echo(
                f'   ! Spec "{action.spec_slug}" / Issue #{action.issue_number}: "{action.title}"'
            )

    return actions_executed


@app.command()
def sync(
    dry_run: bool = typer.Option(
        False, "--dry-run", "-n", help="Preview changes without applying them"
    ),
    no_git: bool = typer.Option(
        False, "--no-git", help="Skip git pull/commit/push operations"
    ),
    no_cleanup: bool = typer.Option(
        False, "--no-cleanup", help="Skip branch cleanup after sync"
    ),
):
    """
    Bidirectional sync between GitHub issues and local specs.

    By default, syncs in both directions:
    - Inbound: GitHub issues with 'mem-spec' label -> Local specs
    - Outbound: Local specs -> GitHub issues

    Also handles git operations:
    - Pulls latest changes before syncing
    - Commits and pushes .mem/ changes after syncing

    Use --dry-run to preview changes without applying them.
    Use --no-git to skip git operations.
    """
    # Auto-switch to dev if on main or test
    switched, switch_msg = ensure_on_dev_branch()
    if switched and switch_msg:
        typer.echo(f"⚠️  {switch_msg}")

    typer.echo("🔄 Synchronizing with GitHub...")

    try:
        # 1. Git pull (unless --no-git or --dry-run)
        if not no_git and not dry_run:
            typer.echo("   Pulling latest changes...")
            success, message = git_fetch_and_pull()
            if not success:
                typer.echo(f"\n❌ {message}", err=True)
                raise typer.Exit(code=1)

        # 2. Setup GitHub client and repo
        client = get_github_client()
        repo_owner, repo_name = get_repo_from_git(ENV_SETTINGS.caller_dir)
        repo_full_name = f"{repo_owner}/{repo_name}"
        repo = client.get_repo(repo_full_name)

        # 3. Fetch all data
        typer.echo("   Fetching GitHub issues...")
        github_issues = list_repo_issues(repo, state="open")

        typer.echo("   Loading local specs...")
        local_specs = specs.get_all_specs()

        # 4. Build sync plan
        typer.echo("   Building sync plan...")
        plan = build_sync_plan(repo, local_specs, github_issues)

        # 5. Execute or preview
        if dry_run:
            print_sync_plan(plan)
        else:
            if not plan.has_changes:
                typer.echo("\n✅ Everything is in sync. No changes needed.")
            else:
                actions_executed = execute_sync_plan(plan, repo)

                # 6. Git commit and push (unless --no-git)
                if not no_git and actions_executed > 0:
                    if git_has_mem_changes():
                        typer.echo("\n📤 Committing and pushing changes...")
                        commit_message = build_sync_commit_message(
                            plan, actions_executed
                        )
                        success, message = git_commit_and_push(commit_message)
                        if not success:
                            typer.echo(f"   ⚠ {message}", err=True)
                        else:
                            typer.echo("   ✓ Changes pushed to remote")

                typer.echo("\n" + "=" * 60)
                typer.echo("✅ Sync complete!")
                typer.echo("=" * 60)
                typer.echo(f"   📊 Actions executed: {actions_executed}")
                if plan.conflicts:
                    typer.echo(f"   ⚠️ Conflicts skipped: {len(plan.conflicts)}")

        # 7. Run cleanup (unless --no-cleanup or --dry-run)
        if not no_cleanup and not dry_run:
            from src.commands.cleanup import run_cleanup

            deleted, _ = run_cleanup(dry_run=False, silent=True)
            if deleted > 0:
                typer.echo(f"\n🧹 Cleaned up {deleted} stale branch(es)")

    except Exception as e:
        typer.echo(f"\n❌ Sync failed: {e}", err=True)
        raise typer.Exit(code=1)


if __name__ == "__main__":
    app()
