#!/usr/bin/env python3
"""
mem - A command-line utility for managing project context in agentic coding workflows

Usage:
    mem cli                                           # Print "Hi there" (test command)
    mem ui                                            # Open Textual TUI
    mem init                                          # Initialize mem in current project
    mem sync                                          # Synchronize with GitHub
    mem onboard                                       # Build context for AI agent
    mem log                                           # Create work session log
    mem spec new "feature name"                       # Create new spec
    mem spec list                                     # List all specs
    mem spec show <slug>                              # Show spec details
    mem spec activate <slug>                          # Activate a spec
    mem spec deactivate                               # Deactivate current spec
    mem spec complete <slug> <message>                # Mark spec complete
    mem task new "task description" --spec <slug>     # Create new task
    mem task list --spec <slug>                       # List all tasks
    mem task update <task> --spec <slug>              # Update task
    mem task complete <task> --spec <slug>            # Mark task complete

"""

import typer
from textual.app import App, ComposeResult
from textual.widgets import Footer, Header, Static

from src.commands.cleanup import cleanup as cleanup_command
from src.commands.docs import app as docs_app
from src.commands.init import init as init_command
from src.commands.log import log as log_command
from src.commands.merge import merge as merge_command
from src.commands.migrate import migrate as migrate_command
from src.commands.onboard import onboard as onboard_command
from src.commands.spec import app as spec_app
from src.commands.sync import sync as sync_command
from src.commands.task import app as task_app

# Create the main Typer app
app = typer.Typer(
    help="A command-line utility for managing project context in agentic coding workflows"
)

# Register sub-commands
app.command(name="init", help="Initialize mem in current project")(init_command)
app.command(name="sync", help="Synchronize with GitHub")(sync_command)
app.command(name="onboard", help="Build context for AI agent")(onboard_command)
app.command(name="log", help="Create work session log")(log_command)
app.command(name="merge", help="Merge pull requests for completed specs")(merge_command)
app.command(name="cleanup", help="Remove stale branches from completed specs")(
    cleanup_command
)
app.command(name="migrate", hidden=True)(migrate_command)
app.add_typer(spec_app, name="spec", help="Manage specifications")
app.add_typer(task_app, name="task", help="Manage tasks")
app.add_typer(docs_app, name="docs", help="Manage technical documentation")


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
