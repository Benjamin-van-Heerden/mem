# Work Log - Teams Demo History Passing Fix (Team.arun List[Message] Flattening)

## Spec File: `agent_rules/spec/s_20251129_benjamin_van_heerden__agent_infrastructure_migration.md`

## Overarching Goals

Establish a correct, repeatable persistence + restoration workflow for Agno Teams such that:
- The Team Leader’s final responses are stored and can be used on subsequent runs.
- Restored conversation history is actually interpreted by Agno as structured chat history (roles preserved).
- The demo provides a reliable basis for designing the eventual database schema and SSE streaming layer.
- We avoid context window bloat by controlling what is restored, but do so in a way that still behaves correctly.

The immediate issue in this session: the team kept re-delegating in Phase 2, even after Phase 1 had completed and we believed we were “passing history”. We needed to prove whether history was being passed properly and fix it deterministically.

## What Was Accomplished

### 1) Identified the Real Bug: `Team.arun()` with `List[Message]` Does NOT Pass Chat History

We inspected Agno internals and found the key behavior:

- `Team.arun(input=...)` calls `_validate_input(input)`.
- If there is no `input_schema` (our case), `_validate_input` returns the input unchanged.
- In the streaming path (`_arun_stream`), Agno builds run messages via `_aget_run_messages(...)`.
- `_aget_run_messages` ultimately builds a *single user message* using `_aget_user_message` / `_get_user_message`.

**Critical internal mechanic:**
When `input_message` is a `list` of `Message`, Agno treats it as “list input” and converts it to a single user message by extracting text content:

- In `Team._get_user_message(...)`:
  - If `input_message` is a list and the first element is a `Message`, it calls `get_text_from_message(input_message)` and sets that as the *content* of a new `Message(role="user", content=...)`.

This means that passing `team.arun([msg1, msg2, msg3])` **does not preserve roles** as discrete chat history messages. Instead, the list becomes a single `role="user"` message containing flattened text.

Practical consequence:
- Even if the restored list contains a prior Team answer (`role="assistant"`), it is not being supplied to the model as an assistant message; it is being merged into a single user-content blob.
- This breaks continuity and strongly encourages the model to “do the workflow again” (delegate again).

This is a concrete mechanical bug in our demo usage, not “LLMs gonna LLM”.

### 2) Implemented Correct History Injection Using `Team.additional_input`

Agno’s `_get_run_messages` / `_aget_run_messages` builds messages in a defined order:
1. system message
2. `additional_input` messages (if present)
3. history from DB (only if `add_history_to_context=True` and a db-backed TeamSession exists)
4. user message (built from `input_message`)

Given we are not using Agno’s DB/session store in this demo, the correct mechanism to inject our reconstructed messages is:

- Put restored history messages into `team.additional_input`
- Call `team.arun()` with ONLY the new user message as the input (a single `Message`), not a list.

This ensures the model sees structured history with correct roles.

Implementation details in our demo:
- Phase 1: call `team.arun(user_msg)` rather than `team.arun([user_msg])`.
- Phase 2:
  - Build `restored_messages` from stored records.
  - Set:
    - `previous_additional_input = team.additional_input`
    - `team.additional_input = restored_messages`
  - Call `team.arun(user_msg_2)` (only the new message)
  - Restore `team.additional_input` afterward in a `finally` block.

This is intentionally model-agnostic: it relies on Agno’s message construction behavior, not a provider-specific payload format.

### 3) Added “from_history” Tagging to Restored Messages

Agno sets `from_history=True` on DB-loaded messages when it loads history through `session.get_messages(...)` with `add_history_to_context=True`. Since we are not using Agno’s DB history, we tag restored messages ourselves.

In `stored_to_message(...)` we now include:
- `"from_history": True` when reconstructing `Message.from_dict(...)`.

This aligns our reconstructed messages with what Agno itself does for “history” messages, and makes it easier to reason about the run input and debug behavior.

### 4) Narrowed the Demo to Phase 1 + Phase 2 Only (Focus on Correctness)

Per instruction to focus on the hard part first, we commented out phases 3+ so the demo is minimal and targeted:
- Phase 1: initial weather + fx question
- Phase 2: “repeat that” follow-up

This reduces noise while validating the correct persistence + restoration semantics.

### 5) Improved Debug Output So We Can Verify Exactly What Is Being Passed

We kept debug-style logging (as requested), but shifted it toward *verifying the real issue*:

- Phase 2 prints which messages are being injected as history via `team.additional_input`.
- Phase 2 explicitly logs that `team.arun()` is called with ONLY the new user message.
- We added debug output at `TeamRunCompleted` to confirm the team’s final response was accumulated and stored:
  - `content_len`, `tool_calls`, `is_stored`.

This removes ambiguity about “what went in” and “what came out”.

## Key Files Affected

- `scripts/random/teams_demo2.py`
  - Fixed incorrect history passing that flattened `List[Message]` into a single user message.
  - Updated `process_team_events` to accept a generic input object and pass it directly to `team.arun(...)`.
  - Phase 1 now passes a single `Message` to `team.arun`, not a list.
  - Phase 2 now injects restored history via `team.additional_input` and calls `team.arun()` with only the new user message.
  - Restored messages now set `from_history=True` in `stored_to_message`.
  - Kept and improved debug logging, and commented out phases 3+ to focus on core behavior.

