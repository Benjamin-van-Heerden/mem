---
title: Update docs list command for core docs
status: completed
created_at: '2026-01-12T10:04:07.549027'
updated_at: '2026-01-12T10:10:08.394219'
completed_at: '2026-01-12T10:10:08.394212'
---
Update src/commands/docs.py to show core docs separately in 'mem docs list' output. Core docs should be marked as 'core' rather than indexed/unindexed, and should not appear in the indexing workflow.

## Completion Notes

Updated mem docs list to show core docs separately from indexed docs