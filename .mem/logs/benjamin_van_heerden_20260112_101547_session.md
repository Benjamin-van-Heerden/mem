---
created_at: '2026-01-12T10:15:47.846212'
username: benjamin_van_heerden
spec_slug: core_documents_auto_included_in_onboard
---
# Work Log - Implement core documents feature

## Overarching Goals

Add a new "core documents" feature where markdown files placed in `.mem/docs/core/` are automatically included in full in `mem onboard` output, without requiring indexing. This is for essential reference material that should always be visible to AI agents.

## What Was Accomplished

### 1. Updated init to create core docs directory
- Added `_get_core_docs_dir()` function in `src/utils/docs.py`
- Updated `ensure_docs_dirs()` to create `.mem/docs/core/` directory
- Updated init output message to mention the core directory

### 2. Added core docs utility functions
Added the following functions to `src/utils/docs.py`:
- `list_core_doc_files()` - Lists all markdown files in the core docs directory
- `get_core_doc_slug()` - Extracts slug from core document file path
- `get_core_doc_path()` - Gets path to a core document by slug
- `read_core_doc()` - Reads core document content by slug

### 3. Added core documentation section to onboard
Added a "CORE DOCUMENTATION" section to `src/commands/onboard.py` that:
- Appears before the "TECHNICAL DOCUMENTATION" section
- Reads all files from `.mem/docs/core/`
- Displays each document's full content with its filename as heading

### 4. Updated docs list command
Modified `mem docs list` in `src/commands/docs.py` to:
- Show core docs in a separate section labeled "CORE DOCUMENTS (auto-included in onboard)"
- Show indexed docs in a separate section labeled "INDEXED DOCUMENTS (summaries in onboard)"
- Provide helpful guidance when no documents exist

### 5. Verified with tests
- All 14 docs tests pass
- All 2 init tests pass
- Manual verification confirmed the feature works as expected

## Key Files Affected

- `src/utils/docs.py` - Added core docs utility functions and updated `ensure_docs_dirs()`
- `src/commands/init.py` - Updated output message to mention core/ directory
- `src/commands/onboard.py` - Added CORE DOCUMENTATION section
- `src/commands/docs.py` - Updated `list_docs` command to show core docs separately

## What Comes Next

All spec tasks have been completed. The spec is ready for completion and PR creation.
