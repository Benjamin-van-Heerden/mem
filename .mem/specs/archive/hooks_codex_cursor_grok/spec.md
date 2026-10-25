---
title: Hooks for Codex, Cursor and Grok
status: abandoned
assigned_to: benjamin_van_heerden
created_at: "2026-10-25T15:29:17+02:00"
updated_at: "2026-10-25T15:38:49+02:00"
completed_at: "2026-10-25T15:38:49+02:00"
abandon_reason: Narrowed to Cursor only; Codex and Grok are deferred to a todo.
---

## Overview

mem reaches coding agents in two ways. The first is the managed block and memories in `AGENTS.md`. The second is two hooks, kept in `.claude/settings.json` and written only for Claude Code:

- A `SessionStart` hook with the `compact` matcher runs `mem hook compact`. It catches up after compaction and prints a digest.
- A `UserPromptSubmit` hook runs `mem hook prompt`. It prints the work-log nudge with each user message.

This spec extends mem to OpenAI Codex (its CLI, IDE extension and the ChatGPT desktop app, which share one configuration), Cursor and Grok. Each agent gets mem's context wherever its hook system allows it. mem's hooks must also do no harm where an agent runs them without being able to use them.

Research done on 2026-10-25 (sources in Notes) found:

| What mem needs | Claude Code | Codex (CLI 0.164; 0.153.4 installed here) | Cursor (3.24) | Grok Build (`grok` 1.0.67) |
|---|---|---|---|---|
| Persistent instructions | `AGENTS.md` | `AGENTS.md`, kept through compaction; reading stops at `project_doc_max_bytes` (default 32 KiB) | `AGENTS.md`, root and nested | `AGENTS.md`, `CLAUDE.md` and others, read in full; restored into the compacted history |
| Context after compaction | `SessionStart` `compact` | The same event and matcher; plain stdout becomes developer context | Not possible: `preCompact` can only message the user, and `sessionStart` does not fire again | Not possible: `PreCompact`/`PostCompact` are passive, and `SessionStart` never fires on compaction |
| Context with each message | `UserPromptSubmit` | The same event; plain stdout becomes developer context | Not possible: `beforeSubmitPrompt` can only allow or block | Not possible: `UserPromptSubmit` output is discarded |
| Hook file | `.claude/settings.json` | `.codex/hooks.json`, the same JSON shape as Claude's `hooks` object | `.cursor/hooks.json`; Cursor also runs `.claude/settings.json` hooks | `.grok/hooks/*.json`; Grok also runs `.claude/settings.json` and `.cursor/hooks.json` hooks |

## Goals

- Codex runs mem's compaction catch-up and work-log nudge from a committed `<repo>/.codex/hooks.json`, which `init`, `import` and onboard maintain the same way as `.claude/settings.json`.
- Codex reads all of `AGENTS.md`: mem sets `project_doc_max_bytes` in `<repo>/.codex/config.toml`, and onboard warns when `AGENTS.md` grows large enough to crowd any agent.
- mem's hooks do nothing when Cursor runs them. Cursor runs `.claude/settings.json` hooks, but it cannot take plain-text output from them, and it may fire the compaction catch-up (a fetch and a push) at every session start.
- The same holds for Grok Build, which runs `.claude/settings.json` hooks and discards their output. It runs `mem hook prompt` with every message.
- The two hook switches become agent-neutral: `[claude] compact_hook`/`prompt_hook` move to `[agent_hooks] compact`/`prompt` through a schema 3 patch, and they govern every agent's hook file.

## Technical Approach

### Hook files (`internal/claude` becomes `internal/agenthooks`)

- Rename the package `internal/claude` to `internal/agenthooks`. Today it only maintains `.claude/settings.json`, so the new name covers both files.
- Generalise `SyncHooks(root, want)` to take the target file:
  - `ClaudeSettings` has path `.claude/settings.json`; mem's entries sit under the `hooks` key among other settings.
  - `CodexHooks` has path `.codex/hooks.json`, which uses the same `{"hooks": {"<Event>": [{"matcher", "hooks": [{type, command, timeout}]}]}}` shape.
  - Both files keep the ordered `object` and the entry matching by marker as they work now. Keep the other top-level keys of `.codex/hooks.json`, such as `description`.
  - A target file that does not exist and has no hooks to write stays absent.
