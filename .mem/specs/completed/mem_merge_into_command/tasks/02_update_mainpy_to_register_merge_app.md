---
title: Update main.py to register merge app
status: completed
created_at: '2026-01-12T11:09:49.894544'
updated_at: '2026-01-12T11:27:00.578021'
completed_at: '2026-01-12T11:27:00.578012'
---
Change main.py to import and register the merge Typer app instead of the single merge function. Use app.add_typer() like spec and docs commands do.

## Completion Notes

Changed import from merge function to merge app, updated registration to use add_typer instead of command