---
created_at: "2026-09-28T12:05:44+02:00"
user: benjamin_van_heerden
---

# Work Log - mem manages itself; templates, self-update and facts-only logs

<!-- A work log records what happened in this session, as fact. It is not updated later. Anything still to be done, including blockers and decisions waiting on the user, belongs in a todo, not here. -->

## Overarching Goals

Start developing mem with mem (dogfooding), fixing problems and unclear stdout found along the way; build project templates with promotion between projects; add a self-update and release regularly; rethink work logs so they stop carrying stale "pending" items.

## What Was Accomplished

### mem manages its own repository (2026-09-25)

- `mem init` now creates missing branches in promotion order (staging from production, development from staging, tracking the remote where a branch exists), publishes them and switches to development, as the Python harness setup did. The managed `<mem>` block now leads AGENTS.md.
- This checkout was initialized: `dev` → `test` → `main`, `protect = true`. `rename-mem` was deleted; the per-phase `rework` history is kept as the local tag `archive/rework`. The hand-written duplicate of the general guidelines was removed from AGENTS.md.
- `.mem/structure.md` was written.

### Project templates (spec `project_templates`, archived)

- `internal/templates`: library clone in the user cache (`<repo>-<hash>` names, short for Windows), `Sync`/`Status` with `.mem/templates.lock`, `Promote`, `Reset`. Hashes ignore carriage returns so CRLF checkouts match.
- Commands: `init --template`, `template use|list|promote|reset`; template sync at onboard with publishing.
- Library github.com/Benjamin-van-Heerden/mem-templates (public) seeded with nextjs-web, tanstack-start, python, rust and phoenix, from the old harness guides turned into skills; the one studium module name in the Phoenix guide was replaced with MyAppWeb.
- The user-level config file was dropped in favour of a built-in default library URL (`templates.DefaultSource`).
- Verified against GitHub with a temporary library repo (deleted afterwards): memory, doc and skill promoted from one project arrived in another at onboard; later promotions updated unedited items and flagged a two-sided edit; `reset` resolved it; promoting to octocat/Hello-World was refused with 403 and left no trace.

### Self-update and releases

- `internal/selfupdate`: latest tag from the `releases/latest` redirect, checksum-verified download, atomic replace; onboard updates and re-runs itself (`MEM_NO_UPDATE=1` stops it); `mem update` on demand; dev builds never update.
- Released v0.3.0, v0.3.1, v0.3.2, v0.4.0 and v0.4.1 via `mem promote staging`, `mem promote production --notes`, then a semver tag. Verified live: v0.3.0 → v0.3.1 at onboard, v0.3.1 → v0.3.2 via `mem update`, v0.3.2 → v0.4.0 at onboard, v0.4.0 → v0.4.1 via `mem update`.

### Facts-only logs and `mem log commit` (spec `logs_record_facts_log_commit_closes_the_session`, archived)

- Log template without a future-work section; `log new` lists open todos and prompts to delete completed ones and record open work as todos.
- `mem log commit` refuses unfilled placeholders, commits `.mem/` records, syncs, pushes and reports uncommitted work outside `.mem/`.
- Onboard shows todo age and claimer, the user's latest log in full and other logs of the last 14 days by title, and summarizes open work from specs, todos and release status.

### Fixes found by dogfooding

- Block tags in AGENTS.md are matched only on their own lines (a backticked mention had broken onboard).
- Onboard no longer reports commits it has just pushed as unpushed; mid-session checks omit the unpushed nudge.
- mem's own pushes time out after 30 s and a failed push is reported as such, not as a failed commit.
- `task complete` names the task record to commit and prints the next task in full; `todo claim` says how to finish a todo.
- Onboard tells the agent to re-read AGENTS.md when it changed it.
- Slugs split on slashes, dots and colons.
- Promote links hand-made skills into `.claude/skills`, lists every path to commit, and explains refused pushes.

## Decisions

- Templates live in a separate Git repository; promotion is a push to it, so only people with write access (the user) can promote. Others fork and set `[templates] source`.
- Skills are installed in `.agents/skills/<name>` with a relative `.claude/skills/<name>` symlink; framework guides are skills rather than docs because docs are printed at every onboard.
- Template sync adds, updates unedited items and flags two-sided edits, never deletes; deleting an item records an opt-out in `[templates] exclude`.
- No user-level config: the default library URL is compiled in; the project records its library.
- Work logs are statements of fact; open work, blockers and pending decisions are todos. The prompt to close todos at session end is a light instruction, not an enforced step. The log is kept because `mem log commit` is where the session converges with the shared codebase.
- mem's justification rests on shared, reviewable state across people, agents and projects, and on Git convergence, not on working around context limits.
- mem releases: date tags from `mem promote production` plus a manual `vX.Y.Z` tag; the release workflow builds only semver tags with `GORELEASER_CURRENT_TAG` set.

## Key Files Affected

- `internal/cli/init.go`: `ensureBranches`, `--template`, `--template-source`.
- `internal/cli/template.go`, `internal/templates/{library,sync,promote}.go` and tests: templates.
- `internal/selfupdate/selfupdate.go`, `internal/cli/update.go`: self-update.
- `internal/cli/onboard.go`: auto-update, template sync, `afterUpdates`, log selection, todo ages, instruction changes.
- `internal/cli/log.go`, `internal/work/log.go`: facts-only template, `log commit`, `Unfilled`, `Heading`, `LatestLog`.
- `internal/git/git.go`: `Commit`, `Push` with `PushTimeout`, `ErrPush`.
- `internal/agentsmd/agentsmd.go`, `instructions.md`: block placement and whole-line tags; instruction updates.
- `internal/converge/converge.go`: `Unpushed`.
- `.github/workflows/release.yml`: semver-only trigger, `GORELEASER_CURRENT_TAG`.
- `docs/design.md`, `docs/status.md`, `README.md`, `.mem/structure.md`, `AGENTS.md`.

## Errors and Barriers

- The first v0.3.0 release failed: GoReleaser picked the date tag on the same commit. Fixed with `GORELEASER_CURRENT_TAG`; the failed tag was deleted and re-created.
- Windows CI failed twice: the library cache path exceeded the path length limit (full URL as directory name), and CRLF checkouts made file comparisons and hashes differ. Both fixed.
- One chain of commits stalled for over two minutes on a push with no time limit, which led to the push timeout.
- `~/.config/mem/` still holds dead symlinks from the old Python mem (`config.toml`, `templates/*.md`) pointing into the dotfiles; left untouched.
