# Codebase and Structure

## Overview

mem is a single Go command-line executable (`cmd/mem`) that builds context for coding agents and keeps work records in a Git repository. Each project it manages holds its state as plain files: `.mem/` (config, structure doc, docs, runnables, specs, todos, logs) and a managed block in `AGENTS.md`. mem shells out to the `git` executable for all repository operations and uses HTTPS only to update itself from GitHub releases; it has no database or server.

## Tech Stack

- Go 1.27.1 (`go.mod`, `.tool-versions`)
- `github.com/spf13/cobra` for the command tree
- `github.com/pelletier/go-toml/v2` for `.mem/config.toml`
- `go.yaml.in/yaml/v3` for Markdown frontmatter in work records
- The `git` CLI, invoked through `internal/git`
- GoReleaser (`.goreleaser.yaml`) and GitHub Actions (`.github/workflows/`) for CI and releases

## Directory Layout

```text
cmd/mem/main.go          entry point: runs cli.New() with an interrupt-aware context
internal/
  cli/                   one file per command group; all stdout and agent instructions live here
  agentsmd/              the managed <mem> block and the memories block in AGENTS.md
    instructions.md      source of the managed block, embedded into the binary
    install.md           source of .mem/install.md, embedded into the binary
  project/               .mem/config.toml, project loading, identity, schema patches
  git/                   thin wrapper around the git executable
  converge/              fetch, fast-forward/rebase and drift nudges
  structure/             .mem/structure.md template, drift measurement, file tree
  work/                  specs, tasks, todos and logs as Markdown with YAML frontmatter
  release/               promotion planning/execution, release notes drafts and release status
  github/                GitHub REST calls for promotion pull requests
  templates/             template library clone, item sync with .mem/templates.lock, promote and reset, setup files
  selfupdate/            latest-release check and verified binary replacement
  hooks/                 pre-push / pre-commit (protection) and post-commit (nudges) hook install, protection checks
  checkpoint/            work since the last log or completed task, unpushed commits, stale structure doc
  claude/                mem's entry in .claude/settings.json (the post-compaction hook)
  runnables/             runs .mem/runnables/* for onboard
  importer/              conversion from the Python harness (.agent_core/)
  output/                heading, section and instruction formatting
  buildinfo/             version and commit, set by ldflags or module info
docs/                    design.md (target design), status.md (implemented vs planned); roadmap.md is gitignored
old/                     gitignored legacy Python harness, reference only
.mem/                    this repository's own mem state (mem manages its own source)
dist/                    gitignored local builds, e.g. dist/mem-dev
```

## Key Modules

