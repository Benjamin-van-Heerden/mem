# Work Log - Pivot to Agno-Managed MySQL Sessions (Teams) + DB-Backed Demo Validation

## Spec File: `agent_rules/spec/s_20251129_benjamin_van_heerden__agent_infrastructure_migration.md`

## Overarching Goals

- Resolve correctness issues in custom team persistence (ordering, segmentation, concurrency) by relying on Agno’s canonical session storage.
- Prove that DB-backed Team sessions provide resumability and recall across separate process runs.
- Establish a clean “known-good” demo path that aligns with how the rest of the project will operate (MySQL-backed sessions, history injection via `add_history_to_context`).

## What Was Accomplished

### 1) Identified a Persistence/Ordering Risk in the Custom Event → Message State Machine
- Observed that while `teams_demo2.py` could store key messages, the persisted ordering could diverge from what the user actually experienced during streaming.
- This mismatch becomes increasingly risky under:
  - multi-step reasoning (multiple assistant segments),
  - interleaved tool calls,
  - concurrent member execution.

This reinforced that building a bespoke state machine to reconstruct a “perfect” timeline from streamed events is non-trivial and error-prone, especially given that Agno already has a canonical “sessions + runs” persistence model.

### 2) Pivot Decision: Use Agno-Managed Storage and Sessions as the Canonical Source of Truth
Decision made to move away from the custom persistence approach (structured `chat_events` + custom reconstruction) as the primary persistence mechanism for Teams, and instead rely on Agno-managed session storage (backed by MySQL) for:
- correct ordering and session continuity,
- correct history injection semantics,
- resumability across invocations.

The earlier critical Teams insight remains relevant but becomes less central operationally once we rely on Agno sessions:
- Do not pass `List[Message]` to `Team.arun()` expecting it to behave as structured chat history.
- Prefer Agno DB-backed history (`add_history_to_context=True`) over manual history injection.
- Manual history injection via `team.additional_input` remains a viable fallback if needed.

### 3) Added SQLAlchemy Support for Agno DB Backends
- Agno database backends require SQLAlchemy; this was missing initially and prevented using `MySQLDb`.
- SQLAlchemy was added to the project dependencies.

### 4) Implemented and Validated a Minimal DB-Backed Recall Demo (teams_demo3.py)
- Created a new demo script `scripts/random/teams_demo3.py` to validate Agno-managed persistence using `MySQLDb`.
- Proved persistence across separate runs using a fixed `session_id`:
  - Run 1 (“set”): store the secret word (“bird”) and confirm it is persisted.
  - Run 2 (“recall”): new process, same `session_id`, successful recall of “bird”.
- Printed Agno session chat history via `team.get_session(session_id).get_chat_history()` to confirm the session stored expected user/assistant turns.

### 5) Replicated teams_demo2 “Phase 1 / Phase 2” Behavior with DB-Backed Sessions
- Updated `teams_demo3.py` to mirror the essential behavior from `teams_demo2.py` but using Agno’s DB persistence:
  - Phase 1: Delegation to Weather + Finance members (tool usage)
  - Phase 2: “repeat that” follow-up should NOT re-delegate
- Verified behavior:
  - Phase 1 delegates as expected.
  - Phase 2 delegates zero times (uses stored history) when run with the same `session_id`.
- Added session replay printing to inspect:
  - chat history (user+assistant),
  - full messages (including tool + member messages).

### 6) Standardized MySQL Connection URL Usage via Project Helper
- The demo uses the project’s `src/utils/db/connection.py::db_url()` to produce a `mysql+mysqlconnector://...` SQLAlchemy URL, matching the project’s MySQL connector preference.

## Key Files Affected

- `scripts/random/teams_demo3.py`
  - New DB-backed Teams demo using Agno `MySQLDb` session storage.
  - Includes:
    - recall proof across runs (initial minimal validation),
    - Phase 1/Phase 2 delegation + “repeat” behavior demonstration,
    - event streaming output and basic counting,
    - session replay printing from Agno session object.

- `pyproject.toml`
  - Added `sqlalchemy` dependency to support Agno DB backends.

- `src/utils/db/connection.py`
  - `db_url()` provides a `mysql+mysqlconnector://...` SQLAlchemy URL for consistency across the project.

## Errors and Barriers

- Initial blocker: Agno DB backend imports failed due to missing SQLAlchemy.
  - Resolved by adding SQLAlchemy to project dependencies.

## What Comes Next

1) **Continue building on the DB-backed approach**
   - Treat Agno session storage as the canonical persistence layer for Teams and Agents.
   - Use `session_id`/`user_id` consistently across runs to ensure resumability.

2) **Expand teams_demo3 parity with teams_demo2 advanced scenarios (but DB-backed)**
   - Re-enable/implement:
     - image handling validation (confirm storage behavior, and ensure it aligns with desired constraints),
     - recall test + forgetting test by controlling `num_history_runs` and adding filler runs,
     - delegation/event storage exploration using `store_events=True` + `events_to_skip=[]` where helpful.

3) **Update the spec direction**
   - In `agent_rules/spec/s_20251129_benjamin_van_heerden__agent_infrastructure_migration.md`, adjust “Message Persistence Research” next steps to reflect the pivot:
     - prioritize Agno-managed sessions for correctness and continuity,
     - defer/limit custom storage primitives unless required for analytics or external UX needs beyond Agno session APIs.

4) **Isolate “known-good” primitives for the rest of the project**
   - Standardize on `MySQLDb(db_url=...)` and Agno session access methods (`get_session`, `get_chat_history`, `get_messages`) as the primary interfaces.
   - Keep any custom persistence/event handling only as an optional enhancement if/when requirements emerge that Agno sessions cannot satisfy.