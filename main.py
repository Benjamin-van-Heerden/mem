#!/usr/bin/env python3
"""
mem - A command-line utility for managing project context in agentic coding workflows

Usage:
    mem cli                                     # Print "Hi there" (test command)
    mem ui                                      # Open Textual TUI
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

import typer
from textual.app import App, ComposeResult
from textual.widgets import Footer, Header, Static

from src.commands.init import init as init_command
from src.commands.spec import app as spec_app
from src.commands.subtask import app as subtask_app
from src.commands.task import app as task_app

# Create the main Typer app
app = typer.Typer(
    help="A command-line utility for managing project context in agentic coding workflows"
)

# Register sub-commands
app.command(name="init", help="Initialize mem in current project")(init_command)
app.add_typer(spec_app, name="spec", help="Manage specifications")
app.add_typer(task_app, name="task", help="Manage tasks")
app.add_typer(subtask_app, name="subtask", help="Manage subtasks")


# Simple CLI command for testing
@app.command()
def cli():
    """Print a simple greeting (test command)"""
    print("Hi there")


# Textual TUI App
class MemApp(App):
    """A Textual TUI for mem."""

    TITLE = "mem"
    SUB_TITLE = "Project Context Manager"

    CSS = """
    Screen {
        align: center middle;
    }

    Static {
        width: auto;
        height: auto;
        padding: 2 4;
        background: $primary;
        border: heavy $accent;
    }
    """

    BINDINGS = [
        ("q", "quit", "Quit"),
        ("d", "toggle_dark", "Toggle dark mode"),
    ]

    def compose(self) -> ComposeResult:
        """Create child widgets."""
        yield Header()
        yield Static(
            "Welcome to mem TUI!\n\nPress 'q' to quit, 'd' to toggle dark mode."
        )
        yield Footer()

    def action_toggle_dark(self) -> None:
        """Toggle dark mode."""
        self.dark = not self.dark


# UI command that launches the Textual app
@app.command()
def ui():
    """Launch the Textual TUI interface"""
    mem_app = MemApp()
    mem_app.run()


if __name__ == "__main__":
    app()
