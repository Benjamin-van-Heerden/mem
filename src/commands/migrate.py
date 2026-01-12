"""
Migrate command for converting agent_rules/ format to .mem/ format.

This is a one-off migration command for converting projects using the old
agent_rules/ context system to the mem format.
"""

from pathlib import Path

import typer

from src.utils.migrate import run_migration


def migrate(
    target_dir: Path = typer.Argument(
        ".",
        help="Path to project directory containing agent_rules/ folder (defaults to current directory)",
        exists=True,
        file_okay=False,
        dir_okay=True,
        resolve_path=True,
    ),
    dry_run: bool = typer.Option(
        False,
        "--dry-run",
        help="Preview migration without making changes",
    ),
) -> None:
    """Migrate agent_rules/ format to .mem/ format.

    This is a one-off migration for converting projects using the old
    agent_rules/ context system to the mem format.
    """
    agent_rules_dir = target_dir / "agent_rules"
    if not agent_rules_dir.exists():
        print(f"Error: No agent_rules/ directory found in: {target_dir}")
        raise typer.Exit(1)

    run_migration(target_dir, dry_run)
