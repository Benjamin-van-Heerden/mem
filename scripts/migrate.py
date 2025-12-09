#!/usr/bin/env python3
"""
Database Migration CLI Script for mem

Usage:
    # Generate a new migration file (dev use)
    python scripts/migrate.py --generate "create_initial_schema"

    # Apply migrations to the current project's .mem/mem.db
    python scripts/migrate.py --up

    # Rollback migrations
    python scripts/migrate.py --rollback -n 1

    # Show migration status
    python scripts/migrate.py --status
"""

import argparse
import logging
import sys
from pathlib import Path

# Add parent directory to path so we can import from src
sys.path.insert(0, str(Path(__file__).resolve().parent.parent))

from src.env_settings import ENV_SETTINGS
from src.utils.migrations_runner import MigrationTool

# Configure logging
logging.basicConfig(
    level=logging.INFO,
    format="%(levelname)s - %(message)s",
    handlers=[logging.StreamHandler()],
)

logger = logging.getLogger(__name__)


def main():
    parser = argparse.ArgumentParser(description="Database Migration Tool")
    group = parser.add_mutually_exclusive_group(required=True)

    group.add_argument(
        "--generate", metavar="NAME", help="Generate new migration files"
    )
    group.add_argument("--up", action="store_true", help="Run all pending migrations")
    group.add_argument("--rollback", action="store_true", help="Rollback migrations")
    group.add_argument("--status", action="store_true", help="Show migration status")

    parser.add_argument(
        "-n", type=int, default=1, help="Number of migrations to rollback (default: 1)"
    )

    parser.add_argument(
        "--db",
        type=Path,
        default=ENV_SETTINGS.db_file,
        help=f"Path to sqlite database (default: {ENV_SETTINGS.db_file})",
    )

    args = parser.parse_args()

    # Create migration tool instance
    migration_tool = MigrationTool(
        db_path=args.db, migrations_dir=ENV_SETTINGS.migrations_dir
    )

    try:
        if args.generate:
            migration_tool.generate_migration(args.generate)
        elif args.up:
            migration_tool.run_up()
        elif args.rollback:
            migration_tool.run_rollback(args.n)
        elif args.status:
            migration_tool.show_status()
    except Exception as e:
        logger.error(f"Error: {e}")
        sys.exit(1)


if __name__ == "__main__":
    main()
