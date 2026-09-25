# mem

mem builds context for coding agents and standardizes how work is done in a repository. It is one executable on your PATH; each repository keeps its own state in `.mem/` and `AGENTS.md`, as ordinary files in Git.

It is made for solo developers and small teams in daily contact, and it is built around one idea: everyone works on the same codebase. mem keeps checkouts converged automatically where that is safe and nudges firmly where it is not.

## Install

With Go installed:

```sh
go install github.com/Benjamin-van-Heerden/mem/cmd/mem@latest
```

This puts `mem` in Go's binary directory (`~/go/bin` by default), which must be on your PATH. Without Go, download the binary for your platform from the [latest release](https://github.com/Benjamin-van-Heerden/mem/releases/latest), make it executable and move it to a directory on your PATH as `mem`.

`mem version` shows the installed version. Everyone on a team should run the same release; onboard warns when a teammate's newer mem has written the project's instructions.

Releases are published by pushing a `v*` tag to GitHub.

## Set up a project

In an existing Git repository:

```sh
mem init
```

This writes `.mem/config.toml`, adds mem's instructions and a memories section to `AGENTS.md` (existing content is kept), ignores `.mem/local/`, creates any missing development, staging and production branches (from the remote where they exist, otherwise from production), publishes them, switches to development and installs the Git hooks. Commit the result on the development branch and push it.

Branch names default to `dev`, `test` and `main`:

```sh
mem init --development development --staging test --production production
```

Solo projects that do not want the hooks: `mem init --protect=false`.

A project that uses the Python coding harness (`.agent_core/`) is converted instead:

```sh
mem import agent-core
```

The import converts config, memories, docs, specs, todos and logs, replaces the old instruction block in `AGENTS.md`, and leaves `.agent_core/` in place for you to remove once you are happy with the result.

## How it works

Agents learn the workflow from the mem block in `AGENTS.md` and from the instructions mem commands print. The main pieces:

- **`mem onboard`** starts every session. It fetches and brings the checkout up to date (fast-forwarding, or rebasing unpushed commits, when that is safe), refreshes mem's instructions, and prints the project context for the agent. Anything that needs attention, such as diverged history or unpushed work, is flagged with ⚠️ for the agent to raise with you. `mem sync` does the Git part on demand.
- **Specs, tasks and todos** track planned work. A spec is written up with the user, broken into ordered tasks and implemented task by task; todos are standalone matters. None of them are needed for ordinary coding.
- **Work logs** carry context between sessions: what was done, what failed and what comes next. Onboard shows the most recent ones.
- **Memories** are lasting project conventions, kept in `AGENTS.md`.
- **The structure doc**, `.mem/structure.md`, is a living map of the codebase that onboard includes. `mem structure` creates it and later lists what has changed since it was last updated; mem warns when it falls behind the code.

## Releases

```text
dev ──mem promote staging──▶ test ──mem promote production──▶ main
```

Work lands on the development branch, which does not deploy. `mem promote staging` fast-forwards staging for a preview release (`--to <commit>` leaves unfinished work out). `mem promote production` fast-forwards production to exactly what staging previewed and tags it with release notes (`v2026.09.24.1`). Staging and production never get commits of their own, so there is always a single history. CI deploys staging and production on push.

The hooks refuse direct pushes to staging and production and commits made on them, pointing to `mem promote` instead.

## Project layout

```text
AGENTS.md                  mem block, your content, project memories
.mem/
  config.toml
  structure.md             codebase and structure map
  docs/*.md                documents included in onboard
  runnables/*              executables whose output is included in onboard
  specs/<slug>/            spec.md and tasks/; completed specs move to specs/archive/
  todos/<slug>.md
  logs/<user>_<date>_<time>.md
  local/                   ignored; generated onboard output
```

```toml
schema = 1
name = "my-app"
description = "One line about the project"

[git]
remote = "origin"
development = "dev"
staging = "test"
production = "main"
protect = true

[structure]
ignore = ["migrations/**"]   # optional: paths that do not count as code changes
```

Runnables are for context that is better extracted than pasted whole, such as design tokens, UI components or database models. Each executable in `.mem/runnables/` runs at onboard from the repository root, in name order, with a 15 second limit.

## Developing mem

Use Go 1.27.1 or later. Run the source with `go run ./cmd/mem` or build `dist/mem-dev`; the `mem` on your PATH may be a different build. Try changes in disposable repositories, not in this checkout. See [AGENTS.md](AGENTS.md) for contributor conventions, [docs/design.md](docs/design.md) for the design and [docs/status.md](docs/status.md) for what is implemented.
