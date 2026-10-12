---
created_at: "2026-10-12T11:02:02+02:00"
user: benjamin_van_heerden
---

# Work Log - Size-based work-log nudges, a prompt hook, path-scoped task commits and checkpoint-shaped logs

## Overarching Goals

Make work logs real checkpoints: nudge for them by the size of the work rather than the number of commits, reach the agent with each user message as well as after commits, keep parallel sessions in one checkout out of each other's task commits, and shape the log template and onboard for frequent logs.

## What Was Accomplished

### Evidence that led to it

- Work between logs, measured from Git: in this repository, before the post-commit nudges, gaps of 18 commits / 2,405 lines, 13 / 1,587 and 12 / 1,509; since about 2026-10-08, 1–3 commits and 9–653 lines. praxis-app had 4 commits, 46 files and 2,922 lines since mem was set up and no log: the commit count (3 suggest, 5 require) never required one.

### Path-scoped task commits (452408d)

- `mem task complete <task> "<notes>" [<path>...]`: with paths, `commitTask` commits them with the task record through `git.Commit` (`git commit -- <paths>`, so other staged changes stay out), then `leftovers` lists what remains (`git status --porcelain --untracked-files=all`), new files first, and the instruction asks for a follow-up commit `Add files missed from <task title>` for anything of the agent's own. Without paths it still uses `git.CommitAll`.
- Instructions (`internal/agentsmd/instructions.md`) and `spec start`'s instruction tell agents to name the files they changed, since another session may work in the same checkout.

### Size-based nudges and the prompt hook (5a34994)

- `checkpoint.Since` returns `Work{Commits, Files, Lines, LatestIsWork}`: the user's work commits since their last checkpoint and the relevant files and lines those commits (numstat) and the uncommitted changes (`structure.ChangesSince(HEAD)`) touched; teammates' commits and `.mem/` never count. `history` reads 100 commits first, then up to 1,000 when no checkpoint is among them.
- `Work.Level`: `Suggested` from 5 commits, 10 files or 500 lines; `Due` from 10, 20 or 1,200. `MidTurnLine` (post-commit hook, compaction digest) never says "now": "write a work log at the next natural stopping point" / "a work log is due: … such as when the current change is done". `TurnStartLine` (prompt hook) at Due: write one before taking up substantial new work, finish a change in progress first, answer a quick question first, stop if asked. Every line says a log does not end the turn.
- `mem hook prompt` prints `TurnStartLine` and nothing below the thresholds. `claude.SyncHooks(root, Hooks{Compact, Prompt})` replaces `SyncCompactHook` and manages the `SessionStart`/`compact` and `UserPromptSubmit` entries (timeout 10 s); `[claude] prompt_hook = false` removes the latter. Onboard reports "Updated mem's Claude Code hooks in .claude/settings.json: …".
- `[structure] include` brings back files the built-in rules leave out (for example `*.md`); `.mem/`, `AGENTS.md` and `CLAUDE.md` never count, and `ignore` wins over `include`. Patterns without a slash match file names; others match whole paths with `**` for any number of directories (`globMatch`). The same filter drives structure-doc drift.
- Measured: the prompt hook takes about 0.2 s in mem and praxis-app (0.65 s on a cold cache), almost all of it the one `git log` call.

### Checkpoint-shaped logs (6b3ddfe)

- Template: `## Goal` (one or two sentences) replaces Overarching Goals; `## Notes` (optional: errors, barriers, approaches tried and dropped) replaces Errors and Barriers; What Was Accomplished keeps its name because release notes read its headings.
- `checkpoint.Cover` lists the user's own commits since the checkpoint (oldest first) and every file outside `.mem/` they and the uncommitted changes touched (`structure.Diff`, unfiltered, split out of `ChangesSince`). `log new` writes them to the frontmatter (`commits: [..]`, flow style) and under Key Files Affected as `` - `path` (N lines) `` (at most 30), followed by a placeholder asking the agent to annotate, remove other sessions' files and add files outside the repository.
- Onboard's `recentLogs` shows up to three of the user's logs from the last 14 days in full while they stay within 12,000 bytes (the latest always), labelled "Your latest log." and "An earlier log of yours.", and lists the rest.
- Verified with focused tests in `cli`, `checkpoint`, `claude`, `structure` and `work`, `go vet`, and `dist/mem-dev` end to end in a throwaway repository (init installs both hooks; prompt hook and `log new` output as intended).

