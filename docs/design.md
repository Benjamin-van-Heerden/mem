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
base = "main"           # optional, on a long-lived development branch built on released code, such as a partner's own main
# single_branch = true  # instead of staging and production (which must then be absent): no releases, see Single-branch projects

[release]
production_pr = true    # optional: release production through a pull request

[claude]
compact_hook = false    # optional: leave out the Claude Code post-compaction hook

[cursor]
stop_hook = false       # optional: leave out the Cursor work-log follow-up at the end of a turn

[structure]
ignore = ["migrations/**"]   # optional: files that do not count as work (drift detection, work-log nudges)
include = ["*.md"]           # optional: files that count although the built-in rules leave them out
```

Identity is `git config user.name`, slugified.

## Onboard

0. **Leave staging and production.** A fresh clone checks out the remote's default branch, which may be production. Before anything else, onboard switches a checkout on staging or production to the development branch, creating it from the remote's copy when it only exists there, and warns about commits on the branch it leaves that no remote branch has. If the switch fails, for example on conflicting uncommitted changes, onboard installs the protection hooks but commits and pushes nothing, and tells the agent to resolve it with the user. A detached checkout and a single-branch project's base branch only get the usual nudges. Onboard installs the protection hooks before its first commit.
1. **Converge.** Fetch with a time limit. On failure, warn prominently and continue with local context.

   | State | Behavior |
   | --- | --- |
   | Up to date | Nothing |
   | Behind | Fast-forward. Works with uncommitted changes unless incoming commits touch the same files. |
   | Unpushed local commits and remote moved | Clean tree: rebase onto upstream. Conflict: abort and instruct the agent to raise it with the user. |
   | Local commits include a merge and the remote or development moved | Never rebase, since that would flatten the merge into copies: nudge to bring the incoming commits in with `git merge --no-ff`, then `git push`. |
   | Upstream rewritten (force-pushed elsewhere) | Rebase with `--fork-point`, replaying only the commits made here, so commits the rewrite dropped do not come back. |
   | Ahead only | Nudge to push at the next sensible point; `mem sync`, record completion and `log commit` push |
   | Push rejected because others pushed first | Fetch, catch up as above and push again, up to three attempts; a conflict ends with the conflict nudge. |
   | Uncommitted changes block convergence | Nudge: commit, then `mem sync` |
   | Feature branch, development moved on | Clean tree: rebase onto `origin/<development>` and say to run the tests. Already pushed with only the user's commits on its remote copy: force-push with a lease pinned to the fetched commit. Commits by others: nudge, never rewrite. Conflict: abort and name the files. Development, staging, production and the `[git] base` branch are never rebased onto development or force-pushed. |
   | Feature branch ahead of development | Nudge to merge it back soon; completing a spec on it instructs merging now (a fast-forward, since it follows development) |

2. **Update.** Refresh the managed `AGENTS.md` block, apply pending project patches, and commit and push these mem-owned paths.
3. **Build context,** in this order: project, structure doc (with a drift warning when stale), docs, runnable output, active specs in full with pending tasks, other open specs and todos as one-liners, a ⏳ CHECK THESE list of what has probably stopped being true (active specs whose record has not changed for 14 days, remote branches not merged into development and idle for 14 days), recent logs (current user first), git summary, and a final state-specific agent instruction. Memories are not repeated; they are already in `AGENTS.md`. Memories and skills that changed during this onboard's sync are shown instead, since the running agent loaded the older copies. Output over ~14k characters goes to `.mem/local/onboard.md` with an instruction to read all of it.

**Long-lived branches built on released code.** A branch such as a partner's own main (`fundi` in studii-next) is its checkout's development branch, with its own `.mem/config.toml`, and is built on released code rather than on the shared development branch. `[git] base = "main"` says so. On that branch, onboard reports under 🌿 SHARED CODEBASE how many commits of `origin/main` it lacks and the rule: released commits come in only when the user asks, by a merge commit (`git merge --no-ff origin/main`, then `git push`); the branch is never rebased or force-pushed, since it is published and deployed, and unreleased work is never merged into it. When commits are missing, the agent instruction asks to mention them. Short-lived branches, by contrast, follow the development branch (see the table above). The base branch itself is never one of them: a checkout on it is not rebased onto development, and sync nudges off it. In a staged project it stays unguarded otherwise, unless it is also staging or production: the protection hooks do not refuse commits on it, and `mem sync` pushes it like any other branch.

`mem sync` is the mid-session catch-up: it performs step 1, pushes when the branch is then ahead and not behind (never staging or production), syncs template items, and reports what others pushed since the checkout last fetched (their commits; specs, tasks and todos opened, started, completed or closed; changed memories and skills) and a newer mem release, without replacing the binary. Other commands print a short divergence nudge when local Git state shows drift.

**After compaction.** In Claude Code, a `SessionStart` hook with the `compact` matcher runs `mem hook compact` after every compaction. It performs the same catch-up as `mem sync` and prints a short digest (under ~3 KB) that joins the agent's context next to the compaction summary: branch state and what the sync did, the user's active spec and next task, the user's work since their last work log or completed task (with the post-commit hook's wording, and once a log is due an instruction to write one at the next natural stopping point while the summary still holds the work), incoming changes, changed memories and nudges, and asks the agent to check the session's title against the work in progress and retitle it for the current work where its environment can rename sessions (see Session titles). It does not restate session progress, the structure doc, docs or logs; compaction keeps those. It never commits the user's work; like `mem sync`, it pushes committed work when the fetch succeeded and the branch is ahead and not behind (never staging or production), so a long session shares its commits at every compaction. It exits cleanly on any error. `init`, `import` and onboard keep the entry in `.claude/settings.json`, preserving other settings; `[claude] compact_hook = false` removes it.

**At each user message.** A `UserPromptSubmit` hook runs `mem hook prompt`, whose output Claude Code adds to the agent's context with the message. It is silent until a work log is called for (see Checkpoint nudges), then prints one line: at the suggest level, write a log at the next natural stopping point; once a log is due, write one before taking up substantial new work, finishing a change in progress first, answering a quick question first and stopping if asked to. Both say that a log does not end the turn. It reads local state only (about 0.2 s) and exits cleanly on any error. The same commands keep it in `.claude/settings.json`; `[claude] prompt_hook = false` removes it.

**In Cursor.** Cursor reads `AGENTS.md`, but its hooks cannot do what mem's Claude Code hooks do:
- **After compaction:** `preCompact` only shows the user a message, and `sessionStart` does not fire again.
- **With each message:** `beforeSubmitPrompt` can only allow or block the prompt.
- **At session start:** `sessionStart`'s `additional_context` is dropped by a race in current releases, and `AGENTS.md` already asks for `mem onboard`.

Cursor's one way to put mem's words in front of the agent is the `stop` hook, whose `followup_message` Cursor submits as the next user message. `init`, `import` and onboard keep `{"command": "mem hook stop", "timeout": 30}` under `stop` in `.cursor/hooks.json`, preserving other keys and hooks; `[cursor] stop_hook = false` removes it. The command has no POSIX guard, because Cursor runs hook commands through PowerShell on Windows; where mem is missing, the hook fails and Cursor carries on.

`mem hook stop` reads Cursor's input and always answers with a JSON object. It sends a follow-up only when all of these hold:
- the turn completed, so not after the user stopped it or after an error;
- it is not already answering its own follow-up (`loop_count` 0);
- a work log is due;
- the work summary differs from the one it last asked about, which it keeps in `.mem/local/stop_nudge`, so an agent that put the log off is asked again only after more work.

The follow-up says it comes from mem and not the user. It asks for the log now if the turn finished a piece of work. If the turn ended with a question for the user or mid-change, it asks the agent to say in one line that a log is due and end the turn, and not to start other work.

Cursor also runs the hooks in `.claude/settings.json`. It rejects their plain-text output and may fire the `compact` entry at every session start, so `mem hook compact` and `mem hook prompt` return at once when `CURSOR_VERSION` is set. (`CLAUDE_PROJECT_DIR` cannot tell the agents apart, since Cursor sets it too.) The post-commit hook's nudge reaches Cursor's agent in its shell output, as for any agent.

**Session titles.** The Claude desktop app titles sessions from their first message, often "Let's get to work", so mem asks agents to give sessions a short title for the work wherever they can rename them, at three points. The title leaves out the project: the app groups sessions by project folder. The onboard instruction asks for it as soon as the user has said what to work on: at once when their first message already did, otherwise as the first thing after their next one. When the context is long enough to go to `.mem/local/onboard.md`, the file ends with the same request under 🏷️ SESSION TITLE: the agent reads the file after the instruction, and an agent that ended its reading on the work log had forgotten the request by the user's next message. After compaction and at each work log, when the session may have moved on from what it was named for, the digest and `log new` ask the agent to look up the current title, compare it with the current work and rename the session if it no longer fits. Each instruction says that the desktop app's session tools address the current session as `"self"` and asks the agent to correct and retry a rename that fails: an agent once called the rename tool without a session id and moved on when it was refused. These are instructions in mem's output, not hooks.

**Commit rhythm.** Commit each coherent, working change (typically one per task or fix). `task complete` makes the task's commit: the paths the agent names (or, without paths, every change in the working tree, ignored files excepted) together with the task record, with the task title as subject, the completion note as body and a `Mem-Task: <spec>/<task>` trailer that resets the work-log count. Agents are told to name the files they changed, since another session may be working in the same checkout, and to commit anything of theirs that mem lists as left uncommitted as a follow-up; and to keep secrets and local artifacts ignored. `spec complete` commits the archived record on its own, warning when other changes stay uncommitted. Both then run the `mem sync` catch-up and push. `log commit` pushes at each work log, and `mem sync` pushes in between, for example before a promotion. Uncommitted work over 15 code files or 800 lines triggers a nudge to commit the finished parts.

## Work records

- **Spec:** larger planned work. Statuses `draft → active → completed | abandoned`. Body template: Overview, Goals, Technical Approach, Success Criteria, Notes. A started spec is worked through to the end without stopping for approval between tasks; `spec start`, `task complete` and the instructions say so, and the agent stops only for decisions only the user can make, continuing meanwhile with tasks that do not depend on them. `spec complete` first syncs and refuses while the fetch fails, the branch is still behind its upstream (an unresolved conflict) or a feature branch lacks development commits; the spec stays active until those are resolved.
  - `spec new "title"`, `spec list`, `spec show <slug>`
  - `spec start <slug>`: assigns the current user and marks it active; commits and pushes the change
  - `spec complete <slug>`: requires all tasks done; archives the spec, commits the archive, syncs and pushes; tells the agent to summarize the spec and offer a log (the Success Criteria are checked before, at the last `task complete`)
  - `spec abandon <slug> --reason "..."`: archives the spec
- **Task:** ordered steps within a spec. `task new "title" "description" [--spec]`, `task list [--spec]`, `task complete <slug> "notes" [<path>...] [--spec]`. `--spec` defaults to the user's single active spec. Completion commits the named paths with the task record and lists the changes it left uncommitted, new files first, so the agent commits any of its own as a follow-up while another session's changes in the same checkout stay out; without paths it commits every change in the working tree. The commit has the task title as subject, the notes as body and a `Mem-Task` trailer; completion then syncs, pushes, prints remaining tasks and directs the agent to continue.
- **Todo:** open work that is not part of a spec, including blockers and decisions waiting on someone. `todo new "title" "description"`, `todo list`, `todo show`, `todo delete` (when done or no longer relevant). A todo has no states: it is open until it is done, then deleted. Larger work that someone takes on is a spec, which `spec start` assigns.
- **Memory:** a lasting convention in the `AGENTS.md` memories block. `memory set <name> "<instruction>"`, `memory list`, `memory remove <name>`. Some agent runtimes read only part of `AGENTS.md` (Codex reads the first 32 KiB by default), and the memories come after mem's block, so they are cut first: when `AGENTS.md` is over 32 KiB, `memory set` and onboard warn and tell the agent to stop and agree with the user which memories to merge, shorten or retire.

Arguments accept a slug or an unambiguous title.

## Session continuation

Two layers:

1. **Records** hold open work: specs with their pending tasks, and todos. Onboard renders them, with each todo's age, and derives release status from Git, so the next session knows what remains without reading any note. Records live on the branch they are changed on: specs, tasks and todos changed on a feature branch reach development, and teammates, when the branch merges, and mem says so when it commits one there.
2. **Logs** are statements of fact about a stretch of work since the previous log, as things stand: its goal, what was done, decisions, the files affected, and notes such as approaches tried and dropped. They are written at checkpoints (a completed spec, a release, the work-log nudges, the user wrapping up), not tied to the end of a session, which in practice often never comes. They are never updated, so they never carry open work. `log new [--spec]` creates the templated file, with the user's own commits since their last checkpoint in its frontmatter (`commits`, short hashes, oldest first) and Key Files Affected listing every file outside `.mem/` that those commits and the uncommitted changes touched, with changed lines (at most 30), for the agent to annotate; it lists open todos and prompts the agent to delete the ones the work completed and to record anything left open as todos. `log commit [<log>] [<.mem/ path>...]` refuses while placeholders remain, commits the log with the `.mem/` records the agent names (the todos it added or deleted for the log, the structure doc), lists other changed records without committing them, since another session may be working in the same checkout, syncs the branch with its upstream, pushes, and reports uncommitted work outside `.mem/`. Onboard shows the user's latest log in full, with up to two more of theirs from the last 14 days while the three stay under 12,000 bytes, since each covers only a stretch of work, and lists the other logs of the last 14 days by title. `log list`, `log show`.

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
- After a release pushes staging or production (`mem promote`, `mem deploy`, or completing a production pull request), mem fast-forwards the local copy of that branch, which `mem init` created, to the released commit, so it does not fall behind the remote. It leaves a local copy that is checked out in a worktree or has commits the release lacks, with a ⚠️ line, and never creates one.
- Onboard shows release status: the latest production tag, how far staging and development are ahead and how long ago the oldest unreleased development commit was made, with a ⚠️ when staging or production has commits development lacks, such as a pull request merged into production directly.

**Single-branch projects.** A project that never deploys from branches, such as a personal fork that builds locally from its development branch, sets `[git] single_branch = true` and has no `staging` or `production` (loading refuses a config that sets them, or `[release] production_pr` or `notes`, alongside it). The mode is never inferred: `mem init --single-branch` sets it, taking the branch the user is on as the development branch and creating no others; it refuses a detached HEAD and the flags that name or configure other branches. An existing project converts by editing its config; the next onboard rewrites the managed block and the hooks. In this mode mem makes no releases: `mem promote` and `mem deploy` refuse before fetching or pushing anything, so no tag is ever pushed (a fork's inherited release workflow would fire on one), onboard has no release status, and the managed block has no Releases instructions. `[git] base` still works: the base branch takes the place of staging and production as the branch mem never commits to or pushes, `mem sync` neither pushes it nor rebases it onto development, and onboard nudges off it.

**Hooks.** With `protect = true`, `init` and onboard install `pre-push` and `pre-commit` hooks in the repository's hooks directory. They refuse pushes to staging or production that do not come from `mem promote`, and commits made on those branches; in a single-branch project they refuse pushes to and commits on the base branch instead, from checkouts that hold the mem config (a base such as an upstream mirror has none, so on it only the managed block's rule applies). The hooks call `mem hook <name>` and do nothing where mem is not installed. Existing non-mem hooks are never overwritten; onboard explains how to add the call instead. `protect = false` removes the hooks for solo projects. `git push --no-verify` and web-UI merges bypass hooks; promotion detects the resulting divergence.

**Checkpoint nudges.** mem measures the user's work since their last work log or completed task: their work commits, and the files and lines that count as work changed by those commits and by uncommitted changes in the checkout. Files count as work by the structure doc's rules (`[structure] ignore` and `include`); work commits are the user's commits that change such a file, other than onboard's project-file commit; teammates' commits and mem's records never count, and a completed task resets the measure like a log, since its note records what was done. Whichever measure reaches its threshold first sets the level: a log is suggested from 5 commits, 10 files or 500 lines, and due from 10 commits, 20 files or 1,200 lines. The nudges never ask the agent to drop the work at hand. Every project, protected or not, gets a `post-commit` hook running `mem hook post-commit`. From local state only, so a commit stays fast, it prints up to three `mem:` lines: after each of the user's work commits, the measure, with "write a work log at the next natural stopping point" once a log is suggested and "a work log is due" once it is due; commits not pushed once there are 3 or more; and a stale structure doc. The prompt hook (see At each user message) and the compaction digest report the same measure. Every Git command mem runs carries `MEM_GIT=1`, and the hook script exits at once for those, so mem's own commits and rebases stay silent. The script also exits quietly when the `mem` on PATH lacks the subcommand.

## Templates

A template library is an ordinary Git repository with one directory per template, recognized by its `template.toml` (`description = "…"`), holding `memories/<name>.md`, `skills/<name>/`, `docs/<name>.md` and `runnables/<name>`, and optionally `changes/<name>.md` change notes. A project names its library and templates in `.mem/config.toml`:

```toml
[templates]
source = "https://github.com/Benjamin-van-Heerden/mem-templates.git"
use = ["base", "nextjs-web"]      # later templates win when two provide the same item
exclude = ["skill:old-guide"]     # opt-outs, written by mem
```

- `mem init` asks for templates first: without `--template` or `--no-template` it changes nothing, lists the library's templates and tells the agent to ask the user which fit before running it again.
- `mem init --template <name>` (repeatable) and `mem template use <name>` draw a template's items into the project: memories into `AGENTS.md`, skills into `.agents/skills/<name>/` with a `.claude/skills/<name>` link, docs into `.mem/docs/`, runnables into `.mem/runnables/` with LF line endings and the executable bit, so onboard runs a template's checks in every project that uses it. Without `--template-source`, the library is github.com/Benjamin-van-Heerden/mem-templates; the project's `[templates] source` records whichever library it uses.
- mem keeps one clone of each library in the user cache and pulls it at onboard (not with `--offline`).
- `.mem/templates.lock` records, per item, the content it last shared with its template. Onboard compares project copy, template copy and that record: it adds missing items, updates items the project has not edited, reports local-only edits as promotion candidates, flags items edited on both sides, keeps items a template no longer provides, and turns a deleted item into an `exclude` entry. It never deletes project files, and it publishes its changes unless those paths already had uncommitted edits.
- A memory only guides new work, so when a sync adds a template memory to a project or replaces one with the template's new version, it also records a todo ("Apply the template convention <name>", with the memory's text and, for a change, its previous text) asking for the existing code to be checked against it. The todo is committed with the sync, so it is recorded once for everyone; a later change to a memory with an open todo rewrites that todo. Onboard, `mem sync` and the compaction digest mark it with 📌 and tell the agent to raise it with the user. The todo also covers a convention that relies on tooling or setup the project has not adopted (portless in a project whose `dev` script is plain `next dev`): the agent says so and asks the user whether to adopt it, as its own todo or spec, or drop it with `mem memory remove <name>`, and updates or removes a project memory it contradicts rather than leaving both. mem does not detect adoption; the todo's wording carries this. A project still in its template setup gets these todos too: `init` copies the template's setup once and never updates it, so a convention that arrives while the setup is pending may describe something its steps never build. Such a todo says so, and the agent is told to follow the convention in the remaining setup steps and check the code once the setup is done. The memories `init` installs with the setup come from the same template revision, so they record none. `mem template use` records one todo for all the memories the template brings.
- Some template changes need action in existing projects that no sync performs, such as a setting in a starter file the project copied long ago. The template's author records such a change as a change note, `changes/<name>.md` (a dated name such as `2026-10-29-react-compiler`, a first-line `# <title>`, then what changed and what an existing project should do). A note is not an item: it is never installed. `.mem/templates.lock` records each note the project has seen (`[[change]]`, by template and name), and a sync turns every unseen note into a todo ("Apply the <template> template change: <title>", quoting the note) asking for it to be applied to the existing code, committed with the sync and marked with 📌 like the convention todos. `mem init` and `mem template use` record the notes of the templates they install as seen without todos, since the project starts from the template as it is; the notes the project's other templates gained meanwhile still become todos. While a setup is pending, the todo asks for the change to be applied in the remaining setup steps. Records of notes the library no longer has are dropped, and a published note is never re-read, so a correction is a new note.
- `mem template promote <memory|skill|doc|runnable> <name> [--to <template>]` copies a project item into its template, commits and pushes the library, so every project using the template receives it at its next onboard. Promoting to a template that does not exist creates it.
- `mem template reset <kind> <name>` takes the template's copy, and brings back an excluded item.
- `mem template list` shows the library's templates, which of them have a setup, and the state of each of the project's template items.