## Errors and Barriers

### 1) Misunderstanding Agno Input Semantics (Resolved)
Initial assumption: passing `List[Message]` to `team.arun()` would behave like “chat history”.
Reality: Agno converts a list of `Message` objects into a single `role="user"` message by calling `get_text_from_message(...)`.

This was the primary cause of “re-delegation” even when we believed history was present.

### 2) Confusion Around “Team response not stored” vs “history not interpreted”
The Team response was being stored in the mock DB, but because history was being flattened incorrectly, the model was not being given a structured conversation context. This manifested as repeated delegation and looked like a persistence failure.

The fix is to pass history in the way Agno expects (additional_input or DB-backed history), not to “add more prompt instructions”.

## What Comes Next

### Immediate Next Steps (Demo Validation)
1. Re-run `scripts/random/teams_demo2.py` multiple times and confirm Phase 2:
   - Does not delegate to members
   - Emits only Team events (TeamRunStarted/Content/Completed)
   - Produces a repeat/summarization directly

2. Keep debug output until we are fully confident in:
   - event ordering under delegation
   - persistence trigger points
   - correct restoration semantics

### Re-introduce Advanced Scenarios (After confirming Phase 1 + Phase 2 stability)
Once the Phase 1/2 mechanism is stable, re-enable the following scenarios incrementally:
1. Image input with base64 stripping (store metadata + system note only).
2. Recall test with a “remember bird” interaction.
3. Context limiting by restoring only last N interactions.

Important: when we re-enable context limiting, we should apply it to what we inject via `team.additional_input`, not by passing a list into `team.arun()`.

### Schema Direction (High Confidence)
Given Agno’s internals, a robust storage approach should support reconstructing actual `Message` objects with correct role separation. For persistence primitives:
- Store enough to reconstruct messages faithfully.
- Avoid storing image base64; store metadata + system note.
- Store tool calls and tool results as separate message entries (role="tool") in chronological order.

### Agno Internal Mechanisms (Quick Reference)
- `Team.arun(input=...)` does not automatically interpret `List[Message]` as history.
- `_get_user_message` converts a `List[Message]` into a single `Message(role="user")` with flattened content.
- True “history injection” happens via:
  - `add_history_to_context=True` AND a DB-backed `TeamSession` (Agno-managed), or
  - `team.additional_input` (explicit list of `Message` objects inserted before the user message).

### Spec Progress
This work advances the “Message Persistence Research” task by:
- validating the correct method of passing restored history to Teams
- documenting a key Agno internal gotcha that would otherwise cause repeated re-delegation and confusion

Next spec work should proceed with:
- finalizing the schema (likely dual strategy: flattened columns + canonical message dict for perfect reconstruction if needed)
- implementing actual storage primitives (`src/agents/storage.py`) once the demo is stable
- then building the SSE API layer only after persistence correctness is locked in

## Addendum - Agno Code Breadcrumbs, Do/Don't Rules, and Next-Session Playbook

This addendum is intentionally detailed so that the next session does not require re-reading Agno internals.

### A) Agno Internals: The Exact Mechanism That Broke “History Passing”

The issue was not “model behavior”. It was a deterministic transformation performed by Agno when `Team.arun()` is called with a list input.

#### A.1 `Team.arun()` entrypoint (streaming path)
Reference: `agno/team/team.py` (`Team.arun`) (user provided selection around lines 2682-2988).

Key call chain (streaming):
- `Team.arun(input=...)`
  - `validated_input = self._validate_input(input)`
  - if streaming: `self._arun_stream(input=validated_input, ...)`

#### A.2 `_validate_input` does not change the input when no schema is set
Reference: `agno/team/team.py` (`Team._validate_input`) (user provided selection 767-815).

If `self.input_schema is None`, it returns the input unchanged.

Meaning:
- If you pass `List[Message]`, it remains `List[Message]` at this stage.

#### A.3 History/Prompt construction happens inside `_arun_stream` via `_aget_run_messages`
Reference: `agno/team/team.py` (`Team._arun_stream`) (user provided selection 2356-2680).

Inside `_arun_stream`:
- `run_messages = await self._aget_run_messages(... input_message=run_input.input_content, ...)`

Important:
- This is where Agno builds the actual message list passed to the model.
- If we “pass history incorrectly”, it is already too late after this step.

#### A.4 `_aget_run_messages` / `_get_run_messages` ordering
Reference: `agno/team/team.py` (`Team._aget_run_messages` / `Team._get_run_messages`) (user provided selection 6070-6489).

Agno builds the list of messages in this order:
1. system message
2. `additional_input` messages (if present)
3. history from DB (`session.get_messages(...)`) only if `add_history_to_context=True`
4. user message (built from `input_message`)

This means:
- If you are not using Agno’s DB session history, `add_history_to_context` will not help.
- The correct hook for injecting reconstructed history in a “custom persistence” world is `team.additional_input`.

