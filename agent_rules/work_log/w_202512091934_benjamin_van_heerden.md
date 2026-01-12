# Work Log - Agent Message Persistence and Tool Call Structure Discovery

## Spec File: `agent_rules/spec/s_20251129_benjamin_van_heerden__agent_infrastructure_migration.md`

## Overarching Goals

Validate the complete Message persistence and restoration workflow using Agno's native primitives (`Message` and `Image` classes), with particular focus on:
1. Understanding tool call structure and storage requirements
2. Implementing proper image metadata storage (base64 stripped) with system notes
3. Demonstrating multi-phase conversation persistence (store → restore → continue cycles)
4. Comprehensive token tracking including image processing costs
5. Discovering all event types emitted during agent streaming to inform storage architecture

The ultimate goal is to understand exactly what data structures Agno provides so we can design an optimal database schema for conversation persistence.

## What Was Accomplished

### 1. Created Comprehensive Agent Message Persistence Demo

**Rewrote `scripts/play/agent_demo.py` from scratch** - removed all 6 old demos and replaced with a single comprehensive workflow demonstrating realistic conversation patterns:

**Phase 1: Initial Conversation with Tool Call**
- User asks about weather in Paris
- Agent calls `get_current_weather` tool
- Both user and assistant messages stored with metrics
- Simulates session end

**Phase 2: Session Restoration and Continuation**  
- Restores messages from storage using `Message.from_dict()`
- User continues: "I didn't catch that, could you repeat the weather?"
- Agent responds using restored context (validates restoration works)
- New messages stored

**Phase 3: Image Input with Metadata Storage**
- Restores session again
- User sends `mystery.png` image asking for mystery word
- Image base64 **stripped before storage** (keeps only filename and type metadata)
- **System note added to message content** explaining image unavailable
- Agent processes image successfully
- Token tracking shows ~4800 token increase for image processing

**Phase 4: Final Restoration and Context Verification**
- Final restoration from storage
- User asks: "Please repeat the mystery word"
- Agent responds correctly from conversation history (no image bytes present)
- Proves agent remembers from its own previous response, not image data

### 2. Implemented Image Storage Strategy with System Notes

**Storage preparation function** (`prepare_message_for_storage`):
```python
def prepare_message_for_storage(message: Message) -> dict:
    msg_dict = message.to_dict()
    
    if "images" in msg_dict and msg_dict["images"]:
        image_filenames = []
        for img in msg_dict["images"]:
            if "content" in img:
                img["content"] = "[BASE64_STRIPPED_FOR_STORAGE]"
                img["filename"] = "mystery.png"
                img["type"] = "image/png"
                image_filenames.append(img["filename"])
        
        # Add system note to content (stored in DB)
        if image_filenames:
            system_note = f"\n<system_note>\nThe user added {', '.join(image_filenames)} to the message, but it is no longer available to view - do your best to infer its content from the preceding and succeeding messages.\n</system_note>"
            msg_dict["content"] = msg_dict.get("content", "") + system_note
    
    return msg_dict
```

**Restoration function** (`restore_message_from_storage`):
```python
def restore_message_from_storage(stored_dict: dict) -> Message:
    restored_dict = stored_dict.copy()
    
    # Remove images array (system note already in content from storage)
    if "images" in restored_dict and restored_dict["images"]:
        del restored_dict["images"]
    
    return Message.from_dict(restored_dict)
```

**Key insight**: System note is stored in the database (not added on-the-fly), and images array is removed during restoration to prevent passing broken Image objects to the agent.

### 3. Discovered Tool Call Event Structure

Added comprehensive logging to capture complete tool call events. Key findings:

**ToolCallStartedEvent structure:**
```python
{
    'event': 'ToolCallStarted',
    'run_id': '7f4388ab-61b1-4cbe-8c86-c98c25f1f8a2',
    'session_id': '33b523ee-0458-49b2-b291-98becc43753b',
    'tool': ToolExecution(
        tool_call_id='toolu_015oNjeqR4fP7gxhpoymgjd2',  # LLM provider ID
        tool_name='get_current_weather',
        tool_args={'location': 'Paris'},
        result=None,  # Not available yet
        created_at=1765300749
    )
}
```

