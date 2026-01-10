# Work Log - Initial Agent SSE API + Agent Registry Refactor (BaseAIRequest)

## Spec File: `agent_rules/spec/s_20251129_benjamin_van_heerden__agent_infrastructure_migration.md`

## Overarching Goals
- Establish the first “small unit” of the new chat system: a single Agent (not a Team yet) reachable through a dedicated `/agents` API router.
- Implement a minimal SSE streaming endpoint that can be validated via curl before building any UI.
- Move toward an agent-owned request validation pattern so that each agent can define and validate its own input schema.
- Keep persistence aligned with the current direction: Agno DB-backed sessions using shared MySQL tables.

## What Was Accomplished

### 1) Added shared auth dependencies for API key + WCU JWT
- Implemented reusable auth dependencies in `src/api/routes/utils/auth.py`:
  - API key auth via `X-API-Key` header
  - WCU JWT auth via `Authorization: Bearer <jwt>` (validated via WCU API)
  - `auth_any_dependency` to accept either auth mode and return a small auth context (`user_id`, auth type)

### 2) Implemented initial Agent API router with session creation + SSE streaming run
- Created `src/api/routes/agents/router.py` and wired it into `src/api/main.py`.
- Endpoints implemented:
  - `GET /agents` → list available agents
  - `POST /agents/{agent_name}/sessions` → generate server-side UUIDv7 session id (via `uuid_extensions.uuid7str`) after validating “create session” payload for that agent
  - `POST /agents/{agent_name}/run` → supports:
    - `stream=false` → non-streaming JSON response
    - `stream=true` → SSE (`text/event-stream`) streaming of deltas + tool events + done/error markers
- SSE event format is “real SSE frames”:
  - `event: meta|delta|tool_start|tool_end|done|error`
  - `data: <json payload>`

### 3) Introduced shared MySQLDb helper for canonical Agno session storage
- Added `src/utils/agno/db.py` to construct a shared `MySQLDb` using the canonical shared tables:
  - `ai_sessions`, `ai_metrics`, `ai_memories`
- Refactored to a simple cached singleton using `functools.lru_cache(maxsize=1)`.

### 4) Refactored agent registration toward a TaskConfig-style list and agent-owned request validation
- Implemented `src/agents/core.py` with:
  - `BaseAIRequest` (common API/agent request primitive: `session_id`, `message`, `stream`)
  - `BaseAgent` (agent interface: `NAME`, `RequestModel`, `validate_request`, `validate_create_session`, `validate_run`, `build_agent`)
- Refactored `src/agents/config.py` to a TaskConfig-like list (`AgentConfig.AGENTS = [ExampleAgent, ...]`) and central helpers:
  - `list_agent_names()`
  - `get_agent_class()`
  - `validate_create_session_request()`
  - `validate_run_request()`
  - `build_agno_agent()` (builds an Agno Agent after agent-level payload validation)

### 5) Implemented a minimal Example agent for validation + streaming
- Updated `src/agents/example/agent.py` to define:
  - `ExampleAIRequest` inheriting `BaseAIRequest` with basic field validation
  - `ExampleAgent` implementing `BaseAgent` and returning an Agno `Agent` configured with:
    - shared `MySQLDb`
    - `add_history_to_context=True`, `num_history_runs=5`
    - `store_media=False` (avoid DB JSON bytes issues)

## Key Files Affected
- `src/api/routes/utils/auth.py`
  - Added API key + JWT auth dependencies and shared auth context.
- `src/utils/agno/db.py`
  - Added shared `MySQLDb` helper with `lru_cache`.
- `src/agents/core.py`
  - Added `BaseAIRequest` + `BaseAgent` primitives for agent-owned request validation and construction.
- `src/agents/config.py`
  - Switched to a list-based registry (TaskConfig-style) and helper functions for the API router.
- `src/agents/example/agent.py`
  - Implemented `ExampleAgent` and request schema.
- `src/api/routes/agents/models.py`
  - Added SSE event models and non-streaming run response model.
- `src/api/routes/agents/router.py`
  - Implemented `/agents` routes and SSE/non-streaming run behavior.
- `src/api/main.py`
  - Included the agents router.

## Errors and Barriers
- Static diagnostics tooling reported parse/type issues around:
  - generator/SSE streaming endpoint structure
  - union return types / response model inference
  - generic typing for `BaseAgent`
- Mitigations applied:
  - Simplified `BaseAgent` to non-generic base class (kept agent-owned request model pattern)
  - Disabled FastAPI response-model inference on the streaming endpoint (`response_model=None`)
  - Confirmed the app can import and routes are registered via a quick `uv run python -c ...` check.

## What Comes Next
- Validate the agent API via curl:
  - Create a session with an agent-specific payload (message/session omitted)
  - Run with `stream=false` and confirm JSON response
  - Run with `stream=true` and confirm SSE frames and incremental deltas
- Add session listing + session history endpoints using Agno `MySQLDb.get_sessions(...)` and `get_session(...)` so the frontend can show history.
- Begin cloning/rebuilding the chat UI component (new version of AgentInterface) against this API once curl verification is solid.
- Once Agents are stable, expand the same patterns to Teams under `src/api/routes/teams`.