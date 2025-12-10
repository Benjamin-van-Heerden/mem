"""
Database utility functions for mem
"""

import sqlite3
from pathlib import Path
from typing import Any, Optional

from env_settings import ENV_SETTINGS


class DBCursorCtx:
    def __init__(self, db_path: Optional[Path] = None):
        """
        Initialize database cursor context manager.

        Args:
            db_path: Path to database file. If None, uses .mem/mem.db in current directory.
        """
        self.db_path = db_path or ENV_SETTINGS.db_file
        self.connection: sqlite3.Connection | None = None
        self.cursor: sqlite3.Cursor | None = None

    def __enter__(self):
        if not self.db_path.exists():
            raise FileNotFoundError(
                f"Database not found at {self.db_path}. Have you run 'mem init'?"
            )

        self.connection = sqlite3.connect(self.db_path, autocommit=True)
        self.connection.row_factory = sqlite3.Row
        self.cursor = self.connection.cursor()
        return self.cursor

    def __exit__(self, exc_type: Any, exc_val: Any, exc_tb: Any) -> None:
        if self.cursor:
            self.cursor.close()
        if self.connection:
            self.connection.close()


def ensure_mem_initialized() -> bool:
    """
    Check if mem is initialized in the current directory.

    Returns:
        True if .mem/mem.db exists, False otherwise
    """
    return ENV_SETTINGS.db_file.exists()
