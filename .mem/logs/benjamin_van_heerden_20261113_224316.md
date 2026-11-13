---
created_at: "2026-11-13T22:43:16+02:00"
user: benjamin_van_heerden
commits: [b53bc6e, 40cdada, f075779, b7c829d]
---

# Work Log - .mem/README.md guide, checks at promotion, tool configuration rule

## Goal

Give every project a guide to how mem works, run the project's checks at promotion instead of before each commit, add the agent tool configuration rule to the managed instructions, and clear the todos the user settled.

## What Was Accomplished

### Todos

The user settled the open todos. `add_rust_gpui_and_go_templates`, `check_cursor_hooks_real`, `cloud_agent_support`, `check_codex_hooks_windows` and `hook_support_grok` were deleted, because Cursor and cloud agents were confirmed to work and the rest were dropped. `consider_replacing_mem_precommit_checks` and `agent_tool_configuration_rule_managed` were done and deleted with their work. `move_work_finance_folder_mode` stays open.

### Checks at promotion (b53bc6e)

- `mem precommit` became `mem check` (`internal/cli/check.go`), and `[precommit]` became `[checks]` (`project.ChecksConfig`). `runChecks` now takes a `retry` phrase for its failure instruction, such as "run `mem promote staging --to <commit>` again".
- Schema 3: `renamePrecommitToChecks` is a no-op patch. `ParseConfig` moves a legacy `Precommit *ChecksConfig` (`toml:"precommit"`) into `Checks` and clears it, so `Upgrade`'s `WriteConfig` writes the config without it. An unpatched project still runs its checks.
- Single-branch projects: `task complete` (before the record is written) and `spec complete` (before the sync) run the checks and refuse on failure.
- The managed instructions no longer ask for checks before each commit. A `.Single` branch names who runs them, and `mem check` runs them on request. The test-suite rule now refers to the project's checks.
- Tests: `TestOnboardRenamesTheSchema2PrecommitTableToChecks`, `TestSingleBranchCompletionRunsTheChecksAndRefusesWhenTheyFail` (toggled with `git rev-parse --verify --quiet refs/tags/ok` so it works in every shell), and the existing check and upgrade tests updated for schema 3.

### .mem/README.md (40cdada)

- Authored in `internal/agentsmd/readme.md` and embedded. `agentsmd.Guide{Path, About}` with `Write` replaces `WriteGuide`/`GuidePath`. `Guides = {InstallGuide, Readme}`. Onboard and `applyFolderUpdates` loop over them, and `WriteGuides` serves `init`, `init --folder` and `import`.
- Contents: the `.mem/` layout, AGENTS.md's three parts, every config key, onboard's steps, the structure doc, docs, runnables (limits, shebang, an example), specs, tasks, todos, logs and the nudge thresholds, memories and skills, templates, checks, branches and releases, agent and Git hooks, `.mem/local/`, and folder projects.
- The managed block and the folder block both point to it after the `--help` line.

### Agent tool configuration rule (f075779)

A new `## Agent Tool Configuration` section in General Principles, from the todo's draft, with `npm run dev` as the example script.

### Memory (b7c829d)

`mem-guide`: keeping `internal/agentsmd/readme.md` current is imperative. Every behaviour change updates it in the same commit, and it holds implemented behaviour only.

### Verified

Focused tests, then `mem precommit` (installed v0.14.1, still the old name here) before each code commit, all passing. `dist/mem-dev` in scratch repositories rendered the checks sentence for staged and single-branch projects, wrote `.mem/README.md` at init and onboard, and rendered the new rule.

## Decisions

- **`[checks]` + `mem check`** (user's choice). The command stays for on-request runs and after a failed promotion.
- **Single-branch projects check at `task complete` and `spec complete`** (user's choice), because they have no promotion. Ordinary commits outside a spec get no checks there.
- **One guide for Git and folder projects**, with a closing section on what folder projects leave out, rather than a second file to maintain.
- **The guide is managed and rewritten at onboard**, like `install.md`, so project edits to it are lost.

## Key Files Affected

- `internal/agentsmd/readme.md`: the new guide
- `internal/agentsmd/agentsmd.go`: `Guide`, `Guides`, `WriteGuides`
- `internal/agentsmd/instructions.md`, `instructions_folder.md`: checks wording, guide pointer, tool configuration rule
- `internal/cli/check.go` (was `precommit.go`), `check_test.go` (was `precommit_test.go`)
- `internal/cli/task.go`, `spec.go`: single-branch checks
- `internal/cli/promote.go`, `deploy.go`, `root.go`: renamed command and retry phrases
- `internal/cli/onboard_updates.go`, `folder.go`, `init.go`, `import.go`: write both guides
- `internal/project/project.go`, `patches.go`: `[checks]`, schema 3
- `internal/cli/onboard_test.go`, `single_branch_test.go`, `log_test.go`, `folder_test.go`
- `README.md`, `docs/design.md`, `docs/status.md`, `.mem/structure.md`
- `AGENTS.md`: the developer section's checks line and the `mem-guide` memory

## Notes

- This repository's own `.mem/config.toml` keeps `[precommit]`, and its `.mem/README.md` does not exist yet. The installed v0.14.1 reads `[precommit]`, and onboard with the next release applies the schema 3 patch and writes the guide.
- The managed block in this repository's `AGENTS.md` still says `mem precommit` until that onboard rewrites it.
