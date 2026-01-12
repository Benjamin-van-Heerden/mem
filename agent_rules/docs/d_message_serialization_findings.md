# Message Serialization Findings

## Summary

This document captures the key findings from exploring Agno's `Message` and `Image` classes through CLI demos. These findings inform our database schema design for storing agent/team conversations.

**Date:** 2024-12-09  
**Demos:** `scripts/play/agent_demo.py`, `scripts/play/teams_demo.py`

---

## 1. Message.to_dict() Structure

### Text-Only Message
```json
{
  "id": "e78cbc35-74a1-4e5a-9fdf-4a3b12b7bdc6",
  "content": "What is the capital of France?",
  "from_history": false,
  "stop_after_tool_call": false,
  "role": "user",
  "created_at": 1765280936
}
```

### Key Fields
- **`id`**: Auto-generated UUID for each message
- **`content`**: String content of the message
- **`role`**: `"user"`, `"assistant"`, `"system"`, or `"tool"`
- **`from_history`**: Boolean flag (likely used during reconstruction)
- **`stop_after_tool_call`**: Boolean flag for tool handling logic
- **`created_at`**: Unix timestamp (integer, not ISO string!)

---

## 2. Image Serialization

Images are stored in a separate `images` array, **not in the content field**:

```json
{
  "id": "0908fd92-56dc-4e80-855d-03d268830350",
  "content": "What word do you see in this image?",
  "from_history": false,
  "stop_after_tool_call": false,
  "role": "user",
  "images": [
    {
      "id": "f6e052ed-c857-4683-a54c-df7e38c8d27e",
      "content": "iVBORw0KGgoAAAANSUhEUgAA... (full base64)"
    }
  ],
  "created_at": 1765280936
}
```

### Key Points
- Each image has its own UUID
- Base64 content stored in `image.content` field (not `image.base64`)
- Multiple images supported (array)
- Images are **separate from text content**

---

## 3. Metrics in Messages

Metrics can be attached to Message objects (typically assistant messages):

```json
{
  "id": "3fb809a5-67b3-4e98-8746-608d47f9928c",
  "content": "2 + 2 equals 4.",
  "from_history": false,
  "stop_after_tool_call": false,
  "role": "assistant",
  "metrics": {
    "input_tokens": 82,
    "output_tokens": 22,
    "total_tokens": 104
  },
  "created_at": 1765280937
}
```

### Metrics Fields Available
- `input_tokens`: Token count for input
- `output_tokens`: Token count for output
- `total_tokens`: Sum of input + output
- `cache_read_tokens`: Tokens read from cache
- `cache_write_tokens`: Tokens written to cache
- `reasoning_tokens`: Tokens used for reasoning (if applicable)

**Important:** Metrics are ONLY available at the `run_completed` event, not during streaming.

---

## 4. Message Reconstruction

### Creating Messages
```python
from agno.models.message import Message
from agno.media import Image

# Text message
msg = Message(role="user", content="Hello")

# Message with image
image = Image.from_base64(base64_data, media_type="image/png")
msg = Message(role="user", content="Describe this", images=[image])
```

### Serialization & Reconstruction
```python
# Store to database
msg_dict = msg.to_dict()
# Store msg_dict as JSONB in database

# Reconstruct from database
stored_dict = # ... fetch from database
reconstructed_msg = Message.from_dict(stored_dict)
```

**Validation:** Reconstruction is **perfect** - all fields preserved, including images.

---

## 5. Agent Input Format

Agents accept `List[Message]` for conversation history:

```python
from agno.agent import Agent

# Single message
response = await agent.arun("Hello")

# Or as Message object
msg = Message(role="user", content="Hello")
response = await agent.arun([msg])

# Multiple messages (conversation history)
messages = [
    Message(role="user", content="What is 2+2?"),
    Message(role="assistant", content="4"),
    Message(role="user", content="What about 3+3?")
]
response = await agent.arun(messages)
```

**Note:** Images passed via `Message.images`, not via separate `images` parameter when using Message objects.

---

## 6. Teams Behavior

**Key Finding:** Teams work identically to agents regarding Message handling.

```python
from agno.team import Team

# Teams accept List[Message] just like agents
messages = [Message(role="user", content="Hello team")]
response = await team.arun(messages)

# Message.to_dict() structure is identical for team responses
# Images work the same way
# Metrics attached to messages the same way
```

