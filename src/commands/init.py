"""
Init command - Initialize mem in a project
"""

import git
import typer
from typing_extensions import Annotated

from env_settings import ENV_SETTINGS
from src.utils.migrations_runner import MigrationTool

# Create a Typer app for the init command
app = typer.Typer()


def check_git_repo() -> bool:
    """
    Check if the current directory is a git repository root.

    Returns:
        bool: True if git repo exists in current directory, False otherwise
    """
    try:
        git.Repo(ENV_SETTINGS.caller_dir, search_parent_directories=False)
        return True
    except git.InvalidGitRepositoryError:
        return False


def create_default_config():
    """Create a default config.toml file"""
    config_content = """# mem configuration file
[project]
name = "My Project"
description = "Project description"
type = "python"  # or "typescript", "rust", etc.

[context]
# Files to always include in onboard
important_files = [
    "README.md",
]

# Directories to scan for context
scan_directories = ["src/"]

[rules.python]
# Python-specific coding patterns and rules
style = "Follow PEP 8"
type_hints = "Always use type hints"
testing = "Use pytest for testing"

[github]
enabled = false
# repository = "owner/repo"
# token_env = "GITHUB_TOKEN"
# default_assignee = "username"

[git]
branch_prefix = "feature"
auto_create_branch = true
require_assignment = true
"""

    with open(ENV_SETTINGS.config_file, "w") as f:
        f.write(config_content)

    typer.echo(f"✓ Created config file: {ENV_SETTINGS.config_file_stripped}")


@app.command()
def init(
    force: Annotated[
        bool,
        typer.Option("--force", "-f", help="Force reinitialization without prompting"),
    ] = False,
):
    """
    Initialize mem in the current project.

    This command creates the .mem directory structure and initializes the database.
    A git repository must exist before initializing mem.
    """

    # Check for git repository first
    if not check_git_repo():
        typer.echo(
            "❌ Error: No git repository found in current directory.",
            err=True,
        )
        typer.echo("\nPlease initialize a git repository first:", err=True)
        typer.echo("  git init", err=True)
        typer.echo("  git add .", err=True)
        typer.echo('  git commit -m "Initial commit"', err=True)
        raise typer.Exit(code=1)

    # Check if already initialized
    if ENV_SETTINGS.mem_dir.exists() and ENV_SETTINGS.db_file.exists():
        typer.echo("⚠️  mem is already initialized in this directory.")

        if not force:
            response = typer.confirm(
                "Reinitialize? This will keep existing data.",
                default=False,
            )
            if not response:
                typer.echo("Initialization cancelled.")
                raise typer.Exit(code=0)

    # Create .mem directory
    ENV_SETTINGS.mem_dir.mkdir(exist_ok=True)
    typer.echo(f"✓ Created directory: {ENV_SETTINGS.mem_dir_stripped}")

    # Create subdirectories
    ENV_SETTINGS.specs_dir.mkdir(exist_ok=True)

    ENV_SETTINGS.logs_dir.mkdir(exist_ok=True)
    typer.echo("✓ Created subdirectories")

    # Create config file if it doesn't exist
    if not ENV_SETTINGS.config_file.exists():
        create_default_config()
    else:
        typer.echo(f"✓ Config file already exists: {ENV_SETTINGS.config_file_stripped}")

    # Run migrations
    typer.echo("Running database migrations...")
    migration_tool = MigrationTool(
        db_path=ENV_SETTINGS.db_file, migrations_dir=ENV_SETTINGS.migrations_dir
    )

    try:
        migration_tool.run_up(silent=True)
        typer.echo("✓ Database initialized")
    except Exception as e:
        typer.echo(f"❌ Error running migrations: {e}", err=True)
        raise typer.Exit(code=1)

    typer.echo("\n✨ mem initialized successfully!")
    typer.echo("\nNext steps:")
    typer.echo(f"1. Edit {ENV_SETTINGS.config_file_stripped} to configure your project")
    typer.echo("2. Run 'mem spec new \"your feature\"' to create a specification")
    typer.echo("3. Run 'mem onboard' to build context for AI agents")


if __name__ == "__main__":
    app()
