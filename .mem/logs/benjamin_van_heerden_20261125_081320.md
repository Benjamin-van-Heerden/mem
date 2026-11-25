---
created_at: "2026-11-25T08:13:20+02:00"
user: benjamin_van_heerden
commits: [7219ffb, c2c4324, fa1b555, ef55cf9]
---

# Work Log - Leaner managed block, skills for situational memories; staging for v0.17.0

## Goal

Make room in AGENTS.md for project memories by trimming mem's managed block. Steer conventions that matter only for some tasks into skills. Make sure Claude Code sees every project skill. Then release these changes.

## What Was Accomplished

### Trimmed the managed block (c2c4324)

- `internal/agentsmd/instructions.md` drops what it said twice: onboard at session start, the development-branch and push rules, the staging/production rule, the communication and abstraction guidance, and the repeated testing advice. It also drops detail that command output and help already give, such as how `task complete` and `spec complete` behave.
- Functional Approach is folded into Code Quality. The generic Performance Considerations are gone. File Size, Goal-Driven Execution and Testing Philosophy are shorter.
- The staged block went from 19.4 KB to 15.8 KB. `instructions_folder.md` got the matching trims, and `docs/cloud_agents.md` was updated to match.

### Size warning suggests skills (fa1b555, 3cdc656)

- `agentsSizeWarning(text, folder)` in `internal/cli/memory.go` now also proposes turning memories that matter only for some tasks into skills. It says where a skill goes: `.agents/skills/<name>/` linked from `.claude/skills/<name>`, or `.claude/skills/<name>/` in a folder project. It then asks for `mem memory remove <name>`. Conventions that must hold in every session stay memories.
- Callers pass `p.Folder()`: `memory set`, `onboard.go` and `folder.go`. `folder_test.go` checks the folder wording.

### Onboard links every project skill (ef55cf9)

- `templates.LinkSkills(root)` in `internal/templates/sync.go` links `.claude/skills/<name>` to each `.agents/skills/<name>/SKILL.md` that has no link yet, reusing `linkSkill`. Before this, only skills installed from a template got a link, so a skill an agent wrote by hand stayed invisible to Claude Code.
- `applyUpdates` (`onboard_updates.go`) reports each new link and handles it one of three ways:
  - If `.gitignore` ignores the link, it stays local, and onboard makes it in every checkout.
  - If the skill is not committed yet, onboard asks the agent to commit the link with it.
  - Otherwise the link is added to `publishPaths`.
- The managed block's Memories section now says that a convention needed only for some tasks fits a skill, and that onboard links it.
- `TestOnboardLinksSkillsAnAgentWroteSoClaudeCodeFindsThem` covers all of this. A teammate's committed skill gets its link published. A local uncommitted skill gets linked with the commit-it-together line. A second onboard makes no links.

### Other

- `7219ffb` pins Go 1.27.4 in `.tool-versions`.
- `31c7c49` is the installed mem's onboard updating this repository's own files: `.mem/README.md` added, `.mem/config.toml` upgraded, and the AGENTS.md block refreshed.

### Release

- `mem promote staging` ran `go vet ./...` and `go test ./...` (both passed) and moved `test` from `26c337f` to `ef55cf9`.
- Production goes out as v0.17.0, a minor bump because onboard has new behaviour.

### Checked: idle compaction

Claude Code's idle compaction summarizes a long idle conversation and applies the summary on the next message. Three earlier session transcripts in this project show `SessionStart:compact` running `mem hook compact` when the summary was applied, not when it was written. So the digest and catch-up reflect the repository as it is when the user comes back. No change was needed.

## Decisions

- **Situational conventions go in skills, not memories.** A skill costs no context unless its description fits the task, which keeps AGENTS.md under Codex's 32 KiB read limit.
- **Onboard links skills for every project, not only those from templates,** and commits a link only when its skill is already committed.
- **Idle compaction needs no handling in mem**, because the compaction hook already runs when the summary is applied.

## Key Files Affected

- `internal/agentsmd/instructions.md`, `instructions_folder.md`, `readme.md`: trimmed block, skill guidance, guide text
- `internal/cli/memory.go`, `onboard.go`, `folder.go`: size warning with skill advice, folder wording
- `internal/cli/onboard_updates.go`, `internal/templates/sync.go`: `LinkSkills` and how each new link is reported
- Tests: `internal/cli/onboard_test.go`, `memory_test.go`, `folder_test.go`
- `docs/design.md`, `docs/status.md`, `docs/cloud_agents.md`, `.mem/structure.md`
- `.tool-versions`
