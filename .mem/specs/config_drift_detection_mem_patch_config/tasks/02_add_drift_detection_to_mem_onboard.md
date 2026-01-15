---
title: Add drift detection to mem onboard
status: todo
created_at: '2026-01-15T12:54:00.206337'
updated_at: '2026-01-15T12:54:00.206337'
completed_at: null
---
In mem onboard, detect when .mem/config.toml drifts from the canonical supported config schema. Do not modify files; print a concise warning and instruct to run 'mem patch config'. Keep onboard output short and compatible with temp-file output behavior.