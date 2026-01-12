# Work Log - Message Serialization Exploration and Demo Refinement

## Spec File: `agent_rules/spec/s_20251129_benjamin_van_heerden__agent_infrastructure_migration.md`

## Overarching Goals

Explore Agno's `Message` and `Image` classes to understand their serialization behavior and determine the correct database storage strategy. Update CLI demos to use proper Agno primitives instead of manually constructed dicts, and document findings to inform database schema design.

## What Was Accomplished

### 1. Fixed Import Paths for Agno Classes

**Corrected imports in both demo files:**
- `Image` class is in `agno.media`, not `agno.models.image`
- `Message` class correctly imported from `agno.models.message`

```python
from agno.media import Image
from agno.models.message import Message
```

### 2. Created New Demo: Agent Message Serialization

**Added Demo 7 to `scripts/play/agent_demo.py`:**
- Explores `Message.to_dict()` output structure for text-only messages
- Tests `Image.from_base64()` and image serialization in Message objects
- Validates `Message.from_dict()` reconstruction fidelity
- Demonstrates attaching metrics to Message objects
- Shows agent accepting `List[Message]` as input

**Key discovery - Message.to_dict() structure:**
```json
{
  "id": "auto-generated-uuid",
  "content": "message text",
  "from_history": false,
  "stop_after_tool_call": false,
  "role": "user",
  "images": [
    {
      "id": "image-uuid",
      "content": "base64-data-here"
    }
  ],
  "metrics": {
    "input_tokens": 82,
    "output_tokens": 22,
    "total_tokens": 104
  },
  "created_at": 1765280936
}
```

**Critical findings:**
- `created_at` is Unix timestamp (integer), not ISO string
- Images stored in separate `images` array, NOT in content
- Each image has its own UUID
- Metrics can be attached to assistant messages after `run_completed`
- Reconstruction with `Message.from_dict()` is perfect (all fields preserved)

### 3. Created New Demo: Team Message Serialization

**Added Demo 5 to `scripts/play/teams_demo.py`:**
- Validates that teams accept `List[Message]` identically to agents
- Confirms `Message.to_dict()` structure is unchanged in team context
- Tests image handling in team context (works the same)
- Shows metrics attachment for team responses

**Key finding:** Message handling is **identical** for teams and agents. No special cases needed.

### 4. Pruned Redundant Demos

**Removed from agent_demo.py (4 demos):**
- `demo_basic_streaming` - covered by `demo_full_event_streaming`
- `demo_multiple_tool_calls` - same pattern as single tool call
- `demo_event_attributes` - event structure covered elsewhere
- `demo_image_input_with_tokens` - covered by `demo_message_serialization`

**Removed from teams_demo.py (3 demos):**
- `demo_event_object_inspection` - event structure covered by demo 1
- `demo_message_reconstruction` - Message pattern covered by demo 5
- `demo_team_image_input_with_tokens` - image handling covered by demo 5

**Remaining essential demos:**
- **agent_demo.py:** 3 demos (event streaming, token tracking, Message serialization)
- **teams_demo.py:** 2 demos (team delegation, team Message serialization)

### 5. Added Comprehensive Documentation

**Created `agent_rules/docs/d_message_serialization_findings.md`:**
- Complete `Message.to_dict()` structure documentation
- Image serialization format (base64 in images array)
- Metrics attachment patterns
- Reconstruction validation results
- Database schema recommendations
- Token metrics handling (only at `run_completed`)
- Agent input format (`List[Message]`)
- Team behavior documentation

### 6. Defined Image Storage Strategy

**Decision made: Metadata-only storage with future S3 migration path**

**Current approach (MVP):**
- During conversation: Send full base64 to LLM (required for processing)
- For storage: Store only `filename` and `type` in message images array
- For UI: Display icon (📷/📄) + filename in conversation history
- Image context NOT preserved across logout/login sessions

**Storage format:**
```json
{
  "images": [
    {
      "id": "img-uuid",
      "filename": "screenshot.png",
      "type": "image/png"
      // NO base64 content - stripped before storage
    }
  ]
}
```

**Future enhancement:**
- Upload to S3 instead of inline base64
- Store `url` field pointing to S3 object
- Enables image persistence across sessions
- Supports image regeneration/display in UI

**Benefits:**
- Lean database (no huge base64 blobs)
- Fast history loading
- Clear UI representation
- No immediate need for external storage infrastructure

## Key Files Affected

**Modified:**
- `scripts/play/agent_demo.py` (462 lines → 620 lines)
  - Fixed Image import path
  - Added Demo 7: Message and Image Object Serialization
  - Pruned 4 redundant demos
  - Added header documentation explaining demo purposes

