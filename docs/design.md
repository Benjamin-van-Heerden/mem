# mem design

Target design for the Go rewrite. Where the implementation differs, [status](status.md) says so.

## Purpose

mem automates context building for coding agents and standardizes how work is done in a repository. It is one executable on PATH; each repository carries its own state in `.mem/` and `AGENTS.md`.

Audience: solo developers and small teams in daily contact. mem guides and nudges; it does not police every eventuality or make deliberate actions difficult.

## Principles

- **One shared codebase.** Everyone works on substantially the same code at all times, apart from uncommitted changes. mem converges automatically where that is safe (fast-forward, rebasing unpublished commits) and nudges firmly where it is not (divergence, long-lived branches, unpushed work).
- **Onboard builds context.** One command syncs, applies updates and produces everything an agent needs to start working.
- **Stdout directs the agent.** Clear headings, concrete commands, state-specific instructions. No disclaimers, no generic footers, no commands that do not exist.
- **Agents act autonomously within the user's request.** No approval flags or per-task confirmation loops.
- **Plain files, ordinary Git.** Markdown with YAML frontmatter, readable slugs, no locks, no databases, no GitHub issue mirroring.

## Repository layout

```text
AGENTS.md                           managed <mem> block, <memories> block, user content
.mem/
  config.toml
  structure.md                      living codebase and structure map
  docs/*.md                         project documents, included in onboard
  runnables/*                       executables whose stdout is included in onboard
  specs/<slug>/spec.md
  specs/<slug>/tasks/NN_<slug>.md
  specs/archive/<slug>/...          completed and abandoned specs
  todos/<slug>.md
  logs/<user>_<YYYYMMDD>_<HHMMSS>.md
  local/                            ignored; generated onboard output
```

```toml
schema = 1
name = "praxis-app"
description = "One line about the project"

[git]
remote = "origin"
development = "dev"
staging = "test"
production = "main"
protect = true          # install hooks that keep staging and production promotion-only

[structure]
ignore = ["migrations/**"]   # optional extra exclusions from drift detection
```

Identity is `git config user.name`, slugified.

## Onboard

1. **Converge.** Fetch with a time limit. On failure, warn prominently and continue with local context.

   | State | Behavior |
   | --- | --- |
   | Up to date | Nothing |
   | Behind | Fast-forward. Works with uncommitted changes unless incoming commits touch the same files. |
   | Unpushed local commits and remote moved | Clean tree: rebase onto upstream. Conflict: abort and instruct the agent to raise it with the user. |
   | Ahead only | Nudge to push at the next sensible point |
   | Uncommitted changes block convergence | Nudge: commit, then `mem sync` |
   | Not on the development branch | Report lag behind `origin/<development>`; nudge to integrate soon |

2. **Update.** Refresh the managed `AGENTS.md` block, apply pending project patches, and commit and push these mem-owned paths.
3. **Build context,** in this order: project, structure doc (with a drift warning when stale), docs, runnable output, active specs in full with pending tasks, other open specs and todos as one-liners, recent logs (current user first), git summary, and a final state-specific agent instruction. Memories are not repeated; they are already in `AGENTS.md`. Output over ~14k characters goes to `.mem/local/onboard.md` with an instruction to read all of it.

`mem sync` performs step 1 on demand. Other commands print a short divergence nudge when local Git state shows drift.

**Commit rhythm.** Commit each coherent, working change (typically one per task or fix); `task complete` asks for it. Push at the end of each session with the work log, after completing a spec and before a promotion. Uncommitted work over 15 code files or 800 lines triggers a nudge to commit the finished parts.

## Work records

- **Spec:** larger planned work. Statuses `draft → active → completed | abandoned`. Body template: Overview, Goals, Technical Approach, Success Criteria, Notes.
  - `spec new "title"`, `spec list`, `spec show <slug>`
  - `spec start <slug>`: assigns the current user and marks it active; commits and pushes the change
  - `spec complete <slug>`: requires all tasks done; tells the agent to verify the Success Criteria against the code, write a log and commit the work; archives the spec
  - `spec abandon <slug> --reason "..."`: archives the spec
