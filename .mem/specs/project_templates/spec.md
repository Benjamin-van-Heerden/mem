---
title: Project templates
status: active
assigned_to: benjamin_van_heerden
created_at: "2026-09-25T13:57:57+02:00"
updated_at: "2026-09-25T14:29:14+02:00"
---

## Overview

Project templates give a new or existing mem project a starting set of memories, skills and docs for its kind of project (for example `nextjs-web`), and keep that set flowing between projects. Templates live in a separate Git repository, the template library. A project names the templates it uses in `.mem/config.toml`. `mem init --template <name>` draws the items in, every onboard adds what is missing and updates what the project has not changed, and `mem template promote` sends a project's memory, skill or doc back to its template so the next onboard of every similar project picks it up.

## Goals

- `mem init --template nextjs-web` produces a project with that template's memories in `AGENTS.md`, its skills in `.agents/skills/` (linked from `.claude/skills/`) and its docs in `.mem/docs/`.
- An existing project can adopt a template with `mem template use <name>`.
- Onboard pulls the library and brings the project up to date with its templates: adds missing items, updates items the project has not edited, and flags items edited on both sides. It never deletes project files.
- Deleting a template item in a project records an opt-out, so onboard does not bring it back.
- `mem template promote <memory|skill|doc> <name>` copies a project item into its template, commits and pushes the library. Other projects receive it at their next onboard.
- Everything is ordinary Git: the library is a normal repository, project state is committed files, no locks or services.

## Technical Approach

### Template library

A Git repository, for example `github.com/Benjamin-van-Heerden/mem-templates` (private or public; Git handles authentication). Layout:

