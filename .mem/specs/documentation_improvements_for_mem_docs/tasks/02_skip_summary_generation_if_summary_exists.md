---
title: Skip summary generation if summary exists
status: todo
created_at: '2026-01-13T18:07:48.070067'
updated_at: '2026-01-13T18:07:48.070067'
completed_at: null
---
In src/commands/docs.py index() command, before generating a summary for a document, check if the summary already exists using docs.get_summary_path(slug).exists(). If it exists, print 'Summary already exists (skipped)' and skip the generation. This avoids redundant AI calls when re-indexing documents.