## Decisions

- Logs keep the name "work log": "checkpoint" says when one is written, not what it is; Claude Code already uses "checkpoints" for rewind; and a rename would touch commands, `.mem/logs/` in every project, the importer and release notes.
- Logs stay honest records of work as things stand; they still never hold open work, so a mid-task log states facts without promising what remains.
- A file left out of a path-scoped task commit goes in a follow-up commit after the pushed task commit, rather than a confirmation step before committing; a briefly red CI run on a forgotten new file is the accepted cost.
- Two sessions editing the same file cannot be separated by mem; per-session edit tracking through Claude Code hooks was rejected (misses Bash-made changes, no reliable session identity), as was any session locking.
- One filter (`[structure]`) defines what counts as work for both the structure doc and the nudges; no new config section, since no project under yhat/ sets `[structure]`.
- Thresholds are constants, like the structure doc's staleness rule; uncommitted changes from another session in the checkout count toward the nudge, accepted since it is only a nudge.
- No spec: the user judged the change small enough to do directly. Log search stays unplanned (not pressing).

## Key Files Affected

- `internal/cli/task.go`, `task_test.go`: path arguments, `commitTask`, `leftovers`; parallel-session test.
- `internal/checkpoint/checkpoint.go`, `checkpoint_test.go`: `Since`, `Work`, `Level`, `MidTurnLine`, `TurnStartLine`, `Cover`, `sinceCheckpoint`; numstat/summary history with hashes.
- `internal/structure/structure.go`, `structure_test.go`: `Relevant(file, StructureConfig)` with include/ignore and `**`, `Diff`, exported `RenamedTo`.
- `internal/project/project.go`: `StructureConfig.Include`, `ClaudeConfig.PromptHook`.
- `internal/claude/settings.go`, `settings_test.go`: `SyncHooks`, `syncEntry`, `PromptCommand`.
- `internal/cli/hook.go`, `hook_test.go`: `mem hook prompt`; `compact.go`, `compact_test.go`: digest on `Work`; `onboard_updates.go`, `init.go`, `import.go`: `SyncHooks` and `claudeHooksLine`.
- `internal/work/log.go`, `work_test.go`: template, `LogMeta.Commits`, `NewLog(..., commits, files)`, `keyFiles`.
- `internal/cli/log.go`, `log_test.go`, `onboard.go`: `log new` fills from `Cover`; `recentLogs` returns several logs in full.
- `internal/agentsmd/instructions.md`, `internal/cli/spec.go`, `docs/design.md`, `docs/status.md`, `README.md`, `.mem/structure.md`.

## Errors and Barriers

- The first session title included the project prefix ("mem · …"), picked up from an older log's title that describes the dropped `<project> · <the work>` convention; renamed to "Checkpoint log nudges". The source has no prefix.
- `TestLogNewListsTheFilesAndCommitsSinceTheLastLog` failed intermittently: a short hash made only of digits is quoted in YAML (`"5223436"`), so comparing the raw frontmatter text failed; the test now reads the frontmatter back.
- The cli test helper gives every clone the same email, so a "teammate" commit counted as the user's until the test committed with `-c user.email=ana@example.com`.
- Claude Code's docs do not say whether `UserPromptSubmit` fires for a message sent while the agent is mid-turn (community reports suggest it may), so the Due wording also covers being mid-change.
