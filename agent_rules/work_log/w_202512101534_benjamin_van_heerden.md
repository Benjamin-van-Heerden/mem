# Work Log - Teams Demo Debugging and Event Flow Analysis

## Spec File: `agent_rules/spec/s_20251129_benjamin_van_heerden__agent_infrastructure_migration.md`

## Overarching Goals

Debug and finalize the `teams_demo.py` script to validate the complete message persistence workflow for Agent Teams. The primary goal was to ensure that the Team Leader's final response (synthesizing member outputs) is correctly captured and stored, preventing the "Groundhog Day" loop where the team repeatedly delegates tasks because it lacks a record of having answered the user.

## What Was Accomplished

### 1. Refactored Teams Demo for Parallel Execution
- Replaced linear `active_agent_name` tracking with a `dict[run_id, ContextState]` approach.
- Implemented `ContextState` (Pydantic model) to track `agent_name`, `current_msg`, `is_stored`, and `metrics` per execution run.
- This correctly handles the interleaved event stream where the Team Leader delegates to multiple members (Weather/Finance) who execute concurrently.

### 2. Implemented Robust Event Handling
- Added auto-registration for Team contexts: If a `Team*` event fires (e.g., `TeamRunContent`) and the `run_id` is missing from `active_contexts`, it is automatically created. This fixed "Unknown" agent logs.
- Added a catch-all handler for `TeamRunCompleted` to bypass potential class mismatch issues with `RunCompletedEvent`.

### 3. Debugged Missing Team Leader Response
- **Symptom:** The Team Leader delegates, members answer, but the final Team response is never stored. This causes the team to re-delegate endlessly upon restoration.
- **Investigation:**
  - Added comprehensive event logging (`Event Progression`) which confirmed `TeamRunContent` events *are* firing (e.g., 25+ chunks).
  - Added debug prints which revealed `content_len=0` at `TeamRunCompleted`.
  - **Critical Finding:** Despite `TeamRunContent` events firing, the `event.content` attribute appears to be empty or `None` in the handler, or the `run_id` lookup logic is subtly failing in a way that prevents accumulation.
  - The logs show `('Info Team', 'TeamRunContent'): 25`, confirming the events exist and are attributed to the correct agent.
  - However, the final storage logic sees an empty message and skips storage.

### 4. Prompt Engineering
- Updated team instructions to explicitly handle image inputs: *"If the user provides an image, DO NOT DELEGATE. Analyze it yourself to answer the question."*
- Despite this, the team continued to delegate, likely due to the missing history of its own previous answers.

## Key Files Affected

- `scripts/play/teams_demo.py`:
  - Major refactor of `run_team_and_store` to use `active_contexts`.
  - Added `ContextState` model.
  - Added robust event logging and progression analysis.
  - Added debug prints for content accumulation.
  - Updated team instructions.

## Errors and Barriers

### Unresolved: Missing Team Leader Content
The core issue remains: **The Team Leader's final response text is not being captured.**
- `TeamRunContent` events fire (25-30 times).
- `TeamRunCompleted` fires.
- But `ctx.current_msg.content` remains empty at completion.
- Debug prints suggest we might not be entering the content accumulation block, or `event.content` is indeed empty/None.
- If `event.content` is empty, where is the text? It might be in `event.delta` or another field, or the event structure for Teams differs from Agents in this version of Agno.

## What Comes Next

### 1. Fix Content Capture
- Investigate `TeamRunContent` event structure. Print `dir(event)` or `event.model_dump()` to find where the text is hiding.
- It is possible `TeamRunContent` uses `delta` instead of `content`, or the content is being streamed differently.
- Once content is captured, the final message will be stored, breaking the delegation loop.

### 2. Finalize Teams Demo
- Verify that Phase 2 (Restoration) and Phase 3 (Image) work correctly once the leader's response is stored.
- Ensure "Ozymandias" is correctly identified by the leader (or member if delegated).

### 3. Database Schema & Migration
- Once the demo validates the data structures, proceed to create the `chat_events` table migration.
- Implement the actual storage backend (`src/agents/storage.py`) based on the demo's `prepare_message_for_storage` logic.

## Deep Dive: The "Groundhog Day" Delegation Loop

The persistence failure creates a critical UX loop:
1. **User asks question.**
2. **Team delegates** (stores delegation tool call).
3. **Members answer** (stores member responses).
4. **Team Leader synthesizes final answer** (streams content but FAILS TO STORE it).
5. **Session ends.**
6. **Session restored.** History contains: User Question -> Delegation -> Member Answers. **Missing: Final Team Answer.**
7. **User asks follow-up.**
8. **Team Leader looks at history.** Sees unanswered question (or at least, no final resolution from itself).
9. **Team Leader decides to delegate again** to get the answer it thinks is missing.
10. **Loop repeats.**

This confirms that **perfect fidelity in message storage** (including the final synthesis) is mandatory for multi-turn team conversations. The fix likely involves identifying the correct attribute on `TeamRunContent` events (e.g., `delta` vs `content`) or ensuring the `run_id` mapping is robust enough to capture the stream.

