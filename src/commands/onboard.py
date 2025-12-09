"""
Onboard command - Build context for AI agents

This command reads project configuration, important files, and active specs/tasks
to provide comprehensive context for AI agents working on the project.
"""

import logging
import sqlite3
import sys
from pathlib import Path
from typing import Optional

try:
    import tomllib
except ImportError:
    import tomli as tomllib

from src.env_settings import ENV_SETTINGS
from src.utils.db import ensure_mem_initialized, get_db_connection

logger = logging.getLogger(__name__)


def read_config() -> dict:
    """Read the .mem/config.toml file"""
    if not ENV_SETTINGS.config_file.exists():
        logger.warning(f"No {ENV_SETTINGS.config_file} found. Using defaults.")
        return {}

    try:
        with open(ENV_SETTINGS.config_file, "rb") as f:
            return tomllib.load(f)
    except Exception as e:
        logger.error(f"Error reading config: {e}")
        return {}


def read_file_safely(file_path: Path) -> Optional[str]:
    """Read a file safely, returning None if it doesn't exist or can't be read"""
    try:
        if file_path.exists():
            return file_path.read_text()
    except Exception as e:
        logger.warning(f"Could not read {file_path}: {e}")
    return None


def get_active_specs(conn: sqlite3.Connection) -> list[dict]:
    """Get all active specs with their associated tasks"""
    cursor = conn.cursor()

    # Get active specs
    cursor.execute("""
        SELECT id, title, file_path, status, created_at, updated_at
        FROM specs
        WHERE status = 'active'
        ORDER BY updated_at DESC
    """)

    specs = []
    for row in cursor.fetchall():
        spec = dict(row)

        # Get tasks for this spec
        cursor.execute(
            """
            SELECT id, title, file_path, status, parent_id, created_at
            FROM tasks
            WHERE spec_id = ? AND parent_id IS NULL
            ORDER BY created_at
        """,
            (spec["id"],),
        )

        tasks = []
        for task_row in cursor.fetchall():
            task = dict(task_row)

            # Get subtasks for this task
            cursor.execute(
                """
                SELECT id, title, file_path, status, created_at
                FROM tasks
                WHERE parent_id = ?
                ORDER BY created_at
            """,
                (task["id"],),
            )

            task["subtasks"] = [dict(st) for st in cursor.fetchall()]
            tasks.append(task)

        spec["tasks"] = tasks
        specs.append(spec)

    return specs


def get_open_todos(conn: sqlite3.Connection) -> list[dict]:
    """Get all open todos"""
    cursor = conn.cursor()
    cursor.execute("""
        SELECT id, title, description, created_at
        FROM todos
        WHERE status = 'open'
        ORDER BY created_at DESC
    """)

    return [dict(row) for row in cursor.fetchall()]


def get_recent_logs(conn: sqlite3.Connection, limit: int = 5) -> list[dict]:
    """Get recent work logs"""
    cursor = conn.cursor()
    cursor.execute(
        """
        SELECT id, file_path, created_at
        FROM work_logs
        ORDER BY created_at DESC
        LIMIT ?
    """,
        (limit,),
    )

    return [dict(row) for row in cursor.fetchall()]


def format_specs_context(specs: list[dict]) -> str:
    """Format specs and tasks for context output"""
    if not specs:
        return "No active specifications."

    output = []
    output.append("# Active Specifications\n")

    for spec in specs:
        output.append(f"## Spec #{spec['id']}: {spec['title']}")
        output.append(f"Status: {spec['status']}")
        output.append(f"File: {spec['file_path']}")

        # Read spec file content if it exists
        spec_content = read_file_safely(Path(spec["file_path"]))
        if spec_content:
            output.append("\n### Specification Details:")
            output.append("```")
            output.append(spec_content.strip())
            output.append("```")

        # List tasks
        if spec["tasks"]:
            output.append("\n### Tasks:")
            for task in spec["tasks"]:
                status_icon = "✓" if task["status"] == "completed" else "○"
                output.append(
                    f"  {status_icon} Task #{task['id']}: {task['title']} ({task['status']})"
                )

                # List subtasks
                if task["subtasks"]:
                    for subtask in task["subtasks"]:
                        sub_icon = "✓" if subtask["status"] == "completed" else "○"
                        output.append(
                            f"    {sub_icon} Subtask #{subtask['id']}: {subtask['title']} ({subtask['status']})"
                        )
        else:
            output.append("\n### Tasks: None")

        output.append("")

    return "\n".join(output)


