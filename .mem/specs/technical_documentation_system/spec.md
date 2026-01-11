---
title: Technical Documentation System
status: merge_ready
assigned_to: Benjamin-van-Heerden
issue_id: 37
issue_url: https://github.com/Benjamin-van-Heerden/mem/issues/37
branch: dev-benjamin_van_heerden-technical_documentation_system
pr_url: https://github.com/Benjamin-van-Heerden/mem/pull/38
created_at: '2026-01-11T12:40:06.142758'
updated_at: '2026-01-11T14:53:48.803234'
completed_at: null
last_synced_at: '2026-01-11T12:43:30.128492'
local_content_hash: 3f0172a980a4b4da5642fc24cd17fd0b5e09736c1694b904024343c9ab140e4a
remote_content_hash: 3f0172a980a4b4da5642fc24cd17fd0b5e09736c1694b904024343c9ab140e4a
---
## Overview

Add a documentation subsystem to mem that allows users to store, index, and semantically search technical documentation. Documents are markdown files stored in `.mem/docs/`, indexed into a ChromaDB vector store using VoyageAI embeddings, and AI-summarized for quick context during onboarding.

## Goals

- Allow users to add technical documentation (API docs, framework references, etc.) to their project
- Provide semantic search over documentation via vector embeddings
- Generate AI summaries for quick onboarding context
- Handle worktree symlinks so docs are accessible during spec work
- Make the system git-friendly (markdown tracked, vector DB gitignored)

## Commands

```
mem docs index       # Index unindexed docs, generate summaries, clean up orphaned entries
mem docs list        # List all indexed documents with their summaries
mem docs read <slug> # Read the full document content
mem docs search <query> [--doc <slug>]  # Semantic search across docs
mem docs delete <slug>  # Delete a document and its index/summary
```

## Directory Structure

```
.mem/docs/
├── my_api_guide.md              # Raw markdown files (git tracked)
├── framework_reference.md
├── summaries/                   # AI-generated summaries (git tracked)
│   ├── my_api_guide_summary.md
│   └── framework_reference_summary.md
└── data/                        # Gitignored - local only
    ├── .doc_hashes.json         # Maps doc_slug -> content_hash
    └── chroma/                  # ChromaDB persistent storage
```

## Technical Approach

### 1. Document Indexing (`mem docs index`)

