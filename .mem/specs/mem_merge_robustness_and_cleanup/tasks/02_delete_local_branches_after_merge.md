---
title: Delete local branches after merge
status: todo
subtasks: []
created_at: '2026-01-06T15:43:40.877193'
updated_at: '2026-01-06T15:43:40.877193'
completed_at: null
---
Currently only remote branches are deleted after merge. Need to also delete local branches.

After each successful PR merge in merge.py:
1. Delete remote branch (already implemented with delete_branch())
2. Delete local branch if it exists:
   - subprocess.run(['git', 'branch', '-d', branch_name])
   - Use -D if -d fails (force delete)
3. Prune stale remote tracking refs:
   - subprocess.run(['git', 'remote', 'prune', 'origin'])

Add helper function in sync.py or a new git_utils.py:
- delete_local_branch(branch_name) -> bool
- prune_remote_refs() -> None

Call these after delete_branch() succeeds in merge.py.