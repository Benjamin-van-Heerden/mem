# Work Log - BaseAgent Generic Request + Tool Config Refactor (Team-Ready)

## Spec File: `agent_rules/spec/s_20251129_benjamin_van_heerden__agent_infrastructure_migration.md`

## Overarching Goals
- Refine the agent foundation to support both standalone Agents and Team member Agents without leaking request/session/user context into Team members.
- Establish a simple, enforceable pattern for tool currying that works in both standalone and Team settings using validated “tool config” models.
- Improve type-safety and developer experience via generics and typed Pydantic models while keeping the framework easy to extend.

## What Was Accomplished

### 1) Reworked `BaseAgent` around generic Request + Tool Config contracts
- Implemented a `BaseAgent` API that is classmethod-driven and generic over:
  - `ReqT` (bound to `BaseAIRequest`)
  - `ToolT` (bound to `pydantic.BaseModel`)
- Established two explicit validation pathways:
  - `validate_request(raw_request) -> ReqT`
  - `validate_tool_config(raw_tool_config) -> ToolT`
- Added a shared `_format_validation_error(...)` helper to produce clean, readable error messages from Pydantic validation failures.

### 2) Made tool currying Team-compatible via “tool config” input
- Defined the core mechanism for Teams to pass only a minimal, validated config to member Agents:
  - Team creates `raw_tool_config` (derived from Team request interpretation)
  - Member validates and curries its tools from `ToolT`
- Introduced two explicit agent construction entry points:
  - `get_agent(raw_request, raw_tool_config)` for standalone usage (wires Agno DB session persistence using `get_agno_mysql_db()` and uses `request.session_id`/`request.user_id`)
  - `get_agent_for_team(raw_tool_config)` for Team member usage (builds a stateless member agent without DB/session/user wiring)

### 3) Clarified the “Team owns persistence” boundary
- Ensured that Team member agents constructed via `get_agent_for_team(...)` do not receive DB/session/user wiring.
- Reinforced the intended boundary:
  - Team is responsible for interpreting the request and binding/currying required parameters for member tools.
  - Members validate tool config and expose tools derived from that config.

## Key Files Affected
- `src/agents/core.py`
  - Refactored `BaseAIRequest`
  - Added `_format_validation_error(...)`
  - Introduced generic `BaseAgent[ReqT, ToolT]` with:
    - `request_model()` / `tool_config_model()`
    - `validate_request(...)` / `validate_tool_config(...)`
    - `get_agent(...)` / `get_agent_for_team(...)`
  - Standardized naming to “tool config” for the Team/member interface.

## Errors and Barriers
- Type-checker friction was encountered during experimentation around generics and overrides when subclasses were not parameterized correctly.
- Resolved by explicitly using `BaseAgent[ConcreteReq, ConcreteToolConfig]` in subclasses and keeping method signatures consistent with the abstract base definitions.

## What Comes Next
- Update `src/agents/config.py` to align with the new `BaseAgent` API (registry + construction helpers must call the new methods and provide tool configs explicitly).
- Update `src/api/routes/agents/router.py` to:
  - inject `user_id` into request payloads before validation,
  - provide/derive `raw_tool_config` for standalone agent creation,
  - use `get_agent(...)` for standalone runs and preserve SSE streaming behavior.
- Review and simplify `src/api/routes/agents/models.py` to remove unused models and keep only what is needed for the current API surface.
- Update `src/agents/example/agent.py` to implement:
  - `request_model()`, `tool_config_model()`, `tools(...)`, `name()`, `role()`, `instructions()`
  - and ensure it works in both standalone and Team member contexts.
