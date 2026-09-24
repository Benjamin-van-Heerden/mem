# memr

memr builds context for coding agents and standardizes how work is done in a repository. It is one executable on your PATH; each repository keeps its own state in `.memr/` and `AGENTS.md`, as ordinary files in Git.

It is made for solo developers and small teams in daily contact, and it is built around one idea: everyone works on the same codebase. memr keeps checkouts converged automatically where that is safe and nudges firmly where it is not.

## Install

With Go installed:

```sh
go install github.com/Benjamin-van-Heerden/memr/cmd/memr@latest
```

This puts `memr` in Go's binary directory (`~/go/bin` by default), which must be on your PATH. Without Go, download the binary for your platform from the [latest release](https://github.com/Benjamin-van-Heerden/memr/releases/latest), make it executable and move it to a directory on your PATH as `memr`.

`memr version` shows the installed version. Everyone on a team should run the same release; onboard warns when a teammate's newer memr has written the project's instructions.

Releases are published by pushing a `v*` tag to GitHub.

## Set up a project

In an existing Git repository:

```sh
memr init
```

This writes `.memr/config.toml`, adds memr's instructions and a memories section to `AGENTS.md` (existing content is kept), ignores `.memr/local/`, creates the development branch if needed and installs the Git hooks. Commit the result on the development branch and push it.

Branch names default to `dev`, `test` and `main`:

```sh
memr init --development development --staging test --production production
```

Solo projects that do not want the hooks: `memr init --protect=false`.

A project that uses the Python coding harness (`.agent_core/`) is converted instead:

```sh
memr import agent-core
```

The import converts config, memories, docs, specs, todos and logs, replaces the old instruction block in `AGENTS.md`, and leaves `.agent_core/` in place for you to remove once you are happy with the result.

## How it works

Agents learn the workflow from the memr block in `AGENTS.md` and from the instructions memr commands print. The main pieces:

- **`memr onboard`** starts every session. It fetches and brings the checkout up to date (fast-forwarding, or rebasing unpushed commits, when that is safe), refreshes memr's instructions, and prints the project context for the agent. Anything that needs attention, such as diverged history or unpushed work, is flagged with ⚠️ for the agent to raise with you. `memr sync` does the Git part on demand.
- **Specs, tasks and todos** track planned work. A spec is written up with the user, broken into ordered tasks and implemented task by task; todos are standalone matters. None of them are needed for ordinary coding.
- **Work logs** carry context between sessions: what was done, what failed and what comes next. Onboard shows the most recent ones.
- **Memories** are lasting project conventions, kept in `AGENTS.md`.
- **The structure doc**, `.memr/structure.md`, is a living map of the codebase that onboard includes. `memr structure` creates it and later lists what has changed since it was last updated; memr warns when it falls behind the code.

## Releases

```text
dev ──memr promote staging──▶ test ──memr promote production──▶ main
```

Work lands on the development branch, which does not deploy. `memr promote staging` fast-forwards staging for a preview release (`--to <commit>` leaves unfinished work out). `memr promote production` fast-forwards production to exactly what staging previewed and tags it with release notes (`v2026.09.24.1`). Staging and production never get commits of their own, so there is always a single history. CI deploys staging and production on push.

The hooks refuse direct pushes to staging and production and commits made on them, pointing to `memr promote` instead.

## Project layout

```text
AGENTS.md                  memr block, your content, project memories
.memr/
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

Runnables are for context that is better extracted than pasted whole, such as design tokens, UI components or database models. Each executable in `.memr/runnables/` runs at onboard from the repository root, in name order, with a 15 second limit.

## Developing memr

Use Go 1.27.1 or later. Run the source with `go run ./cmd/memr` or build `dist/memr-dev`; the `memr` on your PATH may be a different build. Try changes in disposable repositories, not in this checkout. See [AGENTS.md](AGENTS.md) for contributor conventions, [docs/design.md](docs/design.md) for the design and [docs/status.md](docs/status.md) for what is implemented.
