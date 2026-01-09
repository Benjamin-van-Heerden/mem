"""
Task command - Manage tasks within specs
"""

from typing import Optional

import typer
from typing_extensions import Annotated

from env_settings import ENV_SETTINGS
from src.utils import specs, tasks, worktrees

app = typer.Typer(help="Manage tasks")


def _get_active_spec_slug() -> str:
    """Get active spec slug or raise error with helpful message."""
    active = specs.get_active_spec()
    if not active:
        typer.echo("Error: No active spec.", err=True)
        typer.echo("\nYou must be on a spec branch to work with tasks.", err=True)
        typer.echo("Options:", err=True)
        typer.echo("  1. Switch to a spec branch: mem spec activate <slug>", err=True)
        typer.echo("  2. Specify a spec: mem task <command> --spec <slug>", err=True)

        # List available specs
        todo_specs = specs.list_specs(status="todo")
        if todo_specs:
            typer.echo("\nAvailable specs:", err=True)
            for s in todo_specs:
                typer.echo(f"  - {s['slug']}: {s['title']}", err=True)

        raise typer.Exit(code=1)
    return active["slug"]


def _resolve_spec_slug(spec_slug: Optional[str]) -> str:
    """Resolve spec slug, using active spec if not provided."""
    if spec_slug:
        spec = specs.get_spec(spec_slug)
        if not spec:
            raise ValueError(f"Spec '{spec_slug}' not found")

        # Check if we're in the main repo but a worktree exists for this spec
        current_dir = ENV_SETTINGS.caller_dir
        if not worktrees.is_worktree(current_dir):
            # We're in the main repo - check if spec has a worktree
            worktree = worktrees.get_worktree_for_spec(current_dir, spec_slug)
            if worktree:
                typer.echo(
                    f"⚠️  Warning: Spec '{spec_slug}' has a worktree at:", err=True
                )
                typer.echo(f"   {worktree.path}", err=True)
                typer.echo("", err=True)
                typer.echo(
                    "   You are in the main repo. Tasks should be created in the worktree.",
                    err=True,
                )
                typer.echo(f"   Run: cd {worktree.path}", err=True)
                typer.echo("   Then: mem task new ...", err=True)
                raise typer.Exit(code=1)

        return spec_slug

    return _get_active_spec_slug()


def _find_task_by_title(spec_slug: str, title: str) -> str:
    """Find task filename by title (case-insensitive partial match)."""
    filename = tasks.find_task_by_title(spec_slug, title)
    if not filename:
        raise ValueError(f"Task '{title}' not found in spec '{spec_slug}'")
    return filename


@app.command()
def new(
    title: Annotated[str, typer.Argument(help="Title of the task")],
    description: Annotated[
        str, typer.Argument(help="Description of what this task involves")
    ],
    spec_slug: Annotated[
        Optional[str],
        typer.Option("--spec", help="Spec slug (uses active spec if not provided)"),
    ] = None,
):
    """
    Create a new task.

    Requires an active spec (be on a spec branch) unless --spec is provided.
    """
    try:
        resolved_slug = _resolve_spec_slug(spec_slug)
        task_file = tasks.create_task(resolved_slug, title, description)

        typer.echo(f"Created task: {task_file.name}")
        typer.echo(f"  Spec: {resolved_slug}")
        typer.echo("")
        typer.echo("Hints:")
        typer.echo(
            f'  Complete with: mem task complete "{title}" "detailed completion notes"'
        )
        typer.echo(
            f'  For complex tasks, break into subtasks: mem subtask new "subtask title" --task "{title}"'
        )
        typer.echo("")
        typer.echo("Refinement options:")
        typer.echo(f'  Rename task: mem task rename "{title}" "new title"')
        typer.echo(
            f'  Amend after completion: mem task amend "{title}" "additional requirements"'
        )
        typer.echo(
            "    (Amend resets status to todo, enabling iterative refinement cycles)"
        )

    except ValueError as e:
        typer.echo(f"Error: {e}", err=True)
        raise typer.Exit(code=1)


def _truncate(text: str, max_len: int) -> str:
    """Truncate text to max_len, adding ... if truncated."""
    if len(text) <= max_len:
        return text
    return text[: max_len - 3] + "..."


