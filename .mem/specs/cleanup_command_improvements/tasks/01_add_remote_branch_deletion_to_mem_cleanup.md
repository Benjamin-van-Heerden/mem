---
title: Add remote branch deletion to mem cleanup
status: completed
subtasks: []
created_at: '2026-01-07T11:17:51.104581'
updated_at: '2026-01-07T11:19:54.090436'
completed_at: '2026-01-07T11:19:54.090431'
---
Update cleanup.py to delete remote branches on GitHub for completed/abandoned specs. Use delete_branch from src/utils/github/api.py. Should handle cases where remote branch doesn't exist gracefully.

## Completion Notes

Added get_remote_branches() function and updated run_cleanup() to delete both local and remote branches via GitHub API