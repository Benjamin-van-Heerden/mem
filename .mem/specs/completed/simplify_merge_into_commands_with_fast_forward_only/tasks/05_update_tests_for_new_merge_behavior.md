---
title: Update tests for new merge behavior
status: completed
created_at: '2026-01-13T13:52:17.816250'
updated_at: '2026-01-13T14:15:15.339539'
completed_at: '2026-01-13T14:15:15.339531'
---
Update test_merge_into.py to reflect simplified flow. Remove tests for back-merge behavior. Ensure all tests pass.

## Completion Notes

Updated tests/test_merge_into.py: (1) Changed module docstring from back-merge to fast-forward only, (2) Updated test_into_main_dry_run_shows_all_steps to check for ff-only output instead of back-merge output, (3) Renamed test_into_main_all_branches_at_same_commit to test_into_main_test_and_main_at_same_commit and updated assertions since dev is no longer back-merged. All 12 tests pass.