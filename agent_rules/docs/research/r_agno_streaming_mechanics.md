# Agno Streaming Mechanics Research

## Overview

This document provides comprehensive research on how Agno's agent streaming works, including event types, structures, and patterns needed to implement our custom streaming wrapper.

## Core Streaming Pattern

### Basic Streaming Invocation

```python
# Synchronous streaming
stream: Iterator[RunOutputEvent] = agent.run("message", stream=True)

# Asynchronous streaming
async_stream: AsyncIterator[RunOutputEvent] = agent.arun("message", stream=True)

# With all events (not just content)
stream = agent.run("message", stream=True, stream_events=True)
```

**Key Parameters:**
- `stream=True` - Enable streaming response
- `stream_events=True` - Include all events (tool calls, reasoning, etc.), not just content

## Event Types and Structure

### RunEvent Enum

The `RunEvent` enum defines all possible event types:

```python
from agno.agent import RunEvent

# Main lifecycle events
RunEvent.run_started          # Run begins
RunEvent.run_completed        # Run finishes successfully
RunEvent.run_cancelled        # Run was cancelled
RunEvent.run_content          # Content chunk

# Tool execution events
RunEvent.tool_call_started    # Tool call initiated
RunEvent.tool_call_completed  # Tool call finished

# Reasoning events (for thinking models)
RunEvent.reasoning_step       # Reasoning/thinking step

# Human-in-the-loop events
RunEvent.run_paused           # Awaiting user input/confirmation
```

### RunOutputEvent Structure

Each yielded event is a `RunOutputEvent` with these key attributes:

```python
class RunOutputEvent:
    event: str                    # Event type (from RunEvent enum)
    run_id: str                   # Unique run identifier
    content: str | None           # Text content (for content events)
    status: RunStatus | None      # Run status (pending, running, completed, cancelled, error)
    
    # Tool-related attributes (when event is tool_call_*)
    tool: ToolExecution | None    # Tool execution details
    
    # Pause-related attributes (when event is run_paused)
    is_paused: bool
    tools_requiring_user_input: List[ToolExecution]
    tools_awaiting_external_execution: List[ToolExecution]
```

### ToolExecution Structure

When tools are called, the event includes:

```python
class ToolExecution:
    tool_name: str               # Name of the tool
    tool_args: Dict[str, Any]    # Arguments passed to tool
    result: Any | None           # Tool execution result (when completed)
    tool_call_id: str            # Unique identifier for this tool call
```

## Event Flow Patterns

### Normal Successful Run

```python
1. RunEvent.run_started         # Run begins
2. RunEvent.tool_call_started   # (if tools used)
3. RunEvent.tool_call_completed # (if tools used)
4. RunEvent.run_content         # Content chunk 1
5. RunEvent.run_content         # Content chunk 2
6. RunEvent.run_content         # Content chunk N
7. RunEvent.run_completed       # Run finishes
```

### Run with Multiple Tool Calls

```python
1. RunEvent.run_started
2. RunEvent.tool_call_started   # Tool A starts
3. RunEvent.tool_call_completed # Tool A completes
4. RunEvent.tool_call_started   # Tool B starts
5. RunEvent.tool_call_completed # Tool B completes
6. RunEvent.run_content         # Agent response using tool results
7. RunEvent.run_completed
```

### Run with Cancellation

```python
1. RunEvent.run_started
2. RunEvent.run_content         # Partial content
3. RunEvent.run_content         # Partial content
4. RunEvent.run_cancelled       # Cancellation detected
# Stream ends, no run_completed event
```

### Run with Pause (Human-in-the-Loop)

```python
1. RunEvent.run_started
2. RunEvent.tool_call_started   # Tool requiring confirmation
3. (stream pauses, event.is_paused = True)
# User provides confirmation
# Use agent.continue_run() to resume
4. RunEvent.tool_call_completed
5. RunEvent.run_content
6. RunEvent.run_completed
```

