---
title: Fix mem merge worktree cleanup
status: completed
subtasks: []
created_at: '2026-01-10T11:39:18.079610'
updated_at: '2026-01-10T12:02:27.255076'
completed_at: '2026-01-10T12:02:27.255070'
---
mem merge is not properly cleaning up worktree directories after specs are completed. Completed specs like config_important_infos_field and task_amendments_and_work_log_improvements still have their worktree directories remaining in ../mem-worktrees/

## Completion Notes

Added worktree cleanup to merge command. After merging a PR and deleting branches, the merge command now extracts the spec slug from the branch name and removes the associated worktree directory if it exists.