### Concrete Example: Phase 1 (Initial Delegation)

**Console output (abridged):**
```text
--- PHASE 1: Initial Conversation with Delegation ---

USER: What is the weather in Paris and the USD to EUR exchange rate?
✓ Stored User Message

TEAM PROCESSING (Initial delegation)...
INFO Agent 'Weather Agent' inheriting model from Team: claude-haiku-4-5
INFO Agent 'Finance Agent' inheriting model from Team: claude-haiku-4-5

👤 MEMBER STARTED: Weather Agent | run_id: ...
👤 MEMBER STARTED: Finance Agent | run_id: ...

[MEMBER TOOL CALL (Weather Agent): get_current_weather]
[MEMBER TOOL CALL COMPLETED (Weather Agent)]
✓ Stored Request (Weather Agent): { ... "agent_name": "Weather Agent" }
✓ Stored Tool Result: { ... "content": "The weather in Paris is sunny and 22°C", "agent_name": "Weather Agent" }

[MEMBER TOOL CALL (Finance Agent): get_exchange_rate]
[MEMBER TOOL CALL COMPLETED (Finance Agent)]
✓ Stored Request (Finance Agent): { ... "agent_name": "Finance Agent" }
✓ Stored Tool Result: { ... "content": "1 USD = 0.85 EUR", "agent_name": "Finance Agent" }

The current USD to EUR exchange rate is **1 USD = 0.85 EUR**.

This means that 1 US Dollar is equivalent to 0.85 Euros.
[MEMBER RUN COMPLETED (Finance Agent)]
✓ Stored Final Message (Finance Agent)

Here's the current weather information for Paris, France:

- **Temperature:** 22°C (approximately 72°F)
- **Conditions:** Sunny

The weather in Paris is currently pleasant with sunny skies and mild temperatures. It's a great day to be outdoors!
[MEMBER RUN COMPLETED (Weather Agent)]
✓ Stored Final Message (Weather Agent)

[DEBUG] Hit TeamRunCompleted catch-all. run_id=...
[DEBUG] Found context for TeamRunCompleted. stored=False content_len=0
[TEAM RUN COMPLETED (Info Team)]
```

**Event progression for this phase:**
```text
Event Progression (Initial delegation):
('Info Team', 'TeamRunStarted'): 1
('Info Team', 'TeamRunContent'): 8
('Info Team', 'TeamToolCallStarted'): 2
('Weather Agent', 'RunStarted'): 1
('Finance Agent', 'RunStarted'): 1
('Finance Agent', 'RunContent'): 1
('Finance Agent', 'ToolCallStarted'): 1
('Finance Agent', 'ToolCallCompleted'): 1
('Finance Agent', 'RunContent'): 10
('Finance Agent', 'RunContentCompleted'): 1
('Finance Agent', 'RunCompleted'): 1
('Weather Agent', 'RunContent'): 1
('Weather Agent', 'ToolCallStarted'): 1
('Weather Agent', 'ToolCallCompleted'): 1
('Weather Agent', 'RunContent'): 23
('Weather Agent', 'RunContentCompleted'): 1
('Weather Agent', 'RunCompleted'): 1
('Info Team', 'TeamToolCallCompleted'): 2
('Info Team', 'TeamRunContent'): 25
('Info Team', 'TeamRunContentCompleted'): 1
('Info Team', 'TeamRunCompleted'): 1
```

**Key observations:**
- We clearly see **Info Team** emitting `TeamRunContent` 8 + 25 times and `TeamRunCompleted` once.
- Debug output for `TeamRunCompleted`:
  - `stored=False content_len=0`
- This means that, by the time `TeamRunCompleted` fires, the Team Leader’s `ContextState.current_msg.content` is empty, so the final Info Team message is **not stored**.

**Final stored messages after Phase 1 (abridged JSON):**
```json
[
  {
    "role": "user",
    "content": "What is the weather in Paris and the USD to EUR exchange rate?",
    "agent_name": null
  },
  {
    "role": "assistant",
    "tool_calls": [{ "name": "get_current_weather", ... }],
    "agent_name": "Weather Agent"
  },
  {
    "role": "tool",
    "content": "The weather in Paris, France is sunny and 22°C",
    "agent_name": "Weather Agent"
  },
  {
    "role": "assistant",
    "tool_calls": [{ "name": "get_exchange_rate", ... }],
    "agent_name": "Finance Agent"
  },
  {
    "role": "tool",
    "content": "1 USD = 0.85 EUR",
    "agent_name": "Finance Agent"
  },
  {
    "role": "assistant",
    "content": "The current USD to EUR exchange rate is:\n\n**1 USD = 0.85 EUR**\n\nThis means that 1 US Dollar is currently worth 0.85 Euros.",
    "agent_name": "Finance Agent"
  },
  {
    "role": "assistant",
    "content": "Here's the current weather information for Paris, France:\n\n**Temperature:** 22°C ...",
    "agent_name": "Weather Agent"
  }
]
```

**What is missing:**
- There is **no message** with `agent_name: "Info Team"` representing the Team Leader’s synthesized response, despite the event progression showing `TeamRunContent` and `TeamRunCompleted`.

