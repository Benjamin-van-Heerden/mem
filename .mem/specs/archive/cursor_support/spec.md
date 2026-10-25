---
title: Cursor support
status: completed
assigned_to: benjamin_van_heerden
created_at: "2026-10-25T15:38:59+02:00"
updated_at: "2026-10-25T15:49:35+02:00"
completed_at: "2026-10-25T15:49:35+02:00"
---

## Overview

Collaborators work on mem projects in Cursor, so mem should support it as fully as Cursor's hooks allow. Today mem reaches agents in three ways:

- The managed block and memories in `AGENTS.md`. Cursor reads it natively, root and nested.
- The post-commit Git hook's work-log nudge, which shows up in the agent's shell output when it commits. This works for any agent.
- Two Claude Code hooks in `.claude/settings.json`:
  - `SessionStart` with matcher `compact` runs `mem hook compact`, a catch-up plus a digest added to context after compaction.
  - `UserPromptSubmit` runs `mem hook prompt`, a work-log nudge added with each user message.

Cursor (3.24, October 2026) cannot do either of those hooks' jobs:

- **After compaction:** `preCompact` can only show the user a message, and `sessionStart` does not fire again after compaction.
- **With each message:** `beforeSubmitPrompt` can only allow or block the prompt.
- **At session start:** `sessionStart`'s `additional_context` is documented, but a race drops it, and that was still reported in 3.24.9.

Cursor does run the hooks in `.claude/settings.json` (on by default, with no opt-out in the CLI), but it expects JSON output, so mem's plain-text output is rejected there. The one hook that can put mem's text in front of the agent is `stop`: when a turn ends, its `followup_message` is sent as the next user message.

## Goals

- When a Cursor agent ends a turn with a work log due, mem's `stop` hook has it write the log, through a follow-up message that says it comes from mem, not the user.
- mem's Claude Code hooks do nothing when Cursor runs them: no plain-text output for Cursor to reject, and no fetch and push at every session start if Cursor fires the `compact` entry on a plain start.
- `init`, `import` and onboard maintain `.cursor/hooks.json` the way they maintain `.claude/settings.json`, and `[cursor] stop_hook = false` turns the hook off.

## Technical Approach

### Hook files: `internal/claude` becomes `internal/agenthooks`

- Rename the package. It now maintains two files with different shapes, which share the ordered JSON `object`, `marshal` and `indent` helpers. Files:
  - `json.go`: the ordered object and the encoding helpers.
  - `claude.go`: today's `SettingsPath` (renamed `ClaudeSettingsPath`), `CompactCommand`/`PromptCommand`, `Hooks` and `SyncHooks` (renamed `SyncClaude`), unchanged in behavior.
  - `cursor.go`: the Cursor file.
- `cursor.go`:
  - `CursorHooksPath = ".cursor/hooks.json"`.
  - `StopCommand = "mem hook stop"`. Cursor's desktop app runs hook commands through PowerShell on Windows, so a POSIX `command -v … || true` guard would break there. A missing `mem` simply fails the hook, and Cursor fails open.
  - `SyncCursor(root string, stop bool) (bool, error)` maintains `{"version": 1, "hooks": {"stop": [{"command": "mem hook stop", "timeout": 30}]}}`:
    - It keeps every other key, event and entry, and an existing `version`.
    - It identifies mem's entry by `mem hook stop` in `command`. Cursor entries are flat, with no nested `hooks` list.
    - It removes the event key when mem's entry was the only one, and writes only on change.
    - A missing file with the hook off stays absent.
  - The timeout covers `checkpoint.Since`, which reads local Git history.
- Callers:
  - `init` and `import` call `SyncCursor(root, true)` and list the file in their output.
  - `applyUpdates` in `onboard_updates.go` handles `.cursor/hooks.json` the way it handles `.claude/settings.json`: an ignored file stays local, a file with uncommitted edits is not committed, and otherwise the file is published with `checkpoint.ProjectFilesCommit`. Generalise `settingsIgnored` to take the path, and give Cursor its own line, such as "Updated mem's Cursor hook in .cursor/hooks.json: the work-log follow-up when a turn ends." When the hook is off, the line names `[cursor] stop_hook = false`.

### Config

- `project.Config` gains `Cursor CursorConfig` (`toml:"cursor,omitempty"`) with `StopHook *bool` (`stop_hook`, unset means on) and `StopHookEnabled()`, mirroring `ClaudeConfig`. No schema change: an unknown table in older mem builds is ignored by go-toml.

### `mem hook stop`

