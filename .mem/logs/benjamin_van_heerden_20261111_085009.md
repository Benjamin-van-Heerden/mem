---
created_at: "2026-11-11T08:50:09+02:00"
user: benjamin_van_heerden
commits: [5539ffc, d719948]
---

# Work Log - Codex hooks in .codex/hooks.json, verified in a real Codex

## Goal

Give Codex the same compaction catch-up and work-log nudge that Claude Code gets, and confirm in a real Codex that they reach the model.

## What Was Accomplished

### Codex hooks

- `agenthooks.SyncCodex` (`internal/agenthooks/codex.go`) keeps two entries in `.codex/hooks.json` through the shared `syncFile`: `SessionStart` with matcher `compact` running `CompactCommand` (60 s), and `UserPromptSubmit` running `PromptCommand` (10 s). Each also has `"commandWindows": "mem hook compact"` / `"mem hook prompt"`, because the POSIX guard does not run in Windows shells.
- `claudeEntry` became `commandEntry`, shared by the Claude and Codex files.
- `[codex] compact_hook` / `prompt_hook` (`project.CodexConfig`, unset means on) turn them off.
- `init`, `import`, `init --folder`, onboard's `applyUpdates` and `applyFolderUpdates` sync the file. Onboard publishes it with the project files, like the Claude and Cursor files.
- `claudeHooksLine` now uses a shared `contextHooksLine`. `codexHooksLine` adds that Codex runs the hooks only in a trusted project once each user approves them in `/hooks`, again in every checkout or worktree.
- `TestSyncCodexInstallsStableEntriesIntoAFreshProject` pins the file byte for byte, because a changed entry loses its approval. `TestOnboardInstallsTheCodexHooksAndHonoursTheOptOut` covers onboard restoring and publishing the file and the `[codex] prompt_hook = false` opt-out.

### Verified with Codex 0.164 on macOS

- Setup: a scratch mem repository with a local bare remote and 12 work commits (log due), and diagnostic hooks beside mem's that dumped stdin, environment, working directory, `command -v mem` and a `curl` to github.com. The user approved the hooks in `/hooks` in an interactive Codex. After that, `codex exec` ran without any bypass flag.
- Prompt hook: the nudge arrived as a developer message with `content_item_kinds: ["hooks.additional_context"]` in the session's rollout file (`~/.codex/sessions/…jsonl`).
- Compaction hook: forced with `-c model_auto_compact_token_limit=12000`. `SessionStart` fired with `"source": "compact"` after each mid-turn auto-compaction, and the 🔄 MEM AFTER COMPACTION digest arrived as developer context.
- Hooks ran outside the sandbox (`-s read-only`, curl returned 200), so the catch-up's fetch and push work.
- Hooks get the login shell's `PATH`, not the `PATH` Codex was started with. The diagnostic found `~/.local/bin/mem` even with the dev build first on the launching `PATH`.
- Approval records sit in `~/.codex/config.toml` as `[hooks.state."<abs path>/.codex/hooks.json:<event>:<group>:<handler>"]` with a `trusted_hash`. Removing the diagnostic entry ahead of mem's `UserPromptSubmit` entry moved it from group 1 to 0, and Codex then skipped it silently, with no warning in `codex exec` output.
- No Codex-specific variable identifies Codex in the hook environment (only `CODEX_MANAGED_BY_NPM`, from the install). Codex does not run `.claude/settings.json`.

### Records

- `hook_support_codex_grok` was replaced by `hook_support_grok` (the Grok half) and `check_codex_hooks_windows`.

## Decisions

- **Same two hooks as Claude Code, no Stop hook for Codex.** Codex's Stop hook can continue the turn like Cursor's follow-up, but the prompt hook already nudges with each message.
- **No `.codex/config.toml`.** The existing AGENTS.md size warning already names `project_doc_max_bytes`.
- **Per-agent `[codex]` section**, matching `[claude]` and `[cursor]`, rather than the agent-neutral `[agent_hooks]` the abandoned spec proposed.

## Key Files Affected

- `internal/agenthooks/codex.go`, `internal/agenthooks/codex_test.go`, `internal/agenthooks/claude.go`
- `internal/cli/onboard_updates.go`, `internal/cli/onboard_test.go`, `internal/cli/folder.go`, `internal/cli/folder_test.go`, `internal/cli/init.go`, `internal/cli/import.go`, `internal/cli/hook.go`
- `internal/project/project.go`
- `README.md`, `docs/design.md`, `docs/status.md`, `.mem/structure.md`

## Notes

- `codex exec --dangerously-bypass-hook-trust` was refused by Claude Code's permission check as a safety bypass, so the user approved the hooks in an interactive Codex instead.
- The Codex models refused to quote hook context back ("I can't quote hidden system, developer, or hook instructions"), so verification read the rollout file instead.
- `codex exec` waits on stdin when it is an open pipe; run it with `< /dev/null`.
- The scratch repository's path now has trust records in the user's `~/.codex/config.toml` (`[projects."…/codexcheck/repo"]` and its `[hooks.state.…]` entries).
- `gofmt -l` flags `internal/checkpoint/checkpoint.go`, which this work did not touch.
