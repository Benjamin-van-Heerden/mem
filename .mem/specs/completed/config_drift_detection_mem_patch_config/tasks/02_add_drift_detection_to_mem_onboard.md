---
title: Add drift detection to mem onboard
status: completed
created_at: '2026-01-15T12:54:00.206337'
updated_at: '2026-01-15T15:17:39.181170'
completed_at: '2026-01-15T15:17:39.181162'
---
In mem onboard, detect when .mem/config.toml drifts from the canonical supported config schema. Do not modify files; print a concise warning and instruct to run 'mem patch config'. Keep onboard output short and compatible with temp-file output behavior.

## Completion Notes

Drift detection implemented and verified. Fixed missing import by using generate_default_config_toml directly instead of create_config_with_discovery.