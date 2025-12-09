"""
Database utility functions for mem
"""

import sqlite3
from pathlib import Path
from typing import Optional

from env_settings import ENV_SETTINGS


def get_db_connection(db_path: Optional[Path] = None) -> sqlite3.Connection:
    """
    Get a connection to the mem database.

    Args:
        db_path: Path to database file. If None, uses .mem/mem.db in current directory

    Returns:
        SQLite connection with row_factory set to Row
    """
    if db_path is None:
        db_path = ENV_SETTINGS.db_file

    if not db_path.exists():
        raise FileNotFoundError(
            f"Database not found at {db_path}. Have you run 'mem init'?"
        )

    conn = sqlite3.connect(db_path)
    conn.row_factory = sqlite3.Row
    return conn


def ensure_mem_initialized() -> bool:
    """
    Check if mem is initialized in the current directory.

    Returns:
        True if .mem/mem.db exists, False otherwise
    """
    return ENV_SETTINGS.db_file.exists()
