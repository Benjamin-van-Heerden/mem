# Agent Infrastructure Migration - Systematic Eval+Delete Approach

%% Status: In Progress %%

## Description

This spec outlines the systematic migration of existing agent functionality from the legacy `__wcu_agent_server` repository into the new clean `wcu_ai_server` codebase. The migration follows an "eval+delete" loop methodology, where existing code is evaluated, simplified according to KISS principles, and re-implemented without over-engineering or unnecessary complexity.

The legacy implementation contains functional AI agent infrastructure that has been deployed and tested, but requires careful evaluation to ensure only essential, well-designed components are migrated. This is not a direct copy-paste operation, but rather a thoughtful reimplementation that eliminates technical debt and unnecessary abstractions.

**Core Components to Migrate:**
- FastAPI streaming API for Agents and Teams
- Meetings Agent with its tools and configuration
- WCU API helper functions for authentication and API interaction (COMPLETED - migrated to `src/utils/weconnectu/` and `src/utils/models/wcu_*`)
- Frontend `<AgentInterface />` component (Nuxt/Vue)
- Critical integration and workflow tests (dramatically reduced scope from original)

**Key Principles:**
- Apply KISS principle ruthlessly - eliminate complexity at every opportunity
- Evaluate each component individually before migration
- Simplify and reduce abstractions where possible
- Dramatically reduce testing scope - focus on workflow tests, not exhaustive edge cases
- Canonical persistence via Agno DB-backed MySQL sessions (`MySQLDb`) for Agents and Teams
- Any additional structured tables are optional projections for analytics/costing (not canonical conversation state)
- Code should be self-explanatory without comments

**Architecture Decisions:**
- **Canonical Persistence**: Use Agno DB-backed MySQL sessions (`MySQLDb`) as the source of truth for Agents and Teams (session continuity and history injection via `add_history_to_context=True` + `num_history_runs`).
- **Streaming**: Use `agent.run(..., stream=True, stream_events=True)` and surface raw Agno events via SSE; avoid bespoke wrapper types or state machines.
- **Session Management**: Use server-generated session identifiers and Agno session storage; avoid separate diskcache session managers for canonical conversation state.
- **Simplified BaseAgent**: Agents are auth-agnostic; routes handle auth. Agents define request + config schemas using primitives (`BaseAIRequest`, `BaseAgentConfig`) and support both standalone and Team usage.
- **Team-Ready Architecture**: Teams own request interpretation and persistence; Team members are constructed without DB/session/user wiring and receive only validated config for currying tools.

**Implementation Order:**

The tasks should be implemented in the following sequence to maintain clean dependencies:

1. **Message Persistence Research** - CURRENT: Understanding event structures and storage patterns
2. **Base Agent Architecture and Context Management** - Foundation for all agents
3. **Custom Storage and Session Management** - Database schema and CRUD operations
4. **Agent Streaming Wrapper** - Core streaming logic with interruption handling
5. **Meetings Agent Migration** - First concrete agent implementation (COMPLETED)
6. **FastAPI Agent Routes** - API layer that ties everything together
7. **Frontend Component Migration** - UI integration with new backend
8. **Essential Tests Migration** - Workflow tests validating end-to-end functionality

Note: WCU API Integration Utilities task is marked COMPLETED.

## Tasks

### Task: Message Persistence Research (COMPLETED)

**Status:** COMPLETED - Pivoting to Agno-managed MySQL session storage as the canonical persistence layer; `teams_demo3.py` proves DB-backed recall and Phase 2 “repeat that” does not re-delegate when using the same `session_id`.

#### Key Findings
- Agno’s `add_history_to_context=True` works reliably with `MySQLDb` for restoring conversation context.
- Custom persistence logic (reconstructing event streams) is fragile and unnecessary for the core chat loop.
- `store_media=False` is required to avoid JSON serialization issues with image bytes in sessions.
- Teams handle delegation state correctly when sessions are restored.

#### Remaining Work
- None for this task. Persistence strategy is set (ADR 2025-12-15).

---

### Task: Base Agent Architecture and Context Management (COMPLETED)

- [x] Remove BaseAgent authentication and playground mode (routes handle auth and env concerns)
- [x] Implement `BaseAgent[ReqT, ConfigT]` with explicit request + config contracts (using `BaseAIRequest`, `BaseAgentConfig`)
- [x] Define a simple config mechanism that supports tool currying and model selection (`ModelPicker`)
- [x] Implement `get_agent(raw_request, raw_config)` for standalone usage (DB/session/user wiring lives here)
- [x] Implement `get_agent_for_team(raw_config)` for Team member usage (no DB/session/user wiring)
- [x] Ensure agents are Team-ready by design: Teams interpret requests and provide configs; members validate configs and curry tools

