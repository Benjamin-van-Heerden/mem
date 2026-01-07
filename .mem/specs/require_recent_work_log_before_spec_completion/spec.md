---
title: Require recent work log before spec completion
status: merge_ready
assigned_to: Benjamin-van-Heerden
issue_id: 17
issue_url: https://github.com/Benjamin-van-Heerden/mem/issues/17
branch: dev-benjamin_van_heerden-require_recent_work_log_before_spec_completion
pr_url: null
created_at: '2026-01-07T12:33:26.358977'
updated_at: '2026-01-07T12:44:34.767968'
completed_at: null
last_synced_at: '2026-01-07T12:35:00.017118'
local_content_hash: aace85264d07705b122d0058500d009310661f7fc42c723912eddbd25a6e17da
remote_content_hash: aace85264d07705b122d0058500d009310661f7fc42c723912eddbd25a6e17da
---
## Overview

Enforce that a work log is created immediately before completing a spec, ensuring developers document their work while it's fresh. Add a timing check to `mem spec complete` that verifies a log was created within the last 3 minutes.

## Goals

- Ensure work is documented before spec completion
- Provide escape hatch via `--no-log` flag for edge cases
- Give clear feedback when a recent log is missing

## Technical Approach

1. In `spec complete`, check for work logs linked to the spec with `created_at` within the last 3 minutes
2. If no recent log found, display error message prompting user to create one
3. Add `--no-log` flag to bypass this check for edge cases
4. Use the new `created_at` datetime field in log frontmatter for timing comparison

## Success Criteria

- `mem spec complete` fails if no work log created in last 3 minutes for the spec
- Error message clearly instructs user to run `mem log`
- `--no-log` flag bypasses the timing check
- Existing work log validation (at least one log exists) remains in place

## Notes

This builds on the multi-log support added in the previous spec, which changed logs to use `created_at` datetime instead of just `date`.
