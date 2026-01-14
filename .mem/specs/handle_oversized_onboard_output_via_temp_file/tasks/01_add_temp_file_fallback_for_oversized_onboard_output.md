---
title: Add temp-file fallback for oversized onboard output
status: todo
created_at: '2026-01-14T14:11:51.740119'
updated_at: '2026-01-14T14:11:51.740119'
completed_at: null
---
Update 'mem onboard' so that if the final output exceeds 30KB UTF-8 bytes, it writes the full content to .mem/tmp/onboard_<timestamp>.md and prints a short instruction containing the path and a command to read it. Keep current stdout behavior under the threshold.