---
title: Core documents auto-included in onboard
status: completed
assigned_to: Benjamin-van-Heerden
issue_id: 40
issue_url: https://github.com/Benjamin-van-Heerden/mem/issues/40
branch: dev-benjamin_van_heerden-core_documents_auto_included_in_onboard
pr_url: https://github.com/Benjamin-van-Heerden/mem/pull/42
created_at: '2026-01-12T10:03:05.317359'
updated_at: '2026-01-12T10:19:33.210488'
completed_at: '2026-01-12T10:19:33.209014'
last_synced_at: '2026-01-12T10:05:24.517668'
local_content_hash: b63edb0de3ca619a2846a2dedcad802ddf2916ea9c747a65e6bd3004674939d6
remote_content_hash: b63edb0de3ca619a2846a2dedcad802ddf2916ea9c747a65e6bd3004674939d6
---
## Overview

Add a new document class called "core" documents. Documents placed in `.mem/docs/core/` are automatically included in full in the `mem onboard` output, without requiring indexing via `mem docs index`. This is for essential reference material that should always be visible to AI agents.

The existing indexed docs (in `.mem/docs/`) continue to work as before - they require `mem docs index` and only their summaries appear in onboard output.

## Goals

- Create a `.mem/docs/core/` directory for core documents
- Automatically read and include all core documents in `mem onboard` output (full content, not summaries)
- No indexing required for core docs - they're read directly from disk
- Keep the existing indexed docs behavior unchanged

## Technical Approach

### Directory Structure

```
.mem/docs/
├── core/              # NEW: Core docs (auto-included in full)
│   └── *.md           # Any markdown files here appear in onboard
├── summaries/         # Existing: AI-generated summaries
├── data/              # Existing: ChromaDB, hashes (gitignored)
└── *.md               # Existing: Indexed docs (summaries in onboard)
```

### Implementation

1. **Update `src/commands/init.py`**: Create `.mem/docs/core/` directory during `mem init`

2. **Update `src/utils/docs.py`**: Add functions to work with core docs:
   - `_get_core_docs_dir() -> Path`: Return `.mem/docs/core/` path
   - `list_core_doc_files() -> list[Path]`: List all markdown files in core dir
   - `read_core_doc(slug: str) -> str | None`: Read a core document by slug
   - Update `list_doc_files()` to exclude the `core/` subdirectory (so core docs aren't mixed with indexed docs)

3. **Update `src/commands/onboard.py`**: Add "CORE DOCUMENTATION" section before "TECHNICAL DOCUMENTATION":
   - Read all files from `.mem/docs/core/`
   - Display each document's full content with its filename as heading
   - No summary, no indexing status - just raw content

4. **Update `src/commands/docs.py`**: 
   - Update `mem docs list` to show core docs separately (marked as "core" rather than indexed/unindexed)
   - Core docs should NOT appear in `mem docs index` workflow (they don't need indexing)

## Success Criteria

- `.mem/docs/core/` directory is created by `mem init`
- Markdown files in `.mem/docs/core/` appear in full in `mem onboard` output
- Core docs are separate from indexed docs in `mem docs list`
- Core docs don't require or get affected by `mem docs index`
- Existing indexed docs behavior unchanged

## Notes

Use case: Reference material that's always needed (API docs, architecture decisions, coding standards) that the agent should see in full rather than searching for.
