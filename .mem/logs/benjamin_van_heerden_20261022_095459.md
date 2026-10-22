---
created_at: "2026-10-22T09:54:59+02:00"
user: benjamin_van_heerden
commits: [f03eb07]
---

# Work Log - mem v0.8.3: init asks for templates first; studii-next re-initialised with nextjs-web

## Goal

`mem init` in `~/Documents/studii/studii-next` set the project up without templates although it is meant to be a Next.js app, because nothing asked. Fix that project, and make init ask for templates before it sets anything up.

## What Was Accomplished

### studii-next re-initialised with nextjs-web

The first init had committed nothing: `.mem/`, `AGENTS.md`, `.gitignore` and `.claude/` were all untracked and all init's, and the repository held only the GitHub `Initial commit` (README.md). Those files were deleted and `mem init --template nextjs-web` run again with the installed v0.8.2. It added 9 memories, 16 skills (`.agents/skills/` with `.claude/skills/` links), `.mem/templates.lock` and `.mem/setup.md`. `dev`, `test` and `main` already existed on origin and were left as they were. Nothing there is committed yet.

### init asks for templates (f03eb07)

`internal/cli/init.go`:
- New `--no-template` flag. Without it or `--template`, `RunE` calls `askForTemplates(ctx, out, librarySource(templateSource, ""))` right after the "already has .mem/config.toml" check, before `ensureBranches` or any file write, and returns.
- `askForTemplates` opens the library (pulling), prints `🧩 MEM INIT: CHOOSE TEMPLATES` with "Nothing has been set up yet.", the library URL, a TEMPLATE/SETUP/DESCRIPTION table, and an instruction: show the user the templates, ask which fit, rerun the same `mem init` with `--template <name>` per choice or `--no-template`. An empty library gets only the `--no-template` instruction. A library that cannot be opened returns the error with "run `mem init --no-template` to set up without templates".
- `--template` with `--no-template` is an error.

Tests: `TestInitWithoutTemplateChoiceListsTemplatesAndChangesNothing` (`template_test.go`) checks the output offers `nextjs-web` and `--no-template`, `git status --porcelain --ignored` is empty and the bare remote has no `dev`. The three tests that ran a plain `init` (`log_test.go`, `onboard_test.go`, `template_test.go`'s later project) now pass `--no-template`. `go test ./internal/cli/` and `go vet ./internal/cli/` passed; `dist/mem-dev init` was checked in a scratch repository against the real library.

Docs: README (init section; `--protect=false` example gains `--no-template`), `docs/design.md`, `docs/status.md`, `.mem/structure.md` (Templates data flow).

### Release v0.8.3

- `origin/dev` fast-forwarded to f03eb07 (`git push origin HEAD:dev` from this worktree branch, which now tracks `origin/dev`). CI run 39741709851 green on ubuntu, macOS and Windows.
- `mem promote staging`, `mem promote production --confirm` (date tag `v2026.10.22.1`), `v0.8.3` tagged on `origin/main` with the confirmed notes (`--cleanup=whitespace`). Release run 39743454542 built 7 assets; `mem update` moved the installed mem to v0.8.3.

## Decisions

- Init does not prompt on a terminal: agents run it without one, so it stops with a template list and an agent instruction, the same pattern as other mem commands. `--no-template` is the explicit way to set up without templates.
- Released as a patch (v0.8.3) at Benjamin's request, although `--no-template` is a new flag.

## Key Files Affected

- `internal/cli/init.go` (48 lines): `--no-template`, `askForTemplates`
- `internal/cli/template_test.go` (24 lines): new init test; later project passes `--no-template`
- `internal/cli/log_test.go`, `internal/cli/onboard_test.go` (2 lines each): `--no-template`
- `README.md` (6 lines), `docs/status.md` (2 lines), `docs/design.md` (1 line), `.mem/structure.md`
- Outside this repository: `~/Documents/studii/studii-next` (uncommitted mem setup with nextjs-web)

## Notes

- After the Release workflow finished, `mem update` reported v0.8.2 as latest for about three minutes although `curl -I …/releases/latest` already redirected to v0.8.3; repeated retries picked it up. Same lag as at v0.8.1, longer this time.