#### Implementation Details

**Generic BaseAgent Pattern:**
```python
class BaseAgent(ABC, Generic[ReqT, ConfigT]):
    @classmethod
    @abstractmethod
    def request_model(cls) -> type[ReqT]: ...

    @classmethod
    @abstractmethod
    def config_model(cls) -> type[ConfigT]: ...

    @classmethod
    @abstractmethod
    def tools(cls, config: ConfigT) -> list[Callable]: ...

    @classmethod
    def get_agent(cls, raw_request: dict[str, Any], raw_config: dict[str, Any]) -> Agent:
        request = cls.validate_request(raw_request)
        config = cls.validate_config(raw_config)
        
        # Model selection via ModelPicker
        model = ModelPicker(model=config.model).llm_model
        
        return Agent(
            model=model,
            tools=cls.tools(config),
            db=get_agno_mysql_db(), 
            session_id=request.session_id, 
            user_id=request.user_id,
            add_history_to_context=True
        )
```

**Key Changes from Legacy:**
- No authentication in BaseAgent (routes handle this)
- Teams own persistence and request interpretation
- Standalone Agents wire DB/session/user; Team members do not
- Tool currying is based on validated `BaseAgentConfig` rather than passing request context to Team members
- `ModelPicker` utility handles provider selection and API keys cleanly

---

### Task: FastAPI Agent Routes (COMPLETED)

- [x] Design simplified route structure using primitives (`BaseAIRequest`, `BaseAgentConfig`)
- [x] Implement POST /agents/{agent_name}/run endpoint with streaming (SSE)
- [x] Implement POST /agents/{agent_name}/sessions endpoint for session creation
- [x] Implement GET /agents/{agent_name} endpoint for schema discovery
- [x] Handle authentication at router level (JWT validation dependency)
- [x] Integrate with ModelPicker and BaseAgent architecture
- [x] Implement GET /agents/{agent_name}/sessions endpoint for session listing
- [x] Implement GET /agents/{agent_name}/sessions/{session_id} endpoint for session restoration (history)
- [x] Implement DELETE /agents/{agent_name}/sessions/{session_id} endpoint for session deletion

#### Implementation Details

**Route Architecture:**

Create `src/api/routes/agents/router.py` with:

**POST /agents/{agent_name}/run** - Main chat endpoint with streaming:
```python
async def run_agent(
    agent_name: str,
    request: RouterAIRequest = Body(...),
    config: RouterAgentConfig = Body(default_factory=RouterAgentConfig),
    stream: bool = Query(default=True),
    auth: AuthContext = Depends(auth_any_dependency),
):
    # 1. Validation handled by Pydantic + primitives
    # 2. Build agent with request + config
    agent = build_agent(agent_name, raw_request=request.model_dump(), raw_config=config.model_dump())
    
    # 3. Stream events via SSE (AsyncGenerator)
    # Events: Raw Agno events (RunResponse, ToolCallStarted, etc.)
```

**Response Models:**
- `AgentRunResponse` (non-streaming): Returns `RunOutput` directly or simple wrapper.
- Streaming: Raw Agno events serialized as SSE `event: type\ndata: json\n\n`.

**Auth:**
- Uses `APIRouter(dependencies=[Depends(auth_any_dependency)])`
- Supports both `Authorization: Bearer <jwt>` and `X-API-Key`

**Additional Endpoints:**
- `GET /agents/{agent_name}/sessions` (List sessions)
- `GET /agents/{agent_name}/sessions/{session_id}` (Restore session)
- `DELETE /agents/{agent_name}/sessions/{session_id}` (Delete session)

#### Testing Outline
- [x] Verify agent listing (`GET /agents`)
- [x] Verify schema discovery (`GET /agents/{agent_name}`)
- [x] Verify session creation (`POST /sessions`)
- [x] Verify run with streaming (`stream=true`) + persistence recall
- [x] Verify run without streaming (`stream=false`) + tool calls in response
- [x] Verify session listing and history endpoints
- [x] Verify session deletion endpoint

---

### Task: Teams Backend Infrastructure (COMPLETED)

- [x] Implement `BaseTeam` primitive with member hydration logic
- [x] Implement `TeamConfig` registry and factory
- [x] Implement `src/api/routes/teams` endpoints (CRUD + Streaming)
- [x] Integrate Teams with Agno MySQL persistence (canonical)

#### Implementation Details
- `BaseTeam` mirrors `BaseAgent` but handles member instantiation via `build_agent_for_team` to ensure safety.
- Routes mirror `/agents` structure but point to `/teams`.
- Uses `BaseTeamConfig` to route configuration to specific team members.

