---
title: Add mem cleanup command for stale branches
status: completed
subtasks: []
created_at: '2026-01-06T15:43:51.325673'
updated_at: '2026-01-07T10:28:48.115697'
completed_at: '2026-01-07T10:28:48.115692'
---
Add a cleanup command to remove stale branches from completed/abandoned specs.

New command: mem cleanup (or add to mem sync with --cleanup flag)

Logic:
1. Get list of local branches matching 'dev-*' pattern
2. For each branch, extract the spec slug from branch name
3. Check if spec exists in completed/ or abandoned/
4. If yes, delete the local branch
5. Run git remote prune origin to clean tracking refs
6. Report what was deleted

Location: Create src/commands/cleanup.py or add to sync.py

Output example:
  Cleaning up stale branches...
  Deleted: dev-user-old_spec (spec completed)
  Deleted: dev-user-another_spec (spec abandoned)
  Pruned 3 stale remote tracking refs
  Done.

## Completion Notes

Created src/commands/cleanup.py with cleanup command, registered in main.py