"""
Task command - Manage tasks within specs
"""

from typing import Optional

import typer
from typing_extensions import Annotated

from src.utils import specs, tasks

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
        typer.echo("Hint: For complex tasks, break them into subtasks:")
        typer.echo(f'  mem subtask new "subtask title" --task "{title}"')

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
            typer.echo(f'  mem spec complete {resolved_slug} "commit message"')

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
        typer.echo(f"Completed task: {title}")

        # Check if all tasks are now complete
        task_list = tasks.list_tasks(resolved_slug)
        pending = [t for t in task_list if t["status"] != "completed"]

        typer.echo("")
        if not pending and task_list:
            typer.echo("All spec tasks are complete!")
            typer.echo(
                f'Spec ready for completion via: mem spec complete {resolved_slug} "commit message"'
            )
        else:
            if pending:
                typer.echo("")
                typer.echo(f"Remaining tasks: {len(pending)}")
        typer.echo("Stop here AT ONCE and Report what was accomplished")
        typer.echo(
            "Await review of what you have done and further instructions. DO NOT continue with the next task or action."
        )

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


if __name__ == "__main__":
    app()