## Streaming Response Examples

### Example 1: Basic Content Streaming

```python
async for event in agent.arun("Write a story", stream=True):
    if event.event == RunEvent.run_content:
        # event.content contains text chunk
        print(event.content, end="", flush=True)
    elif event.event == RunEvent.run_completed:
        print("\nDone!")
```

### Example 2: Full Event Handling

```python
async for event in agent.arun("Search and summarize", stream=True, stream_events=True):
    if event.event == RunEvent.run_started:
        print(f"Run started: {event.run_id}")
    
    elif event.event == RunEvent.tool_call_started:
        print(f"Tool: {event.tool.tool_name}")
        print(f"Args: {event.tool.tool_args}")
    
    elif event.event == RunEvent.tool_call_completed:
        print(f"Result: {event.tool.result}")
    
    elif event.event == RunEvent.run_content:
        print(event.content, end="", flush=True)
    
    elif event.event == RunEvent.run_completed:
        print("\nCompleted!")
    
    elif event.event == RunEvent.run_cancelled:
        print("\nCancelled!")
```

### Example 3: Handling Paused Runs

```python
run_response = agent.run("Send email to user", stream=False)

if run_response.is_paused:
    # Tool requires user input
    for tool in run_response.tools_requiring_user_input:
        # Collect user input for required fields
        for field in tool.user_input_schema:
            field.value = input(f"Enter {field.name}: ")
    
    # Continue execution
    final_response = agent.continue_run(run_response=run_response)
```

## Cancellation Mechanism

### How Cancellation Works

```python
# In one thread/task: Start the run
async for event in agent.arun("long task", stream=True):
    # Store run_id when first event arrives
    if event.run_id:
        stored_run_id = event.run_id
    
    # Process events...
    if event.event == RunEvent.run_cancelled:
        print("Run was cancelled!")
        break

# In another thread/task: Cancel the run
success = agent.cancel_run(run_id)
# Returns True if cancellation was successful
```

**Key Points:**
- Cancellation is asynchronous - may take a moment to take effect
- When cancelled, a `RunEvent.run_cancelled` event is emitted
- No `run_completed` event after cancellation
- Partial content before cancellation is accessible

## Error Handling

### Exception in Stream

```python
try:
    async for event in agent.arun("message", stream=True):
        # Process events
        pass
except Exception as e:
    # Handle streaming errors
    print(f"Streaming error: {e}")
    # Store partial content if available
```

**Common Error Scenarios:**
- Network interruption during streaming
- Tool execution failures
- Model API errors
- Invalid tool arguments

### Tool Execution Errors

When a tool fails, the error is typically captured in the tool result:

```python
if event.event == RunEvent.tool_call_completed:
    if isinstance(event.tool.result, Exception):
        print(f"Tool failed: {event.tool.result}")
    # Or tool may return error string
    elif "error" in str(event.tool.result).lower():
        print(f"Tool error: {event.tool.result}")
```

## Client Disconnection Detection

Agno doesn't explicitly detect client disconnections. Our wrapper needs to handle this:

```python
async def stream_to_client():
    try:
        async for event in agent.arun("message", stream=True):
            # Attempt to yield to client
            yield format_sse(event)
    except asyncio.CancelledError:
        # Client disconnected
        print("Client disconnected")
        # Mark message as interrupted in database
    except Exception as e:
        # Other streaming errors
        print(f"Streaming error: {e}")
```

## Team Streaming

Teams have similar streaming patterns with `TeamRunOutputEvent`:

```python
from agno.team import Team, TeamRunEvent

async for event in team.arun("message", stream=True, stream_events=True):
    if event.event == TeamRunEvent.run_content:
        print(event.content, end="")
    elif event.event == TeamRunEvent.member_run_started:
        print(f"\nMember {event.member_name} started")
    elif event.event == TeamRunEvent.member_run_completed:
        print(f"\nMember {event.member_name} completed")
```

