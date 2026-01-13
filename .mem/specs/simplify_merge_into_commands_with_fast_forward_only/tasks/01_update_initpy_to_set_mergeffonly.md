---
title: Update init.py to set merge.ff=only
status: completed
created_at: '2026-01-13T13:52:08.194964'
updated_at: '2026-01-13T14:03:44.654305'
completed_at: '2026-01-13T14:03:44.654298'
---
Change configure_merge_settings() to set merge.ff=only instead of merge.ff=false. Update the success message.

## Completion Notes

Changed merge.ff from false to only in configure_merge_settings function. Updated docstring and echo message to reflect the new behavior.