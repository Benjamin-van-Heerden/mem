# Work Log - Agent and Team Streaming Validation with Database Schema Research

## Spec File: `agent_rules/spec/s_20251129_benjamin_van_heerden__agent_infrastructure_migration.md`

## Overarching Goals

Validate Agno's streaming mechanics for both individual agents and teams through practical CLI demos, with specific focus on:
1. Understanding exact event types, structures, and token reporting mechanisms
2. Testing image input handling with base64 encoding
3. Inspecting actual object shapes returned by Agno
4. Determining the correct database schema to support both agents and teams
5. Understanding Agno's `Message` class structure for proper message storage/reconstruction

## What Was Accomplished

### 1. Created Comprehensive Agent Streaming Demo

**Created `scripts/play/agent_demo.py`:**
- **Demo 1: Basic Content Streaming** - Validates content chunk accumulation
- **Demo 2: Full Event Streaming** - Shows tool call events (started/completed)
- **Demo 3: Multiple Tool Calls** - Sequential tool calls with separate events
- **Demo 4: Event Attribute Inspection** - Deep inspection of event object attributes using `getattr()`
- **Demo 5: Image Input with Token Tracking** - Base64 image encoding and token metrics inspection
- **Demo 6: Token Accumulation Detail** - Proves tokens only appear in `run_completed` event

**Key technical discovery - Claude API image format:**
```python
{
    "type": "image",
    "source": {
        "type": "base64",
        "media_type": "image/png",
        "data": "<base64_string>"
    }
}
```

### 2. Created Comprehensive Team Streaming Demo

**Created `scripts/play/teams_demo.py`:**
- **Demo 1: Team with Member Tools** - Shows team-level vs member-level events, delegation via `delegate_task_to_member`
- **Demo 2: Event Object Introspection** - Automated inspection of all event attributes across different event types
- **Demo 3: Message Reconstruction** - Demonstrates how to collect and store messages following "tool calls are user role" convention
- **Demo 4: Team Image Input with Token Tracking** - Image handling in team context with member tool calls

**Critical findings:**
- Teams emit both `TeamRunEvent` and `RunEvent` (member-level)
- Member events include `agent_id` attribute for tracking which agent acted
- Tool calls from members are tracked separately from team-level tool calls
- Token metrics appear only on `TeamRunCompleted` / `RunCompleted` events

### 3. Token Metrics Deep Dive

**Token tracking insights from demos:**
- Tokens are ONLY available in `run_completed` event's `metrics` attribute
- No per-event token counts during streaming - only final totals
- Metrics object structure:
  ```python
  metrics.input_tokens: int
  metrics.output_tokens: int
  metrics.total_tokens: int
  metrics.cache_read_tokens: int
  metrics.cache_write_tokens: int
  metrics.reasoning_tokens: int
  metrics.provider_metrics: dict  # e.g., {"service_tier": "standard"}
  ```
- Image input consumed 120 tokens (vs ~50 for text-only)
- Tool calls in Demo 6: 2608 input tokens, 202 output tokens = 2810 total

### 4. Image Input Validation

Successfully tested image input with `scripts/play/mystery.png`:
- Image correctly identified word "Ozymandias"
- Base64 encoding: 7108 characters for test image
- Proper format validation for Claude API
- Same image handling works for both agents and teams

### 5. Critical Discovery: Agno Message Class

**User provided critical Agno source code references:**

**`Message` class structure** (`agno/models/message.py`):
```python
class Message(BaseModel):
    id: str
    role: str  # system, user, assistant, or tool
    content: Optional[Union[List[Any], str]]
    name: Optional[str]
    tool_call_id: Optional[str]
    tool_calls: Optional[List[Dict[str, Any]]]
    
    # Media handled separately, NOT in content
    audio: Optional[Sequence[Audio]]
    images: Optional[Sequence[Image]]
    videos: Optional[Sequence[Video]]
    files: Optional[Sequence[File]]
    
    # Metrics built-in
    metrics: Metrics = Field(default_factory=Metrics)
    created_at: int  # Unix timestamp
    
    # Serialization methods
    def to_dict(self) -> Dict[str, Any]: ...
    def from_dict(cls, data: Dict[str, Any]) -> "Message": ...
```

**Agent `.arun()` signature** (`agno/agent/agent.py`):
```python
def arun(
    self,
    input: Union[str, List, Dict, Message, BaseModel, List[Message]],
    *,
    stream: Optional[bool] = None,
    images: Optional[Sequence[Image]] = None,  # Images passed separately!
    audio: Optional[Sequence[Audio]] = None,
    videos: Optional[Sequence[Video]] = None,
    files: Optional[Sequence[File]] = None,
    stream_events: Optional[bool] = None,
    ...
) -> Union[RunOutput, AsyncIterator[RunOutputEvent]]:
```

