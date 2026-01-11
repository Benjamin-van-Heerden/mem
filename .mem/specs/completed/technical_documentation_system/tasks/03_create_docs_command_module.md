---
title: Create docs command module
status: completed
created_at: '2026-01-11T12:41:48.579939'
updated_at: '2026-01-11T13:42:20.006971'
completed_at: '2026-01-11T13:42:20.006963'
---
Create src/commands/docs.py with Typer app:

Commands:
1. index() - Main indexing command:
   - Check for VOYAGE_AI_API_KEY, error if missing
   - List all doc files, compute hashes
   - Load existing hashes, compare
   - For new/changed docs: chunk, index to ChromaDB, generate summary
   - For deleted docs: remove from index, delete summary
   - Save updated hashes
   - Print summary of actions taken

2. list_docs_cmd() - List documents:
   - Show all docs with indexed status
   - Show summary preview (first 100 chars) if available

3. read(doc_slug) - Read full document:
   - Display content of .mem/docs/{slug}.md
   - Error if not found

4. search(query, --doc) - Semantic search:
   - Query ChromaDB collection
   - Calculate n_results for ~2000 words
   - Apply doc filter if provided
   - Display chunks with source file info

5. delete(doc_slug) - Delete document:
   - Confirm with user
   - Delete md file, summary, index entries
   - Update hashes

## Completion Notes

Created src/commands/docs.py with index, list, read, search, and delete commands using Typer