**Setup.** A template may also have `setup.md` at its root: a one-time setup for new projects, written as ordered steps with checkbox headings (`## [ ] 1. Scaffold the app`), each ending in a "Done when" line; steps that need the user (logging in to a service, secrets, design choices) are marked `(you)`, and the agent asks for what they name. `mem init --template` joins the setups of the chosen templates, in order, into `.mem/setup.md`. The file is the setup's only state: while it exists, onboard shows it first under 🏗️ SETUP with its progress and makes working through it, ticking and committing step by step, the session's first instruction, and the compaction digest reminds the agent of it. When every step is ticked, the agent deletes the file. Setup is not a template item: it is never synced, locked or excluded, and `mem template use` does not install it on an existing project; it only mentions that the template has one.

Framework guides belong in templates as skills rather than docs: docs are printed in full at every onboard, skills load when they are relevant.

## Updates

- The managed `AGENTS.md` block is refreshed from the executable on every onboard. It opens with Ground Rules that hold in every session whatever a task description or a tool's defaults say: onboard first or stop; work on the development branch and push only through mem, with no pull requests or other branches unless the user asks; never touch staging or production, force-push or rewrite published history; release only when the user asks in the conversation; regenerate generated files rather than editing them, also on conflict; one package manager. The block is the same on every branch of a project, and in every project apart from single-branch ones, whose variant (rendered from the same `instructions.md`) leaves out the release rule, release status and the Releases instructions, says that mem makes no releases there, and asks agents never to commit to or push the base branch. Branch-specific guidance, such as the base rule above, is printed by onboard instead, so a long-lived branch's AGENTS.md never conflicts with the development branch's when released code is merged in.
- Project patches are numbered migrations keyed by `schema`, applied once at onboard.
- Onboard checks the latest GitHub release (from the redirect of `releases/latest`, no API token) and, when it is newer than a release build of mem, downloads the binary for the platform, verifies it against `checksums.txt`, replaces the executable and re-runs onboard with it. `mem update` does the same on demand. `MEM_NO_UPDATE=1` turns the automatic update off; development builds never replace themselves.

## Import

`mem import agent-core` converts a Python coding-harness project: config, memories, docs, specs and tasks, todos, logs and its structure doc map almost directly onto this layout. Logs get a `# Work Log - <title>` heading where the harness wrote a plain title line or none, and lose their "What Comes Next" sections, which would otherwise present long-finished work as pending; the import prints the newest one so the agent can review it with the user and record what is still open as todos. Records keep their GitHub issue links, and the harness's GitHub usernames become mem identities through its `user_mappings.toml`, in logs as in specs and todos. The `coding_general.md` and `coding_testing.md` docs of earlier harness versions are left out; the general principles in the mem block replace them. The original `.agent_core/` is left in place for the user to remove, along with references to it in docs and scripts; the import's instructions cover both. Release notes leave out imported records, since their timestamps predate the release range.
