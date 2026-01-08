"""
Log command - Create work session logs
"""

import typer

from env_settings import ENV_SETTINGS
from src.utils import logs, specs


def log():
    """
    Create a work log for this session.

    Creates a log file from the template. Multiple logs per day are supported.
    The AI agent should then fill in the {placeholders} based on what was accomplished.
    """
    # Get active spec if any
    active_spec = specs.get_active_spec()
    spec_slug = active_spec["slug"] if active_spec else None

    try:
        log_file = logs.create_log(spec_slug=spec_slug)
        relative_path = log_file.relative_to(ENV_SETTINGS.caller_dir)

        typer.echo(f"Created log file: {relative_path}")
        typer.echo("")
        typer.echo(
            "Please read the file and fill in the {placeholders} based on our current interaction session."
        )

        if active_spec:
            typer.echo("")
            typer.echo("If this is the LAST log before completing the spec:")
            typer.echo(
                "  No action needed - `mem spec complete` handles git automatically."
            )
            typer.echo("")
            typer.echo("Otherwise, commit and push your changes:")
            typer.echo(
                "  git add -A && git commit -m '<describe what was done>' && git push"
            )

    except Exception as e:
        typer.echo(f"Error: {e}", err=True)
        raise typer.Exit(code=1)
