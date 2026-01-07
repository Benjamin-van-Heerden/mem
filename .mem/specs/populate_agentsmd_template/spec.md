---
title: Populate AGENTS.md template
status: merge_ready
assigned_to: Benjamin-van-Heerden
issue_id: 19
issue_url: https://github.com/Benjamin-van-Heerden/mem/issues/19
branch: dev-benjamin_van_heerden-populate_agentsmd_template
pr_url: null
created_at: '2026-01-07T12:53:11.198548'
updated_at: '2026-01-07T13:28:07.538680'
completed_at: null
last_synced_at: '2026-01-07T12:54:45.931675'
local_content_hash: 7511840a95781e4d4cbf3a90634c9546dd7fd394130c26ac8a45b4b4911ab00b
remote_content_hash: 7511840a95781e4d4cbf3a90634c9546dd7fd394130c26ac8a45b4b4911ab00b
---
## Overview

Populate the empty AGENTS.md template file with comprehensive onboarding content for AI agents. This template is copied to projects when they run `mem init` and serves as the primary entry point for agents to understand how to work with mem-enabled projects.

## Goals

- Provide clear instructions for AI agents on how to work with mem
- Explain the `mem onboard` workflow and when to run it
- Document core mem concepts (specs, tasks, subtasks, work logs)
- Include essential commands agents will use frequently
- Set expectations for workflow patterns (task completion, work logs, etc.)

## Technical Approach

Write the AGENTS.md template in `src/templates/AGENTS.md` with content structured for quick agent comprehension:
1. Opening statement about mem integration
2. First action: run `mem onboard`
3. Core workflow patterns
4. Key commands reference
5. Important behavioral expectations

## Success Criteria

- AGENTS.md template contains actionable guidance for AI agents
- Content is concise but comprehensive
- Workflow expectations are clear (mark tasks complete immediately, create work logs, etc.)
- Template works standalone without requiring agents to read other docs first

## Notes

This is part of a set of housekeeping improvements. The user mentioned there are other changes they want to introduce alongside this.
