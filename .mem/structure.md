# Codebase and Structure

## Overview

mem is a single Go command-line executable (`cmd/mem`) that builds context for coding agents and keeps work records in a Git repository. Each project it manages holds its state as plain files: `.mem/` (config, structure doc, docs, runnables, specs, todos, logs) and a managed block in `AGENTS.md`. mem shells out to the `git` executable for all repository operations; it has no database or server.

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
  project/               .mem/config.toml, project loading, identity, schema patches
  git/                   thin wrapper around the git executable
  converge/              fetch, fast-forward/rebase and drift nudges
  structure/             .mem/structure.md template, drift measurement, file tree
  work/                  specs, tasks, todos and logs as Markdown with YAML frontmatter
  release/               promotion planning/execution and release status
  hooks/                 pre-push / pre-commit hook install and checks
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

- **`internal/cli`**: builds the cobra tree in `root.go` (`New`, `app{dir}` for the global `--dir` flag). Commands: `init.go`, `onboard.go`, `sync.go`, `spec.go`, `task.go`, `todo.go`, `log.go`, `memory.go`, `structure.go`, `promote.go`, `import.go`, `version` (in `root.go`) and the hidden `hook.go`. Shared helpers in `root.go`: `publish` (commit and push mem-owned paths via `git.CommitPaths`), `table`, `readAgents`/`writeAgents`. Depends on every other internal package; nothing depends on it except `cmd/mem`.
- **`internal/project`**: `Config`/`GitConfig`/`StructureConfig` (TOML), `Load` (finds the Git toplevel and reads `.mem/config.toml`, rejecting newer schemas), `WriteConfig`, `User` (slug of `git config user.name`), `Slugify`, and `Project.Path`/`Rel` helpers rooted at `.mem/`. `patches.go` holds `Upgrade` with an empty per-schema patch map (`Schema = 1`).
- **`internal/agentsmd`**: `Install` (block first, then memories, then existing content), `ReplaceLegacy` (swaps the Python harness block in place), `Refresh` (rewrites the block from the embedded `instructions.md` unless the stamp `<!-- Managed by mem X -->` is a newer semver), and memory CRUD (`Memories`, `SetMemory`, `RemoveMemory`) over `## name` sections inside `<memories>`. `span` locates each block by its tags and requires exactly one of each.
- **`internal/converge`**: `Sync` (fetch with a 20 s limit, then fast-forward or rebase unpushed commits, aborting on conflict) and `Local` (inspect cached refs only). Both return a `Report` with `Done` and `Nudges`: unpushed commits, drift from the development branch, missing remote, being on staging/production, and uncommitted work above 15 code files or 800 lines.
- **`internal/structure`**: `Template` for the doc, `Measure`/`ChangesSince` (code changes since the last commit touching `.mem/structure.md`, including untracked files; `Stale` at more than 5 code files or 1,000 changed lines), `Relevant` (excludes Markdown, lockfiles, binaries, vendored/build dirs and `[structure] ignore` patterns), and `Tree` (tracked and unignored files, capped at 300 entries).
- **`internal/work`**: record types `Spec`, `Task`, `Todo`, `Log` with frontmatter structs, create/find/list/complete/claim/archive functions, and the templates for specs and logs. `markdown.go` has `ReadMarkdown`/`WriteMarkdown` and `resolve`, which matches a reference by exact slug, then case-insensitive title, then unique slug prefix.
- **`internal/release`**: `Prepare` (fetch, compute commits, completed specs, divergence and the next date tag `vYYYY.MM.DD.N`), `Execute` (one atomic push of the branch and, for production, an annotated tag, with `MEM_PROMOTE=1` so the hook allows it), `CurrentStatus` for onboard.
- **`internal/hooks`**: `Sync` installs or removes `pre-push`/`pre-commit` scripts per `protect`, never overwriting non-mem hooks. The scripts run `mem hook <name>` and exit 0 when no mem with hook support is on PATH. `PrePush` and `PreCommit` implement the checks.
- **`internal/runnables`**: `Run` executes each file in `.mem/runnables/` from the repo root in name order, 15 s timeout, output capped at 20,000 bytes.
- **`internal/importer`**: `Import` converts `.agent_core/` config, memories, docs, specs and tasks, todos, logs, structure doc and old files/tree_dirs/runnables settings; it leaves the originals in place.
- **`internal/git`**: `Run`, `RunEnv` (adds `GIT_TERMINAL_PROMPT=0`), `Toplevel`, `CurrentBranch`, `UserName`, `CommitPaths` (commits only the given paths and pushes when there is an upstream).

