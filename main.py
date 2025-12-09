#!/usr/bin/env python3
"""
mem - A command-line utility for managing project context in agentic coding workflows

Usage:
    mem init                                    # Initialize mem in current project
    mem onboard                                 # Build context for AI agent
    mem spec new "feature name"                 # Create new spec
    mem spec list                               # List all specs
    mem spec show <id>                          # Show spec details
    mem spec update <id>                        # Update spec
    mem spec complete <id>                      # Mark spec complete
    mem task new "task description"             # Create new task
    mem task list                               # List all tasks
    mem task show <id>                          # Show task details
    mem task update <id>                        # Update task
    mem task complete <id>                      # Mark task complete
    mem subtask new "subtask" --parent <id>     # Create subtask
    mem subtask list --parent <id>              # List subtasks for task
    mem subtask complete <id>                   # Mark subtask complete
    mem todo new "reminder"                     # Create todo
    mem todo list                               # List todos
    mem todo complete <id>                      # Mark todo complete
    mem todo delete <id>                        # Delete todo
    mem log new                                 # Create work log entry
    mem log list                                # List recent logs
    mem log show <id>                           # Show specific log
"""

import argparse
import logging
import sys
from pathlib import Path

# Configure logging
logging.basicConfig(
    level=logging.INFO,
    format="%(levelname)s - %(message)s",
    handlers=[logging.StreamHandler()],
)

logger = logging.getLogger(__name__)


def main():
    parser = argparse.ArgumentParser(
        prog="mem",
        description="A command-line utility for managing project context in agentic coding workflows",
    )

    subparsers = parser.add_subparsers(dest="command", help="Available commands")

    # Parse arguments
    args = parser.parse_args()

    # Show help if no command provided
    if not args.command:
        parser.print_help()
        sys.exit(1)

    # Execute the command
    try:
        args.func(args)
    except AttributeError:
        # Command without subcommand (e.g., just "mem spec")
        parser.print_help()
        sys.exit(1)
    except Exception as e:
        logger.error(f"Error: {e}")
        sys.exit(1)


if __name__ == "__main__":
    main()
