---
title: Update config model
status: completed
subtasks: []
created_at: '2026-01-09T16:21:34.362294'
updated_at: '2026-01-09T16:22:27.251387'
completed_at: '2026-01-09T16:22:27.251382'
---
Add optional important_infos: str | None field to the project section in src/models.py

## Completion Notes

No Pydantic config model exists - config is parsed directly from TOML with tomllib. The field is documented in the config template instead.