### Concrete Example: Phase 2 (Continuation)

Phase 2 repeats the same pattern when the user asks:
```text
USER: Could you summarize that for me?
```

The console again shows:
```text
👤 MEMBER STARTED: Weather Agent | run_id: ...
👤 MEMBER STARTED: Finance Agent | run_id: ...

[MEMBER TOOL CALL (Weather Agent): get_current_weather]
...
[MEMBER RUN COMPLETED (Weather Agent)]
✓ Stored Final Message (Weather Agent)

[MEMBER TOOL CALL (Finance Agent): get_exchange_rate]
...
[MEMBER RUN COMPLETED (Finance Agent)]
✓ Stored Final Message (Finance Agent)

[DEBUG] Hit TeamRunCompleted catch-all. run_id=...
[DEBUG] Found context for TeamRunCompleted. stored=False content_len=0

[TEAM RUN COMPLETED (Info Team)]
```

**Event progression (Continuation):**
```text
Event Progression (Continuation):
('Info Team', 'TeamRunStarted'): 1
('Info Team', 'TeamRunContent'): 8
('Info Team', 'TeamToolCallStarted'): 2
('Weather Agent', 'RunStarted'): 1
('Finance Agent', 'RunStarted'): 1
('Finance Agent', 'RunContent'): 1
('Finance Agent', 'ToolCallStarted'): 1
('Finance Agent', 'ToolCallCompleted'): 1
('Finance Agent', 'RunContent'): 9
('Finance Agent', 'RunContentCompleted'): 1
('Finance Agent', 'RunCompleted'): 1
('Weather Agent', 'RunContent'): 12
('Weather Agent', 'RunContentCompleted'): 1
('Weather Agent', 'RunCompleted'): 1
('Info Team', 'TeamToolCallCompleted'): 2
('Info Team', 'TeamRunContent'): 28
('Info Team', 'TeamRunContentCompleted'): 1
('Info Team', 'TeamRunCompleted'): 1
```

**Resulting stored messages (abridged) again show:**
- User question
- Delegation tool calls (recorded as assistant/tool role messages for Weather/Finance)
- Member final answers
- **Still no Info Team message**

### How This Causes the Loop

After Phase 1 and Phase 2, the restored history used for Phase 3 (or any subsequent run) looks like:

```json
[
  { "role": "user", "content": "What is the weather in Paris and the USD to EUR exchange rate?" },
  { "role": "assistant", "agent_name": "Weather Agent", "tool_calls": [ "get_current_weather" ] },
  { "role": "tool", "agent_name": "Weather Agent", "content": "The weather in Paris, France is sunny and 22°C" },
  { "role": "assistant", "agent_name": "Finance Agent", "tool_calls": [ "get_exchange_rate" ] },
  { "role": "tool", "agent_name": "Finance Agent", "content": "1 USD = 0.85 EUR" },
  { "role": "assistant", "agent_name": "Finance Agent", "content": "The current USD to EUR exchange rate is ..." },
  { "role": "assistant", "agent_name": "Weather Agent", "content": "Here's the current weather information for Paris, France ..." },
  { "role": "user", "content": "Could you summarize that for me?" },
  { "role": "assistant", "agent_name": "Weather Agent", "...": "another get_current_weather cycle" },
  { "role": "assistant", "agent_name": "Finance Agent", "...": "another get_exchange_rate cycle" }
]
```

From the Team Leader’s perspective:
- It **never sees itself** as having produced a final, user-facing answer.
- It only sees:
  - User questions
  - Its own delegation tool calls (indirectly via member activity)
  - Member-level answers
- So each new run it behaves as if:
  - “User asked weather + FX”
  - “I have no record of *me* answering this”
  - “I should delegate again to be safe”

This is why even a stronger prompt (“If there is no need to delegate, answer the question yourself.” and “If the user provides an image, DO NOT DELEGATE.”) does **not** stop delegation: the missing Info Team messages in storage convince the model that the task remains unresolved.

### Summary of the Technical Bug

- **Events prove** the Team Leader is active:
  - `TeamRunStarted`, `TeamToolCallStarted/Completed`, `TeamRunContent`, `TeamRunCompleted`.
- **Our state tracking (`ContextState`) never accumulates text for the team:**
  - At `TeamRunCompleted`, `content_len` is always 0.
- **JSON history never contains Info Team messages**:
  - Only member (Weather/Finance) messages are stored.
- **Consequences:**
  - The team behaves as though it never answered the user directly.
  - Every subsequent run re-delegates, re-fetches weather/FX, and re-streams essentially the same answers.

A future fix should:
- Inspect the raw `TeamRunContent` event objects (e.g., via `event.model_dump()` or `vars(event)`) to determine where the textual content actually lives for team runs (it may differ from the agent `RunContentEvent` structure).
- Adjust the `RunContentEvent` handling branch in `teams_demo.py` to read from the correct attribute(s) and accumulate the Info Team’s final answer into `ContextState.current_msg`, ensuring it is stored on `TeamRunCompleted`.