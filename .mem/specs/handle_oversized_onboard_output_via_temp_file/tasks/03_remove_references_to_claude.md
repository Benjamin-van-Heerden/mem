---
title: Remove references to claude
status: completed
created_at: '2026-01-14T14:25:13.744521'
updated_at: '2026-01-14T15:11:12.949167'
completed_at: '2026-01-14T15:11:12.949160'
---
Find references to 'claude' across the repo and remove/replace them. In src/commands/init.py, do not create CLAUDE.md or a CLAUDE.md symlink; only create AGENTS.md. Update any user-facing messages/docs that mention running 'claude' to instead say 'your preferred agent'.

## Completion Notes

Removed user-facing references to 'claude' in CLI output and stopped mem init from creating CLAUDE.md (AGENTS.md only).