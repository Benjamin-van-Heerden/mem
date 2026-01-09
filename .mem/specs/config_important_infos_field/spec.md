---
title: Config important infos field
status: todo
assigned_to: Benjamin-van-Heerden
issue_id: 28
issue_url: https://github.com/Benjamin-van-Heerden/mem/issues/28
branch: dev-benjamin_van_heerden-config_important_infos_field
pr_url: null
created_at: '2026-01-09T16:19:36.032353'
updated_at: '2026-01-09T16:20:33.139435'
completed_at: null
last_synced_at: '2026-01-09T16:20:24.396450'
local_content_hash: 2dc7a548c621f03b36c6a8250d3a9c7a0114c8972a2a1012dd5f6489fe19b663
remote_content_hash: 2dc7a548c621f03b36c6a8250d3a9c7a0114c8972a2a1012dd5f6489fe19b663
---
## Overview

Add an optional "Important Infos" field to the project config (`config.toml`) that displays custom information at the bottom of the `mem onboard` output.

## Goals

- Add optional `important_infos` field to config schema
- Display the content at the bottom of onboard output when present

## Technical Approach

1. Update config model in `src/models.py` to include optional `important_infos: str | None` field in the project section
2. Update `src/templates/config.toml` to document the new field
3. Update `src/commands/onboard.py` to render the important infos section at the bottom of output when the field is set

## Success Criteria

- Config with `important_infos` field parses correctly
- Onboard output shows the info at the bottom when field is set
- Onboard output unchanged when field is not set (backwards compatible)