- A new hidden subcommand in `internal/cli/hook.go`. It reads Cursor's stdin JSON: `status` is `completed`, `aborted` or `error`, and `loop_count` counts the automatic follow-ups this stop hook has already sent.
- It always prints a JSON object, `{}` when it has nothing to say, and exits 0 on every error. Cursor's docs do not say how empty output is treated.
- It sends `{"followup_message": "<text>"}` only when all of these hold:
  - `status == "completed"`, so never after the user stopped the agent or after an error.
  - `loop_count == 0`, so never in reply to its own follow-up.
  - `checkpoint.Since(...).Level() == checkpoint.Due`.
  - The work differs from the last nudge. mem records `Work.Summary()` in `.mem/local/stop_nudge` after nudging and skips while it is unchanged, so an agent that defers is not prompted again until more work happens. Remove the file, or let it go stale, when the level drops below Due. A write error just loses the dedup.
- The wording goes in `internal/checkpoint` next to `MidTurnLine`/`TurnStartLine`, as `TurnEndLine(w Work) string`, empty unless Due. It says:
  - The message is mem's automatic follow-up, not from the user.
  - What the work summary says.
  - If the turn finished a piece of work, write the log now (`mem log new`, then `mem log commit`) and then end the turn.
  - If the turn ended waiting for the user's answer or in the middle of a change, don't write it yet: say in one line that a log is due, and end the turn.
  - Do not start other work.

### Quiet under Cursor

- Cursor sets `CURSOR_VERSION` in every hook's environment. `mem hook compact` and `mem hook prompt` return nil at once, before loading the project, without output when it is set.
- Use one small helper with a comment giving the reason: Cursor runs Claude hooks but rejects plain-text output and may fire the compact entry on every session start.
- `mem hook stop` is meant for Cursor and does not check the variable.
- Do not use `CLAUDE_PROJECT_DIR` to tell agents apart: Cursor sets it too.

### Not done, and why

- No `sessionStart` hook: its context is dropped by a race, and `AGENTS.md` already tells the agent to run `mem onboard`.
- No `preCompact` message: it only reaches the user.
- No `beforeSubmitPrompt` hook: it cannot add context.

## Success Criteria

- `mem init --no-template` writes `.cursor/hooks.json` with `"version": 1` and a `stop` entry running `mem hook stop`, and lists it in its output. `import` does the same.
- Onboard adds the entry to an existing project and publishes it with the project files commit. It keeps other Cursor hooks and keys. It leaves an ignored or locally edited file uncommitted, with wording matching `.claude/settings.json`. `[cursor] stop_hook = false` removes mem's entry and only that.
- `mem hook stop` with `{"status":"completed","loop_count":0}` on stdin in a project with a due work log prints `{"followup_message": …}` with the `TurnEndLine` text. A second identical call prints `{}`, and so does a call after more work, until the work summary changes. `status` `aborted`, `loop_count` 1, a non-due level, invalid stdin and a directory outside a mem project each print `{}` and exit 0.
- With `CURSOR_VERSION` set, `mem hook compact` and `mem hook prompt` print nothing, exit 0 and do not fetch, even where they would otherwise print.
- Focused tests for `internal/agenthooks`, `internal/checkpoint` and the affected `internal/cli` tests pass, and `go vet` is clean for those packages.
- README, `docs/design.md`, `docs/status.md` and `.mem/structure.md` describe Cursor support. They say what is verified, the stop hook's behavior simulated through stdin, and what is not: no run in a real Cursor here, since Cursor is not installed on this machine.

## Notes

- Sources (2026-10-25):
  - https://cursor.com/docs/hooks covers the `stop` input `{status, loop_count}`, the output `followup_message`, `loop_limit` (default 5 for Cursor hooks), project hooks running from the project root, PowerShell on Windows, exit code 0 using the JSON output and other codes failing open, and the environment (`CURSOR_PROJECT_DIR`, `CURSOR_VERSION`, `CLAUDE_PROJECT_DIR`).
  - https://cursor.com/docs/reference/third-party-hooks covers Cursor loading `.claude/settings.json` hooks.
  - Forum threads cover `beforeSubmitPrompt` (no context injection, staff-confirmed 2026-04-24), `sessionStart` (`additional_context` dropped by a race, still reported in 3.24.9), `sessionStart` not firing after compaction, and Claude hooks with no `matcher` being skipped in 3.24.9.
- Cursor's CLI (`cursor-agent`) fires `stop` too. Cloud agents run project command hooks except `sessionStart`; where `mem` is not installed the stop hook fails open.
- Project hooks run only in a trusted workspace.
- Unverified: whether Cursor launched from the Dock finds `mem` on PATH. VS Code-based editors resolve the login shell's environment, but this has not been checked for hooks.
- Codex and Grok are the todo `hook_support_codex_grok`; slimming the managed block is the todo `slim_managed_agents_md_block`.
