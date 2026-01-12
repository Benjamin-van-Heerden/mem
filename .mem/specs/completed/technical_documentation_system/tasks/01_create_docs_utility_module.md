---
title: Create docs utility module
status: completed
created_at: '2026-01-11T12:41:27.155967'
updated_at: '2026-01-11T13:33:30.015709'
completed_at: '2026-01-11T13:33:30.015700'
---
Create src/utils/docs.py with core functionality:
- get_docs_dir(), get_summaries_dir(), get_data_dir() helpers
- compute_doc_hash(path) - SHA256 hash of file content
- load_hashes() / save_hashes() - read/write .doc_hashes.json
- get_chroma_client() - PersistentClient at .mem/docs/data/chroma/
- get_docs_collection(client) - get_or_create_collection with VoyageAI embedding function
- list_doc_files() - list .md files in docs dir (excluding summaries/ and data/)
- get_doc_slug(path) - extract slug from filename
- chunk_document(path) - use agno MarkdownChunking to chunk document
- index_document(collection, doc_path, chunks) - upsert chunks to ChromaDB
- remove_document_from_index(collection, doc_slug) - delete all chunks for a doc
- search_docs(collection, query, n_results, doc_filter) - query collection

## Completion Notes

Created src/utils/docs.py with functions for document hashing, chunking, ChromaDB indexing, search, and hash-based change detection