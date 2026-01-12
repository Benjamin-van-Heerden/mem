---
title: Convert merge.py to Typer sub-app
status: completed
created_at: '2026-01-12T11:09:44.175999'
updated_at: '2026-01-12T11:26:34.759624'
completed_at: '2026-01-12T11:26:34.759617'
---
Convert src/commands/merge.py from exporting a single function to using a Typer app pattern (like spec.py). Create 'app = typer.Typer()' and make the existing merge function a command on it. Keep all existing functionality intact.

## Completion Notes

Added Typer app to merge.py, converted merge function to use @app.callback with invoke_without_command=True pattern