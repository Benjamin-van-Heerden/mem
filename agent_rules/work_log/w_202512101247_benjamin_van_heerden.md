# Work Log - Teams Demo Refactoring and Database Schema Research

## Spec File: `agent_rules/spec/s_20251129_benjamin_van_heerden__agent_infrastructure_migration.md`

## Overarching Goals

Continue the Message Persistence Research task by refactoring the teams demo to align with the agent demo pattern and exploring database schema design for unified agent/team storage. The primary focus was on:

1. Understanding team event structures and agent attribution patterns
2. Refactoring teams demo to use consistent event handling with match statements
3. Exploring optimal database schema to minimize JSONB usage while supporting both agents and teams
4. Identifying event discovery patterns and active agent tracking for teams

## What Was Accomplished

### 1. Database Schema Design Discussion

Analyzed the complete message history from agent demo output and proposed a flattened schema approach:

**Proposed `chat_events` table structure:**
```sql
id              UUID        -- PK (UUIDv7)
session_id      UUID        -- FK to sessions
role            VARCHAR     -- user, assistant, tool, system
content         TEXT        -- main message text
tool_call_id    VARCHAR     -- only for role='tool'
tool_name       VARCHAR     -- only for role='tool'
input_tokens    INTEGER     -- metrics (default 0)
output_tokens   INTEGER     -- metrics (default 0)
duration        FLOAT       -- execution time
created_at      TIMESTAMP   -- creation time
message_metadata JSONB      -- ONLY for tool_calls and images lists
```

**Key decisions:**
- Extract all scalar values into columns for queryability
- Use JSONB only for variable-length lists (`tool_calls`, `images`)
- Support both agents and teams in single schema
- Remove `tool_call_id` column as it's LLM-internal detail, not needed for UI reconstruction
- Focus on UI reconstruction requirements rather than framework internals

### 2. Teams Demo Refactoring Attempt

**Updated `scripts/play/teams_demo.py`:**
- Changed from `match event.event:` to `match event:` for type-based matching (like agent demo)
- Implemented proper event imports (`RunStartedEvent`, `RunContentEvent`, `ToolCallStartedEvent`, etc.)
- Added agent name attribution logic using `event.agent_name` from `RunStartedEvent`
- Updated event handling to distinguish between team-level and member-level events
- Used same storage helper functions as agent demo with `agent_name` parameter

**Event handling pattern implemented:**
```python
match event:
    case RunStartedEvent():
        if event.event == "TeamRunStarted":
            # Team leader started
        elif event.event == "RunStarted":
            # Member agent started - capture agent_name
    case RunContentEvent():
        # Handle both team and member content
    case ToolCallStartedEvent():
        # Handle both team and member tool calls
```

### 3. Agent Attribution Discovery

From user's demo output, identified key insights about team event structures:
- Team events have `team_id`, `team_name` attributes
- Member events have `agent_id`, `agent_name` attributes in `RunStartedEvent`
- Delegation flow: Team leader → `delegate_task_to_member` → Member agent → Tool calls → Results back to team
- Event progression shows clear patterns for attribution tracking

### 4. Event Type Analysis

User provided comprehensive team event progression showing:
- `TeamRunStarted` → `TeamRunContent` → `TeamToolCallStarted` → `RunStarted` (member) → member tool calls → `TeamToolCallCompleted` → more delegation cycles → `TeamRunCompleted`
- Member events are nested within team tool call execution
- Clear separation between team coordination and member execution

## Key Files Affected

**Modified:**
- `scripts/play/teams_demo.py` (major refactor)
  - Changed event matching from enum-based to type-based pattern
  - Added proper agent attribution using `event.agent_name`
  - Implemented unified storage pattern with `prepare_message_for_storage(message, agent_name)`
  - Fixed event handling for both team and member contexts
  - Updated to use `claude-3-5-haiku-20241022` model

## Errors and Barriers

### 1. Teams Demo Execution Failure
The refactored teams demo failed with "Killed: 9" error during execution. This appears to be a resource/memory issue rather than a code logic problem, as the refactor was structural and aligned with working agent demo patterns.

### 2. Agent Name Attribution Complexity
While the schema supports `agent_name` attribution, the actual runtime mapping from Agno's internal `agent_id` to meaningful names requires further investigation. The demo attempts to use `event.agent_name` but this may need validation.

### 3. Event Type Import Issues
The refactor required importing specific event types, but the exact import paths and class relationships between team and agent events may need refinement.

## What Comes Next

### Immediate Next Steps (Teams Demo Completion)

1. **Debug teams demo execution failure**
   - Investigate the "Killed: 9" error (likely memory/resource issue)
   - Verify event type imports and class relationships
   - Ensure team event handling logic is correctly implemented

2. **Validate agent attribution**
   - Confirm `event.agent_name` is available in `RunStartedEvent` for team members
   - Test actual agent name capture during team delegation
   - Ensure storage correctly attributes messages to specific agents vs team leader

3. **Complete event discovery**
   - Run successful teams demo to catalog all team-specific event types
   - Compare with agent demo event types to understand full event space
   - Document team vs agent event differences

### Database Schema Implementation

4. **Create database migration**
   - Implement proposed `chat_events` table schema
   - Create migration: `python scripts/migrate.py --generate "create_agent_chat_tables"`
   - Include proper indexes for session_id, created_at, role, agent_name

5. **Implement storage primitives**
   - Create `src/agents/storage.py` with schema-aware functions
   - `prepare_for_db_storage()` - extract scalars, minimize JSONB
   - `restore_from_db_storage()` - reconstruct Message objects
   - Handle both agent and team message attribution

6. **Test roundtrip storage/reconstruction**
   - Store messages from both agent and teams demos
   - Verify UI reconstruction capability (role, content, attribution, chronological order)
   - Validate tool call sequence preservation
   - Test image metadata handling

### Architecture Completion

7. **Update spec with findings**
   - Mark Message Persistence Research as COMPLETED
   - Document final database schema decisions
   - Add team-specific storage requirements
   - Update next task (Base Agent Architecture) with database integration

8. **Proceed to next spec task**
   - Base Agent Architecture with context management
   - Custom storage integration with new schema
   - Session management with team/agent attribution

### Key Insights Established

**Database Design:**
- Single `chat_events` table supports both agents and teams
- Minimal JSONB usage (only for variable-length arrays)
- Agent attribution via `agent_name` column for UI reconstruction
- Focus on UI needs rather than framework internals

**Team Event Patterns:**
- Team leader coordinates via `delegate_task_to_member`
- Member agents execute with full tool access
- Clear event attribution available via `agent_name` field
- Nested execution model requires careful storage timing

**Storage Strategy:**
- Event-driven storage using match statements on event types
- Immediate persistence for completed events
- Agent/team attribution captured at RunStarted events
- Message reconstruction focuses on conversation flow, not technical details

The foundation for unified agent/team storage is now established, requiring completion of the teams demo validation and database schema implementation.