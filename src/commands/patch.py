"""
Patch command - Update configuration files to match current schema
"""

import tomllib
from typing import Annotated

import typer

from env_settings import ENV_SETTINGS
from src.config.main_config import (
    find_unknown_key_paths,
    generate_default_config_toml,
    load_and_validate_local_config,
)
from src.config.models import MemLocalConfig

app = typer.Typer(help="Patch configuration files")


def _extract_known_values(raw: dict) -> dict:
    """
    Extract known values from raw config that should be preserved.

    Returns a flat dict with keys like 'project.name', 'worktree.symlink_paths', etc.
    """
    values = {}

    project = raw.get("project", {})
    if "name" in project:
        values["project.name"] = project["name"]
    if "description" in project:
        values["project.description"] = project["description"]
    if "generic_templates" in project:
        values["project.generic_templates"] = project["generic_templates"]

    if "files" in raw:
        values["files"] = raw["files"]

    worktree = raw.get("worktree", {})
    if "symlink_paths" in worktree:
        values["worktree.symlink_paths"] = worktree["symlink_paths"]

    return values


def _filter_valid_files(files: list[dict]) -> list[dict]:
    """Filter files list to only include entries with valid 'path' keys."""
    valid = []
    for f in files:
        if isinstance(f, dict) and "path" in f:
            valid.append({"path": f["path"], "description": f.get("description")})
    return valid


@app.command(name="config")
def patch_config(
    dry_run: Annotated[
        bool,
        typer.Option(
            "--dry-run", "-n", help="Show what would change without modifying"
        ),
    ] = False,
):
    """
    Update .mem/config.toml to match the current schema.

    This command:
    - Removes unknown keys (not part of the supported schema)
    - Adds missing keys with sensible defaults
    - Preserves user-set values for known keys
    - Is idempotent (running twice produces no further changes)
    """
    config_path = ENV_SETTINGS.config_file

    if not config_path.exists():
        typer.echo(f"❌ Config file not found: {ENV_SETTINGS.config_file_stripped}")
        typer.echo("Run 'mem init' to create one.")
        raise typer.Exit(code=1)

    result = load_and_validate_local_config(config_path)
    raw = result.raw

    if not raw:
        typer.echo(f"❌ Could not parse {ENV_SETTINGS.config_file_stripped}")
        raise typer.Exit(code=1)

    unknown_paths = find_unknown_key_paths(raw, MemLocalConfig)
    known_values = _extract_known_values(raw)

    has_changes = False
    changes: list[str] = []

    if unknown_paths:
        has_changes = True
        changes.append(f"Remove {len(unknown_paths)} unknown key(s):")
        for path in unknown_paths:
            changes.append(f"  - {path}")

    project_name = known_values.get("project.name")
    project_desc = known_values.get("project.description")
    generic_templates = known_values.get("project.generic_templates")
    files = known_values.get("files", [])
    symlink_paths = known_values.get("worktree.symlink_paths")

    missing_keys = []
    if "project" not in raw:
        missing_keys.append("[project] section")
    if "worktree" not in raw:
        missing_keys.append("[worktree] section")
    if "project" in raw and "generic_templates" not in raw.get("project", {}):
        missing_keys.append("project.generic_templates")
    if "worktree" in raw and "symlink_paths" not in raw.get("worktree", {}):
        missing_keys.append("worktree.symlink_paths")

    if missing_keys:
        has_changes = True
        changes.append(f"Add {len(missing_keys)} missing key(s):")
        for key in missing_keys:
            changes.append(f"  + {key}")

    if not has_changes:
        typer.echo("✅ Config is already up to date. No changes needed.")
        return

    if not project_name:
        project_name = ENV_SETTINGS.caller_dir.name
        changes.append(f'  + project.name = "{project_name}" (default)')

    if not project_desc:
        project_desc = "Add your project description here."
        changes.append("  + project.description (default)")

    valid_files = _filter_valid_files(files) if files else None

    new_config = generate_default_config_toml(
        project_name=project_name,
        project_description=project_desc,
        generic_templates=generic_templates,
        important_files=valid_files,
        symlink_paths=symlink_paths,
    )

    if dry_run:
        typer.echo("Dry run - the following changes would be made:\n")
        for change in changes:
            typer.echo(change)
        typer.echo(
            f"\nNew config would be written to: {ENV_SETTINGS.config_file_stripped}"
        )
        typer.echo("\nRun without --dry-run to apply changes.")
        return

    config_path.write_text(new_config)

    typer.echo(f"✅ Updated {ENV_SETTINGS.config_file_stripped}\n")
    for change in changes:
        typer.echo(change)
