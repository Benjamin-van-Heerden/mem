# Work Log - Document Parsing Spec and Test Fixes

## Spec File: `agent_rules/spec/s_20260104_benjamin_van_heerden__document_parsing_and_knowledge_base.md`

## Overarching Goals

1. Fix broken tests caused by tool function renames from the previous session
2. Create a comprehensive spec for the document parsing and knowledge base feature (Phase 3 of the agent platform expansion)

## What Was Accomplished

### Test Fixes for Tool Name Changes

Updated `test_meeting_agent.py` and `test_offenses_agent.py` to avoid magic strings when checking tool calls. Instead of hardcoding tool names, the tests now import the factory functions and use `.__name__` to get the actual function name:

**test_meeting_agent.py:**
```python
from src.agents.meetings.tools import make_get_meetings

# In test:
searching_for_meetings = make_get_meetings("", "", [])
assert searching_for_meetings.__name__ in tool_calls
```

**test_offenses_agent.py:**
```python
from src.agents.offenses.tools import make_get_offense_setups

# In test:
retrieving_offense_rules = make_get_offense_setups("", "")
assert retrieving_offense_rules.__name__ in tool_calls
```

This pattern ensures tests stay in sync with any future tool name changes.

### Document Parsing and Knowledge Base Spec

Created a detailed spec covering the full document parsing pipeline:

1. **Database Schema** - `parsed_documents` table with file_name index and content_hash for deduplication
2. **Download & Parsing Utilities** - Modular utilities for downloading documents, converting to images (PDF/DOCX), and VLM-based markdown extraction using Gemini 3 Flash
3. **Document Parse Task** - Database task extending `BaseDatabaseTask` to orchestrate the pipeline
4. **Vector Embedding Pipeline** - pgvector + VoyageAI `voyage-3.5` embeddings + Agno's MarkdownChunking
5. **BaseAgent Knowledge Integration** - Optional `knowledge()` and `search_knowledge()` methods in BaseAgent
6. **Environment & Dependencies** - pdf2image, python-docx, voyageai, system deps (poppler, libreoffice)

Key design decisions documented:
- All document types converted to images for consistency (handles figures/charts)
- pgvector chosen over Qdrant (already have PostgreSQL)
- Document name extracted from URL/metadata (no separate payload field)
- Knowledge integrated at BaseAgent level for clean agent-knowledge coupling

## Key Files Affected

| File | Change |
|------|--------|
| `tests/agents/test_meeting_agent.py` | Import factory, use `.__name__` instead of magic string |
| `tests/agents/test_offenses_agent.py` | Import factory, use `.__name__` instead of magic string |
| `agent_rules/spec/s_20260104_benjamin_van_heerden__document_parsing_and_knowledge_base.md` | NEW - Full spec for document parsing feature |

## What Comes Next

Begin implementation of the document parsing spec, starting with:

1. **Task 1**: Create database migration for `parsed_documents` table
2. **Task 2**: Implement document download and parsing utilities
3. **Task 3**: Create `DocumentParseTask` database task
4. **Task 4**: Set up pgvector and embedding pipeline
5. **Task 5**: Integrate knowledge base into BaseAgent
6. **Task 6**: Add dependencies and environment configuration

The spec file contains detailed implementation guidance for each task. Work should proceed sequentially as later tasks depend on earlier ones.
