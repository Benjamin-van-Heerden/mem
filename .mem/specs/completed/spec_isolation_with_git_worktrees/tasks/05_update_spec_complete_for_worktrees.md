---
title: Update spec complete for worktrees
status: completed
subtasks: []
created_at: '2026-01-09T13:39:28.765423'
updated_at: '2026-01-09T14:02:52.427375'
completed_at: '2026-01-09T14:02:52.427367'
---
Modify spec complete to work from main repo or worktree. After creating PR, remove the worktree (keep branch). Update to handle worktree cleanup gracefully.

## Completion Notes

Updated complete command: now checks active spec via get_active_spec() instead of branch comparison. Removed 'switch back to dev' step - user stays in worktree, cleanup happens later via mem merge. Updated next steps message.