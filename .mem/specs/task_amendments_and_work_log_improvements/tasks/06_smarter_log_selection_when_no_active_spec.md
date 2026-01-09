---
title: Smarter log selection when no active spec
status: completed
subtasks: []
created_at: '2026-01-09T15:50:07.622096'
updated_at: '2026-01-09T16:22:19.009721'
completed_at: '2026-01-09T16:22:19.009713'
---
When showing last 3 logs and all 3 are from same spec, show only 2 from that spec and use 3rd slot for most recent log from any other spec

## Completion Notes

Added diversity logic: when all 3 recent logs are from the same spec, keeps 2 from that spec and replaces 3rd with most recent log from a different spec