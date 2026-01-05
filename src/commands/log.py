"""
Log command - Create work session logs
"""

import typer

from env_settings import ENV_SETTINGS
from src.utils import logs, specs


def log():
    """
    Create a work log for today's session.

    Creates a log file from the template. The AI agent should then
    fill in the {placeholders} based on what was accomplished.
    """
    # Get active spec if any
    active_spec = specs.get_active_spec()
    spec_slug = active_spec["slug"] if active_spec else None

    try:
        log_file = logs.create_log(spec_slug=spec_slug)
        relative_path = log_file.relative_to(ENV_SETTINGS.caller_dir)
        template_path = "src/templates/log.md"

        typer.echo(f"Created log file: {relative_path}")
        typer.echo("")
        typer.echo(
            f"Please read {template_path} and fill in the {{placeholders}} based on this session."
        )

    except ValueError as e:
        # Log already exists for today
        today_log = logs.get_today_log()
        if today_log:
            log_path = ENV_SETTINGS.logs_dir / today_log["filename"]
            relative_path = log_path.relative_to(ENV_SETTINGS.caller_dir)
            typer.echo(f"Log already exists for today: {relative_path}")
            typer.echo("")
            typer.echo("Update the existing log file if needed.")
        else:
            typer.echo(f"Error: {e}", err=True)
            raise typer.Exit(code=1)
    except Exception as e:
        typer.echo(f"Error: {e}", err=True)
        raise typer.Exit(code=1)