**Critical implications:**
1. **DO NOT construct message dicts manually** - use `Message` objects
2. **Images are NOT embedded in content** - they go in the `images` parameter as `Image` objects
3. **Use `List[Message]` for conversation history** - proper typed objects
4. **Serialization is built-in** - `Message.to_dict()` and `Message.from_dict()` exist
5. **Database should store what `Message.to_dict()` produces** - ensures perfect reconstruction

### 6. Database Schema Research Progress

**User's proposed schema (mermaid ERD format):**
```
users {uuid id PK}
teams {uuid id PK, text name}
agents {uuid id PK, text name}
sessions {
    uuid id PK,
    uuid user_id FK,
    uuid team_id FK (indexed+nullable),
    uuid agent_id FK (indexed+nullable),
    timestamp created_at,
    timestamp updated_at
}
chat_events {
    uuid id PK,
    uuid session_id FK,
    text model_provider_slug,
    text role,
    text agent_name,
    jsonb message_parts,
    integer input_tokens,
    integer output_tokens,
    decimal cost_in_rands,
    timestamp created_at (indexed)
}
```

**Key requirements identified:**
- UUID v7 for all IDs (time-ordered for unfragmented indexes)
- JSONB `message_parts` should match `Message.to_dict()` output
- Support both individual agents and teams in same schema
- Cost calculation built-in using existing exchange rate + provider cost files
- `agent_name` field to track which agent (or team member) created the message

**Outstanding questions:**
1. What exactly goes in `message_parts` JSONB? Should it be the full `Message.to_dict()` output?
2. How to handle base64 image storage - in DB or external reference?
3. Tool calls attribution: when a team member calls a tool, how do we track that in a single `chat_events` row?
4. Do we need separate tables for `users`, `teams`, `agents` or just VARCHAR references?

## Key Files Affected

**Created:**
- `scripts/play/agent_demo.py` (462 lines) - Six comprehensive agent streaming demos
- `scripts/play/teams_demo.py` (547 lines) - Four comprehensive team streaming demos
- `scripts/play/mystery.png` - Test image with word "Ozymandias"

**Modified:**
- `scripts/play/agent_demo.py` - Fixed Claude API image format (source.type, source.data)
- `scripts/play/teams_demo.py` - Fixed Claude API image format
- Both demos - Removed premature conclusion print statements

## What Comes Next

### Immediate Next Steps

1. **Update demos to use Agno `Message` objects properly**
   - Replace manual dict construction with `Message(role="user", content="...", images=[Image(...)])`
   - Use `Image.from_base64()` for proper image handling
   - Inspect what `Message.to_dict()` actually outputs for storage design
   - Validate that `Message.from_dict()` can reconstruct messages from DB

2. **Finalize database schema for `chat_events`**
   - Determine exact structure of `message_parts` JSONB
   - Decide: store full `Message.to_dict()` or subset?
   - Decide: base64 images in DB or external storage with references?
   - Determine how to handle tool calls from team members (agent attribution)

3. **Run teams demo to validate team behavior**
   ```bash
   uv run python scripts/play/teams_demo.py
   ```
   - Validate team-level vs member-level events
   - Inspect team message structures
   - Understand member delegation patterns

4. **Create database migration for agent tables**
   - Generate migration: `python scripts/migrate.py --generate "create_agent_tables"`
   - Implement tables: `users`, `teams`, `agents`, `sessions`, `chat_events`
   - Use UUID v7 for all primary keys
   - Create appropriate indexes (session_id, created_at)

5. **Implement storage layer**
   - File: `src/agents/storage.py`
   - Functions: `create_session()`, `store_message()`, `load_session_messages()`
   - Use Agno `Message.to_dict()` / `Message.from_dict()` for serialization
   - Handle JSONB storage of message data

6. **Prototype message storage/reconstruction**
   - Create a simple script that:
     1. Creates `Message` objects with images and tool calls
     2. Calls `message.to_dict()` and stores in mock JSONB
     3. Retrieves and reconstructs with `Message.from_dict()`
     4. Validates reconstruction is identical
   - This will validate the storage approach before building full system

### Architecture Decisions Needed

1. **Message storage granularity**: Store one `Message` per `chat_events` row, or pack multiple messages?
2. **Image storage strategy**: Inline base64 in JSONB (simple but large) vs external storage with references (complex but efficient)?
3. **Cost calculation timing**: At storage time (locked-in historical costs) vs query time (current rates)?
4. **Token attribution for teams**: How to split input/output tokens between team leader and members?

### Reference for Next Session

**Critical Agno classes to use:**
- `agno.models.message.Message` - DO NOT construct dicts manually
- `agno.models.image.Image` - Use `Image.from_base64()` for images
- `Message.to_dict()` - Serialization for DB storage
- `Message.from_dict()` - Deserialization for session reconstruction

**User's key insight:**
> "We cannot be constructing things like cavemen. Let's adopt the List[Message] as input structure."

The database schema MUST support storing and reconstructing `Message` objects perfectly so we can pass `List[Message]` to `.arun()` for session continuation.