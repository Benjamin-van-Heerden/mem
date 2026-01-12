---
title: Add usage hints and confirmations
status: completed
created_at: '2026-01-12T11:14:03.514645'
updated_at: '2026-01-12T12:00:01.325160'
completed_at: '2026-01-12T12:00:01.325152'
---
Add progressive hints between merge commands:

1. After 'mem merge' completes successfully, print hint: 'Next step: mem merge into test (requires confirmation)'

2. After 'mem merge into test' completes successfully, print hint: 'Next step: mem merge into main'

3. 'mem merge into main' should run in dry-run mode by default, showing what would happen and asking for confirmation. User must run 'mem merge into main --force' to actually execute. Do NOT expose the --force flag in any previous hints - only show it in the dry-run output of 'mem merge into main' itself.

All commands must work when run from the dev branch. The 'into' commands should handle switching to the required branch (test/main) automatically.

## Amendments

Clarification on hints and confirmations:

The 'confirmation' is for AI agents, not interactive CLI prompts. Use [AGENT INSTRUCTION] sections in the output.

**After mem merge completes:**
Print hint with [AGENT INSTRUCTION] section:
- Show: 'Next step: mem merge into test'
- Agent instruction: 'Ask the user before performing this action'

**After mem merge into test completes:**
Print hint with [AGENT INSTRUCTION] section:
- Show: 'Next step: mem merge into main'
- Agent instruction: 'Ask the user before performing this action'

**mem merge into main (dry-run by default):**
- Show what would happen
- Print instruction: 'Run mem merge into main --force to execute'
- This is the ONLY place --force is mentioned

The [AGENT INSTRUCTION] pattern is used elsewhere in mem (e.g., onboard command) to guide AI agents to pause and confirm with the user before taking action.

## Completion Notes

Added progressive hints after merge commands and fixed test_merge.py to use CliRunner