def format_todos_context(todos: list[dict]) -> str:
    """Format todos for context output"""
    if not todos:
        return "No open todos."

    output = []
    output.append("# Open Todos\n")

    for todo in todos:
        output.append(f"- Todo #{todo['id']}: {todo['title']}")
        if todo["description"]:
            output.append(f"  {todo['description']}")

    return "\n".join(output)


def format_important_files(config: dict) -> str:
    """Format important files content"""
    important_files = config.get("context", {}).get("important_files", [])

    if not important_files:
        return ""

    output = []
    output.append("# Important Project Files\n")

    for file_path_str in important_files:
        file_path = Path(file_path_str)
        content = read_file_safely(file_path)

        if content:
            output.append(f"## {file_path}")
            output.append("```")
            output.append(content.strip())
            output.append("```")
            output.append("")

    return "\n".join(output)


def format_project_info(config: dict) -> str:
    """Format project information from config"""
    project = config.get("project", {})

    if not project:
        return ""

    output = []
    output.append("# Project Information\n")

    if project.get("name"):
        output.append(f"**Name:** {project['name']}")

    if project.get("description"):
        output.append(f"**Description:** {project['description']}")

    if project.get("type"):
        output.append(f"**Type:** {project['type']}")

    output.append("")

    return "\n".join(output)


def format_rules(config: dict) -> str:
    """Format project-specific rules"""
    rules = config.get("rules", {})

    if not rules:
        return ""

    output = []
    output.append("# Project-Specific Rules\n")

    for rule_category, rule_content in rules.items():
        output.append(f"## {rule_category.title()} Rules")

        if isinstance(rule_content, dict):
            for key, value in rule_content.items():
                output.append(f"- **{key}:** {value}")
        else:
            output.append(str(rule_content))

        output.append("")

    return "\n".join(output)


def onboard():
    """
    Main onboard command - builds and outputs context for AI agents
    """
    if not ensure_mem_initialized():
        logger.error("mem is not initialized in this directory. Run 'mem init' first.")
        sys.exit(1)

    # Read configuration
    config = read_config()

    # Connect to database
    try:
        conn = get_db_connection()
    except Exception as e:
        logger.error(f"Error connecting to database: {e}")
        sys.exit(1)

    # Gather context
    try:
        specs = get_active_specs(conn)
        todos = get_open_todos(conn)
        recent_logs = get_recent_logs(conn)
    except Exception as e:
        logger.error(f"Error gathering context: {e}")
        sys.exit(1)
    finally:
        conn.close()

    # Build output
    output_sections = []

    # Header
    output_sections.append("=" * 80)
    output_sections.append("# MEM ONBOARD - Project Context")
    output_sections.append("=" * 80)
    output_sections.append("")

    # Project info
    project_info = format_project_info(config)
    if project_info:
        output_sections.append(project_info)

    # Important files
    important_files = format_important_files(config)
    if important_files:
        output_sections.append(important_files)

    # Rules
    rules = format_rules(config)
    if rules:
        output_sections.append(rules)

    # Active specs and tasks
    output_sections.append(format_specs_context(specs))
    output_sections.append("")

    # Todos
    output_sections.append(format_todos_context(todos))
    output_sections.append("")

    # Recent activity
    if recent_logs:
        output_sections.append("# Recent Work Logs\n")
        for log in recent_logs:
            output_sections.append(
                f"- Log #{log['id']}: {log['file_path']} ({log['created_at']})"
            )
        output_sections.append("")

    # Footer
    output_sections.append("=" * 80)
    output_sections.append("# How to Use mem")
    output_sections.append("=" * 80)
    output_sections.append("")
    output_sections.append("Use the following commands to interact with mem:")
    output_sections.append("")
    output_sections.append(
        '- `mem spec new "feature name"` - Create a new specification'
    )
    output_sections.append('- `mem task new "task description"` - Create a new task')
    output_sections.append(
        "- `mem task update <id> --status in_progress` - Update task status"
    )
    output_sections.append("- `mem task complete <id>` - Mark task as complete")
    output_sections.append("- `mem log new` - Create a work log entry")
    output_sections.append("- `mem onboard` - Refresh this context")
    output_sections.append("")

    # Print all sections
    print("\n".join(output_sections))
