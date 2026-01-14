---
title: Support --stdout flag for mem onboard
status: completed
created_at: '2026-01-14T14:17:27.558609'
updated_at: '2026-01-14T15:09:27.121500'
completed_at: '2026-01-14T15:09:27.121492'
---
Add a CLI flag (e.g. --stdout) that forces mem onboard to print full project context to stdout instead of writing it to a temp file. Default behavior remains temp-file output.

## Completion Notes

Added --stdout option to mem onboard to force printing full onboard context to stdout; default behavior unchanged (temp file fallback remains when output is large).