def _get_first_lines(body: str, max_chars: int = 150) -> str:
    """Get first portion of body text, truncated."""
    if not body:
        return ""
    # Remove any heading markers and clean up
    lines = body.strip().split("\n")
    text = " ".join(
        line.strip() for line in lines if line.strip() and not line.startswith("#")
    )
    return _truncate(text, max_chars)


@app.command("list")
def list_tasks_cmd(
    spec_slug: Annotated[
        Optional[str],
        typer.Option("--spec", help="Spec slug (uses active spec if not provided)"),
    ] = None,
    verbose: Annotated[
        bool,
        typer.Option("--verbose", "-v", help="Show full task descriptions"),
    ] = False,
):
    """
    List tasks for a specification.

    Shows task status, title, description preview, and subtask summary.
    Use --verbose for full descriptions.
    """
    try:
        resolved_slug = _resolve_spec_slug(spec_slug)
        task_list = tasks.list_tasks(resolved_slug)

        if not task_list:
            typer.echo(f"No tasks found for spec '{resolved_slug}'.")
            return

        typer.echo(f"\nTasks for '{resolved_slug}':\n")

        for task in task_list:
            # Status display
            status = task["status"]
            if status == "completed":
                status_display = "[completed]"
            else:
                status_display = "[todo]"

            # Title
            typer.echo(f"{status_display} {task['title']}")

            # Description (body)
            body = task.get("body", "")
            if verbose and body:
                typer.echo(f"       {body.strip()}")
            elif body:
                preview = _get_first_lines(body)
                if preview:
                    typer.echo(f"       {preview}")

            # Subtasks summary
            subtask_list = task.get("subtasks", [])
            if subtask_list:
                completed_count = sum(
                    1 for s in subtask_list if s["status"] == "completed"
                )
                total_count = len(subtask_list)
                typer.echo(f"       Subtasks: {completed_count}/{total_count} complete")

                if verbose:
                    for sub in subtask_list:
                        sub_icon = "[x]" if sub["status"] == "completed" else "[ ]"
                        typer.echo(f"         {sub_icon} {sub['title']}")

            # Created date
            created = task.get("created_at", "")
            if created:
                # Just show the date part
                date_part = created.split("T")[0] if "T" in created else created
                typer.echo(f"       Created: {date_part}")

            typer.echo()  # Blank line between tasks

        # Summary
        completed = sum(1 for t in task_list if t["status"] == "completed")
        total = len(task_list)
        typer.echo(f"Total: {total} task(s), {completed} completed")

        # Hints based on state
        pending = [t for t in task_list if t["status"] != "completed"]
        if pending:
            typer.echo("")
            typer.echo(f"Next task: {pending[0]['title']}")
            typer.echo(
                f'  Complete with: mem task complete "{pending[0]["title"]}" "notes"'
            )
        elif total > 0:
            typer.echo("")
            typer.echo("All tasks complete! Spec ready for completion:")
            typer.echo(f'  mem spec complete {resolved_slug} "detailed commit message"')

    except ValueError as e:
        typer.echo(f"Error: {e}", err=True)
        raise typer.Exit(code=1)


