# Document Parsing and Knowledge Base Integration

`%% Status: Complete %%`

## Description

Implement a document parsing pipeline that converts PDF and Word documents to markdown, stores the parsed content, and makes it searchable via vector embeddings. This enables the Coach Team agents (Writer, LevelUp Content, Support Articles) to retrieve relevant document context when responding to user queries.

**High-level flow:**
1. User submits a database task with a document URL and collection name
2. Task runner downloads the document, converts each page to an image, and uses a VLM (Gemini 3 Flash) to extract markdown
3. Parsed markdown is stored in the database with a content hash for deduplication
4. Content is chunked using markdown-aware splitting and embedded using VoyageAI
5. Embeddings are stored in PostgreSQL via pgvector
6. Agents can search the knowledge base via Agno's `Knowledge` integration in `BaseAgent`

**Key design decisions:**
- All document types (PDF, DOCX) are converted to images for consistency and to preserve visual elements (figures, charts, tables)
- pgvector for vector storage (already have PostgreSQL)
- VoyageAI `voyage-3.5` for embeddings
- Agno's `Knowledge` object integrated into `BaseAgent` for clean agent-knowledge coupling
- Document name extracted from URL or content metadata (no separate field needed in payload)

## Tasks

### Task 1: Database Schema for Parsed Documents

- [x] Create migration for `parsed_documents` table
- [x] Include content hash index for deduplication checks

#### Implementation Details

Create a new table to store parsed documents:

```sql
CREATE TABLE parsed_documents (
    id BIGSERIAL PRIMARY KEY,
    file_name VARCHAR(512) NOT NULL,
    source_url TEXT NOT NULL,
    content_hash VARCHAR(64) NOT NULL,
    parsed_markdown TEXT NOT NULL,
    collection_name VARCHAR(128) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_parsed_documents_file_name ON parsed_documents(file_name);
CREATE INDEX idx_parsed_documents_content_hash ON parsed_documents(content_hash);
CREATE INDEX idx_parsed_documents_collection ON parsed_documents(collection_name);
```

The `file_name` is the primary lookup key (same file can be added via different URLs). The `content_hash` is SHA-256 of the raw document bytes for deduplication.

#### Testing Outline

- Migration applies and rolls back cleanly
- Indexes are created correctly

> Relevant existing files: [scripts/migrate.py, migrations/]
> New files: [migrations/20260105102314_create_parsed_documents_table_UP.sql, migrations/20260105102314_create_parsed_documents_table_DOWN.sql]
> Tests: []

## Completion Report and Documentation
Migration created and applied successfully. Table `parsed_documents` with columns: id, file_name, source_url, content_hash, parsed_markdown, collection_name, created_at, updated_at. Indexes on file_name, content_hash, and collection_name.

---

### Task 2: Document Download and Parsing Utilities

- [x] Create utility to download document from URL into memory
- [x] Create utility to detect document type (PDF vs DOCX)
- [x] Create PDF-to-images converter using pdf2image
- [x] Create DOCX-to-images converter (convert to PDF first, then to images)
- [x] Create VLM-based page-to-markdown parser using Gemini 3 Flash

#### Implementation Details

**File structure:**
```
src/document_processing/
    download.py      # URL download utilities
    convert.py       # Document to images conversion
    parse.py         # VLM-based markdown extraction
```

**Download utility (`download.py`):**
- Use `httpx` to fetch document bytes from URL
- Extract filename from URL path or Content-Disposition header
- Compute SHA-256 hash of document bytes
- Return tuple of `(bytes, filename, content_hash)`

**Conversion utilities (`convert.py`):**
- `pdf_to_images(pdf_bytes: bytes, dpi: int = 80) -> list[bytes]` - uses pdf2image
- `docx_to_images(docx_bytes: bytes, dpi: int = 80) -> list[bytes]` - convert DOCX to PDF first (using python-docx + reportlab or libreoffice headless), then to images
- Both return list of PNG image bytes

**Parsing (`parse.py`):**
- Create the parser agent with Gemini 3 Flash via OpenRouter
- `parse_document_images(images: list[bytes]) -> str` - iterate through images, build markdown progressively with page context
- Use the instruction prompt provided (preserving headings, tables, figures, lists)

**Dependencies to add:**
- `pdf2image` (requires poppler)
- `python-docx`
- Potentially `subprocess` call to `libreoffice --headless` for DOCX→PDF, or `docx2pdf` library

#### Testing Outline

- Download utility handles various URL formats
- PDF conversion produces correct number of page images
- DOCX conversion produces correct number of page images
- Parser produces valid markdown from test document

