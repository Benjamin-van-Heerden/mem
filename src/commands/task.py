"""
Task command - Manage tasks and subtasks
"""

from typing import Optional

import typer

from src.utils.db import DBCursorCtx

app = typer.Typer(help="Manage tasks")


@app.command()
def new(
    title: str = typer.Argument(..., help="Title of the task"),
    spec_id: Optional[int] = typer.Option(None, "--spec", help="Link to spec ID"),
    detail: Optional[str] = typer.Option(
        None, "--detail", help="Additional task details"
    ),
):
    """
    Create a new task.

    Optionally link to a spec and add details.
    """
    with DBCursorCtx() as cursor:
        # Validate spec_id if provided
        if spec_id is not None:
            cursor.execute("SELECT id FROM specs WHERE id = ?", (spec_id,))
            if not cursor.fetchone():
                typer.echo(f"❌ Spec ID {spec_id} not found.", err=True)
                raise typer.Exit(code=1)

        # Insert task
        cursor.execute(
            """
            INSERT INTO tasks (title, spec_id, detail)
            VALUES (?, ?, ?)
            """,
            (title, spec_id, detail),
        )
        task_id = cursor.lastrowid

    typer.echo(f"✓ Created task (ID: {task_id})")
    if spec_id:
        typer.echo(f"  Linked to spec {spec_id}")
    if detail:
        typer.echo(f"  Detail: {detail[:50]}{'...' if len(detail) > 50 else ''}")

    typer.echo("\n✨ Task created successfully!")
    typer.echo("\nNext steps:")
    typer.echo("1. mem task list [--spec <id>]")
    typer.echo(f"2. mem task update {task_id} --status in_progress")
    typer.echo(f"3. mem subtask new 'subtask' --parent {task_id}")


if __name__ == "__main__":
    app()