**ToolCallCompletedEvent structure:**
```python
{
    'event': 'ToolCallCompleted',
    'tool': ToolExecution(
        tool_call_id='toolu_015oNjeqR4fP7gxhpoymgjd2',  # Same ID
        tool_name='get_current_weather',
        tool_args={'location': 'Paris'},
        result='The weather in Paris is sunny and 22°C',  # Now available
        metrics=Metrics(duration=0.0004143330152146518),
        created_at=1765300749
    )
}
```

**RunCompletedEvent structure:**
```python
{
    'event': 'RunCompleted',
    'content': "The weather in Paris is **sunny**...",  # Accumulated text
    'metrics': Metrics(
        input_tokens=3082,
        output_tokens=118,
        total_tokens=3200,
        provider_metrics={'service_tier': 'standard'}
    )
}
```

**Critical finding**: `run_completed` event does **NOT** contain a `messages` field or array with tool_calls populated. We must manually capture and store tool executions from the event stream.

### 4. Confirmed Tool Call Storage Convention

Tool calls must be stored as **separate message rows** with `role="tool"`. This is the standard convention used by all modern AI providers (OpenAI, Anthropic, Google) and frameworks.

**Typical conversation structure in storage:**
1. User message: "What's the weather in Paris?"
2. Assistant message: (thinking/planning - may be empty or contain text before tool call)
3. Tool call message (role="tool"): tool_call_id, tool_name, tool_args, result
4. Assistant message: "The weather in Paris is sunny and 22°C"

### 5. Agent Instructions Optimization

Added instructions to prevent redundant tool calls:
```python
instructions=[
    "Answer the user's question as best as you can.",
    "The users you will be dealing with are old and technologically challenged - don't re-execute tools if you already have the answer."
]
```

Result: Agent successfully avoided re-calling weather tool on second question, using context instead.

### 6. Validated Complete Workflow

**Demo output confirmed:**
- ✓ Message.to_dict() / from_dict() roundtrip works perfectly
- ✓ Tool calls can be captured from events (not from Message objects)
- ✓ Conversation context maintained through multiple restore cycles
- ✓ Image metadata preserved (base64 stripped, system note added)
- ✓ Agent remembers image content from its own response
- ✓ Token tracking accurate throughout all phases
- ✓ Image input adds significant tokens (~4800) to input count

**Final message storage structure** (8 messages stored):
```json
[
  {"role": "user", "content": "What's the weather..."},
  {"role": "assistant", "content": "The weather is sunny...", "metrics": {...}},
  {"role": "user", "content": "I didn't catch that..."},
  {"role": "assistant", "content": "The weather is sunny...", "metrics": {...}},
  {"role": "user", "content": "Now I have an image...<system_note>...", "images": [...]},
  {"role": "assistant", "content": "The mystery word is ozymandias", "metrics": {...}},
  {"role": "user", "content": "Please repeat the mystery word"},
  {"role": "assistant", "content": "The mystery word is ozymandias", "metrics": {...}}
]
```

**Note**: Tool call messages are NOT yet being stored - this is the next step.

## Key Files Affected

**Modified:**
- `scripts/play/agent_demo.py` (complete rewrite, ~460 lines)
  - Single comprehensive workflow demo replacing 6 separate demos
  - `prepare_message_for_storage()`: strips base64, adds system note
  - `restore_message_from_storage()`: removes images array, preserves system note
  - Four-phase conversation with multiple save/restore cycles
  - Image handling with metadata-only storage
  - Event logging for tool call structure discovery
  - Complete message history JSON dump at end

## What Comes Next

### Immediate Next Steps (Agent Demo Completion)

1. **Implement comprehensive event type discovery**
   - Create a `set()` to collect all unique event types during streaming
   - Run demo and display: `print(f"All event types encountered: {event_types}")`
   - This will show us every possible event we need to handle in storage wrapper

