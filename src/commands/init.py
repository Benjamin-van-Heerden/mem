"""
Init command - Initialize mem in a project
"""

import logging
import sys

from src.env_settings import ENV_SETTINGS
from src.utils.migrations_runner import MigrationTool

logger = logging.getLogger(__name__)


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
"""

    with open(ENV_SETTINGS.config_file, "w") as f:
        f.write(config_content)

    logger.info(f"✓ Created config file: {ENV_SETTINGS.config_file}")


def init_project():
    """Initialize mem in the current project"""

    # Check if already initialized
    if ENV_SETTINGS.mem_dir.exists() and ENV_SETTINGS.db_file.exists():
        logger.warning("mem is already initialized in this directory.")
        response = input("Reinitialize? This will keep existing data. (y/N): ")
        if response.lower() != "y":
            logger.info("Initialization cancelled.")
            sys.exit(0)

    # Create .mem directory
    ENV_SETTINGS.mem_dir.mkdir(exist_ok=True)
    logger.info(f"✓ Created directory: {ENV_SETTINGS.mem_dir}")

    # Create subdirectories
    ENV_SETTINGS.specs_dir.mkdir(exist_ok=True)
    ENV_SETTINGS.tasks_dir.mkdir(exist_ok=True)
    ENV_SETTINGS.logs_dir.mkdir(exist_ok=True)
    logger.info("✓ Created subdirectories")

    # Create config file if it doesn't exist
    if not ENV_SETTINGS.config_file.exists():
        create_default_config()
    else:
        logger.info(f"✓ Config file already exists: {ENV_SETTINGS.config_file}")

    # Run migrations
    logger.info("Running database migrations...")
    migration_tool = MigrationTool(
        db_path=ENV_SETTINGS.db_file, migrations_dir=ENV_SETTINGS.migrations_dir
    )

    try:
        migration_tool.run_up(silent=True)
        logger.info("✓ Database initialized")
    except Exception as e:
        logger.error(f"Error running migrations: {e}")
        sys.exit(1)

    logger.info("\n✓ mem initialized successfully!")
    logger.info("\nNext steps:")
    logger.info(f"1. Edit {ENV_SETTINGS.config_file} to configure your project")
    logger.info("2. Run 'mem spec new \"your feature\"' to create a specification")
    logger.info("3. Run 'mem onboard' to build context for AI agents")
