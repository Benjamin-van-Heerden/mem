# Work Log - Teams Demo Media Persistence Fix + Dev Login Route + Frontend Dev Login Page

## Spec File: `agent_rules/spec/s_20251129_benjamin_van_heerden__agent_infrastructure_migration.md`

## Overarching Goals
- Validate and harden Agno Team persistence using DB-backed sessions (MySQL) while observing real behavior around media (images).
- Add a minimal, KISS local development login flow to support frontend development against the new backend.
- Establish tangible, demoable progress by wiring frontend → API in the new codebase.

## What Was Accomplished

### 1) Validated DB-backed Team persistence and investigated media storage behavior
- Extended `scripts/random/teams_demo3.py` with a new run mode to send an image message (using `scripts/random/mystery.png`) to a DB-backed Agno Team session.
- Observed a persistence failure when media was stored:
  - MySQL session upsert failed with `TypeError: Object of type bytes is not JSON serializable`
  - This prevented session persistence and caused `team.get_session()` to return `None`.
- Implemented the mitigation by setting `store_media=False` on the Team:
  - After this, the image run completed successfully
  - Session persistence succeeded
  - `team.get_session(session_id)` returned a session and replay worked
  - Session payload size remained small (~17–19k chars), confirming that raw media/base64 was not being stored in session JSON.

Key conclusion from the check:
- With `store_media=False`, Agno sessions do not appear to retain any persisted media fields in replayed chat history/messages; only the text content is persisted.
- This is acceptable for now because uploads are not being implemented yet; later, uploads can store external URLs and construct media objects from URLs rather than embedding base64.

### 2) Added a typed dev router with `/wcu_login` convenience endpoint
- Implemented a dev-only `POST /wcu_login` endpoint that returns JWT + user info for frontend dev.
- Route is organized under a dedicated dev router and uses typed request/response models.
- Included the dev router in the FastAPI app only when `ENV_SETTINGS.app_env == "dev"`.

Notes:
- Implemented as a sync route handler to avoid blocking the async event loop with the existing sync `httpx` calls in `src/utils/weconnectu/auth.py::wcu_login`.
- Response includes:
  - `jwt`
  - `user` (user_id, email, name)
  - `refresh_token` (optional)

### 3) Rebuilt `frontend/pages/index.vue` as a minimal dev login UI (KISS)
- Implemented a clean, minimal Nuxt page that:
  - accepts an API base URL
  - logs in via `POST /wcu_login`
  - displays user info and a compact JWT preview
  - provides “Copy JWT” and “Logout” actions
- Fixed UX issues:
  - Added spacing between the API Base URL label and input by wrapping the label+input in the same `.field` pattern.
  - Replaced the full JWT `<textarea>` (which overflowed the container) with a compact preview in a `<p>` tag.
- Added persistence across refresh/HMR:
  - Stores `apiBase`, `jwt`, and `user` in `localStorage`
  - Restores on mount
  - Clears on logout

## Key Files Affected
- `scripts/random/teams_demo3.py`
  - Added image-mode run to test `mystery.png` behavior.
  - Added session payload diagnostics to estimate stored session size.
  - Set `store_media=False` on the Team to prevent JSON serialization failures and DB bloat.
- `src/api/main.py`
  - Included dev router when `ENV_SETTINGS.app_env == "dev"`.
- `src/api/routes/dev/models.py`
  - Added typed Pydantic request/response models for dev login endpoint.
- `src/api/routes/dev/router.py`
  - Added `POST /wcu_login` convenience login endpoint calling `src/utils/weconnectu/auth.py::wcu_login`.
- `frontend/pages/index.vue`
  - Implemented KISS dev login page calling `/wcu_login`.
  - Added spacing improvements, compact JWT display, and localStorage persistence for HMR.

## Errors and Barriers
- Media persistence error when storing images in Agno sessions:
  - `TypeError: Object of type bytes is not JSON serializable` during session upsert
  - Resolved by disabling media persistence via `store_media=False`.
- Tooling/editor warning related to Vue global types generation appeared during diagnostics; not treated as a blocking runtime issue.

## What Comes Next
- Start building the actual agent/team interaction API endpoints (streaming) and a minimal chat UI component in the new frontend (KISS, not porting the old complex `AgentInterface.vue` as-is).
- Maintain `store_media=False` for now and avoid file uploads until there is a dedicated upload pipeline:
  - later: store files externally (S3/URL), reference by URL, and optionally persist metadata for UX replay.
- Continue expanding from the dev login baseline toward a demoable end-to-end flow: login → open chat → send message → stream response.