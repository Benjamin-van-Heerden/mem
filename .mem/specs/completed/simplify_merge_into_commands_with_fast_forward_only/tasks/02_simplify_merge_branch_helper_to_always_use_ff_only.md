---
title: Simplify _merge_branch helper to always use ff-only
status: completed
created_at: '2026-01-13T13:52:10.666357'
updated_at: '2026-01-13T14:05:13.340673'
completed_at: '2026-01-13T14:05:13.340664'
---
Remove the ff_only parameter. Always use --ff-only flag. Remove the -m merge message logic.

## Completion Notes

Removed ff_only parameter from _merge_branch function. Now always uses --ff-only flag. Updated docstring accordingly.