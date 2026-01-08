---
title: Implement git merge rules validation
status: completed
subtasks: []
created_at: '2026-01-08T10:20:55.309750'
updated_at: '2026-01-08T10:41:10.518303'
completed_at: '2026-01-08T10:41:10.518293'
---
Add a mem command or hook that enforces: anything can merge to dev, only dev/hotfix branches can merge to test, only test can merge to main.

## Completion Notes

Added pre-merge-commit git hook enforcing branch rules (anything->dev, dev/hotfix->test, test->main). Added setup_branch_protection() to GitHub API and called it during mem init. Both local and GitHub-side enforcement now in place.