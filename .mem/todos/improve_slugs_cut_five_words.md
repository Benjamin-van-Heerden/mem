---
title: Improve slugs cut at five words
status: claimed
created_at: "2026-09-29T14:07:01+02:00"
claimed_by: benjamin_van_heerden
claimed_at: "2026-09-29T14:54:35+02:00"
---

The five-word cap from v0.4.2 (work.recordSlug in internal/work/markdown.go) cuts titles mid-phrase: 'mem sync catches up mid-session' became mem_sync_catches_up_mid, 'Add mem hook compact with a compact digest' became add_mem_hook_compact_compact, 'Update docs and structure for the sync catch-up' became update_docs_structure_sync_catch. Consider a character budget cut at a word boundary, dropping repeated words, or keeping hyphenated words together.
