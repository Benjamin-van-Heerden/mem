"""
Subtask command - Manage subtasks (tasks with parent_id)
"""

from typing import Optional

import typer

from src.utils.db import DBCursorCtx

app = typer.Typer(help="Manage subtasks")


@app.command()
def new(
    title: str = typer.Argument(..., help="Title of the subtask"),
    parent_id: int = typer.Option(
        ..., "--parent", help="Parent task ID (must be top-level task)"
    ),
    detail: Optional[str] = typer.Option(
        None, "--detail", help="Additional subtask details"
    ),
):
    """
    Create a new subtask under a parent task.

    Subtasks inherit the spec_id from their parent task.
    Parent must be a top-level task (no parent_id).
    """
    with DBCursorCtx() as cursor:
        # Validate parent task exists and is top-level
        cursor.execute(
            """
            SELECT id, title, spec_id
            FROM tasks
            WHERE id = ? AND parent_id IS NULL
            """,
            (parent_id,),
        )
        parent = cursor.fetchone()
        if not parent:
            typer.echo(
                f"❌ Parent task ID {parent_id} not found or is itself a subtask.",
                err=True,
            )
            raise typer.Exit(code=1)

        parent_title = parent["title"]
        parent_spec_id = parent["spec_id"]

        # Insert subtask
        cursor.execute(
            """
            INSERT INTO tasks (title, parent_id, spec_id, detail)
            VALUES (?, ?, ?, ?)
            """,
            (title, parent_id, parent_spec_id, detail),
        )
        subtask_id = cursor.lastrowid

    typer.echo(f"✓ Created subtask (ID: {subtask_id})")
    typer.echo(f"  Parent: Task #{parent_id} '{parent_title}'")
    if parent_spec_id:
        typer.echo(f"  Spec: {parent_spec_id}")
    if detail:
        preview = detail[:50] + "..." if len(detail) > 50 else detail
        typer.echo(f"  Detail: {preview}")

    typer.echo("\n✨ Subtask created successfully!")
    typer.echo("\nNext steps:")
    typer.echo(f"1. mem subtask list --parent {parent_id}")
    typer.echo(f"2. mem subtask complete {subtask_id}")
    typer.echo(f"3. mem task complete {parent_id} (after all subtasks complete)")


if __name__ == "__main__":
    app()