## Data Flow

**Onboard** (`cli/onboard.go`): load project → `converge.Sync` (or `Local` with `--offline`) → `applyUpdates` (schema `Upgrade`, `agentsmd.Refresh`, `publish` of changed mem files, `hooks.Sync`) → `writeContext` into a buffer: structure doc (with drift), `.mem/docs/*.md`, runnable output, active spec and tasks, open specs, open todos, recent logs (up to three of the user's, filled to five) → print shared-codebase report, release status and updates → print the context inline, or write it to `.mem/local/onboard.md` when over 14,000 bytes → print the agent instruction.

**Work records**: `cli` commands call `work` functions that read and write Markdown files under `.mem/specs/`, `.mem/specs/archive/`, `.mem/todos/` and `.mem/logs/`. `spec start` and `todo claim` publish their record; onboard publishes its own updates. `task complete`, `spec complete` and `log new` print `driftNudges` from `converge.Local`.

**Promotion**: `promote staging|production` → `release.Prepare` → print plan (production requires `--notes` to execute) → `release.Execute`.

## Entry Points

- `cmd/mem/main.go`: `cli.New().ExecuteContext(ctx)`; errors print `Error: …` to stderr with exit code 1.
- Development: `go run ./cmd/mem` or `go build -o dist/mem-dev ./cmd/mem`. Local builds report version `dev`.
- Release: GoReleaser builds `./cmd/mem` for darwin/linux/windows × amd64/arm64 with `-X …/buildinfo.Version={{.Tag}}` and `Commit`. `go install …@vX.Y.Z` builds take the version from module info.
- In this repository, the installed `mem` on PATH runs the workflow (onboard, records, hooks) while the source runs via the development commands above.

## Commands and Workflows

- `go test ./internal/<pkg>/ -run <Test>` for focused tests; `go vet ./internal/<pkg>` for affected packages.
- CI (`.github/workflows/ci.yml`): on every push and pull request, `go test ./...`, `go vet ./...` and `go build` on ubuntu, macos and windows.
- Release (`.github/workflows/release.yml`): runs GoReleaser on pushed tags matching `v[0-9]+.[0-9]+.[0-9]+`. mem's own date tags from `mem promote production` do not match.
- Branches for this repository: `dev` → `test` → `main` with `protect = true`.

## External Interfaces

- The `git` executable and the configured remote (default `origin`); network access is limited to `git fetch`/`git push`.
- The filesystem of the target repository: `.mem/`, `AGENTS.md`, `.gitignore` and the Git hooks directory.
- `sh` for hooks and runnables.
- No HTTP APIs or GitHub API calls.

## Tests and Verification

- Colocated `_test.go` files per package: `agentsmd`, `cli` (`init_test.go`), `converge`, `git`, `hooks`, `importer`, `release`, `structure`, `work`. There is no `tests/` directory yet.
- Tests build throwaway repositories in `t.TempDir()`, often with a bare repository as the remote, and drive the real `git` executable. They must not touch GitHub or need credentials.
- For end-to-end checks of `init`, `import` or onboard, run `dist/mem-dev` in a disposable repository, never in this checkout.

## Conventions and Patterns

- stdout is product behavior: use `output.Heading`, `output.Section` and `output.Instruction` for the separator/heading style, and make instructions specific to the current state with concrete commands.
- Commands return errors to cobra rather than exiting; recoverable problems become ⚠️ lines or instruction text.
- Git access goes through `git.Run`/`RunEnv`; helpers such as `refExists` are small local functions inside the package that needs them.
- Record files are Markdown with YAML frontmatter; identifiers are readable slugs (`project.Slugify`), with ordered task files `NN_<slug>.md`.
- Changes to the managed instructions go into `internal/agentsmd/instructions.md`; the copy in `AGENTS.md` is regenerated.
- Package doc comments and short function comments explain intent; code otherwise stays uncommented.
