---
title: Convert merge.py to Typer sub-app
status: todo
created_at: '2026-01-12T11:09:44.175999'
updated_at: '2026-01-12T11:09:44.175999'
completed_at: null
---
Convert src/commands/merge.py from exporting a single function to using a Typer app pattern (like spec.py). Create 'app = typer.Typer()' and make the existing merge function a command on it. Keep all existing functionality intact.