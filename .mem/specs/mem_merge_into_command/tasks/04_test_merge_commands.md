---
title: Test merge commands
status: completed
created_at: '2026-01-12T11:10:04.025209'
updated_at: '2026-01-12T11:41:38.024484'
completed_at: '2026-01-12T11:41:38.024477'
---
Verify: 1) mem merge still works for PR merging, 2) mem merge into test works from dev and fails from other branches, 3) mem merge into main works from test and fails from other branches, 4) After successful merge both branches are at same commit.

## Completion Notes

Added 12 tests in test_merge_into.py covering validation, merge into test, merge into main, and error handling. Fixed _pull_branch to explicitly specify remote/branch and used unique filenames to avoid test conflicts.