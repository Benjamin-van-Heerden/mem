---
created_at: "2026-10-25T15:49:43+02:00"
user: benjamin_van_heerden
spec: cursor_support
---

# Work Log - Cursor support: a stop-hook work-log follow-up, quiet Claude hooks; agent hook research

## Goal

Benjamin's collaborators work in Cursor. The aim was to make mem work there as well as Cursor's hooks allow, after researching what Cursor, OpenAI Codex and Grok offer in place of Claude Code's `SessionStart`/`compact` and `UserPromptSubmit` hooks.

## What Was Accomplished

### Hook research (2026-10-25)

Three research subagents checked current docs, source and forums. Codex was then dropped from this round (see Decisions); its findings and Grok's are in the todo `hook_support_codex_grok` and the abandoned spec `.mem/specs/archive/hooks_codex_cursor_grok/`.

| | Claude Code | Codex CLI 0.164 | Cursor 3.24 | Grok Build 1.0.67 |
|---|---|---|---|---|
| Context after compaction | `SessionStart` `compact` | Same event and matcher, plain stdout | No: `preCompact` only messages the user; `sessionStart` does not refire | No: `PreCompact`/`PostCompact` are passive |
| Context per message | `UserPromptSubmit` | Same event, plain stdout | No: `beforeSubmitPrompt` only allows or blocks | No: output discarded |
| Reads `.claude/settings.json` hooks | yes | no | yes | yes |

- **Cursor's `sessionStart` `additional_context`** is dropped by a race, which was still reported in 3.24.9.
- **Codex** reads `AGENTS.md` only up to `project_doc_max_bytes` (32 KiB). This was verified with codex 0.153.4 using `codex debug prompt-input`, a scratch `CODEX_HOME` and a trusted scratch repository:
  - By default, a 39,813-byte AGENTS.md was cut off at about line 577.
  - With `project_doc_max_bytes = 65536` in the project's `.codex/config.toml`, it was read in full.
- **AGENTS.md sizes:**
  - shannon is 33,456 bytes: the mem block is 16,339 and its 14 memories 17,116.
  - ev-ledger is 31,042 bytes.
  - The mem block alone is 6.9 KB of workflow plus 9.4 KB of general principles.
  - ev-server and praxis-dev-folder are larger, but they are not mem projects: one is Agent Core, the other Praxis.

### `internal/claude` → `internal/agenthooks` (0815bbf)

- `json.go` holds the ordered `object` and a shared `syncFile(root, rel, fresh, []memEntry)` / `syncEntry`. These put mem's entries in place in the lists under a file's `hooks` object, keep everything else in order, write only on change, and create no file when nothing is enabled.
- `claude.go`: `ClaudeSettingsPath`, `ClaudeHooks` and `SyncClaude`, with the same behavior as the old `SyncHooks`.
- `cursor.go`: `SyncCursor(root, stop)` maintains `{"version": 1, "hooks": {"stop": [{"command": "mem hook stop", "timeout": 30}]}}`, matching mem's flat entry by its command.
- `project.CursorConfig` adds `[cursor] stop_hook` (`*bool`, unset means on).
- `init` and `import` write and list `.cursor/hooks.json`. `applyUpdates` loops over both agent files with the same handling for each: ignored, uncommitted edits, or published with the project files commit. `settingsIgnored` takes the path, and `cursorHookLine` words the update.

### `mem hook stop` (6931e94)

- It reads Cursor's `{status, loop_count}` and always prints a JSON object.
- It sends `followup_message` only when the turn has `status: completed`, `loop_count` 0, a due work log, and a `Work.Summary()` different from the one in `.mem/local/stop_nudge`. Below the due level it removes the record.
- The wording, `checkpoint.TurnEndLine`:
  - An automatic follow-up from mem, not the user.
  - If the turn finished a piece of work, write the log now (`mem log new`, then `mem log commit`) and end the turn.
  - Otherwise, say that a log is due and end the turn.
  - Do not start other work.