- Codex entries:
  - Compaction: `SessionStart`, matcher `compact`. Codex matchers are regexes on `source`, which is one of startup, resume, clear, compact or fork, so `compact` matches only compaction.
  - Prompt: `UserPromptSubmit`, no matcher.
  - Both use the same `CompactCommand`/`PromptCommand` strings and timeouts (60 s, 10 s) as Claude. Plain stdout works, and mem's output never starts with `{`, which matters because stdout that looks like JSON but fails to parse makes Codex fail the hook.
  - No `additionalContextLimit`: the digest stays under about 3 KB, below Codex's default preview threshold of about 2,500 tokens.
- **The entries must stay stable once released.** Codex trusts each project hook by a hash of its definition. A changed command, timeout or matcher is skipped until each user approves it again in `/hooks`. Add a comment saying so next to the entry definitions.
- Callers: `init` and `import` sync both files with both hooks on. `applyUpdates` in `onboard_updates.go` loops over both targets:
  - It keeps the per-file handling it has today: an ignored file stays local, a file with uncommitted edits is not committed, and otherwise the file is published with `checkpoint.ProjectFilesCommit`.
  - Generalise `settingsIgnored` and `claudeHooksLine` to take the target and its agent name.
  - When onboard, init or import first creates `.codex/hooks.json` or changes mem's entries in it, the output line says that Codex users must trust the project and then approve the hooks in Codex's `/hooks`.

### Codex's AGENTS.md limit (`.codex/config.toml`)

- Codex stops reading `AGENTS.md` at `project_doc_max_bytes`, 32,768 bytes by default, and mem's memories come last. On 2026-10-25 shannon's AGENTS.md was 33,456 bytes and ev-ledger's 31,042. The mem block alone is 16.3 KB: 6.9 KB of workflow and 9.4 KB of general principles.
- Verified with Codex 0.153.4 using `codex debug prompt-input`, a scratch `CODEX_HOME` and a trusted scratch repository:
  - A 39,813-byte AGENTS.md was cut off at about line 577 by default.
  - With `project_doc_max_bytes = 65536` in the project's `.codex/config.toml`, it was read in full.
- `init`, `import` and onboard ensure `.codex/config.toml` has a top-level `project_doc_max_bytes`:
  - If the key is missing, insert `project_doc_max_bytes = 65536` before the first table header. Create the file if it is absent.
  - A value already present is never changed, whatever it is: the user owns it.
  - Edit the file line by line, so comments and the user's other settings survive; go-toml would drop them.
  - Publish it alongside `.codex/hooks.json`.
  - It is written whatever the hook switches say, since Codex reading all of `AGENTS.md` does not depend on hooks.
- Onboard warns when `AGENTS.md` exceeds 40,000 bytes. That is well under the raised Codex limit and near the size at which Claude Code warns about large instruction files. The warning is one ⚠️ line in the updates section naming the size and the largest memories, and suggests trimming or retiring memories (`mem memory remove`).

### Cursor (`mem hook compact`, `mem hook prompt`)

- Cursor sets `CURSOR_VERSION` (and `CURSOR_PROJECT_DIR`) in every hook's environment. When `CURSOR_VERSION` is set, `mem hook compact` and `mem hook prompt` return at once without output or Git access.
- mem writes no `.cursor/hooks.json`. Cursor gets mem through `AGENTS.md` and the post-commit Git hook's nudge, which appears in the agent's shell output when it commits.
- Not in this spec: work-log nudges through Cursor's `stop` `followup_message` or Grok's `Stop` block. Either one makes the agent carry on after it has finished. The post-commit nudge already covers commits.

### Grok

- Grok Build sets `GROK_HOOK_EVENT` and `GROK_SESSION_ID` in every hook's environment, and also `CLAUDE_PROJECT_DIR`, so that variable cannot identify Claude Code. When `GROK_HOOK_EVENT` is set, `mem hook compact` and `mem hook prompt` return at once, the same as under Cursor.
- Put the check in one helper in `hook.go`, such as `hookCannotInject()`, covering `CURSOR_VERSION` and `GROK_HOOK_EVENT`. Name the reason in its comment: these agents run Claude hooks but discard or reject their output.
- mem writes no `.grok/hooks/` files. Grok gets mem through `AGENTS.md`, which it reads without a size limit and restores after compaction; through the skills in `.agents/skills`/`.claude/skills`; and through the post-commit nudge.
- Grok needs folder trust (`/hooks-trust` or `--trust`) before it loads project instructions or hooks. That is the user's step; the README says so.

### Config (`[agent_hooks]`)

