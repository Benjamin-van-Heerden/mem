---
title: Add completion date ordering to spec list
status: completed
created_at: '2026-01-14T10:48:03.915230'
updated_at: '2026-01-14T11:05:12.233746'
completed_at: '2026-01-14T11:05:12.233737'
---
Modify the list command in src/commands/spec.py. When listing completed specs (-s completed), sort by completed_at datetime. Display the completion date in human-readable format next to each spec.

## Completion Notes

Modified spec list command to sort completed specs by completed_at date (oldest first, most recent at bottom) and display dates in human-readable format (e.g., Jan 13, 2026). Added a separate column layout for completed specs showing Completed date, Title, and Slug.