> Relevant existing files: [env_settings.py]
> New files: [src/document_processing/download.py, src/document_processing/convert.py, src/document_processing/parse.py]
> Tests: []

## Completion Report and Documentation
Created three modules in `src/document_processing/`:
- `download.py`: `download_document()` fetches URL, extracts filename, computes SHA-256 hash. `detect_document_type()` uses extension and magic bytes.
- `convert.py`: `pdf_to_images()` uses pdf2image, `docx_to_images()` uses docx2pdf then pdf2image. Uses `TemporaryDirectory` for automatic cleanup.
- `parse.py`: `parse_document_images()` uses Agno Agent with Gemini 3 Flash, passes previous page context for continuity, outputs structured markdown.

Added `pdf2image` dependency to pyproject.toml. Uses existing `docx2pdf` dependency.

---

### Task 3: Document Parse Database Task

- [x] Create `DocumentParsePayload` model
- [x] Create `DocumentParseTask` extending `BaseDatabaseTask`
- [x] Implement deduplication via content hash check
- [x] Store parsed markdown in `parsed_documents` table
- [x] Register task in task runner config

#### Implementation Details

**Payload model:**
```python
class DocumentParsePayload(BaseModel):
    document_url: str
    collection_name: str  # e.g., "levelup_content", "support_articles", "writer_examples"
```

**Task implementation:**
```python
class DocumentParseTask(BaseDatabaseTask[DocumentParsePayload]):
    @property
    def batch_size(self) -> int:
        return 5  # Conservative due to LLM calls
    
    @property
    def task_timeout_seconds(self) -> int:
        return 600  # 10 minutes for large documents
    
    async def process_task(self, payload: DocumentParsePayload) -> dict:
        # 1. Download document
        doc_bytes, filename, content_hash = await download_document(payload.document_url)
        
        # 2. Check if already processed (by content_hash)
        if await self._document_exists(content_hash):
            return {"status": "skipped", "reason": "duplicate", "content_hash": content_hash}
        
        # 3. Convert to images based on file type
        images = await convert_to_images(doc_bytes, filename)
        
        # 4. Parse to markdown
        markdown = await parse_document_images(images)
        
        # 5. Store in database
        doc_id = await self._store_document(filename, payload.document_url, content_hash, markdown, payload.collection_name)
        
        # 6. Return result
        return {"status": "parsed", "document_id": doc_id, "content_hash": content_hash}
```

#### Testing Outline

- Task processes PDF document end-to-end
- Task processes DOCX document end-to-end
- Duplicate documents are skipped based on content hash
- Task handles download failures gracefully
- Task handles malformed documents gracefully

> Relevant existing files: [src/task_runner/core/base_database_task.py, src/task_runner/config.py]
> New files: [src/task_runner/tasks/database_tasks/document_parse_task.py]
> Tests: []

## Completion Report and Documentation
Created `DocumentParseTask` in `src/task_runner/tasks/database_tasks/document_parse_task.py`:
- Payload: `document_url`, `collection_name`
- Deduplication via `_get_existing_document()` checking content_hash
- Storage via `_store_document()` with RETURNING id
- Registered in `TaskConfig.DATABASE_TASKS`
- batch_size=3, timeout=600s (10 min for large docs)

---

### Task 4: Vector Embedding Pipeline

