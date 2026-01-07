---
title: Create AGENTS.md and CLAUDE.md on mem init
status: completed
subtasks: []
created_at: '2026-01-07T12:59:13.710320'
updated_at: '2026-01-07T13:11:36.812393'
completed_at: '2026-01-07T13:11:36.812387'
---
Modify mem init to:
1. Copy AGENTS.md template from src/templates/ to project root
2. Create CLAUDE.md as a symlink pointing to AGENTS.md in the same directory

This gives one source of truth with two names for different agents.

## Completion Notes

Added create_agents_files function to init.py that copies AGENTS.md template to project root and creates CLAUDE.md as a symlink pointing to it.