@app.command()
def complete(
    title: Annotated[str, typer.Argument(help="Task title (partial match supported)")],
    notes: Annotated[
        str, typer.Argument(help="Completion notes describing what was done")
    ],
    spec_slug: Annotated[
        Optional[str],
        typer.Option("--spec", help="Spec slug (uses active spec if not provided)"),
    ] = None,
):
    """
    Mark a task as complete with notes.

    Appends completion notes to the task body and marks it as completed.
    """
    try:
        resolved_slug = _resolve_spec_slug(spec_slug)
        task_filename = _find_task_by_title(resolved_slug, title)

        # Check for incomplete subtasks
        if tasks.has_incomplete_subtasks(resolved_slug, task_filename):
            typer.echo(
                "Error: Cannot complete task with incomplete subtasks.", err=True
            )
            subtask_list = tasks.list_subtasks(resolved_slug, task_filename)
            incomplete = [s for s in subtask_list if s["status"] != "completed"]
            typer.echo("\nIncomplete subtasks:", err=True)
            for s in incomplete:
                typer.echo(f"  - {s['title']}", err=True)
            raise typer.Exit(code=1)

        tasks.complete_task_with_notes(resolved_slug, task_filename, notes)
        typer.echo(f"✓ Task completed: {title}")
        typer.echo("")

        # Check if all tasks are now complete
        task_list = tasks.list_tasks(resolved_slug)
        pending = [t for t in task_list if t["status"] != "completed"]

        if not pending and task_list:
            typer.echo("All spec tasks are complete!")
            typer.echo(
                f'  Spec ready for completion via: mem spec complete {resolved_slug} "detailed commit message"'
            )
        elif pending:
            typer.echo(f"  Remaining tasks: {len(pending)}")

        typer.echo("")
        typer.echo("[AGENT INSTRUCTION]")
        typer.echo("Your next response must:")
        typer.echo("1. Summarize what was done for this task")
        typer.echo("2. Ask the user if they want to continue")
        typer.echo("Do NOT call any tools. Do NOT start the next task.")

    except ValueError as e:
        typer.echo(f"Error: {e}", err=True)
        raise typer.Exit(code=1)


@app.command()
def delete(
    title: Annotated[str, typer.Argument(help="Task title (partial match supported)")],
    spec_slug: Annotated[
        Optional[str],
        typer.Option("--spec", help="Spec slug (uses active spec if not provided)"),
    ] = None,
):
    """
    Delete a task.
    """
    try:
        resolved_slug = _resolve_spec_slug(spec_slug)
        task_filename = _find_task_by_title(resolved_slug, title)
        tasks.delete_task(resolved_slug, task_filename)
        typer.echo(f"Deleted task: {title}")
    except ValueError as e:
        typer.echo(f"Error: {e}", err=True)
        raise typer.Exit(code=1)


@app.command()
def amend(
    title: Annotated[str, typer.Argument(help="Task title (partial match supported)")],
    notes: Annotated[str, typer.Argument(help="Amendment notes to append")],
    spec_slug: Annotated[
        Optional[str],
        typer.Option("--spec", help="Spec slug (uses active spec if not provided)"),
    ] = None,
):
    """
    Amend a task with additional notes.

    Appends an Amendments section to the task body and resets status to todo.
    This enables iterative refinement: complete a task, then amend it with
    new requirements, complete again, and so on.
    """
    try:
        resolved_slug = _resolve_spec_slug(spec_slug)
        task_filename = _find_task_by_title(resolved_slug, title)
        tasks.amend_task(resolved_slug, task_filename, notes)

        task = tasks.get_task(resolved_slug, task_filename)
        typer.echo(f"Amended task: {task['title']}")
        typer.echo("  Status reset to: todo")
        typer.echo("")
        typer.echo("The task can now be completed again with new completion notes.")
        typer.echo(
            f'  Complete with: mem task complete "{task["title"]}" "completion notes"'
        )

    except ValueError as e:
        typer.echo(f"Error: {e}", err=True)
        raise typer.Exit(code=1)


@app.command()
def rename(
    title: Annotated[
        str, typer.Argument(help="Current task title (partial match supported)")
    ],
    new_title: Annotated[str, typer.Argument(help="New title for the task")],
    spec_slug: Annotated[
        Optional[str],
        typer.Option("--spec", help="Spec slug (uses active spec if not provided)"),
    ] = None,
):
    """
    Rename a task.

    Updates the task title in frontmatter. The filename remains unchanged
    for stability (references and git history are preserved).
    """
    try:
        resolved_slug = _resolve_spec_slug(spec_slug)
        task_filename = _find_task_by_title(resolved_slug, title)

        old_task = tasks.get_task(resolved_slug, task_filename)
        old_title = old_task["title"]

        tasks.rename_task(resolved_slug, task_filename, new_title)

        typer.echo(f"Renamed task:")
        typer.echo(f"  From: {old_title}")
        typer.echo(f"  To:   {new_title}")

    except ValueError as e:
        typer.echo(f"Error: {e}", err=True)
        raise typer.Exit(code=1)


if __name__ == "__main__":
    app()
