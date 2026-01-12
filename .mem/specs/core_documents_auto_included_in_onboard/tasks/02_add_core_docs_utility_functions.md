---
title: Add core docs utility functions
status: todo
created_at: '2026-01-12T10:03:54.977313'
updated_at: '2026-01-12T10:03:54.977313'
completed_at: null
---
Update src/utils/docs.py to add: _get_core_docs_dir(), list_core_doc_files(), read_core_doc(slug). Also update list_doc_files() to exclude the core/ subdirectory so core docs aren't mixed with indexed docs.