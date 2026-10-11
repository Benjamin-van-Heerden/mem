---
created_at: "2026-10-12T00:00:27+02:00"
user: benjamin_van_heerden
---

# Work Log - Starting sessions from the mission-control band, tried and dropped

## Overarching Goals

Let the mission-control band replace the desktop app's sidebar entirely by also starting new sessions.

## What Was Accomplished

### What the app offers

- The app's session tools (`ccd_session_mgmt`, as typed in the mod's `.claude-plugin/types/claude-code-mcp/index.d.ts`) refer to a `start_session` tool (optionally with another folder or `target: "cloud"`), but this build exposes no such tool to the mod or to agents; only `list_sessions`, `get_session`, `send_message`, `detach_session`, `archive_session` and the `set_*` tools are there. Sessions it starts would nest under the starting session (side sessions, archived with it) until `detach_session` moves them to top level.
- No documented `claude://` deep link starts a session; the known form `claude://claude.ai/epitaxy/<session id>` opens an existing one. A link could at best open a new session in a folder the app picks, so it was not tried.

### Cmd+N from the band (built, then removed)

- `~/.claude/mods/mission-control/hooks/register.tsx` got `newSession($)`, which ran `osascript -e 'tell application "Claude" to activate' -e 'tell application "System Events" to keystroke "n" using command down'` and reported a failure with `$.ui.log`, and a plain dim "＋ New session" button beside this session's pill.
- The user pressed it: it opened a new session, as Cmd+N does. It added nothing over pressing Cmd+N, so `newSession`, the button and the `$.ui.log` calls were removed; the pill's row is back to the pill alone. Validation, `tsc -p .` and `claude plugin test` (15 tests) pass.

## Decisions

- The band does not start sessions: the user starts them with Cmd+N. A band button can only press the same shortcut, and `start_session` is not available in this build.

## Key Files Affected

- Outside the repository: `~/.claude/mods/mission-control/hooks/register.tsx` (the new-session button added and removed again; no net change).
