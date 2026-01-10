"""
Subtask command - Manage subtasks within tasks

Subtasks are embedded in task frontmatter, not separate files.
"""

from typing import Optional

import typer
from typing_extensions import Annotated

from src.utils import specs, tasks

app = typer.Typer(help="Manage subtasks")


def _get_active_spec_slug() -> str:
    """Get active spec slug or raise error."""
    active = specs.get_active_spec()
    if not active:
        typer.echo("Error: No active spec.", err=True)
        typer.echo("\nYou must be on a spec branch or use --spec.", err=True)
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
    title: Annotated[str, typer.Argument(help="Title of the subtask")],
    task_title: Annotated[
        str, typer.Option("--task", help="Parent task title (required)")
    ],
    spec_slug: Annotated[
        Optional[str],
        typer.Option("--spec", help="Spec slug (uses active spec if not provided)"),
    ] = None,
):
    """Create a new subtask."""
    try:
        resolved_slug = _resolve_spec_slug(spec_slug)
        task_filename = _find_task_by_title(resolved_slug, task_title)

        tasks.add_subtask(resolved_slug, task_filename, title)
        typer.echo(f"✅ Created subtask: {title}")
        typer.echo(f"  ✏️ Task: {task_title}")
        typer.echo(f"  📋 Spec: {resolved_slug}")

    except ValueError as e:
        typer.echo(f"Error: {e}", err=True)
        raise typer.Exit(code=1)


@app.command()
def complete(
    title: Annotated[str, typer.Argument(help="Title of the subtask")],
    task_title: Annotated[
        str, typer.Option("--task", help="Parent task title (required)")
    ],
    spec_slug: Annotated[
        Optional[str],
        typer.Option("--spec", help="Spec slug (uses active spec if not provided)"),
    ] = None,
):
    """Mark a subtask as complete."""
    try:
        resolved_slug = _resolve_spec_slug(spec_slug)
        task_filename = _find_task_by_title(resolved_slug, task_title)

        tasks.complete_subtask(resolved_slug, task_filename, title)
        typer.echo(f"✅ Completed subtask: {title}")

    except ValueError as e:
        typer.echo(f"Error: {e}", err=True)
        raise typer.Exit(code=1)


@app.command("list")
def list_subtasks_cmd(
    task_title: Annotated[
        str, typer.Option("--task", help="Parent task title (required)")
    ],
    spec_slug: Annotated[
        Optional[str],
        typer.Option("--spec", help="Spec slug (uses active spec if not provided)"),
    ] = None,
):
    """List subtasks for a task."""
    try:
        resolved_slug = _resolve_spec_slug(spec_slug)
        task_filename = _find_task_by_title(resolved_slug, task_title)

        subtask_list = tasks.list_subtasks(resolved_slug, task_filename)

        if not subtask_list:
            typer.echo(f"No subtasks for task '{task_title}'")
            return

        typer.echo(f"\n📝 Subtasks for '{task_title}':\n")
        for sub in subtask_list:
            icon = "[x]" if sub["status"] == "completed" else "[ ]"
            typer.echo(f"  {icon} {sub['title']}")

    except ValueError as e:
        typer.echo(f"Error: {e}", err=True)
        raise typer.Exit(code=1)


@app.command()
def delete(
    title: Annotated[str, typer.Argument(help="Title of the subtask")],
    task_title: Annotated[
        str, typer.Option("--task", help="Parent task title (required)")
    ],
    spec_slug: Annotated[
        Optional[str],
        typer.Option("--spec", help="Spec slug (uses active spec if not provided)"),
    ] = None,
):
    """Delete a subtask."""
    try:
        resolved_slug = _resolve_spec_slug(spec_slug)
        task_filename = _find_task_by_title(resolved_slug, task_title)

        tasks.delete_subtask(resolved_slug, task_filename, title)
        typer.echo(f"🗑️ Deleted subtask: {title}")

    except ValueError as e:
        typer.echo(f"Error: {e}", err=True)
        raise typer.Exit(code=1)


if __name__ == "__main__":
    app()
