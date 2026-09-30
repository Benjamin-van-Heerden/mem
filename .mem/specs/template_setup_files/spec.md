---
title: Template setup files
status: active
assigned_to: benjamin_van_heerden
created_at: "2026-09-30T11:54:52+02:00"
updated_at: "2026-09-30T11:56:41+02:00"
---

## Overview

A template can ship a one-time setup: `<template>/setup.md` at the template's root, a list of steps written as checkbox headings. `mem init --template <name>` copies it to `.mem/setup.md`. While that file exists, onboard makes it the session's first instruction: follow the steps with the user, tick each one, and delete the file when all are done. Deleting it is the completion signal; mem never installs it again. This replaces the `docs/setup.md` convention now used by `nextjs-web`, which relied on docs being printed and on deletion turning into a template exclude.

The flow this serves: create an empty GitHub repository, clone it, run `mem init --template nextjs-web`, and the agent sets the project up before any other work. So `mem init` must also work in a repository without commits.

## Goals

- Templates can provide a structured, ordered setup that an agent follows reliably, resumes after a compaction or a new session, and finishes by deleting one file.
- Setup is not a doc, memory or skill: it is installed once at init, never synced, and needs no lock entry or exclude.
- `mem init` works in a freshly cloned empty repository.
- The `nextjs-web` template in mem-templates uses the new file, with the Next.js agent-file handling made correct from the first step.

## Technical Approach

### Setup file format (authored in the template library)

```markdown
# Setup: nextjs-web

One-line purpose.

## [ ] 1. Scaffold the Next.js app

Instructions…

Done when: `package.json` names the project and `bun install` succeeded.

## [ ] 7. Link Vercel and Neon (you)

…
```

- One `## [ ] N. Title` heading per step; the agent changes `[ ]` to `[x]` when the step's "Done when" holds, and commits.
- Steps marked `(you)` need the user (accounts, secrets, choices); they come last so local development never waits on them.

### mem

- `internal/templates/setup.go` (new):
  - `const SetupPath = ".mem/setup.md"`.
  - `func (l Library) Setup(names []string) (string, error)`: reads `<template>/setup.md` for each named template that has one, in the given order, and joins them with a blank line; empty string when none has one.
  - `func SetupProgress(text string) (done, total int)`: counts headings matching `^## \[( |x|X)\] ` .
  - `Template` gains `HasSetup bool`, set in `Library.Templates()`.
- `mem init` (`internal/cli/init.go`):
  - Empty repository: when `HEAD` does not resolve, point `HEAD` at the production branch (`git symbolic-ref HEAD refs/heads/<production>`) and create an empty commit ("Initial commit"); report it in 🌿 BRANCHES. Replaces the "has no commits yet" error in `ensureBranches`.
  - After `templates.Sync`, when `lib.Setup(templateNames)` is non-empty, write it to `.mem/setup.md` and add the path to the files the instruction tells the agent to commit. Mention it in the 📄 FILES list.
  - The instruction's step 3 (run `mem onboard`) is unchanged; onboard then presents the setup.
- `mem onboard` (`internal/cli/onboard.go`):
  - When `.mem/setup.md` exists, `writeContext` starts with a `🏗️ SETUP` section holding the file's full text (so it lands in `.mem/local/onboard.md` with the rest when the context is long).
  - `renderOnboardInstruction` puts a setup step first: "Setup is pending (N of M steps done). Work through `.mem/setup.md` with the user, in order, before anything else. After each step, when its 'Done when' holds, tick its box and commit. Stop at steps marked (you) and hand them to the user. When every step is ticked, delete `.mem/setup.md`, commit and push." When every box is ticked, the step instead says the setup is finished and the file should be deleted and committed.
  - Keep `onboard.go` under 500 lines: put the setup rendering in `internal/cli/onboard_setup.go`.
- `mem hook compact` (`internal/cli/compact.go`): when `.mem/setup.md` exists, print `Setup pending: N of M steps done in .mem/setup.md.` in the digest and add "Continue the setup in .mem/setup.md." to its instruction.
- `mem template use` (`internal/cli/template.go`): never installs setup. When the added template has a setup, print one line: "<name> has a setup for new projects (`<library>/<name>/setup.md`); this project did not run it. Its steps can guide adding what is missing."
- `mem template list`: mark templates that have a setup.
- Docs: `docs/design.md` (Templates section and the `.mem/` tree), `docs/status.md`, `README.md` (layout and the init flow), and the managed instructions only if a routine agent convention changes (it should not: onboard's instruction carries it).

### mem-templates (separate repository, pushed after review)

- `nextjs-web/setup.md` replaces `nextjs-web/docs/setup.md`; the numbered procedure moves out of the `nextjs-setup` skill into it. The skill keeps the starter files, the "Adding a piece" table, deployment reference and sharp edges, and points to the setup for new projects.
- Step 1 scaffolds with `bunx create-next-app@latest scaffold-tmp … --skip-install --disable-git --no-agents-md` (no AGENTS.md or CLAUDE.md is created), moves the files in, merges only `.gitignore`, and immediately adds `agentRules: false` to `next.config.ts` so an agent-run `next dev` cannot append Next's block before the starter config is copied.
- For a project scaffolded before `mem init`, the steps remove create-next-app's `CLAUDE.md` and the `<!-- BEGIN:nextjs-agent-rules -->…<!-- END:nextjs-agent-rules -->` block from AGENTS.md.
- README: document `setup.md` in the layout and replace the `docs/setup.md` convention.

## Success Criteria

- `templates.Library.Setup` returns the setups of the named templates in order, and `SetupProgress` counts ticked and total step headings; covered by tests in `internal/templates`.
- In a clone of an empty bare repository, `mem init` succeeds: the production branch has one empty commit, the staging and development branches are created from it and published, and the checkout is on development.
- `mem init --template <t>` with a template that has `setup.md` writes `.mem/setup.md` and lists it among the files to commit; with a template without one, no file is written.
- `mem onboard` with `.mem/setup.md` present shows `🏗️ SETUP` with the file's content and puts the setup step first in the instruction with the right progress; with every box ticked it tells the agent to delete the file; without the file nothing about setup is printed.
- `mem hook compact` shows the setup progress line when the file exists.
- `mem template use` never writes `.mem/setup.md`.
- Tests cover the above in `internal/cli` (init, onboard, compact, template use); focused tests and `go vet` pass.
- End to end in a disposable repository against a local mem-templates checkout: empty clone → `mem init --template nextjs-web` → onboard shows the setup → steps followed as written → `bun run build` passes, AGENTS.md has no Next.js block and there is no CLAUDE.md, including after `next dev` run by the agent → `.mem/setup.md` deleted and onboard no longer mentions setup.
- docs/design.md, docs/status.md and README.md describe setup files; mem-templates README and nextjs-web are updated and pushed after user review.

## Notes

- Deliberately not a spec/task per step, and no commands for setup: the file is the state. Checks run by mem ("check: bun run build") were considered and left out.
- No lock entry is needed because only `mem init` installs setup; `templates.Sync` would otherwise drop unknown lock entries.
- Development is dogfooding: note friction in mem's own workflow while implementing this spec and fix small issues on the way.
- Next.js facts checked in next 16.3.7: only `next dev` writes agent files (`server/lib/start-server.js` → `ensureAgentRulesForDev`), only when `@vercel/detect-agent` detects an agent and the block is missing; it appends or replaces only its own block. `agentRules: false` in next.config disables it. create-next-app's `--no-agents-md` skips both AGENTS.md and CLAUDE.md.