- **Task:** ordered steps within a spec. `task new "title" "description" [--spec]`, `task list [--spec]`, `task complete <slug> "notes" [--spec]`. `--spec` defaults to the user's single active spec. Completion prints remaining tasks and directs the agent to continue.
- **Todo:** a standalone matter needing attention. `todo new "title" "description"`, `todo list`, `todo show`, `todo claim` (commits and pushes), `todo delete`.
- **Memory:** a lasting convention in the `AGENTS.md` memories block. `memory set <name> "<instruction>"`, `memory list`, `memory remove <name>`.

Arguments accept a slug or an unambiguous title.

## Session continuation

Two layers:

1. **Records** hold where planned work stands. Onboard renders the active spec and its pending tasks, so the next session knows what remains without any note.
2. **Logs** hold the narrative: what was done, what failed, decisions, what comes next. They cover ad hoc work too. `log new [--spec]` creates a templated file for the agent to fill in; logs are written at the end of a session or when context is getting long, with the user's agreement. `log list`, `log show`.

## Structure doc

`.mem/structure.md` is a living document: agents keep it current incrementally rather than regenerating it.

- `mem structure` without an existing file scaffolds the template, prints a file tree and instructs a full research pass.
- With an existing file, it lists code changes since the baseline and instructs the agent to update only the affected sections.

The baseline is the last commit that touched `.mem/structure.md`; uncommitted edits to it count as up to date. Committing an update moves the baseline, so no stamping step is needed.

Drift is measured from the baseline to the working tree, excluding Markdown, `.mem/`, `AGENTS.md`, lockfiles, binary and generated files, and configured globs. More than 5 changed code files or 1000+ changed lines produces a warning: one line in onboard, and an explicit instruction in `log new` to update the structure doc.

## Runnables

Each executable in `.mem/runnables/` runs at onboard from the repository root with a time limit. Its stdout appears under a heading named after the file. Failures print a short note and never stop onboard. Use them to extract focused context such as design tokens, UI components or database models, instead of including whole files.

## Branches and promotion

Three branch roles, named per project (defaults `dev`, `test`, `main`):

- **Development** is where all work lands. Pushing to it should not trigger deployments.
- **Staging** holds preview releases. CI deploys it.
- **Production** holds releases. CI deploys it.

Staging and production only ever fast-forward to commits that already exist on development, so there is exactly one history and nothing reaches production without having been previewed.

- `mem promote staging [--to <commit>]` fast-forwards staging to `origin/<development>`, or to an earlier development commit to leave unfinished work out.
- `mem promote production` fast-forwards production to `origin/<staging>`. The first run prints what will ship (commits, authors, specs completed in the range, specs still in progress) and asks for release notes; `--notes <file>` then pushes production together with an annotated date tag (`v2026.09.24.1`) carrying the notes, atomically.
- `mem init` creates missing branches in promotion order: staging from production, development from staging, tracking the remote's copy where one exists. It publishes any branch the remote lacks and switches to development.
- If staging or production has commits that are not on development (a hotfix or a web-UI merge), promotion stops and prints the commands to merge them back into development.
- Onboard shows release status: the latest production tag and how far staging and development are ahead.

**Hooks.** With `protect = true`, `init` and onboard install `pre-push` and `pre-commit` hooks in the repository's hooks directory. They refuse pushes to staging or production that do not come from `mem promote`, and commits made on those branches. The hooks call `mem hook <name>` and do nothing where mem is not installed. Existing non-mem hooks are never overwritten; onboard explains how to add the call instead. `protect = false` removes the hooks for solo projects. `git push --no-verify` and web-UI merges bypass hooks; promotion detects the resulting divergence.

## Updates

- The managed `AGENTS.md` block is refreshed from the executable on every onboard.
- Project patches are numbered migrations keyed by `schema`, applied once at onboard.
- A newer-executable check at onboard is planned.

## Import

`mem import agent-core` converts a Python coding-harness project: config, memories, docs, specs and tasks, todos, logs and its structure doc map almost directly onto this layout. The original `.agent_core/` is left in place for the user to remove.