### Team-Specific Considerations
- **Member attribution**: Events include `agent_id` to track which team member acted
- **Delegation**: Teams use `delegate_task_to_member` tool for coordination
- **Storage**: Add `agent_name` or `team_name` field in DB to track context

---

## 7. Database Schema Implications

### Recommended Approach: Store Full Message.to_dict() in JSONB

**Table: `chat_events`**
```sql
CREATE TABLE chat_events (
    id UUID PRIMARY KEY,  -- UUIDv7 for time-ordered indexes
    session_id UUID NOT NULL,
    role VARCHAR(20) NOT NULL,
    agent_name VARCHAR(255),  -- Track which agent/team member
    message_data JSONB NOT NULL,  -- Full Message.to_dict() output
    input_tokens INTEGER,
    output_tokens INTEGER,
    cost_in_rands DECIMAL(10,4),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    INDEX idx_session_created (session_id, created_at)
);
```

### Storage Strategy

**Option A: Inline Base64 (Simple)**
- Store full `Message.to_dict()` including base64 images in `message_data`
- Pros: Simple, single source of truth
- Cons: Large JSONB columns for images

**Option B: External Image Storage (Scalable)**
- Store images in object storage (S3, local files)
- Replace `images[].content` with reference (URL or file_id)
- Pros: Smaller DB, better performance
- Cons: More complexity, two-phase storage

**Recommendation for MVP:** Start with Option A (inline), migrate to Option B when needed.

### Reconstruction Query Pattern
```python
# Fetch messages for session
rows = db.query("SELECT message_data FROM chat_events WHERE session_id = ? ORDER BY created_at", session_id)

# Reconstruct Message objects
messages = [Message.from_dict(row['message_data']) for row in rows]

# Pass to agent
response = await agent.arun(messages, stream=True)
```

---

## 8. Token Metrics Handling

### When Tokens Are Available
- Tokens are **ONLY** reported in the `run_completed` event
- No per-chunk token metrics during streaming
- Must wait for completion to get accurate counts

### Storage Pattern
```python
# During streaming
accumulated_content = []
async for event in agent.arun(message, stream=True):
    if event.event == RunEvent.run_content:
        accumulated_content.append(event.content)
    elif event.event == RunEvent.run_completed:
        # Create assistant message
        assistant_msg = Message(
            role="assistant",
            content="".join(accumulated_content)
        )
        
        # Attach metrics
        if event.metrics:
            assistant_msg.metrics.input_tokens = event.metrics.input_tokens
            assistant_msg.metrics.output_tokens = event.metrics.output_tokens
            assistant_msg.metrics.total_tokens = event.metrics.total_tokens
        
        # Store to database
        db.insert_chat_event(
            session_id=session_id,
            message_data=assistant_msg.to_dict(),
            input_tokens=event.metrics.input_tokens,
            output_tokens=event.metrics.output_tokens
        )
```

---

## 9. Tool Call Handling

**Note:** Tool calls are not yet fully explored in the demos. Further investigation needed:
- How are tool calls represented in Message objects?
- What is the `tool_call_id`, `tool_name`, `tool_args` structure?
- How to reconstruct tool call sequences?

**TODO:** Add demo exploring tool call Message structures.

---

## 10. Key Takeaways

✅ **DO:**
- Store full `Message.to_dict()` output in JSONB
- Use `Message.from_dict()` for reconstruction
- Pass `List[Message]` to `agent.arun()` for conversation history
- Use `Image.from_base64()` for proper image handling
- Attach metrics to assistant messages after `run_completed`

❌ **DON'T:**
- Manually construct message dicts (use Message class)
- Put images in content field (use Message.images)
- Expect tokens during streaming (only at completion)
- Store custom message formats (use Agno's structure)

---

## 11. Next Steps

1. **Finalize database schema** with `message_data` JSONB column
2. **Decide on image storage strategy** (inline vs external)
3. **Explore tool call Message structures** (add demo)
4. **Prototype storage layer** with roundtrip test (Message → DB → Message)
5. **Implement SessionManager** for stop signals
6. **Build streaming wrapper** with incremental storage

---

## References

- Demos: `scripts/play/agent_demo.py` (Demo 2, 3, 7)
- Demos: `scripts/play/teams_demo.py` (Demo 1, 5)
- Agno imports: `from agno.models.message import Message`
- Agno imports: `from agno.media import Image`