**Hash-based change detection:**
- Compute SHA256 hash of each `.md` file in `.mem/docs/` (excluding `summaries/` and `data/`)
- Compare against stored hashes in `.mem/docs/data/.doc_hashes.json`
- Index only new or changed documents
- Remove index entries for deleted documents (hash exists but file doesn't)

**Chunking:**
- Use Agno's `MarkdownChunking` strategy from `agno.knowledge.chunking.markdown`
- Configure with `split_on_headings=2` to split on H1 and H2 headers
- Each chunk gets metadata: `doc_slug`, `chunk_index`, `heading` (if available)

**Vector storage:**
- Use ChromaDB `PersistentClient` with path `.mem/docs/data/chroma/`
- Collection name: `{project_name}_docs` (from config)
- Use VoyageAI embedding function:
  ```python
  from chromadb.utils.embedding_functions import VoyageAIEmbeddingFunction
  embedding_fn = VoyageAIEmbeddingFunction(
      api_key=os.getenv("VOYAGE_AI_API_KEY"),
      model_name="voyage-3-large"
  )
  ```
- Chunk IDs: `{doc_slug}_{chunk_index}` (e.g., `my_api_guide_0`, `my_api_guide_1`)
- Use `collection.upsert()` to handle updates

**Summary generation:**
- For new/changed docs, generate summary using Agno agent with OpenRouter (`google/gemini-3-flash-preview`)
- Store in `.mem/docs/summaries/{doc_slug}_summary.md`
- Summary format: Short description of what the doc covers + core concepts/getting started if applicable
- Keep summaries concise (target ~200-400 words) since they appear in onboard output
- If summary already exists and doc hash unchanged, skip regeneration

### 2. Document Search (`mem docs search <query>`)

- Query the ChromaDB collection with the search query
- Use `n_results` calculated to return approximately 2000 words worth of content
- Optional `--doc <slug>` filter uses ChromaDB's `where` clause: `{"doc_slug": slug}`
- Return format: chunk content with source file and approximate line range

### 3. Document Reading (`mem docs read <slug>`)

- Read and display full content of `.mem/docs/{slug}.md`
- If document doesn't exist, show error

### 4. Document Listing (`mem docs list`)

- List all `.md` files in `.mem/docs/` (excluding summaries/ and data/)
- Show indexed status (check if hash exists in `.doc_hashes.json`)
- Show summary preview if available

### 5. Document Deletion (`mem docs delete <slug>`)

- Delete `.mem/docs/{slug}.md`
- Delete `.mem/docs/summaries/{slug}_summary.md` if exists
- Remove from ChromaDB collection (delete all chunks with `doc_slug` metadata)
- Remove from `.doc_hashes.json`

### 6. Onboard Integration

- Add "TECHNICAL DOCUMENTATION" section to `mem onboard` output
- List each indexed document with its full summary
- Show warning if unindexed documents exist: "⚠️ Unindexed docs found: {names}. Run `mem docs index` to index."

### 7. Init Integration

- Check for `VOYAGE_AI_API_KEY` and `OPENROUTER_API_KEY` environment variables
- Show warning if missing: "⚠️ Document functionality requires VOYAGE_AI_API_KEY and OPENROUTER_API_KEY"
- Create `.mem/docs/` directory structure
- Add `.mem/docs/data/` to `.gitignore`

### 8. Worktree Symlink Support

**Config format in `.mem/config.toml`:**
```toml
[worktree]
symlink_paths = [
    ".mem/docs/data",
]
```

**Implementation in `spec assign`:**
- After creating worktree, read `symlink_paths` from config
- For each path, if it exists in main repo, create symlink in worktree
- Example: Main repo `.mem/docs/data/` → Worktree `.mem/docs/data/` (symlink)
- This ensures vector DB is shared across worktrees without duplication

### 9. Gitignore Updates

Add to project `.gitignore` during `mem init`:
```
# mem docs data (vector DB and hashes)
.mem/docs/data/
```

## Dependencies

Already installed in pyproject.toml:
- `chromadb` - Vector database
- `voyageai` - Embeddings (via chromadb embedding function)
- `agno` - Markdown chunking and AI agents

Environment variables required:
- `VOYAGE_AI_API_KEY` - For VoyageAI embeddings
- `OPENROUTER_API_KEY` - For AI summary generation

## File Locations

New files to create:
- `src/commands/docs.py` - Typer app with all docs commands
- `src/utils/docs.py` - Core docs utilities (indexing, hashing, ChromaDB operations)
- `src/utils/ai/doc_summarizer.py` - AI agent for generating summaries

Files to modify:
- `main.py` - Register docs command app
- `src/commands/onboard.py` - Add documentation section
- `src/commands/init.py` - Add env var checks and gitignore updates
- `src/commands/spec.py` - Add worktree symlink logic in `assign` command
- `src/templates/config.toml` - Add `[worktree]` section template

## Success Criteria

- [ ] `mem docs index` indexes new docs, updates changed docs, removes deleted docs
- [ ] `mem docs search "query"` returns relevant chunks with source info
- [ ] `mem docs read <slug>` displays full document content
- [ ] `mem docs list` shows all docs with indexed status
- [ ] `mem docs delete <slug>` removes doc, summary, and index entries
- [ ] `mem onboard` shows documentation summaries and warns about unindexed docs
- [ ] `mem init` checks for required env vars and sets up gitignore
- [ ] Worktree symlinks work correctly via config

## Notes

- The `agno.knowledge.document.Document` dataclass is used for chunking input
- Markdown files must be well-structured for optimal chunking (user responsibility)
- Summaries are git-tracked so they're available immediately after clone
- Vector DB is local-only, regenerated on each machine via `mem docs index`
- Collection uses `get_or_create_collection` to handle existing collections gracefully
