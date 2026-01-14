---
title: Fix task complete signature in onboard key commands
status: todo
created_at: '2026-01-14T10:52:03.534182'
updated_at: '2026-01-14T10:53:35.499063'
---
In src/commands/onboard.py line 348, the key commands section shows 'mem task complete "title"' but it should be 'mem task complete "title" "notes"' to match the actual required signature. Update this to prevent agents from calling the command incorrectly.

## Amendments

IMPORTANT: Do NOT add any hint about the --accept flag in onboard. The --accept flag should only be revealed to the agent AFTER it attempts to complete a task (via the output of mem task complete without --accept). The onboard command should only show the basic signature: mem task complete "title" "notes"