---
title: Update onboard command
status: completed
subtasks: []
created_at: '2026-01-09T16:21:44.766470'
updated_at: '2026-01-09T16:24:55.635062'
completed_at: '2026-01-09T16:24:55.635055'
---
Update src/commands/onboard.py to render the important infos section at the bottom of output when the field is set

## Completion Notes

Added code to display IMPORTANT INFORMATION section at the bottom of onboard output when important_infos field is set in project config. Verified it displays correctly when set and doesn't appear when not set (backwards compatible).