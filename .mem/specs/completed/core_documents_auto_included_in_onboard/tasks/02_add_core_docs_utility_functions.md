---
title: Add core docs utility functions
status: completed
created_at: '2026-01-12T10:03:54.977313'
updated_at: '2026-01-12T10:09:15.701262'
completed_at: '2026-01-12T10:09:15.701254'
---
Update src/utils/docs.py to add: _get_core_docs_dir(), list_core_doc_files(), read_core_doc(slug). Also update list_doc_files() to exclude the core/ subdirectory so core docs aren't mixed with indexed docs.

## Completion Notes

Added list_core_doc_files(), get_core_doc_slug(), get_core_doc_path(), and read_core_doc() functions