```text
<template>/template.toml          description = "Next.js App Router web app"
<template>/memories/<name>.md     memory body, exactly as it appears under `## <name>` in AGENTS.md
<template>/skills/<name>/...      a skill directory: SKILL.md plus any supporting files
<template>/docs/<name>.md         a project doc
```

A template directory is recognized by its `template.toml`. There is no inheritance; a project that wants shared items lists several templates (for example `["base", "nextjs-web"]`). When two listed templates provide the same item, the later one wins, and `mem template list` says so.

mem keeps one clone per library URL in `os.UserCacheDir()/mem/templates/<slug of URL>`, cloned on first use and updated with `git pull --ff-only` (same time limit as the converge fetch; skipped with `onboard --offline`). The clone is disposable: promote commits into it and pushes immediately, and on a failed pull mem warns and continues with the cached copy.

### Project configuration and state

```toml
[templates]
source = "git@github.com:Benjamin-van-Heerden/mem-templates.git"
use = ["nextjs-web"]
exclude = ["skill:old-router-guide"]     # opt-outs, written by mem
```

The library URL is stored per project so every teammate syncs from the same library. `mem init --template` takes `--template-source <url>`; when it is omitted, init uses `template_source` from a user-level file `os.UserConfigDir()/mem/config.toml`, and fails with an instruction to pass or set it when neither exists. `mem template use` behaves the same way.

`.mem/templates.lock` (TOML, committed) records, for every item mem installed, its template and the content hash that was last in sync:

```toml
[[item]]
kind = "skill"        # memory | skill | doc
name = "next-app-router"
template = "nextjs-web"
hash = "sha256:…"      # memory body; doc file; skill directory (sorted relative paths and contents)
```

### Sync

`internal/templates` provides one function, used by init, `template use` and onboard, that reconciles each item from the listed templates (L = project copy, T = template copy, R = hash in the lock):

| State | Action |
| --- | --- |
| Excluded | Skip |
| L missing, no R | Install T, record R = T |
| L missing, R present | The project deleted it: add to `exclude`, drop R, report the opt-out |
| L = T | Record R = T if different |
| L = R, T ≠ R | Template updated it: overwrite with T, record R = T |
| L ≠ R, T = R | Local edit only: report that it can be promoted |
| L ≠ R, T ≠ R | ⚠️ Edited on both sides: leave L, report with the commands to keep the local version (promote) or take the template's (`mem template reset <kind> <name>`) |
| R present, item gone from template | Keep L, drop R, report |

Kinds:

- **memory**: read and written through `agentsmd.Memories`/`SetMemory`; hash the body.
- **skill**: copied to `.agents/skills/<name>/`; create the relative symlink `.claude/skills/<name>` → `../../.agents/skills/<name>` when it does not exist. If the symlink cannot be created (for example on Windows without symlink permission), print ⚠️ with the manual step and continue.
- **doc**: copied to `.mem/docs/<name>.md`, so onboard includes it as it does today.

Onboard runs the sync inside `applyUpdates`, shows the results under a `🧩 TEMPLATES` section, and publishes the changed paths (`AGENTS.md`, `.agents/skills`, `.claude/skills`, `.mem/docs`, `.mem/templates.lock`, `.mem/config.toml`) with the existing `publish` helper, on the same terms as today's AGENTS.md refresh (not when those paths already had uncommitted edits).

### Commands

- `mem init --template <name>` (repeatable) and `--template-source <url>`: write `[templates]`, run the sync, list what was installed.
- `mem template use <name>`: add a template to an initialized project and sync.
- `mem template list`: templates in the library with descriptions, and for this project each item's state (in sync, local edit, both edited, excluded).
- `mem template promote <memory|skill|doc> <name> [--to <template>]`: copy the project item into the library clone (`--to` required when the project uses more than one template and the item is not already recorded), commit `Promote <kind> <name> from <project>`, push, and record R. A promoted item that did not come from a template becomes tracked.
- `mem template reset <kind> <name>`: replace the local item with the template's copy and record R.

`mem promote` (release promotion) is unchanged; template promotion lives under `mem template` to keep the two apart.

### Documentation

- `internal/agentsmd/instructions.md`: one short entry point under Memories: when the user wants a memory or skill available in other projects, `mem template promote`. Keep usage details in CLI help.
- `docs/design.md`: a Templates section describing the above. `docs/status.md`: move Templates from Planned to Implemented. `README.md`: `--template` in Set up a project.
- `.mem/structure.md`: the new package and commands.

## Success Criteria

- In a disposable repository, `mem init --template nextjs-web --template-source <bare library>` writes `[templates]`, installs the template's memories into `AGENTS.md`, skills into `.agents/skills/<name>/` with `.claude/skills/<name>` symlinks, docs into `.mem/docs/`, and `.mem/templates.lock`.
- Each row of the sync table is covered by a test in `internal/templates` against a bare library repository and passes.
- Onboard in project A after `mem template promote skill <name>` in project B installs the skill in A, without network access to anything but the local bare library.
- Deleting a template skill in a project and running onboard adds `skill:<name>` to `exclude` and does not reinstall it.
- A skill edited in both the project and the template is left unchanged and reported with ⚠️ and the promote/reset commands.
- `mem onboard --offline` does not pull the library and still syncs from the cached clone.
- `go vet` is clean for the affected packages; design, status, README, instructions and the structure doc describe the implemented behavior.

## Notes

- Decided with the user (2026-09-25): separate Git repository for the library; skills in `.agents/skills` with `.claude/skills` symlinks, matching the user's home setup; sync adds, updates unedited items and flags conflicts, never deletes; deletions become recorded opt-outs.
- Proposed, to confirm: the user-level `template_source` default; multiple templates per project instead of inheritance; framework guides as skills rather than docs, because docs are printed in full at every onboard while skills load on demand.
- Seed material: `old/coding/optional_docs/` has guides for Next.js, TanStack, Rust, Python, uv, Elixir and Phoenix (about 3,100 lines), candidates for the first template skills.
- The library repository itself has to be created on GitHub by the user before promote can push to it; tests use local bare repositories only.
