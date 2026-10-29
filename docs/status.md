# Status

What is implemented, what is planned, and known limits. [design.md](design.md) describes the intended behavior.

## Implemented

- `init`, which also creates and publishes missing development, staging and production branches (starting a repository without commits with an empty commit), and `import agent-core` (conversion from the Python coding harness).
- `onboard`: Git convergence, managed `AGENTS.md` refresh with a version stamp, the `.mem/install.md` guide that `AGENTS.md` points to when mem is missing, the `.mem/local/` ignore entry, project patch hook (no patches yet), hook installation, memories and skills changed by the sync, and context: structure doc, docs, runnables, active spec, open specs and todos with their age, the latest work log with a list of other recent ones, release status (flagging commits on staging or production that development lacks). Onboard and `memory set` warn when `AGENTS.md` is over Codex's default 32 KiB limit.
- `sync`: Git convergence and a push when the branch is ahead and not behind, template sync, and a report of incoming commits, work record changes, changed memories and skills, and a newer mem release. `task complete` commits the paths the agent names (or, without paths, all working-tree changes) with the task record (notes as the message) and lists what it left uncommitted and `spec complete` (after a sync that must find the branch in step, conflicts resolved) its archived record, and both run the same catch-up and push; `log new` prints drift nudges.
- Claude Code compaction hook: `init`, `import` and onboard maintain a `SessionStart`/`compact` entry in `.claude/settings.json` that runs `mem hook compact` (the `mem sync` catch-up with a short digest); `[claude] compact_hook = false` removes it.
- Cursor: `init`, `import` and onboard maintain a `stop` entry in `.cursor/hooks.json` that runs `mem hook stop`. When a turn completes with a work log due, it sends a follow-up asking for the log, once per state of the work; `[cursor] stop_hook = false` removes it. `mem hook compact` and `mem hook prompt` do nothing when Cursor (`CURSOR_VERSION`) runs them. This has been verified with simulated hook input, not in a real Cursor. Cursor has no way to add context after compaction or with each message.
- Session titles: agents are asked to give sessions a short title for the work where they can rename them (the Claude desktop app): by the onboard instruction once the user has said what to work on, repeated at the end of a context written to `.mem/local/onboard.md`; the compaction digest and `log new` have the agent look up the current title and rename the session when the work has moved on.
- Specs, tasks, todos, work logs (facts only, closed with `log commit`; `log new` records the commits covered and lists the files affected from Git) and memories.
- `structure` with drift detection against the last commit that touched `.mem/structure.md`.
- Runnables in `.mem/runnables/`.
- Checkpoint nudges: mem measures the user's work since their last log or completed task (work commits, and changed files and lines that count as work, uncommitted ones included, with `[structure] include`/`ignore` patterns) and suggests a log at the next natural stopping point, then says one is due. A `post-commit` hook in every project reports it with unpushed commits and a stale structure doc; a Claude Code `UserPromptSubmit` hook reports it with each user message once a log is called for; the compaction digest reports it too, and the compaction catch-up pushes committed work when safe. Work logs are checkpoints, not session ends.
- Todos have no claims: open until done, then deleted (schema 2; onboard's patch removes old claim fields).
- Onboard lists what has probably stopped being true (⏳ CHECK THESE: active specs unchanged for 14 days, unmerged remote branches idle for 14 days) and the age of unreleased work.
- Worktrees: onboard flags a checkout that is a linked Git worktree, naming the main checkout, and asks the agent to tell the user.
- The managed block opens with Ground Rules (onboard first, development branch and mem pushes only, no PRs unless asked, staging and production untouched, release only when asked, regenerate generated files, one package manager).
- A push rejected because others pushed first is caught up and retried, up to three attempts. Local commits that include a merge are never rebased; mem nudges to merge instead.
- `[git] base`: onboard on a long-lived development branch built on released code reports how many commits of the base it lacks, with the merge-only rule.
- Single-branch projects (`[git] single_branch = true`, set by `init --single-branch`): no staging or production, `promote` and `deploy` refuse, no release status at onboard, a managed-block variant without release instructions, and the base branch guarded in their place by `mem sync` and the protection hooks.
- Feature branches follow development: `mem sync`, onboard, record completion and the compaction catch-up rebase a feature branch onto a newer development branch when that is clean, force-push it with a lease when only the user has committed to it, and otherwise nudge; a rewritten upstream is caught up with `--fork-point`.
- `promote staging|production` with date tags carrying a generated summary, or drafted release notes confirmed with `--confirm` (`[release] notes`), optional production pull requests completed by fast-forward (`[release] production_pr`), `deploy` for one-step releases, and the `pre-push`/`pre-commit` hooks.
- Templates: `init --template` (init without `--template` or `--no-template` only lists the templates for the user to choose), `template use|list|promote|reset`, and template sync at onboard with `.mem/templates.lock` of memories, skills, docs and runnables, recording a todo to check existing code whenever a sync adds or changes a template memory, including while a setup is pending, which also asks the user whether to adopt or drop a convention whose tooling the project has not adopted, and a todo to apply each template change note (`<template>/changes/<name>.md`) the project has not seen, recorded in the lock, except for templates `init` or `template use` install; one-time template setups installed by `init` to `.mem/setup.md` and led by onboard and the compaction digest until the file is deleted.
- Self-update: onboard installs a newer release automatically; `mem update` on demand.
- Releases: published on GitHub with binaries for Linux, macOS and Windows, built by GoReleaser from semver tags.

Tested with disposable repositories and bare remotes; CI runs on Linux, macOS and Windows. mem manages its own source repository and praxis-app. The importer has been run end to end on a copy of Urbion-AI (22 specs, 8 todos, 120 logs from a year of harness use) and of studii-tan (an earlier harness version, three contributors, 38 specs, 111 logs).

## Planned

- **Template content.** The library at github.com/Benjamin-van-Heerden/mem-templates has nextjs-web, tanstack-start, python, rust, rust-gpui and phoenix; profiles for further kinds of project (for example Go) are added as they are needed.

## Known limits

- Hooks and runnables rely on `sh` and executable bits, the compaction hook's command assumes bash (Claude Code falls back to PowerShell on Windows without Git Bash), and template skills are linked into `.claude/skills` with symlinks, and the PowerShell steps in `.mem/install.md` (also the route from Git Bash) have not been run; Windows is supported tentatively and its behavior is untested.
- Not verified: whether Cursor launched from the Dock finds `mem` on PATH for its hooks; how Cursor treats the Claude `SessionStart` `compact` matcher.
- `git push --no-verify` and merges made in a Git host's web UI bypass the hooks. Promotion detects the resulting divergence and prints the recovery steps.
- Work logs imported from the Python harness are interpreted in the importing machine's local time zone.
- A work log's `commits` are recorded when `log new` runs. When `log commit`'s sync then rebases unpushed commits onto a teammate's, their hashes change and the recorded ones no longer resolve.
