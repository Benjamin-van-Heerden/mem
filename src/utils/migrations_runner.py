"""
Database Migration Runner for mem

This module provides the core migration functionality that can be used
by both the CLI script and the `mem init` command.
"""

import logging
import re
import sqlite3
import textwrap
from datetime import datetime
from pathlib import Path

logger = logging.getLogger(__name__)


class MigrationTool:
    """Handles database migrations for mem"""

    def __init__(self, db_path: Path, migrations_dir: Path):
        """
        Initialize the migration tool.

        Args:
            db_path: Path to the SQLite database file
            migrations_dir: Path to the directory containing migration files
        """
        self.db_path = db_path
        self.migrations_dir = migrations_dir

    def get_connection(self):
        """Get SQLite connection"""
        if not self.db_path.parent.exists():
            self.db_path.parent.mkdir(parents=True, exist_ok=True)

        conn = sqlite3.connect(self.db_path)
        conn.row_factory = sqlite3.Row
        return conn

    def ensure_migrations_table(self):
        """Ensure migrations tracking table exists"""
        try:
            with self.get_connection() as conn:
                cursor = conn.cursor()
                cursor.execute("""
                    CREATE TABLE IF NOT EXISTS _migrations (
                        version TEXT PRIMARY KEY,
                        applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
                    )
                """)
                conn.commit()
        except Exception as e:
            logger.error(f"Error creating migrations table: {e}")
            raise

    def get_timestamp(self) -> str:
        """Generate timestamp for migration files"""
        return datetime.now().strftime("%Y%m%d%H%M%S")

    def generate_migration(self, name: str):
        """Generate new migration files"""
        if not self.migrations_dir.exists():
            self.migrations_dir.mkdir(parents=True, exist_ok=True)

        if not re.match(r"^[a-zA-Z0-9_]+$", name):
            raise ValueError(
                "Migration name can only contain letters, numbers, and underscores"
            )

        timestamp = self.get_timestamp()
        up_file = self.migrations_dir / f"{timestamp}_{name}_UP.sql"
        down_file = self.migrations_dir / f"{timestamp}_{name}_DOWN.sql"

        # UP migration template
        up_content = (
            textwrap.dedent("""
            -- Migration: {name}
            -- Created: {created}
            -- Description: Add your migration description here

            -- Add your UP migration SQL here
            -- Example:
            -- CREATE TABLE example_table (
            --     id INTEGER PRIMARY KEY,
            --     name TEXT NOT NULL,
            --     created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
            -- );
            """)
            .format(name=name, created=datetime.now().strftime("%Y-%m-%d %H:%M:%S"))
            .strip()
        )

        # DOWN migration template
        down_content = (
            textwrap.dedent("""
            -- Migration Rollback: {name}
            -- Created: {created}
            -- Description: Rollback for {name} migration

            -- Add your DOWN migration SQL here (reverse of UP migration)
            -- Example:
            -- DROP TABLE IF EXISTS example_table;
            """)
            .format(name=name, created=datetime.now().strftime("%Y-%m-%d %H:%M:%S"))
            .strip()
        )

        # Write files
        with open(up_file, "w") as f:
            f.write(up_content)

        with open(down_file, "w") as f:
            f.write(down_content)

        logger.info("✓ Generated migration files:")
        logger.info(f"  UP:   {up_file}")
        logger.info(f"  DOWN: {down_file}")

        return up_file, down_file

    def get_pending_migrations(self) -> list[str]:
        """Get list of pending migrations"""
        try:
            with self.get_connection() as conn:
                cursor = conn.cursor()
                # Get applied migrations
                cursor.execute("SELECT version FROM _migrations ORDER BY version")
                results = cursor.fetchall()
                applied = {str(row["version"]) for row in results}

            # Get all migration files
            if not self.migrations_dir.exists():
                return []

            up_files = list(self.migrations_dir.glob("*_UP.sql"))
            all_migrations: list[str] = []

            for up_file in up_files:
                # Extract version from filename (timestamp_name_UP.sql)
                filename = up_file.name
                version = filename.replace("_UP.sql", "")
                all_migrations.append(version)

            # Sort migrations by timestamp
            all_migrations.sort()

            # Return pending migrations
            pending = [m for m in all_migrations if m not in applied]
            return pending

        except Exception as e:
            logger.error(f"Error checking migrations: {e}")
            raise

    def get_applied_migrations(self) -> list[str]:
        """Get list of applied migrations"""
        try:
            with self.get_connection() as conn:
                cursor = conn.cursor()
                cursor.execute("SELECT version FROM _migrations ORDER BY version DESC")
                results = cursor.fetchall()
                return [str(row["version"]) for row in results]
        except Exception as e:
            logger.error(f"Error getting applied migrations: {e}")
            raise

    def apply_migration(self, version: str):
        """Apply a single migration"""
        up_file = self.migrations_dir / f"{version}_UP.sql"

        if not up_file.exists():
            raise FileNotFoundError(f"Migration file not found: {up_file}")

        try:
            # Read and execute migration
            with open(up_file, "r") as f:
                sql_content = f.read()

            with self.get_connection() as conn:
                cursor = conn.cursor()

                # We use executescript for the migration content to ensure it runs as a batch
                cursor.executescript(sql_content)

                # Record migration as applied
                cursor.execute(
                    "INSERT INTO _migrations (version) VALUES (?)",
                    (version,),
                )
                conn.commit()

            logger.info(f"✓ Applied migration: {version}")

        except Exception as e:
            logger.error(f"✗ Error applying migration {version}: {e}")
            raise

    def rollback_migration(self, version: str):
        """Rollback a single migration"""
        down_file = self.migrations_dir / f"{version}_DOWN.sql"

        if not down_file.exists():
            raise FileNotFoundError(f"Rollback file not found: {down_file}")

        try:
            # Read and execute rollback
            with open(down_file, "r") as f:
                sql_content = f.read()

            with self.get_connection() as conn:
                cursor = conn.cursor()
                cursor.executescript(sql_content)

                # Remove migration record
                cursor.execute(
                    "DELETE FROM _migrations WHERE version = ?",
                    (version,),
                )
                conn.commit()

            logger.info(f"✓ Rolled back migration: {version}")

        except Exception as e:
            logger.error(f"✗ Error rolling back migration {version}: {e}")
            raise

    def run_up(self, silent: bool = False):
        """
        Run all pending migrations.

        Args:
            silent: If True, suppress "No pending migrations" message
        """
        self.ensure_migrations_table()

        pending = self.get_pending_migrations()

        if not pending:
            if not silent:
                logger.info("✓ No pending migrations")
            return

        logger.info(f"Found {len(pending)} pending migration(s):")
        for migration in pending:
            logger.info(f"  - {migration}")

        logger.info("")
        for migration in pending:
            self.apply_migration(migration)

        logger.info(f"\n✓ Applied {len(pending)} migration(s)")

    def run_rollback(self, count: int):
        """Rollback the last n migrations"""
        self.ensure_migrations_table()

        applied = self.get_applied_migrations()

        if not applied:
            logger.info("✓ No migrations to rollback")
            return

        if count > len(applied):
            raise ValueError(
                f"Cannot rollback {count} migrations, only {len(applied)} applied"
            )

        to_rollback = applied[:count]

        logger.info(f"Rolling back {count} migration(s):")
        for migration in to_rollback:
            logger.info(f"  - {migration}")

        logger.info("")
        for migration in to_rollback:
            self.rollback_migration(migration)

        logger.info(f"\n✓ Rolled back {count} migration(s)")

    def show_status(self):
        """Show migration status"""
        self.ensure_migrations_table()

        applied = self.get_applied_migrations()
        pending = self.get_pending_migrations()

        logger.info("Migration Status:")
        logger.info(f"  Database: {self.db_path}")
        logger.info(f"  Applied: {len(applied)}")
        logger.info(f"  Pending: {len(pending)}")

        if applied:
            logger.info("\nApplied migrations:")
            for migration in reversed(applied):
                logger.info(f"  ✓ {migration}")

        if pending:
            logger.info("\nPending migrations:")
            for migration in pending:
                logger.info(f"  ⏳ {migration}")