- [x] Set up pgvector extension in PostgreSQL
- [x] Create migration for vector embedding table (or rely on Agno's auto-creation)
- [x] Create chunking and embedding utility using Agno's Knowledge + MarkdownChunking + VoyageAIEmbedder
- [x] Integrate embedding step into DocumentParseTask (after storing markdown)

#### Implementation Details

**pgvector setup:**
- Add `CREATE EXTENSION IF NOT EXISTS vector;` to migration or setup script
- Agno's `PgVector` handles table creation automatically

**Embedding integration:**
```python
from agno.knowledge.knowledge import Knowledge
from agno.knowledge.chunking.markdown import MarkdownChunking
from agno.knowledge.document import Document
from agno.vectordb.pgvector import PgVector
from agno.embedder.voyageai import VoyageAIEmbedder

def get_knowledge_base(collection_name: str) -> Knowledge:
    return Knowledge(
        vector_db=PgVector(
            db_url=postgres_db_url(),
            table_name=f"kb_{collection_name}",  # e.g., kb_levelup_content
            embedder=VoyageAIEmbedder(model="voyage-3.5"),
        ),
        chunking_strategy=MarkdownChunking(chunk_size=5000, overlap=0),
    )

async def embed_document(markdown: str, collection_name: str, doc_id: int, filename: str) -> None:
    knowledge = get_knowledge_base(collection_name)
    document = Document(
        content=markdown,
        name=filename,
        meta_data={"document_id": doc_id, "collection": collection_name},
    )
    await knowledge.async_add_documents([document])
```

**Update DocumentParseTask:**
- After storing markdown, call `embed_document()`
- Return embedding stats in result

#### Testing Outline

- pgvector extension installs correctly
- Document is chunked and embedded
- Embeddings are searchable via similarity query

> Relevant existing files: [src/utils/agno/db.py, env_settings.py]
> New files: [src/knowledge/embedding.py]
> Tests: []

## Completion Report and Documentation
Created `src/knowledge/embedding.py` with:
- `get_knowledge_base(collection_name)` returns Knowledge with PgVector + VoyageAI embedder
- `embed_document()` uses `add_content_async()` with MarkdownReader for chunking
- Integrated into DocumentParseTask after markdown storage

Migration `20260105104508_enable_pgvector_extension` enables vector extension.
Docker image changed to `pgvector/pgvector:pg18-trixie`.

---

### Task 5: BaseAgent Knowledge Integration

- [x] Add optional `knowledge` parameter to `BaseAgent`
- [x] Add `search_knowledge` flag similar to Agno's pattern
- [x] Update agent instantiation to optionally include knowledge base
- [x] Ensure knowledge search works in streaming context

#### Implementation Details

**BaseAgent modifications:**
```python
class BaseAgent(ABC, Generic[ReqT, ConfigT]):
    # ... existing code ...
    
    @classmethod
    def knowledge(cls, config: ConfigT) -> Knowledge | None:
        """Override to provide a knowledge base for this agent. Default: None."""
        return None
    
    @classmethod
    def search_knowledge(cls) -> bool:
        """Whether the agent should automatically search knowledge. Default: False."""
        return False
```

**In `get_agent()` and `get_agent_for_team()`:**
```python
agent = Agent(
    # ... existing params ...
    knowledge=cls.knowledge(config) if cls.knowledge(config) else None,
    search_knowledge=cls.search_knowledge(),
)
```

**Example usage in a future agent:**
```python
class LevelUpContentAgent(BaseAgent[...]):
    @classmethod
    def knowledge(cls, config: ConfigT) -> Knowledge | None:
        return get_knowledge_base("levelup_content")
    
    @classmethod
    def search_knowledge(cls) -> bool:
        return True
```

#### Testing Outline

- BaseAgent works without knowledge (backward compatible)
- Agent with knowledge can search and retrieve relevant chunks
- Search results are incorporated into agent context

> Relevant existing files: [src/agents/core.py]
> New files: []
> Tests: []

## Completion Report and Documentation
Added to `BaseAgent`:
- `knowledge(cls, config)` method returning `Knowledge | None` (default None)
- `search_knowledge(cls)` method returning `bool` (default False)
- Both `get_agent()` and `get_agent_for_team()` now pass these to Agent constructor

Future agents can override these methods to enable knowledge-based RAG.

---

### Task 6: Environment and Dependencies

- [x] Add new dependencies to pyproject.toml
- [x] Add VoyageAI API key to env_settings
- [x] Add OpenRouter API key if not already present
- [x] Document poppler installation requirement for pdf2image
- [x] Document libreoffice requirement for DOCX conversion (if used)

#### Implementation Details

**New dependencies:**
```toml
dependencies = [
    # ... existing ...
    "pdf2image>=1.16.0",
    "python-docx>=1.1.0",
    "voyageai>=0.3.0",  # or whatever agno uses
]
```

**New env settings:**
```python
voyage_api_key: str = Field(default="")
# openrouter_api_key should already exist
```

**System dependencies (document in README):**
- `poppler-utils` (for pdf2image)
- `libreoffice` (for DOCX→PDF conversion, headless mode)

#### Testing Outline

- All dependencies install correctly
- Environment variables are loaded

> Relevant existing files: [pyproject.toml, env_settings.py, README.md]
> New files: []
> Tests: []

## Completion Report and Documentation
Dependencies added to pyproject.toml: `pdf2image`, `pgvector`, `voyageai`, `unstructured`, `markdown`, `docx2pdf`
Added `voyage_api_key` to env_settings.py and .env.example
System deps: poppler (for pdf2image), docx2pdf uses MS Word/LibreOffice under the hood

---

# Final Review
All 6 tasks completed. The document parsing and knowledge base pipeline is now implemented:
1. Documents are downloaded, converted to images, parsed to markdown via VLM
2. Markdown is stored in `parsed_documents` table with deduplication
3. Content is chunked and embedded via VoyageAI into pgvector
4. BaseAgent supports optional knowledge integration for RAG-enabled agents