---

### Task: Custom Storage and Session Management (DEPRECATED)

**Status:** DEPRECATED - Replaced by `src/utils/agno/db.py` (canonical Agno MySQL persistence)

---

### Task: Agent Streaming Wrapper (COMPLETED)

**Status:** COMPLETED - Logic integrated directly into `router.py` using `agent.arun(stream=True)` and standard SSE generator pattern. No separate wrapper class needed.

---

### Task: Meetings Agent Migration (COMPLETED)

- [x] Create `MeetingAgentConfig` extending `BaseAgentConfig` with WCU-specific fields
- [x] Implement factory functions for tools with proper currying pattern
- [x] Create `MeetingAgent` class following `BaseAgent[ReqT, ConfigT]` pattern
- [x] Register agent in `AgentConfig.AGENTS`
- [x] Add agent to `ExampleTeam` for team composition testing

#### Implementation Details

**Config Model (`src/agents/meetings/agent.py`):**
```python
class MeetingAgentConfig(BaseAgentConfig):
    wcu_user_jwt: str
    reseller_uid: str
    available_communities: list[str] = ["*"]
```

**Tool Factory Pattern (`src/agents/meetings/tools.py`):**
Each tool is implemented as a factory function that returns a curried callable:
- `make_get_communities(reseller_uid, wcu_user_jwt, available_communities)` → `get_communities(community_name)`
- `make_get_all_communities(...)` → `get_all_communities()`
- `make_get_meetings(...)` → `get_meetings(community_uid, meeting_type, from_date, to_date, limit)`
- `make_get_meeting_detail(...)` → `get_meeting_detail(meeting_uid)`

This pattern ensures:
- Config params are curried at agent creation time
- Docstrings on inner functions are clean (only runtime params)
- Agno sees correct function signatures for LLM tool calling

> Relevant existing files: `src/utils/weconnectu/meetings.py`, `src/utils/weconnectu/communities.py`
> New files: `src/agents/meetings/agent.py`, `src/agents/meetings/tools.py`

---

### Task: Frontend Component Migration (IN PROGRESS)

- [x] Rebuild `<AgentInterface />` using the new `/agents` API
- [x] Implement progressive rendering for SSE deltas
- [x] Implement tool usage display (consuming `ToolCallStarted`/`ToolCallCompleted` events)
- [x] Implement session history sidebar with absolute overlay and blurred backdrop
- [x] Implement session restoration logic (client-side block reconstruction from raw message list)
- [x] Refactor into shared components (`TitleBar`, `SessionHistory`, `ChatInput`, etc.) to support Teams
- [x] Implement session initialization on mount with loading/error states
- [x] Create reusable `Starting.vue` and `Welcome.vue` components
- [x] Implement grouped session history (Today, Yesterday, This Week, Earlier)
- [ ] Implement fully functional `<TeamInterface />` with delegation support
- [ ] Implement run cancellation (Stop button) - **BLOCKED** pending Agno bug fix

#### Session Initialization Flow

On component mount:
1. Show "Starting chat..." with spinner via `AiCommonStarting`
2. Create session and fetch history via `initSession()`
3. On success: show `AiCommonWelcome` with suggested prompts
4. On error: show "Failed to connect to AI Platform" with retry button

#### Shared Components Created

- `AiCommonStarting.vue` - Loading/error state with optional retry button
- `AiCommonWelcome.vue` - Welcome screen with configurable icon, title, and prompts
- `SessionHistory.vue` - Now groups sessions by time period (Today, Yesterday, This Week, Earlier)

#### Run Cancellation Status

**Implementation Complete:**
- Backend: `POST /{agent_name}/sessions/{session_id}/cancel?run_id={run_id}` route using `Agent.cancel_run(run_id)`
- Frontend: Stop button UI commented out, `currentRunId` tracking and `cancelRun()` method remain in `useAiChat.js`

**Blocked by Agno Bug:**
Cancelled runs persist `content` but not `messages`, making them invisible to history injection. Bug report filed with Agno maintainers. Until resolved, the stop functionality will technically work but break conversation continuity (agent won't remember partial responses).

---

### Task: Database Schema Design and Migration System

- [ ] Ensure `ai_sessions` and `ai_metrics` tables are properly migrated/managed
- [ ] Verify `script/migrate.py` handles the new tables if not automatically managed by Agno

---

### Task: Essential Tests Migration

- [ ] Port critical workflow tests to `tests/`
- [ ] Ensure API tests cover auth, validation, streaming, and persistence scenarios

# Final Review

- [ ] Ensure all migrated components follow KISS principles
- [ ] Verify no critical functionality lost from legacy system
- [ ] Confirm documentation/specs are up to date