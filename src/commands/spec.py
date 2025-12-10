"""
Spec command - Manage specifications
"""

from datetime import datetime

import typer
from typing_extensions import Annotated

from env_settings import ENV_SETTINGS
from src.utils.db import DBCursorCtx

app = typer.Typer()


def slugify(text: str) -> str:
    """Convert text to a slug suitable for filenames."""
    return text.lower().replace(" ", "_").replace("-", "_")


@app.command()
def new(
    title: Annotated[str, typer.Argument(help="Title of the specification")],
):
    """
    Create a new specification.

    This creates a database entry and a markdown file in .mem/specs/
    """

    # Generate filename
    date_str = datetime.now().strftime("%Y%m%d")
    slug = slugify(title)
    filename = f"{date_str}_{slug}.md"
    file_path = ENV_SETTINGS.specs_dir / filename

    # Check if file already exists
    if file_path.exists():
        relative_path_str = f"{ENV_SETTINGS.specs_dir_stripped}/{filename}"
        typer.echo(f"❌ Error: Spec file already exists: {relative_path_str}", err=True)
        raise typer.Exit(code=1)

    # Create the markdown file
    spec_content = f"""# {title}

## Overview

Brief description of what this specification aims to achieve.

## Goals

- Goal 1
- Goal 2
- Goal 3

## Technical Approach

Describe the technical approach here.

## Success Criteria

- [ ] Criterion 1
- [ ] Criterion 2
- [ ] Criterion 3

## Notes

Additional notes, considerations, or references.
"""

    try:
        file_path.write_text(spec_content)
        relative_path_str = f"{ENV_SETTINGS.specs_dir_stripped}/{filename}"
        typer.echo(f"✓ Created spec file: {relative_path_str}")
    except Exception as e:
        typer.echo(f"❌ Error creating spec file: {e}", err=True)
        raise typer.Exit(code=1)

    # Insert into database
    # Store relative path from caller_dir
    relative_path = file_path.relative_to(ENV_SETTINGS.caller_dir)

    try:
        with DBCursorCtx() as cursor:
            cursor.execute(
                """
                INSERT INTO specs (title, file_path)
                VALUES (?, ?)
                """,
                (title, str(relative_path)),
            )
            spec_id = cursor.lastrowid

        typer.echo(f"✓ Created spec in database (ID: {spec_id})")
        typer.echo("\n✨ Spec created successfully!")
        typer.echo("\nNext steps:")
        relative_path_str = f"{ENV_SETTINGS.specs_dir_stripped}/{filename}"
        typer.echo(f"1. Edit the spec file: {relative_path_str}")
        typer.echo(
            f'2. Create tasks with: mem task new "task description" --spec {spec_id}'
        )

    except Exception as e:
        typer.echo(f"❌ Database error: {e}", err=True)
        # Clean up the file if database insert failed
        if file_path.exists():
            file_path.unlink()
        raise typer.Exit(code=1)


@app.command()
def list(
    status: Annotated[
        str,
        typer.Option(
            "--status", "-s", help="Filter by status (active, completed, archived)"
        ),
    ] = "active",
):
    """
    List all specifications.

    By default, shows only active specs. Use --status to filter by other statuses.
    """

    try:
        with DBCursorCtx() as cursor:
            cursor.execute(
                """
                SELECT id, title, status, created_at, updated_at
                FROM specs
                WHERE status = ?
                ORDER BY updated_at DESC
                """,
                (status,),
            )
            specs = cursor.fetchall()

        if not specs:
            typer.echo(f"No {status} specs found.")
            return

        # Display specs in a formatted table
        typer.echo(f"\n{status.upper()} SPECS:\n")
        typer.echo(f"{'ID':<4} {'Title':<50} {'Created':<12} {'Updated':<12}")
        typer.echo("=" * 80)

        for spec in specs:
            spec_id = spec["id"]
            title = spec["title"]
            created_at = spec["created_at"]
            updated_at = spec["updated_at"]

            # Format dates to just show date part
            created = created_at.split()[0] if created_at else "N/A"
            updated = updated_at.split()[0] if updated_at else "N/A"

            # Truncate title if too long
            display_title = title[:47] + "..." if len(title) > 50 else title

            typer.echo(f"{spec_id:<4} {display_title:<50} {created:<12} {updated:<12}")

        typer.echo(f"\nTotal: {len(specs)} spec(s)")
        typer.echo("\nTo view a spec: mem spec show <id>")

    except Exception as e:
        typer.echo(f"❌ Database error: {e}", err=True)
        raise typer.Exit(code=1)


if __name__ == "__main__":
    app()