- `project.Config` replaces `Claude ClaudeConfig` (`[claude]`) with `AgentHooks AgentHooksConfig` (`[agent_hooks]`). It has `Compact *bool` (`compact`) and `Prompt *bool` (`prompt`), where unset means on, plus the `CompactEnabled`/`PromptEnabled` helpers.
- `Schema` becomes 3, with patch `2: moveClaudeHooks`:
  - Read `.mem/config.toml` as raw TOML.
  - Move `[claude] compact_hook`/`prompt_hook` to `[agent_hooks] compact`/`prompt`, and drop an empty `[claude]` table.
  - Return the path.
  - `Upgrade` writes `p.Config` after each patch, so the patch must also set `p.Config.AgentHooks`. Otherwise the in-memory config, which has no `[claude]` field, overwrites the move. Adjust `Upgrade`'s patch signature if needed: a patch that changes the config returns the updated config.
- Update the wording that names the switches: `claudeHooksLine`, README, `docs/design.md` and `docs/status.md`.

## Success Criteria

- In a fresh `mem init --no-template` repository, `.codex/hooks.json` holds a `SessionStart` entry with matcher `compact` running `CompactCommand` and a `UserPromptSubmit` entry running `PromptCommand`. `.codex/config.toml` holds `project_doc_max_bytes = 65536`. Both files are listed in init's output.
- Onboard in an existing mem project creates or updates both Codex files and publishes them with the project files commit. It leaves an ignored or locally edited file uncommitted with the same wording as `.claude/settings.json`, and it never changes an existing `project_doc_max_bytes` or the other content of either file.
- `[agent_hooks] compact = false` removes the compaction entry from both `.claude/settings.json` and `.codex/hooks.json`, and `prompt = false` does the same for the prompt entry. A schema 2 project with `[claude] compact_hook = false` is upgraded to schema 3 with `[agent_hooks] compact = false` and no `[claude]` table.
- With `CURSOR_VERSION` set, `mem hook compact` and `mem hook prompt` print nothing, exit 0 and do not fetch.
- Onboard prints a ⚠️ line when `AGENTS.md` is over 40,000 bytes.
- With `GROK_HOOK_EVENT` set, both hook commands behave the same way.
- A real Codex run, `codex exec` in a scratch repository with the user's consent, since it uses their account, shows both hooks' output reaching the model: the prompt nudge from a forced due state, and the digest after `/compact` or with a `--dangerously-bypass-hook-trust` run.
- Focused tests for `internal/agenthooks`, `internal/project` (the patch) and the affected `internal/cli` tests pass; `go vet` is clean for those packages.
- README, `docs/design.md`, `docs/status.md` and `.mem/structure.md` describe the per-agent support, separating what is verified from what is not.

## Notes

- Sources (2026-10-25):
  - Codex: https://learn.chatgpt.com/docs/hooks (developers.openai.com/codex redirects there), https://learn.chatgpt.com/docs/agent-configuration/agents-md, and openai/codex at 970111f, with PRs #14626 (UserPromptSubmit), #19012 (hooks on by default), #21272 (SessionStart `compact`) and #34396 (the hook runs after mid-turn auto-compaction).
  - Grok Build: https://docs.x.ai/build/features/hooks, https://docs.x.ai/build/features/project-rules, github.com/xai-org/grok-build (v1.0.45 public; the hooks guide was read from the installed 1.0.67 binary), xai-org/plugin-marketplace#342 (the open request for `UserPromptSubmit` context).
  - Cursor: https://cursor.com/docs/hooks, https://cursor.com/docs/reference/third-party-hooks, and forum threads on `beforeSubmitPrompt` (no context injection, staff-confirmed 2026-04-24), on `sessionStart` (its `additional_context` dropped by a race, still reported in 3.24.9) and on Claude hooks without a `matcher` being skipped in 3.24.9.
- Codex trust: project `.codex/` layers load only in trusted projects, and each non-managed hook must also be approved in `/hooks` (hash-bound). mem must not write Codex's trust records (`[hooks.state]` in `~/.codex/config.toml`), because that would bypass the user's review.
- Codex cloud runs no command hooks; Cursor's cloud agents run project hooks except `sessionStart`. Both still read `AGENTS.md`.
- Codex reads skills from `.agents/skills`, where mem already installs them. Cursor reads `.claude/skills`. Grok reads both.
- `superagent-ai/grok-cli` (npm `grok-dev`) is a community agent, not xAI's. It shares the `~/.grok` and `.grok` paths and has its own hooks, which are not covered here.
- Slimming the managed block (the 9.4 KB of general principles) is a separate todo, not part of this spec.