- `scripts/play/teams_demo.py` (547 lines → 750 lines)
  - Fixed Image import path
  - Added Demo 5: Team Message Object Serialization
  - Pruned 3 redundant demos
  - Added header documentation

**Created:**
- `agent_rules/docs/d_message_serialization_findings.md` (308 lines)
  - Comprehensive Message.to_dict() structure documentation
  - Image serialization patterns
  - Database schema implications
  - Storage strategy recommendations
  - Token metrics handling
  - Reconstruction patterns

## What Comes Next

### Immediate Next Steps (Database Schema Design)

1. **Finalize `chat_events` table schema**
   - Use JSONB column for `message_data` storing `Message.to_dict()` output
   - Add helper function to strip image base64 before storage
   - Fields: `id` (UUIDv7), `session_id`, `role`, `agent_name`, `message_data` (JSONB), `input_tokens`, `output_tokens`, `cost_in_rands`, `created_at`
   - Indexes: `(session_id, created_at)` for history queries

2. **Create database migration**
   - Generate migration: `python scripts/migrate.py --generate "create_agent_chat_tables"`
   - Tables needed: `sessions`, `chat_events`
   - Consider: `users`, `teams`, `agents` (or just use VARCHAR references to existing WCU entities)

3. **Implement storage layer** (`src/agents/storage.py`)
   ```python
   def prepare_message_for_storage(message: Message) -> dict:
       """Strip image base64, keep metadata only"""
       
   async def store_message(session_id: str, message: Message, ...):
       """Store message with stripped images"""
       
   async def load_session_messages(session_id: str) -> list[Message]:
       """Reconstruct messages from DB (images as metadata only)"""
   ```

4. **Prototype roundtrip test**
   - Create Message with image
   - Store to DB (with base64 stripped)
   - Retrieve from DB
   - Validate metadata preserved (filename, type)
   - Confirm UI can display icon + filename

### Tool Call Exploration (Outstanding)

5. **Add demo exploring tool call Message structures**
   - How are tool calls represented in Message.to_dict()?
   - What is the structure of `tool_call_id`, `tool_name`, `tool_args`?
   - How to reconstruct tool call sequences from stored messages?
   - This is critical for full conversation reconstruction

### Architecture Implementation (Following Spec Order)

6. **SessionManager implementation** (using diskcache)
   - `is_stopped(session_id, run_id) -> bool`
   - `request_stop(session_id, run_id)`
   - `clear_stop(session_id, run_id)`
   - Used by streaming wrapper to check for user interruptions

7. **Streaming wrapper** (`src/agents/runner.py`)
   - `run_agent_streaming(agent, message, session_id, ...)`
   - Yield chunks as they arrive
   - Store user message at start
   - Accumulate assistant message during streaming
   - Attach metrics at `run_completed`
   - Handle interruptions via SessionManager
   - Handle errors and disconnections

8. **BaseAgent simplification** (per spec)
   - Remove authentication (handled in routes)
   - Remove playground mode (not needed)
   - Simple pattern: `__init__(context: BaseModel)` + `get_agent() -> Agent`
   - Context models defined as Pydantic classes

9. **API Routes implementation**
   - POST `/agents/{agent_name}/chat` - Streaming endpoint
   - POST `/agents/{agent_name}/stop` - Interruption endpoint
   - GET `/agents/{agent_name}/sessions` - List sessions
   - GET `/agents/{agent_name}/sessions/{session_id}` - Get history

### Key Architectural Decisions Made

✅ **Use Message primitive for everything** - No manual dict construction  
✅ **Store Message.to_dict() in JSONB** - Single source of truth  
✅ **Strip image base64 before storage** - Keep metadata only (filename, type)  
✅ **Future-proof for S3** - Can add `url` field later without schema changes  
✅ **Image context not preserved on reload** - Acceptable tradeoff for MVP  
✅ **Same pattern for agents and teams** - No special cases needed  

### Reference for Next Session

**Critical principle discovered:**
> "We cannot be constructing things like cavemen. Let's adopt the List[Message] as input structure."

**Everything uses Message objects:**
- Input to agent: `List[Message]`
- Storage format: `Message.to_dict()` in JSONB
- Reconstruction: `Message.from_dict()` from DB
- No manual dict construction anywhere

**Agno imports to use:**
```python
from agno.models.message import Message
from agno.media import Image
```

The foundation is solid. We know exactly how Message serialization works, we have a pragmatic storage strategy, and we're ready to implement the database schema and storage layer.