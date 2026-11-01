---
created_at: "2026-11-01T19:23:14+02:00"
user: benjamin_van_heerden
commits: [24f8dca, 37c2487]
---

# Work Log - Folder structure doc and self-building folder projects

## Goal

Finish folder mode for its first release: give folder projects a structure doc that suits them, and make their instructions push agents to automate recurring work.

## What Was Accomplished

### Folder structure doc

- `structure.FolderTemplate` now has Overview, Layout, Where Things Go, Sources of Truth, and Tools and Commands. The `mem structure` folder instruction asks for patterns rather than an inventory, and points agents at the README, scripts, memories and sample documents.

### Skills and automation in the folder block

- `## Skills` in `instructions_folder.md` became `## Skills and Automation`: the folder is meant to grow into something that largely runs itself, with four rules: look first, create skills without being asked, script what is mechanical, keep skills true.
- A `SKILL.md` description must say when to use the skill, since that is how a later session finds it.
- Work Logs gained a check: before writing a log, create any skill or script that recurring work still lacks.
- Simplicity First and the Surgical Changes test now exempt skills and scripts, which would otherwise have forbidden unrequested work. The exemption is limited to what the work has actually needed.
- `go test ./internal/agentsmd/`, `go test ./internal/cli/ -run 'Folder'` and `go vet` on both pass. A fresh `dist/mem-dev init --folder` writes the new text into `AGENTS.md`.

### Review

- The user reviewed the folder block and approved it for release, so the todo `review_folder_instructions_block` is deleted.

## Decisions

- **Agents create skills unasked and do not wait for a task to happen twice.** The user wants these folders to evolve into self-sustaining entities, so creating a skill is a standing part of the work and the agent tells the user afterwards. Genuine one-offs stay one-offs.
- **The structure doc of a folder project holds what stays true as documents come and go.** Nothing measures its drift, so it describes patterns and filing rules, and live state stays in runnables, todos and logs.
- **Files under `.claude/skills/` stay out of the work-log nudge.** The walk skips dot-paths and this was left as it is; a session that only builds skills gets no file-count nudge, while the log at wrap-up still applies.

## Key Files Affected

- `internal/agentsmd/instructions_folder.md`: Skills and Automation section, log check, exemptions.
- `internal/structure/structure.go`, `internal/cli/structure.go`, `internal/cli/folder_test.go`: folder structure doc template and instruction.
- `docs/design.md`, `README.md`: describe both changes.
- `.mem/todos/move_work_finance_folder_mode.md`: recurring work such as new clients becomes skills during the migration.