- **`internal/cli`**: builds the cobra tree in `root.go` (`New`, `app{dir}` for the global `--dir` flag). Commands: `init.go`, `onboard.go`, `sync.go`, `spec.go`, `task.go`, `todo.go`, `log.go`, `memory.go`, `structure.go`, `promote.go`, `promote_pr.go`, `deploy.go`, `template.go` (`template use|list|promote|reset`), `update.go` (`update`, plus `autoUpdate`/`rerun` used by onboard and `newerRelease` used by sync), `import.go`, `version` (in `root.go`) and the hidden `hook.go` (`hook pre-push|pre-commit|post-commit|compact`; `compact.go` has the compaction digest). Shared helpers in `root.go`: `branchNotice` (on a branch other than development, says a record committed there reaches teammates when the branch merges; used by `publish`, `commitRecord` and `log commit`), `publish` (commit and push mem-owned paths via `git.CommitPaths`, which combines `git.Commit` and `git.Push`), `table`, `readAgents`/`writeAgents`, `removeClaudeLink` (used by `init` and `import`). Depends on every other internal package; nothing depends on it except `cmd/mem`.
- **`internal/project`**: `Config`/`GitConfig`/`StructureConfig`/`TemplatesConfig`/`ClaudeConfig` (TOML; `[claude] compact_hook`, unset means on), `Load` (finds the Git toplevel and reads `.mem/config.toml`, rejecting newer schemas), `WriteConfig`, `User` (slug of `git config user.name`), `Slugify`, and `Project.Path`/`Rel` helpers rooted at `.mem/`. `patches.go` holds `Upgrade` with an empty per-schema patch map (`Schema = 1`).
- **`internal/agentsmd`**: `WriteGuide` (writes the embedded install guide to `GuidePath`, `.mem/install.md`, reporting a change), `Install` (block first, then memories, then existing content), `ReplaceLegacy` (swaps the Python harness block in place), `Refresh` (rewrites the block from the embedded `instructions.md` unless the stamp `<!-- Managed by mem X -->` is a newer semver), and memory CRUD (`Memories`, `SetMemory`, `RemoveMemory`) over `## name` sections inside `<memories>`. `span` locates each block by its tags and requires exactly one of each.
- **`internal/converge`**: `Sync` (fetch with a 20 s limit, then fast-forward or rebase unpushed commits, aborting on conflict), `Push` (after a successful fetch, push a branch that is ahead and not behind, never staging or production, replacing the unpushed nudge with the outcome) and `Local` (inspect cached refs only). Both return a `Report` with `Done` and `Nudges`: unpushed commits, drift from the development branch, missing remote, being on staging/production, and uncommitted work above 15 code files or 800 lines.
- **`internal/structure`**: `Template` for the doc, `Measure`/`ChangesSince` (code changes since the last commit touching `.mem/structure.md`, including untracked files; `Stale` at more than 5 code files or 1,000 changed lines), `Relevant` (excludes Markdown, lockfiles, binaries, vendored/build dirs and `[structure] ignore` patterns), and `Tree` (tracked and unignored files, capped at 300 entries).
- **`internal/claude`**: `SyncCompactHook` adds, updates or removes mem's `SessionStart`/`compact` entry (identified by `mem hook compact` in its command) in `.claude/settings.json`, keeping other keys in order via a small ordered JSON `object`; writes only on change.
- **`internal/work`**: record types `Spec`, `Task`, `Todo`, `Log` with frontmatter structs, create/find/list/complete/claim/archive functions, and the templates for specs and logs. `changes.go` has `RecordChanges`, which compares specs, tasks and todos at two revisions (read through a `Revision` func) by kind and slug and describes them in plain terms. `markdown.go` has `ReadMarkdown`/`parseMarkdown`/`WriteMarkdown` and `resolve`, which matches a reference by exact slug, then case-insensitive title, then unique slug prefix.
- **`internal/release`**: `Prepare` (fetch, compute commits, completed specs, divergence and the next date tag `vYYYY.MM.DD.N`), `Execute` (one atomic, never forced push of the branch and, for production, an annotated tag kept verbatim with `--cleanup=whitespace`, with `MEM_PROMOTE=1` so the hook allows it), `CurrentStatus` for onboard (including commits on staging or production that development lacks). `notes.go`: `DraftNotes` (marker line, specs completed, work log titles and their accomplishment headings, commits; `predates` leaves out records timestamped before the range began, such as imported history), `GeneratedMessage` for `mem deploy` and releases without notes, `DraftCommit`.
- **`internal/github`**: owner/repo from a GitHub remote URL, a token from `GITHUB_TOKEN`, `GH_TOKEN` or `gh auth token`, and the pull request calls promotion needs (create, list open by head prefix, get, reviews, comment, close); `MEM_GITHUB_API` points it at a test server.
- **`internal/hooks`**: `Sync` installs or removes `pre-push`/`pre-commit` scripts per `protect` and installs `post-commit` in every project, never overwriting non-mem hooks. The scripts run `mem hook <name>` and exit 0 when no mem with hook support is on PATH; post-commit also exits at once under `git.InternalEnv`. `PrePush` and `PreCommit` implement the checks.
- **`internal/checkpoint`**: `WorkSinceLog` walks HEAD's non-merge commits (newest first, at most 1,000) counting the user's work commits (changes outside `.mem/`, not `ProjectFilesCommit`) until a checkpoint: their work log, a `Complete task`/`Complete spec` commit of theirs, or the commit that added `.mem/config.toml`. `LogLine` escalates the wording at 3 and 5; `CommitLines` adds unpushed commits (3 or more, from cached refs) and a stale structure doc, for `mem hook post-commit`.
- **`internal/templates`**: `library.go` has `Open` (clone a library URL into `os.UserCacheDir()/mem/templates/<slug>`, `git pull --ff-only` with a 20 s limit; a failed pull is a warning), `Templates` (directories with `template.toml`), `Items` (memories/skills/docs of the used templates, later templates overriding earlier ones) and the `DefaultSource` constant (the mem-templates repository), used when neither `--template-source` nor the project config names a library. `sync.go` has `Sync` and `Status`: `classify` compares project copy, template copy and the `.mem/templates.lock` hash into a state (`StateMissing`, `StateUpdated`, `StateLocalEdits`, `StateBothEdited`, …), and `reconcile` acts on it; installs go to `AGENTS.md` memories, `.agents/skills/<name>` (plus a relative `.claude/skills/<name>` symlink) and `.mem/docs/<name>.md`. `promote.go` has `Promote` (copy into the library clone, commit with the project's Git identity, push) and `Reset`. `setup.go` has `SetupPath` (`.mem/setup.md`), `Library.Setup` (the chosen templates' `setup.md` files joined in order, CRLF normalised; `Template.HasSetup` marks them) and `SetupProgress` (ticked and total `## [ ]` step headings).
- **`internal/selfupdate`**: `Latest` reads the tag from the redirect of `<releases>/latest` (`MEM_RELEASES_URL` overrides the base URL), `Newer`/`IsRelease` compare `vX.Y.Z` versions, `Install` downloads `mem_<tag>_<os>_<arch>[.exe]`, verifies it against `checksums.txt` and renames it over the executable.
- **`internal/runnables`**: `Run` executes each file in `.mem/runnables/` from the repo root in name order, 15 s timeout, output capped at 20,000 bytes.
- **`internal/importer`**: `Import` converts `.agent_core/` config (dropping the harness's placeholder description), memories, docs, specs and tasks (open, `completed/` and `abandoned/`), todos (open and `claimed/`), logs, structure doc and old files/tree_dirs/runnables settings; it leaves the originals in place. `withIssue` keeps a record's GitHub issue link in its body; `userMappings` turns GitHub usernames (case-insensitively) into mem identities for specs, todos and logs, renaming log files to match; `retiredDocs` leaves out the stock `coding_general.md`/`coding_testing.md` of earlier harness versions; `titledLog` gives every log a `# Work Log - <title>` heading (promoting the later harness's plain title line, or naming the session's date and moving top-level section headings down a level); `withoutNextSteps` removes each log's "What Comes Next" section, keeping the newest one in `Summary.NextSteps` for the import command to show.
- **`internal/git`**: `Run`, `RunEnv` (adds `GIT_TERMINAL_PROMPT=0` and `InternalEnv`, `MEM_GIT=1`, which mem's post-commit hook skips), `Toplevel`, `CurrentBranch`, `UserName`, `UserEmail`, `CommitPaths` (commits only the given paths and pushes when there is an upstream), `Push` (its errors match `ErrPush` and describe a rejected push in one line).

## Data Flow

**Onboard** (`cli/onboard.go`, with `onboard_updates.go`, `onboard_knowledge.go` and `onboard_setup.go`): `autoUpdate` (unless `--offline` or `MEM_NO_UPDATE=1`; a newer release replaces the executable and `rerun` hands the invocation to it with `MEM_NO_UPDATE=1`) → load project → `readKnowledge` (memories and skill files) → `converge.Sync` (or `Local` with `--offline`) → `applyUpdates` (schema `Upgrade`, `agentsmd.Refresh`, `agentsmd.WriteGuide`, `ensureIgnored` for `/.mem/local/` with a warning when files there are tracked, `claude.SyncCompactHook` (left unpublished when `.gitignore` ignores `.claude/settings.json`), `publish` of changed mem files, `hooks.Sync`) → `syncTemplates` (open the library, `templates.Sync`, publish unless the paths had uncommitted edits) → `afterUpdates` (refresh ahead/behind and the unpushed nudge) → `writeContext` into a buffer, led by the memories and skills that differ from the first `readKnowledge` (`renderKnowledgeChanges`): a pending `.mem/setup.md` under 🏗️ SETUP (`readSetup`), structure doc (with drift), `.mem/docs/*.md`, runnable output, active spec and tasks, open specs, open todos (with age and claimer), the user's latest log in full plus other logs of the last 14 days by title (`recentLogs`) → print shared-codebase report, release status and updates → print the context inline, or write it to `.mem/local/onboard.md` when over 14,000 bytes → print the agent instruction, which leads with the setup and stops there while setup steps remain.

**Sync and compaction** (`cli/sync.go`, `cli/compact.go`): `catchUp` snapshots `readKnowledge` → `converge.Sync` (which records the upstream revision before and after the fetch in `Report.Before`/`After`) → reload project → `syncTemplates` → `afterUpdates` → `newerRelease` nudge → `readIncoming` (`git log` of `Before..After`, capped at 10, and `work.RecordChanges` over the changed `.mem/specs` and `.mem/todos` paths via `git show`) → snapshot again. `share` (used by `mem sync`, `task complete` and `spec complete`) runs `catchUp`, then `converge.Push`, and prints the report, templates, 📥 INCOMING and changed memories/skills, returning the instruction lines they call for; `mem hook compact` (run by Claude Code's `SessionStart`/`compact` hook) runs `converge.Push` after the catch-up and prints a short digest with the active spec, claimed todos, work commits since the last log or completed task (`checkpoint.WorkSinceLog`, asking for a log from 5), a pending setup's progress, capped incoming changes and changed memories, and exits 0 on any error.

**Work records**: `cli` commands call `work` functions that read and write Markdown files under `.mem/specs/`, `.mem/specs/archive/`, `.mem/todos/` and `.mem/logs/`. `spec start` and `todo claim` publish their record; onboard publishes its own updates. `task complete` and `spec complete` commit the spec's record on its own with `commitRecord` (`Complete task <slug>`, `Complete spec <slug>`; it warns when other files stay uncommitted), then run `share`. `log new` prints `driftNudges` from `converge.Local`. `log commit` (a checkpoint, never the end of a session) refuses logs with unfilled template placeholders (`work.Log.Unfilled`), commits changed `.mem/` paths with `git.Commit`, runs `converge.Sync` and `converge.Push` and reports uncommitted files outside `.mem/`.

**Templates**: `init --template`/`template use` resolve the library (flag, project config, user config), `templates.Open`, then `templates.Sync`; `init` also writes `Library.Setup` to `.mem/setup.md` (`template use` only mentions a setup). In a repository without commits `init` first makes an empty commit on the production branch (`firstCommit`). `template promote` runs `templates.Promote` against the cached clone and pushes the library; other projects receive the item at their next onboard sync.

**Promotion**: `promote staging|production` → `release.Prepare` → print plan → `release.Execute`. Production tags with `GeneratedMessage` unless `[release] notes` is set; then, without `--confirm`, it writes the notes draft (`writeDraft`) and stops, and `--confirm` reads it (`confirmedNotes`). With `production_pr`, `promote_pr.go` first handles an open promotion pull request (report, or complete by fast-forward and tag with its body), and the release opens one (`openPromotionPR`) instead of pushing. `deploy.go`: push development, then staging and production with `GeneratedMessage`; refuses `production_pr` projects unless given the hidden `--force`.

## Entry Points

- `cmd/mem/main.go`: `cli.New().ExecuteContext(ctx)`; errors print `Error: …` to stderr with exit code 1.
- Development: `go run ./cmd/mem` or `go build -o dist/mem-dev ./cmd/mem`. Local builds report version `dev`.
- Release: GoReleaser builds `./cmd/mem` for darwin/linux/windows × amd64/arm64 with `-X …/buildinfo.Version={{.Tag}}` and `Commit`. `go install …@vX.Y.Z` builds take the version from module info.
- In this repository, the installed `mem` on PATH runs the workflow (onboard, records, hooks) while the source runs via the development commands above.

## Commands and Workflows

- `go test ./internal/<pkg>/ -run <Test>` for focused tests; `go vet ./internal/<pkg>` for affected packages.
- CI (`.github/workflows/ci.yml`): on every push and pull request, `go test ./...`, `go vet ./...` and `go build` on ubuntu, macos and windows.
- Release (`.github/workflows/release.yml`): runs GoReleaser on pushed tags matching `v[0-9]+.[0-9]+.[0-9]+`. `GORELEASER_CURRENT_TAG` is set to the pushed tag, because the date tag from `mem promote production` sits on the same commit. mem's own date tags from `mem promote production` do not match. Installed release builds pick up a new release at their next onboard.
- Branches for this repository: `dev` → `test` → `main` with `protect = true`.

## External Interfaces

- The `git` executable and the configured remote (default `origin`); network access is otherwise limited to `git fetch`/`git push`, the self-update, and the GitHub API below.
- The filesystem of the target repository: `.mem/`, `AGENTS.md`, `.gitignore`, `.claude/settings.json` and the Git hooks directory.
- Claude Code runs `mem hook compact` through the `SessionStart` hook after compaction; its stdout joins the agent's context.
- `sh` for hooks and runnables.
- The template library: any Git URL, cloned into the user cache.
- HTTPS to `github.com/Benjamin-van-Heerden/mem/releases` for the self-update (no API token).
- GitHub's REST API (`internal/github`), only in projects with `[release] production_pr = true`: promotion pull requests, with a token from `GITHUB_TOKEN`, `GH_TOKEN` or `gh auth token`.

## Tests and Verification

- Colocated `_test.go` files per package: `agentsmd`, `claude`, `cli` (`init`, `log`, `onboard`, `onboard_setup`, `sync`, `compact`, `template`, `promote`, `promote_pr` (a fake GitHub API backed by the test's bare repository) and `deploy` tests driving `New()` end to end), `converge`, `git`, `github` (an `httptest` API), `hooks`, `importer`, `release`, `selfupdate` (an `httptest` release server), `structure`, `templates`, `work`. There is no `tests/` directory yet.
- Tests build throwaway repositories in `t.TempDir()`, often with a bare repository as the remote or template library, and drive the real `git` executable. Tests that touch the user cache set `HOME` (and clear `XDG_CACHE_HOME`) to a temp directory. They must not touch GitHub or need credentials.
- For end-to-end checks of `init`, `import` or onboard, run `dist/mem-dev` in a disposable repository, never in this checkout. Checks against real GitHub use throwaway repositories that are deleted afterwards.
- CI runs on Windows too, and catches what macOS does not (CRLF checkouts, path lengths): check the run after every push to `dev`, not only before a release.

## Conventions and Patterns

- stdout is product behavior: use `output.Heading`, `output.Section` and `output.Instruction` for the separator/heading style, and make instructions specific to the current state with concrete commands.
- Commands return errors to cobra rather than exiting; recoverable problems become ⚠️ lines or instruction text.
- Git access goes through `git.Run`/`RunEnv`; helpers such as `refExists` are small local functions inside the package that needs them.
- Record files are Markdown with YAML frontmatter; identifiers are readable slugs (`work.recordSlug`: title split on whitespace and punctuation but not hyphens, filler and repeated words dropped, at most five words), with ordered task files `NN_<slug>.md`.
- Changes to the managed instructions go into `internal/agentsmd/instructions.md`; the copy in `AGENTS.md` is regenerated.
- Text read from Git checkouts, such as template library files, may have CRLF line endings on Windows; normalise to LF before writing it into a project.
- Package doc comments and short function comments explain intent; code otherwise stays uncommented.
