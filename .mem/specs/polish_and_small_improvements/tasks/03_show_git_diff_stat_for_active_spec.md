---
title: Show git diff stat for active spec
status: completed
subtasks: []
created_at: '2026-01-08T12:32:37.276415'
updated_at: '2026-01-08T12:36:38.308916'
completed_at: '2026-01-08T12:36:38.308904'
---
When onboarding with an active spec (or when activating a spec), show git diff --stat against the base branch to summarize what files have already been modified

## Completion Notes

Added get_branch_diff_stat() to specs.py, integrated into onboard.py (shows under active spec section) and spec.py activate command