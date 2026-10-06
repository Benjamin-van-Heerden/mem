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
  install.md                        install guide, written by mem; AGENTS.md points here when mem is missing
  structure.md                      living codebase and structure map
  docs/*.md                         project documents, included in onboard
  runnables/*                       executables whose stdout is included in onboard
  specs/<slug>/spec.md
  specs/<slug>/tasks/NN_<slug>.md
  specs/archive/<slug>/...          completed and abandoned specs
  todos/<slug>.md
  logs/<user>_<YYYYMMDD>_<HHMMSS>.md
  templates.lock                    template items and the content they last shared with their template
  setup.md                          one-time template setup, present until it is finished
.agents/skills/<name>/              skills, from templates or the project; linked from .claude/skills/<name>
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

[release]
production_pr = true    # optional: release production through a pull request

[claude]
compact_hook = false    # optional: leave out the Claude Code post-compaction hook

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
   | Ahead only | Nudge to push at the next sensible point; `mem sync`, record completion and `log commit` push |
   | Uncommitted changes block convergence | Nudge: commit, then `mem sync` |
   | Not on the development branch | Report lag behind `origin/<development>`; nudge to integrate soon |

2. **Update.** Refresh the managed `AGENTS.md` block, apply pending project patches, and commit and push these mem-owned paths.
3. **Build context,** in this order: project, structure doc (with a drift warning when stale), docs, runnable output, active specs in full with pending tasks, other open specs and todos as one-liners, recent logs (current user first), git summary, and a final state-specific agent instruction. Memories are not repeated; they are already in `AGENTS.md`. Memories and skills that changed during this onboard's sync are shown instead, since the running agent loaded the older copies. Output over ~14k characters goes to `.mem/local/onboard.md` with an instruction to read all of it.

`mem sync` is the mid-session catch-up: it performs step 1, pushes when the branch is then ahead and not behind (never staging or production), syncs template items, and reports what others pushed since the checkout last fetched (their commits; specs, tasks and todos opened, claimed, started or completed; changed memories and skills) and a newer mem release, without replacing the binary. Other commands print a short divergence nudge when local Git state shows drift.

**After compaction.** In Claude Code, a `SessionStart` hook with the `compact` matcher runs `mem hook compact` after every compaction. It performs the same catch-up as `mem sync` and prints a short digest (under ~3 KB) that joins the agent's context next to the compaction summary: branch state and what the sync did, the user's active spec and next task, their claimed todos, incoming changes, changed memories and nudges. It does not restate session progress, the structure doc, docs or logs; compaction keeps those. It never commits or pushes the user's work and exits cleanly on any error. `init`, `import` and onboard keep the entry in `.claude/settings.json`, preserving other settings; `[claude] compact_hook = false` removes it.

**Commit rhythm.** Commit each coherent, working change (typically one per task or fix). The agent commits a task's code first; `task complete` and `spec complete` then commit their record on their own, run the `mem sync` catch-up and push, warning when other changes stay uncommitted. `log commit` pushes at the end of each session, and `mem sync` pushes mid-session, for example before a promotion. Uncommitted work over 15 code files or 800 lines triggers a nudge to commit the finished parts.

## Work records

- **Spec:** larger planned work. Statuses `draft → active → completed | abandoned`. Body template: Overview, Goals, Technical Approach, Success Criteria, Notes.
  - `spec new "title"`, `spec list`, `spec show <slug>`
  - `spec start <slug>`: assigns the current user and marks it active; commits and pushes the change
  - `spec complete <slug>`: requires all tasks done; archives the spec, commits the archive, syncs and pushes; tells the agent to summarize the spec and offer a log (the Success Criteria are checked before, at the last `task complete`)
  - `spec abandon <slug> --reason "..."`: archives the spec
- **Task:** ordered steps within a spec. `task new "title" "description" [--spec]`, `task list [--spec]`, `task complete <slug> "notes" [--spec]`. `--spec` defaults to the user's single active spec. Completion commits the spec's record (`Complete task <slug>`), syncs and pushes, then prints remaining tasks and directs the agent to continue.
- **Todo:** open work that is not part of a spec, including blockers and decisions waiting on someone. `todo new "title" "description"`, `todo list`, `todo show`, `todo claim` (commits and pushes), `todo delete` (when done or no longer relevant).
- **Memory:** a lasting convention in the `AGENTS.md` memories block. `memory set <name> "<instruction>"`, `memory list`, `memory remove <name>`.

Arguments accept a slug or an unambiguous title.

## Session continuation

Two layers:

1. **Records** hold open work: specs with their pending tasks, and todos. Onboard renders them, with each todo's age, and derives release status from Git, so the next session knows what remains without reading any note.
2. **Logs** are statements of fact about a finished session: goals, what was done, decisions, files affected, failed approaches. They are never updated, so they never carry open work. `log new [--spec]` creates the templated file, lists open todos and prompts the agent to delete the ones the session completed and to record anything left open as todos. `log commit` closes the session: it refuses while placeholders remain, commits the changed `.mem/` records, syncs the branch with its upstream, pushes, and reports uncommitted work outside `.mem/`. Onboard shows the user's latest log in full and lists the other logs of the last 14 days by title. `log list`, `log show`.

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
- `mem promote production` fast-forwards production to `origin/<staging>`. It prints what will ship (commits, authors, specs completed in the range, specs still in progress) and releases with an annotated date tag (`v2026.09.24.1`) carrying a generated summary of the specs and commits. Most projects need nothing more. Projects that publish release notes set `[release] notes = true` (`mem init --release-notes`): there the first run drafts release notes in `.mem/local/release-notes.md` from the specs completed, the work logs written and the commits in the range; the draft is marked with the commit it describes and kept across re-runs for that commit. After the user has seen the refined notes, `--confirm` pushes production together with an annotated date tag (`v2026.09.24.1`) carrying them, atomically, and removes the draft.
- `mem deploy` is the user's one-step release: it pushes unpushed development commits, fast-forwards staging and then production, and tags production with a generated summary instead of notes. Its output is a report, not agent instructions. Agents are instructed never to run it unless the user explicitly asks.
- With `[release] production_pr = true` (`mem init --production-pr`), production releases go through a pull request: the release pushes a snapshot branch `promotion/<production>/<timestamp>` at the release commit and opens a pull request carrying the generated summary, or, with release notes, the confirmed notes on `--confirm`. Once it is approved, `mem promote production --confirm` fast-forwards production to the pull request's head and tags it with the pull request's description; GitHub marks the pull request merged without a new commit, and mem deletes the snapshot branch (closing the pull request with a comment if GitHub did not mark it merged). Changes requested stop the release. GitHub's merge buttons are not used, because every merge method adds or rewrites commits. GitHub access comes from `GITHUB_TOKEN`, `GH_TOKEN` or `gh auth token`, and is only needed in these projects. `mem deploy` refuses them.
- In a repository without commits, such as a fresh clone of an empty GitHub repository, `mem init` first makes an empty commit on the production branch, built from the empty tree so anything already staged stays staged.
- `mem init` creates missing branches in promotion order: staging from production, development from staging, tracking the remote's copy where one exists. It publishes any branch the remote lacks and switches to development.
- If staging or production has commits that are not on development (a hotfix or a web-UI merge), promotion stops and prints the commands to merge them back into development.
- Onboard shows release status: the latest production tag and how far staging and development are ahead, with a ⚠️ when staging or production has commits development lacks, such as a pull request merged into production directly.

**Hooks.** With `protect = true`, `init` and onboard install `pre-push` and `pre-commit` hooks in the repository's hooks directory. They refuse pushes to staging or production that do not come from `mem promote`, and commits made on those branches. The hooks call `mem hook <name>` and do nothing where mem is not installed. Existing non-mem hooks are never overwritten; onboard explains how to add the call instead. `protect = false` removes the hooks for solo projects. `git push --no-verify` and web-UI merges bypass hooks; promotion detects the resulting divergence.

## Templates

A template library is an ordinary Git repository with one directory per template, recognized by its `template.toml` (`description = "…"`), holding `memories/<name>.md`, `skills/<name>/` and `docs/<name>.md`. A project names its library and templates in `.mem/config.toml`:

```toml
[templates]
source = "https://github.com/Benjamin-van-Heerden/mem-templates.git"
use = ["base", "nextjs-web"]      # later templates win when two provide the same item
exclude = ["skill:old-guide"]     # opt-outs, written by mem
```

- `mem init --template <name>` (repeatable) and `mem template use <name>` draw a template's items into the project: memories into `AGENTS.md`, skills into `.agents/skills/<name>/` with a `.claude/skills/<name>` link, docs into `.mem/docs/`. Without `--template-source`, the library is github.com/Benjamin-van-Heerden/mem-templates; the project's `[templates] source` records whichever library it uses.
- mem keeps one clone of each library in the user cache and pulls it at onboard (not with `--offline`).
- `.mem/templates.lock` records, per item, the content it last shared with its template. Onboard compares project copy, template copy and that record: it adds missing items, updates items the project has not edited, reports local-only edits as promotion candidates, flags items edited on both sides, keeps items a template no longer provides, and turns a deleted item into an `exclude` entry. It never deletes project files, and it publishes its changes unless those paths already had uncommitted edits.
- `mem template promote <memory|skill|doc> <name> [--to <template>]` copies a project item into its template, commits and pushes the library, so every project using the template receives it at its next onboard. Promoting to a template that does not exist creates it.
- `mem template reset <kind> <name>` takes the template's copy, and brings back an excluded item.
- `mem template list` shows the library's templates, which of them have a setup, and the state of each of the project's template items.

**Setup.** A template may also have `setup.md` at its root: a one-time setup for new projects, written as ordered steps with checkbox headings (`## [ ] 1. Scaffold the app`), each ending in a "Done when" line; steps that need the user (logging in to a service, secrets, design choices) are marked `(you)`, and the agent asks for what they name. `mem init --template` joins the setups of the chosen templates, in order, into `.mem/setup.md`. The file is the setup's only state: while it exists, onboard shows it first under 🏗️ SETUP with its progress and makes working through it, ticking and committing step by step, the session's first instruction, and the compaction digest reminds the agent of it. When every step is ticked, the agent deletes the file. Setup is not a template item: it is never synced, locked or excluded, and `mem template use` does not install it on an existing project; it only mentions that the template has one.

Framework guides belong in templates as skills rather than docs: docs are printed in full at every onboard, skills load when they are relevant.

## Updates

- The managed `AGENTS.md` block is refreshed from the executable on every onboard.
- Project patches are numbered migrations keyed by `schema`, applied once at onboard.
- Onboard checks the latest GitHub release (from the redirect of `releases/latest`, no API token) and, when it is newer than a release build of mem, downloads the binary for the platform, verifies it against `checksums.txt`, replaces the executable and re-runs onboard with it. `mem update` does the same on demand. `MEM_NO_UPDATE=1` turns the automatic update off; development builds never replace themselves.

## Import

`mem import agent-core` converts a Python coding-harness project: config, memories, docs, specs and tasks, todos, logs and its structure doc map almost directly onto this layout. Logs get a `# Work Log - <title>` heading where the harness wrote a plain title line or none, and lose their "What Comes Next" sections, which would otherwise present long-finished work as pending; the import prints the newest one so the agent can review it with the user and record what is still open as todos. Records keep their GitHub issue links, and the harness's GitHub usernames become mem identities through its `user_mappings.toml`, in logs as in specs and todos. The `coding_general.md` and `coding_testing.md` docs of earlier harness versions are left out; the general principles in the mem block replace them. The original `.agent_core/` is left in place for the user to remove, along with references to it in docs and scripts; the import's instructions cover both. Release notes leave out imported records, since their timestamps predate the release range.
