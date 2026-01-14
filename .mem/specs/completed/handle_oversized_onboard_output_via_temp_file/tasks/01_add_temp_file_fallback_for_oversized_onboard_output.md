---
title: Add temp-file fallback for oversized onboard output
status: completed
created_at: '2026-01-14T14:11:51.740119'
updated_at: '2026-01-14T15:06:21.957267'
completed_at: '2026-01-14T15:06:21.957260'
---
Update 'mem onboard' so that if the final output exceeds 30KB UTF-8 bytes, it writes the full content to .mem/tmp/onboard_<timestamp>.md and prints a short instruction containing the path and a command to read it. Keep current stdout behavior under the threshold.

## Completion Notes

Onboard now writes large context to /tmp/mem/mem_onboard_{datetime}.md, prunes mem_onboard_* older than 1 hour, and prints a concise pointer with a MUST-read instruction.