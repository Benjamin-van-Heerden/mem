---
title: Update init to create core docs directory
status: completed
created_at: '2026-01-12T10:03:48.953024'
updated_at: '2026-01-12T10:08:54.467868'
completed_at: '2026-01-12T10:08:54.467862'
---
Update src/commands/init.py to create .mem/docs/core/ directory alongside the existing .mem/docs/ setup

## Completion Notes

Added _get_core_docs_dir() function and updated ensure_docs_dirs() to create the core directory