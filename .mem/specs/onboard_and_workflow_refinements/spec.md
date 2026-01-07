---
title: Onboard and workflow refinements
status: merge_ready
assigned_to: Benjamin-van-Heerden
issue_id: 15
issue_url: https://github.com/Benjamin-van-Heerden/mem/issues/15
branch: dev-benjamin_van_heerden-onboard_and_workflow_refinements
pr_url: null
created_at: '2026-01-07T11:37:48.359696'
updated_at: '2026-01-07T12:03:29.727226'
completed_at: null
last_synced_at: '2026-01-07T11:45:44.736098'
local_content_hash: 5d32b8b4666bfa9179b3a7a9f52f59f860bc255eb5d73cee54ecfad51f3dfce7
remote_content_hash: 5d32b8b4666bfa9179b3a7a9f52f59f860bc255eb5d73cee54ecfad51f3dfce7
---
## Overview

Refinements to `mem onboard` output and workflow commands to improve clarity and fix status label ordering issues.

## Goals

- Display full work log content in onboard (no truncation)
- Require at least one work log before spec completion
- Remove installation/prerequisite sections from onboard (fail gracefully if missing)
- Add clear file/section separators in onboard output
- Make sync hard (not dry-run) during onboard
- Fix status label ordering: `mem spec complete` should set 'merge_ready', not `mem merge`

## Technical Approach

1. Update `src/commands/onboard.py` to show full work log content
2. Add work log validation to `mem spec complete`
3. Remove README sections for installation/prerequisites from onboard context
4. Add visual separators (e.g., `=====`) between file contents in IMPORTANT FILES and CODING GUIDELINES
5. Change sync call in onboard from dry-run to actual sync
6. Move 'merge_ready' label sync from merge command to spec complete command

## Success Criteria

- Work logs display in full during onboard
- `mem spec complete` fails if no work logs exist for the spec
- Onboard output has clear visual separation between files/sections
- Sync during onboard performs actual sync (not dry-run)
- `mem spec complete` updates GitHub issue label to 'merge_ready'
- `mem merge` no longer updates status labels (they're already correct)

## Notes

The status label ordering fix is important: currently `mem merge` runs sync which updates labels to 'merge_ready' after the PR is already merged. This is backwards - the label should be set when the spec is completed and PR created.