## Storage Considerations

### What to Store

Based on event types, we should store:

**For each message:**
- `role` - user/assistant/tool
- `content` - Accumulated text from `run_content` events
- `tool_call_id` - If this is a tool call/result
- `tool_name` - Name of tool called
- `tool_arguments` - JSON of tool args
- `created_at` - When message started
- `completed_at` - When message finished (or was interrupted)
- `interrupted` - Boolean flag
- `error` - Error message if any

**For each run:**
- `run_id` - Unique identifier from Agno
- `session_id` - Session this run belongs to
- `user_message_id` - The triggering message
- `assistant_message_id` - The response message
- `started_at` - Run start timestamp
- `completed_at` - Run completion timestamp
- `status` - running/completed/interrupted/error
- `total_tokens` - Token usage (if available)

### Incremental Storage Pattern

```python
# 1. Create run record on run_started
if event.event == RunEvent.run_started:
    run_id = event.run_id
    create_run(run_id, session_id, user_message_id, status='running')
    assistant_message_id = create_message(
        session_id=session_id,
        role='assistant',
        content='',  # Will accumulate
    )

# 2. Accumulate content chunks
accumulated_content = []
if event.event == RunEvent.run_content:
    accumulated_content.append(event.content)
    # Optionally update message in DB periodically

# 3. Store tool calls as separate messages
if event.event == RunEvent.tool_call_completed:
    create_message(
        session_id=session_id,
        role='tool',
        tool_call_id=event.tool.tool_call_id,
        tool_name=event.tool.tool_name,
        tool_arguments=event.tool.tool_args,
        content=str(event.tool.result),
    )

# 4. Finalize on completion
if event.event == RunEvent.run_completed:
    update_message(
        assistant_message_id,
        content=''.join(accumulated_content),
        completed_at=datetime.now(),
    )
    complete_run(run_id, status='completed')
```

## Implementation Checklist for Custom Wrapper

Based on this research, our custom streaming wrapper must:

- [ ] Handle both sync and async streaming (prioritize async)
- [ ] Track run_id from first event
- [ ] Accumulate content chunks from `run_content` events
- [ ] Store tool calls when `tool_call_started` fires
- [ ] Store tool results when `tool_call_completed` fires
- [ ] Check session manager for stop signals on each event
- [ ] Handle `run_cancelled` event (append termination message)
- [ ] Catch exceptions during streaming (network, tool errors)
- [ ] Detect client disconnections (asyncio.CancelledError)
- [ ] Finalize messages and runs on `run_completed`
- [ ] Store partial content when interrupted or errored
- [ ] Yield SSE-formatted events to client
- [ ] Handle pause events if we enable human-in-the-loop later

## Key Takeaways

1. **Events are well-structured** - `RunEvent` enum covers all scenarios
2. **Content comes in chunks** - Must accumulate `run_content` events
3. **Tool calls are separate events** - Started and completed events with full details
4. **Cancellation is explicit** - `run_cancelled` event signals termination
5. **run_id is critical** - Available on all events, needed for cancellation
6. **Streaming is async-friendly** - Use `arun()` for FastAPI integration
7. **No explicit disconnect detection** - Must catch `asyncio.CancelledError`
8. **Final event matters** - `run_completed` vs `run_cancelled` determines outcome
9. **Errors can happen anywhere** - Tool failures, network issues, model errors
10. **Incremental storage is key** - Don't wait for completion to start storing

## References

- Agno Agent Running: https://github.com/agno-agi/agno-docs/blob/main/concepts/agents/running-agents.mdx
- Event Handling Examples: https://github.com/agno-agi/agno-docs/blob/main/examples/concepts/agent/events/
- Cancellation: https://github.com/agno-agi/agno-docs/blob/main/concepts/agents/run-cancel.mdx
- Human-in-the-Loop: https://github.com/agno-agi/agno-docs/blob/main/concepts/hitl/overview.mdx