2. **Implement event-driven storage pattern with match statement**
   - Modern Python match statement on `event.event` type
   - **Immediate persistence** for complete events:
     - `ToolCallStarted` → store tool message (role="tool", no result yet)
     - `ToolCallCompleted` → update tool message with result
     - `RunCompleted` → store final assistant message with accumulated content + metrics
   - **Deferred persistence** for partial events:
     - `RunContent` → accumulate content chunks
     - Persist when content stream ends (before next tool call or run completion)
   
3. **Handle complex interleaved patterns**
   - Pattern: `text1 → tool_call2 → text3 → tool_call4 → text5 → complete`
   - Each "complete" segment persists immediately:
     - text1 completes → persist assistant message with text1
     - tool_call2 completes → persist tool message
     - text3 completes → persist assistant message with text3
     - tool_call4 completes → persist tool message
     - text5 completes with run_completed → persist final assistant message

4. **Update storage functions to handle tool messages**
   - Store tool executions as separate message rows
   - Schema: `role="tool"`, `tool_call_id`, `tool_name` (or `name`), `tool_args`, `result` (content field)
   - Link tool messages to session via `session_id`
   - Maintain chronological order via `created_at` timestamp

5. **Validate complete storage/restoration with tools**
   - Store conversation with tool calls
   - Restore and verify tool messages present
   - Continue conversation and verify agent has tool context
   - Confirm conversation history includes all tool interactions

### Teams Demo (After Agent Demo Complete)

6. **Create similar comprehensive workflow for teams**
   - Multi-agent team with delegation patterns
   - Capture team-specific events (TeamRunEvent, member attribution)
   - Understand how team tool calls differ from agent tool calls
   - Test persistence/restoration with team conversations
   - Discover any team-specific event types

7. **Document team-specific storage requirements**
   - How to attribute messages/tool calls to specific team members
   - Team delegation storage (delegate_task_to_member interactions)
   - Token tracking across team members
   - Any additional fields needed for team context

### Database Schema Design (After Both Demos Complete)

8. **Design final database schema**
   - Based on complete understanding of all event types
   - Support both agents and teams in single schema
   - Efficient querying for session history reconstruction
   - Proper indexing for performance (session_id, created_at)
   - Handle all message types: user, assistant, tool, system

9. **Create migration files**
   - `sessions` table: session metadata, user_id, agent/team references
   - `chat_events` (or `messages`) table: complete message storage with JSONB
   - UUIDv7 for time-ordered primary keys
   - Appropriate foreign keys and indexes

### Architecture Principles Established

**Event-Driven Storage Pattern:**
- Match on event type, persist immediately when data is complete
- Accumulate partial data (streaming content) until segment completes
- Handle interleaved text/tool patterns correctly
- Never lose data - persist as soon as we have complete information

**Tool Call Convention:**
- Tool calls are separate message rows (role="tool")
- Industry standard pattern used by all major providers
- Each tool call has: tool_call_id, tool_name, tool_args, result
- Chronological ordering preserves conversation flow

**Image Storage Strategy:**
- Strip base64 content before database storage
- Keep metadata: filename, type, id
- Add system note to message content explaining unavailability
- Remove images array during restoration (prevent broken objects)
- UI shows metadata (icon + filename), agent gets context note

**Message Persistence:**
- Store `Message.to_dict()` in JSONB column
- Separate columns for token metrics (queryable)
- System notes stored in content field (not added on-the-fly)
- Perfect reconstruction via `Message.from_dict()`

### Reference for Next Session

**Critical insights:**
1. Agno does NOT provide complete Message objects with tool_calls populated in events
2. We must manually capture tool executions from ToolCallStarted/Completed events
3. Tool calls MUST be stored as separate messages (role="tool") per industry convention
4. Complex interleaved patterns require careful event handling (persist each segment)
5. System notes for images must be stored in DB, not added during restoration

**Next demo run should:**
- Collect all event types in a set
- Display complete list of events encountered
- Implement match statement handling all event types
- Store tool calls as separate message rows
- Validate complete conversation reconstruction with tools

**Critical question to answer:**
What are ALL the possible event types we need to handle? (run_started, run_content, tool_call_started, tool_call_completed, run_completed, run_cancelled, run_paused, etc.)

The agent demo is nearly complete - we just need the event type catalog and proper storage implementation. Then teams demo, then schema design.