#### A.5 The core gotcha: `List[Message]` passed as input becomes ONE user message
Reference: `agno/team/team.py` (`Team._get_user_message`) (included in the selection 6070-6489).

When `input_message` is a `list` and the first element is a `Message`, Agno does this:
- extracts text via `get_text_from_message(input_message)`
- constructs a new `Message(role="user", content=input_content, ...)`

Therefore:
- Passing `team.arun(restored_messages + [new_user_msg])` does NOT send role-separated history.
- It flattens the entire list into a single user text blob.
- The model never receives “assistant said X” as an assistant message; it receives “user content contains X”.
- This is why “repeat/summarize” turns into “redo the workflow” and the team delegates again.

### B) Correct “Custom Persistence” Integration Pattern for Teams (Model-Agnostic)

If we are managing persistence ourselves (not using Agno’s DB/session storage), the correct approach is:

1. Reconstruct prior turns into real `Message` objects (role preserved).
2. Inject those `Message` objects via `team.additional_input`.
3. Call `team.arun()` with ONLY the new turn input (preferably a single `Message(role="user", ...)`).

In other words:

- History: `team.additional_input = restored_messages`
- Input: `team.arun(user_msg_2, stream=True, stream_events=True)`

This is model-agnostic:
- It relies on Agno’s message assembly order, not any provider-specific payload structure.

### C) Do / Don’t Rules (Non-Negotiable)

#### Don’t
- DO NOT call:
  - `team.arun(restored_messages + [new_user_msg])`
  - `team.arun([user_msg])` (even for a single message)
- DO NOT assume:
  - “List[Message] == chat history” for `Team.arun`
- DO NOT attempt to fix repeated delegation by only tightening prompts while history is malformed.

#### Do
- DO pass the new turn as a single input:
  - `team.arun(user_msg)` where `user_msg` is `Message(role="user", ...)`
- DO inject restored history using:
  - `team.additional_input = restored_messages`
- DO restore `team.additional_input` afterward (avoid leaking history between test phases):
  - capture previous value and restore in `finally`

### D) Debug/Verification Checklist (How to Prove It’s Correct)

When Phase 2 is correct, you should see:

1. The debug line confirming:
- we inject N history messages via `team.additional_input`
- we call `team.arun()` with ONLY the new user message as input

2. Event stream in Phase 2 should contain ONLY team events:
- `TeamRunStarted`
- `TeamRunContent` (many)
- `TeamRunContentCompleted`
- `TeamRunCompleted`

3. Phase 2 should NOT contain any of:
- `TeamToolCallStarted` / `TeamToolCallCompleted`
- member events like `RunStarted` / `ToolCallStarted` / `RunCompleted` from Weather/Finance agents

If any tool/member events appear in Phase 2:
- The team is delegating, which usually means history injection is missing/incorrect OR the injected history did not include the relevant team answer.

### E) Next-Session Playbook (Step-by-Step)

#### E.1 Confirm the foundational behavior (Phase 1 + Phase 2 only)
1. Keep only Phase 1 and Phase 2 enabled until stable.
2. Run multiple times and confirm Phase 2 never delegates.
3. Confirm the database contains:
   - user message from Phase 1
   - final team answer from Phase 1 (assistant, agent_name == team.name)
   - user message from Phase 2
   - final team answer from Phase 2 (assistant, agent_name == team.name)

#### E.2 Re-enable advanced tests incrementally (do not change history passing pattern)
When re-introducing tests, the rules remain:
- Inject restored history via `team.additional_input`
- Call `team.arun()` with only the new user message

Recommended sequence:
1. Image input:
   - store image metadata only (strip base64)
   - add system note into content at storage time
   - on restore: do not attach images back to the `Message` object (the system note is enough)
2. Recall test:
   - “The word is bird, remember it”
   - “What is the word?”
3. Context limiting:
   - Apply the limit to the history list that you assign to `team.additional_input` (not to the `input` parameter)
   - Example: inject only last N interactions worth of `Message` objects into `team.additional_input`

#### E.3 If/when we move to SSE streaming
The streaming wrapper should treat the persistence boundary like this:
- Persist tool call request + tool result when tool call completes
- Persist final assistant/team message at run completion
- Never attempt to reconstruct state by passing `List[Message]` directly as input to a Team run; always inject via `additional_input` or use Agno’s DB-backed history.

### F) Why This Matters for Schema and Storage Design

Because Teams are sensitive to role-separated context:
- Our persistence must be able to reconstruct a *role-correct* sequence of `Message` objects.
- The storage layer must preserve enough information to reconstruct:
  - `role`
  - `content`
  - tool messages (`role="tool"`, `tool_call_id`, `name`, `content`)
  - tool call requests (assistant messages with `tool_calls` metadata, if we store them)
- For Teams, the “correctness” of restoration is not “we stored something”; it is “Agno receives it as structured messages before the new user message”.

If restoration cannot produce a faithful list of `Message` objects, team behavior will look like a persistence bug even when the DB is technically “storing answers”.
