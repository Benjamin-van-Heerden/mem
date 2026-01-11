---
title: Register docs command in main.py
status: todo
created_at: '2026-01-11T12:41:52.730575'
updated_at: '2026-01-11T12:41:52.730575'
completed_at: null
---
Add to main.py:
- Import docs app from src.commands.docs
- Register with app.add_typer(docs_app, name='docs', help='Manage technical documentation')