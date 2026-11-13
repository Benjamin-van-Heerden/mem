# How mem works

<!-- Managed by mem. Replaced on onboard; change it in the mem source, not here. -->

This project uses mem, a command-line tool that builds context for coding agents and keeps the project's work records in plain files next to the code. This guide explains what mem keeps in the project and how each part behaves. The working rules for agents are in the managed block of `AGENTS.md`. Every command has `mem <command> --help`, and mem's own output says what to do next in the current state.

Some parts apply only to Git projects. A folder project (`mode = "folder"`) has no Git, and the [last section](#folder-projects) says what it leaves out.

## What mem adds to a project

```text
AGENTS.md                  agent instructions: mem's managed block, the project memories, then your own content
.mem/
  README.md                this guide, rewritten by mem at onboard
  install.md               how to install mem; AGENTS.md points here when mem is missing
  config.toml              project configuration (see below)
  structure.md             living map of the codebase, included in onboard
  docs/<name>.md           project documents, printed in full at every onboard
  runnables/<name>         scripts for the commands under [runnables]
  specs/<slug>/spec.md     planned work; its tasks are in specs/<slug>/tasks/NN_<slug>.md
  specs/archive/<slug>/    completed and abandoned specs
  todos/<slug>.md          open matters, one per file
  logs/<user>_<date>_<time>.md   work logs
  templates.lock           template items and the content they last shared with their template
  setup.md                 a template's one-time setup, deleted when every step is done
  local/                   ignored by Git; output for this checkout only
.agents/skills/<name>/     skills, linked from .claude/skills/<name>
.claude/settings.json      Claude Code hooks (mem's entries only; other settings are kept)
.codex/hooks.json          Codex hooks
.cursor/hooks.json         Cursor hook
```

Everything here except `.mem/local/` is committed and shared. Most of these appear only once something uses them: `init` writes the config, the two guides, `AGENTS.md`, the hook files and the `.gitignore` entry, and the rest arrives with `mem structure`, the first spec, todo or log, a template, or a file you add yourself.

### AGENTS.md

`AGENTS.md` holds three parts, in order:

- **The managed block** between `<mem>` and `</mem>`. Onboard rewrites it from the installed mem, so edits to it are lost. Change mem itself to change it.
- **The memories block** between `<memories>` and `</memories>`. These are lasting project conventions, one `## <name>` section each. Change them with `mem memory set <name> "<text>"` and `mem memory remove <name>`; `mem memory list` shows them. Some agents read only the first 32 KiB of `AGENTS.md` (Codex by default), so mem warns when the file is larger.
- **Your own content** after both blocks, which mem never touches.

## config.toml

```toml
schema = 3                  # project format; onboard upgrades older projects
name = "shop"
description = "One line about the project, shown at the top of onboard"

[git]
remote = "origin"
development = "dev"         # where all work lands
staging = "test"            # preview releases; CI deploys it
production = "main"         # releases; CI deploys it
protect = true              # Git hooks keep staging and production promotion-only
# single_branch = true      # instead of staging and production: no releases (see Branches and releases)
# base = "main"             # on a long-lived branch built on released code: take that branch in by merge only

[checks]
commands = ["npm run lint", "npm test"]   # see Checks

[runnables]
routes = "bun scripts/print-routes.ts"    # see Runnables

[structure]
ignore = ["migrations/**"]  # files that do not count as work
include = ["docs/*.md"]     # files that count although the built-in rules leave them out

[templates]
source = "https://github.com/Benjamin-van-Heerden/mem-templates.git"
use = ["base", "nextjs-web"]  # later templates win when two provide the same item
exclude = ["skill:old-guide"] # items you deleted; written by mem

[release]
# notes = true              # production releases wait for release notes the user confirms
# version_tags = true       # name releases with --version vX.Y.Z instead of a date tag
# production_pr = true      # production releases go through a GitHub pull request

[claude]
# compact_hook = false      # leave out the catch-up after compaction
# prompt_hook = false       # leave out the work-log nudge with each message

[codex]
# compact_hook = false
# prompt_hook = false

[cursor]
# stop_hook = false         # leave out the work-log follow-up at the end of a turn
```

Only `schema`, `name`, `description` and `[git]` are written by `init`; every other table is optional. A pattern without a slash matches file names anywhere, and one with a slash matches whole paths, with `**` for any depth. Changes to the config are ordinary commits: commit them so every checkout shares them.

## Onboard

`mem onboard` starts every session. In order, it:

1. Updates mem itself when a newer release exists (not with `--offline`), then runs the new version.
2. Leaves staging or production for the development branch, if the checkout is on one.
3. Fetches and brings the branch up to date: fast-forward, or a rebase of commits that were never pushed. When that is not safe, as with a conflict, it changes nothing and explains the problem.
4. Applies mem's updates: upgrades the project format, rewrites the managed block of `AGENTS.md`, this guide and `install.md`, keeps the hook entries and Git hooks in place, and syncs template items. It commits and pushes what it changed.
5. Prints the context: setup in progress, the structure doc, docs, runnable output, the active spec with its tasks, other open specs and todos, records that look stale, and recent work logs, followed by release status and an instruction for the agent. When the context is long, it goes to `.mem/local/onboard.md` and onboard tells the agent to read all of it.

`mem sync` repeats step 3 at any time, pushes committed work and reports what teammates pushed since the last fetch.

## The structure doc

`.mem/structure.md` is a map of the codebase: overview, layout, key modules, data flow, entry points, commands, conventions. Onboard prints it in full. Agents keep it current as they change what it describes, rather than regenerating it.

- `mem structure` creates it from a template, with a file tree for the first research pass.
- On an existing doc, `mem structure` lists the code changed since the doc was last committed, so only the affected sections are updated.
- Past 5 changed code files or 1,000 changed lines, onboard and `mem log new` flag it as out of date.

Markdown, `.mem/`, `AGENTS.md`, lockfiles, binaries and generated or vendored files do not count as code changes. `[structure] ignore` and `include` adjust this.

## Docs

Every `.mem/docs/<name>.md` is printed in full at every onboard, in name order, so each one costs context in every session. Use docs for what every session needs: product rules, a domain glossary, an integration's constraints. Guides that matter only for some tasks belong in skills, which agents load when they are relevant.

## Runnables

A runnable is a command whose output joins the onboard context under a heading with its name. Use runnables for focused context that changes with the code, such as the database models, the design tokens or the routes, instead of pasting whole files into a doc. They are listed under `[runnables]` in `.mem/config.toml`, one `name = "command"` each:

```toml
[runnables]
routes = "bun scripts/print-routes.ts"
models = "python3 .mem/runnables/models.py"
schema = "npm run --silent db:summary"
```

- At each onboard, every entry runs through the shell (`sh`, or `cmd` on Windows) from the repository root, in name order.
- The command can name any interpreter, call a script in `.mem/runnables/` or elsewhere, or reuse one of the project's own scripts. When the command names the interpreter, the script needs no executable bit or shebang.
- Each runs with a 15-second limit. Its stdout is included, cut off after 20,000 characters. Stderr is not shown, apart from the first line when it fails.
- A failure or timeout prints a ⚠️ line, and onboard carries on.

`.mem/runnables/` is the place for scripts written only for onboard. A file there does not run unless an entry names it. For example, `.mem/runnables/routes.sh`, listed as `routes = "sh .mem/runnables/routes.sh"`:

```sh
# The API routes, so agents know what exists without reading every handler.
grep -rhoE '(get|post|put|delete)\("[^"]+"' src/api | sort -u
```

Keep the output short: it is read at the start of every session. The interpreter a command names must be installed on every machine that runs onboard, including Windows ones if the team has them.

Runnables were files that ran by themselves until schema 3. The upgrade to schema 3 lists each executable file in `.mem/runnables/` as an entry that runs it as before, and records a todo to review the entries with the user.

## Work records

Records are Markdown files with YAML frontmatter, named by a readable slug made from the title. Commands accept the slug or an unambiguous title. They are for planned or tracked work: ordinary coding needs none of them.

### Specs and tasks

A spec is a larger piece of planned work: `.mem/specs/<slug>/spec.md`, with sections for Overview, Goals, Technical Approach, Success Criteria and Notes. Its tasks are ordered steps in `.mem/specs/<slug>/tasks/NN_<slug>.md`.

- A spec moves from `draft` to `active` to `completed` or `abandoned`.
- `mem spec new "<title>"` creates a draft. `mem task new "<title>" "<description>" --spec <spec>` adds a task.
- `mem spec start <spec>` assigns the spec to you, marks it active and publishes that.
- `mem task complete <task> "<notes>" <path>...` commits the named paths together with the task record. The task title is the commit subject and the notes are the body. It then syncs and pushes.
- `mem spec complete <spec>` refuses while tasks are pending or the branch is out of step with the shared codebase. Otherwise it moves the spec to `.mem/specs/archive/`, commits that, syncs and pushes.
- `mem spec abandon <spec> --reason "<why>"` archives a spec without completing it.

Onboard shows the active spec in full, with its pending tasks.

### Todos

A todo is a standalone matter that needs attention, including a blocker or a decision waiting on someone: `.mem/todos/<slug>.md`, with a title, a creation time and a description. It has no states: it is open until it is done or no longer relevant, and then it is deleted with `mem todo delete <todo>`, a removal you commit with the work. Onboard lists open todos with their age.

### Work logs

A work log records a stretch of work as fact: its goal, what was done, decisions, the files affected and notes. It is never updated afterwards, so anything still open goes into a todo instead.

- `mem log new` creates `.mem/logs/<user>_<YYYYMMDD>_<HHMMSS>.md` from a template, listing your commits and the files they touched since your last log or completed task.
- `mem log commit <path>...` refuses a log with unfilled placeholders. Otherwise it commits the log with the `.mem/` records you name, such as todos you added or deleted, then syncs and pushes.
- Logs are written at checkpoints: a completed spec, a release, the work-log nudge, the user wrapping up. A log is never the end of a session.
- mem nudges for a log as your work since the last checkpoint grows. A log is suggested from 5 commits, 10 files or 500 lines, and due from 10 commits, 20 files or 1,200 lines. The nudge comes after commits, with your messages, after compaction and, in Cursor, at the end of a turn.

Onboard shows your latest logs in full and lists other recent ones by title.

## Memories and skills

A memory is a lasting convention, kept in the memories block of `AGENTS.md`, so every agent reads it at the start of every session. It is not for progress on the current work.

A skill is a folder in `.agents/skills/<name>/` with a `SKILL.md` and anything it needs. It is linked from `.claude/skills/<name>`. Agents load a skill when its description fits the task, so a skill can hold a long guide without costing context in every session.

## Templates

A template library is a Git repository with a folder per template, holding memories, skills, docs, runnables and change notes, and optionally a one-time setup. `[templates]` in the config names the library and the templates this project uses.

- `mem init --template <name>` and `mem template use <name>` draw a template's items into the project. A template runnable is a script installed in `.mem/runnables/` with the executable bit, and installing it adds an entry that runs it directly (`name = ".mem/runnables/<name>"`), unless the project already has an entry by that name. A template runnable therefore needs a shebang.
- Each onboard and `mem sync` compares every item with its template. `.mem/templates.lock` records the content each item last shared with its template. mem adds missing items and updates the ones the project has not edited. It reports items edited locally or on both sides, and turns an item you delete into an `exclude` entry.
- A new or changed template memory becomes a todo to check the existing code against it, since a memory only guides new work. A template's change note that the project has not seen also becomes a todo.
- `mem template promote <memory|skill|doc|runnable> <name>` sends a project's item to its template, so other projects receive it at their next onboard. `mem template reset` takes the template's copy back, and `mem template list` shows where each item stands.
- `.mem/setup.md` holds the setup steps of the templates `init` used, with a checkbox heading per step. Onboard leads with it until every step is ticked and the file is deleted.

## Checks

`[checks] commands` lists the project's checks, such as its linter, type checker and tests. They run from the repository root, in order, and stop at the first failure.

- `mem promote staging` and `mem deploy` run the checks and refuse when they fail, so what reaches staging passed on the machine that promoted it. No release waits for CI, and agents need not run the checks before each commit.
- A single-branch project has no promotion, so `mem task complete` and `mem spec complete` run the checks there instead, and a failure leaves the task or spec open.
- `mem check` runs them on request.

## Branches and releases

Work lands on the development branch. Releases only fast-forward the staging and production branches along the development history, and CI deploys them, so there is one history and nothing reaches production without having been on staging.

- `mem promote staging` moves staging to the development branch, or with `--to <commit>` to an earlier development commit. It runs the checks first.
- `mem promote production` moves production to staging and tags the release with a summary of what shipped. `[release] notes`, `version_tags` and `production_pr` change how the release is named and recorded.
- `mem deploy` does all of this in one step without notes or review. It is meant for the user to run.
- Onboard reports the latest release, how far staging and development are ahead, and any commits on staging or production that development lacks.

Other branches are based on the development branch. mem keeps a short-lived branch rebased onto it when that is clean, and nudges to merge it back. A long-lived branch built on released code sets `[git] base` and takes that code in by merge commits only.

A single-branch project (`[git] single_branch = true`) has only its development branch and makes no releases: `mem promote` and `mem deploy` refuse.

## Hooks

**Agent hooks** put mem's state in front of the agent at the right moments. `init` and onboard keep mem's entries in each file and leave the rest alone:

- In **Claude Code** (`.claude/settings.json`), `mem hook compact` runs after compaction. It catches up with the remote like `mem sync` and prints a short digest of the active work. `mem hook prompt` runs with each user message and prints the work-log nudge when one is called for.
- **Codex** (`.codex/hooks.json`) runs the same two hooks, but only in a trusted project, and only after each user has approved them in `/hooks`, in every checkout.
- In **Cursor** (`.cursor/hooks.json`), `mem hook stop` runs at the end of a turn and asks for a work log when one is due.

The `[claude]`, `[codex]` and `[cursor]` settings turn these off.

**Git hooks** are installed in the repository's hooks directory. mem never overwrites a hook it did not write.

- `post-commit` prints the work-log measure after each of your commits. It also notes unpushed commits once there are three or more, and a structure doc that has fallen behind.
- With `protect = true`, `pre-commit` and `pre-push` refuse commits on staging or production, and pushes to them that do not come from `mem promote` or `mem deploy`. In a single-branch project they guard the `base` branch instead.

## Local files

`.mem/local/` is ignored by Git and holds what belongs to one checkout only:

- `onboard.md`: the last onboard context, when it was too long to print.
- `release-notes.md`: a release notes draft waiting for `mem promote production --confirm`.
- `stop_nudge`: the last work-log follow-up sent to Cursor, so it is not repeated.

Anything here can be deleted; mem writes it again when needed.

## Folder projects

A folder project is a plain folder of documents, for example in iCloud Drive, set up with `mem init --folder`. mem never runs Git in it. It keeps the config, `AGENTS.md` with instructions of its own for document work, the structure doc (kept by hand, since mem measures no drift there), docs, runnables, todos, logs, memories, skills and the agent hooks. It has no branches, commits, releases, checks, templates, specs or tasks, and the commands for them refuse. Saving a file is all a record needs. Logs are named by their time alone, and the work-log nudge counts the files changed since the newest log.
