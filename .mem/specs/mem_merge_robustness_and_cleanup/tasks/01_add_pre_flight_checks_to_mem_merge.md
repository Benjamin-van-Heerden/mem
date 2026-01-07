---
title: Add pre-flight checks to mem merge
status: completed
subtasks: []
created_at: '2026-01-06T15:43:31.142487'
updated_at: '2026-01-07T10:28:34.808940'
completed_at: '2026-01-07T10:28:34.808935'
---
Before merging any PRs, mem merge should fail early if the local state isn't clean:

1. Check for uncommitted changes with 'git status --porcelain'
   - If dirty, show error: 'Uncommitted changes detected. Commit or stash before merging.'
   - Exit with code 1

2. Run git fetch and git pull before any GitHub operations
   - If pull fails (conflicts), show error and exit
   - This prevents merging on GitHub then failing to sync locally

3. Only proceed with GitHub PR merges after local state is verified clean

Location: src/commands/merge.py - add checks at start of merge() function

Import git_fetch_and_pull from src.commands.sync and reuse it.

## Completion Notes

Added check_working_directory_clean() and git_fetch_and_pull() at start of merge command