### Quiet Claude hooks under Cursor (ad16cc4)

`underCursor()` returns true when `CURSOR_VERSION` is set. `mem hook prompt` and `mem hook compact` then return before loading the project.

### Docs (72ffab4)

- README: init and a new Agents item.
- `docs/design.md`: an "In Cursor" section and `[cursor]` in the config example.
- `docs/status.md`: implemented, verified only with simulated input, plus the unverified questions.
- `.mem/structure.md`.

### Verification

- Tests:
  - agenthooks: `SyncCursor` on a fresh file, merged with other hooks and keys, removal, and opt-out.
  - `TestOnboardInstallsTheCursorHookAndHonoursTheOptOut`.
  - `TestStopHookAsksForADueWorkLogOncePerStateOfWork`: below due, aborted, error, `loop_count` 1, invalid stdin, the follow-up, dedup, re-asking after more work, and outside a mem project.
  - `TestClaudeHooksStayQuietWhenCursorRunsThem`.
  - `TurnEndLine`'s levels.
- `go test` and `go vet` passed for `internal/agenthooks`, `checkpoint`, `project` and `cli`.
- `dist/mem-dev` in a scratch repository: init wrote `.cursor/hooks.json`, `hook stop` returned the follow-up with 1,300 uncommitted lines, and `hook prompt` under `CURSOR_VERSION` printed nothing.

## Decisions

- **Cursor only for now.** The user first chose Codex hooks by default plus a rename of `[claude]` to `[agent_hooks]`, then narrowed the scope to Cursor. The spec `hooks_codex_cursor_grok` was abandoned, and its research went into todos.
- **Per-agent config sections.** `[cursor] stop_hook` sits next to `[claude]` instead of an `[agent_hooks]` rename, so there is no schema migration.
- **The stop follow-up (user's choice):** only when a log is due, once per state of the work, never after a stop by the user or in reply to itself. It is Cursor's only hook that can put text in front of the agent.
- **No `sessionStart`, `preCompact` or `beforeSubmitPrompt` hooks.** `sessionStart` is broken, `preCompact` only reaches the user, and `beforeSubmitPrompt` cannot add context.
- **`mem hook stop` has no `command -v … || true` guard,** because Cursor runs hook commands through PowerShell on Windows. A missing mem fails open.
- **The hook command never uses `CLAUDE_PROJECT_DIR`** to tell agents apart, because Cursor (and Grok) set it too.

## Key Files Affected

- `internal/agenthooks/` (new; from `internal/claude/`): `json.go`, `claude.go`, `cursor.go`, `claude_test.go`, `cursor_test.go`
- `internal/cli/hook.go`: `hook stop`, `stopFollowUp`, `underCursor`
- `internal/cli/compact.go`: return early under Cursor
- `internal/cli/onboard_updates.go`: loop over the Claude and Cursor files, `cursorHookLine`
- `internal/cli/init.go`, `internal/cli/import.go`: write and list `.cursor/hooks.json`; `settingsIgnored(ctx, root, path)`
- `internal/cli/hook_test.go`, `internal/cli/onboard_test.go`: the stop, quiet and onboard tests
- `internal/checkpoint/checkpoint.go` and its test: `TurnEndLine`
- `internal/project/project.go`: `CursorConfig`
- `README.md`, `docs/design.md`, `docs/status.md`, `.mem/structure.md`
- `.mem/todos/`: `hook_support_codex_grok`, `slim_managed_agents_md_block`, `check_cursor_hooks_real`

## Notes

- Cursor is not installed on this machine, so the hooks are verified only with simulated stdin. A check in a real Cursor is the todo `check_cursor_hooks_real`. The task 4 commit message calls it `check_cursor_hooks_real_cursor`, which is wrong.
- `codex debug prompt-input` renders the model input without a model call, but it does not run hooks. Checking Codex hooks would need `codex exec` with the user's account.
- Another session worked in this checkout at the same time (worktree warnings, c810112 and 4960796). Task commits named their paths, so its changes stayed out of them.
