---
created_at: "2026-10-08T16:20:27+02:00"
user: benjamin_van_heerden
spec: keep_work_current_long_sessions
---

# Work Log - anssum, ascendis and shannon migrations; keeping work current in long sessions and on feature branches; mem v0.7.6

## Overarching Goals

Continue migrating Python-harness projects (anssum-platform, ascendis-transform, shannon), then address the user's observation that sessions rarely end: make mem keep records and branches current in never-ending sessions and on feature branches, simplify todos and task commits, and release.

## What Was Accomplished

### Migrations (pushed to each project's dev)

- **anssum-platform:** description from the README; kept the personal-keys todo; added "Validate recipient assessments before Phase 2"; `.gitignore` ignores only `.claude/settings.local.json`; `.vercelignore`, `.prettierignore`, `eslint.config.mjs` and `tsconfig.json` exclude `.mem` instead of `.agent_core`.
- **ascendis-transform:** consulting conventions kept in front of agents: memory `consulting_record` (read `consulting/README.md`; chronology ordered by date learned with event dates separate; `current-position.md` holds consulting follow-ups; mem todos only for engineering) and runnable `10_current_position` printing the handover at onboard. Wrote a dated "5 October, later" handover entry (all five connector accounts exist; key delivery; Michael's exposed key; Alltech databases). README wording to mem; chronology link to the moved 2 October log. Todos: deleted #1 (covered by #3); added distribution/HTTPS options, `vw_` naming note, Alltech database decision.
- **shannon:** deployment memories `explicit_deployment_trigger_and_procedure` and `keep_deployment_branches_linear` removed (mem covers them); `docs/deployment.md` uses `mem promote` and gained the two rollout cautions; memory `production_deploy_check` points to it; `split_files_over_500_lines` removed. Spec lead_pipeline_v2 closed (tasks 02–05 recorded with evidence, task 05's rest moved to a todo). Todos: deleted four done or superseded; added re-open cold leads, re-qualify legacy leads, remove job-posting leads once cold, follow-ups after a human first touch. `~/.claude/settings.json` auto-mode rule now allows `mem promote` in shannon.

### Spec: Keep work current in long sessions and on feature branches (completed)

- Commit-time nudges: `internal/checkpoint` (work commits since the last log or completed task, escalating at 3 and 5; unpushed at 3; stale structure doc); `post-commit` hook in every project; `git.InternalEnv` (`MEM_GIT=1`) on all of mem's Git calls keeps the hook silent for them.
- Logs at checkpoints: no session-ending wording in log commands, template or instructions; spec completion asks for a log.
- Compaction catch-up runs `converge.Push` and reports the work-log count.
- Records on a feature branch stay there; `branchNotice` says teammates see them when it merges (committing to development from another branch was considered and rejected).
- Feature branches follow development (`converge/branch.go`): clean rebase onto `origin/<dev>`, lease force-push when only the user's commits are on the remote copy, nudges for conflicts and shared branches; rewritten upstreams caught up with `--fork-point`. Bug found in testing and fixed: a checkout holding a commit dropped by a force-push was only "ahead" and would have pushed it back; `Sync` now catches up whenever the upstream was rewritten.
- Onboard ⏳ CHECK THESE: active specs unchanged for 14 days, unmerged remote branches idle for 14 days; release status shows the age of the oldest unreleased commit.

### Further changes in v0.7.6

- Compaction digest asks the agent to title the session for the work in progress where it can rename sessions.
- Todos have no claims (schema 2; `project.Upgrade` patch `dropTodoClaims` strips old fields and onboard commits them; `mem todo claim` removed; importer brings claimed harness todos in as plain todos).
- Started specs run to the end: `spec start`, `task complete` and the instructions say not to stop for approval between tasks, only for decisions only the user can make.
- `task complete` commits every working-tree change (ignored files excepted) with the task record: subject the task title, body the note, `Mem-Task: <spec>/<task>` trailer, which `checkpoint` treats as a checkpoint (old "Complete task" subjects still count). Instructions tell agents to keep secrets and artifacts ignored and commit unrelated changes separately.
- `spec complete` refuses unless `converge.Sync` leaves the branch settled (`Report.Unsettled`: fetched, not behind its upstream, a feature branch containing development), naming what to resolve.
- Install guide: Windows uses the PowerShell steps also from Git Bash; TLS 1.2, no progress bar and basic parsing for Windows PowerShell 5.1.
- Released v0.7.6 (date tag v2026.10.08.1); release workflow built 7 assets; installed mem updated v0.7.5 → v0.7.6 (schema 2).

## Decisions

- OptChat-style chat memory does not fit mem (it needs to own the agent loop, stores raw transcripts, replaces enforced conventions with inferred ones); log search and a log summary tree were not wanted now; one log file per developer was not adopted.
- Session renaming lives in the compaction digest's instruction; hooks cannot rename after compaction (SessionStart's sessionTitle is ignored on compact; PostCompact output does not reach the model).
- Thresholds: unpushed 3; work-log line on every work commit, suggesting at 3, insisting at 5; specs and branches 14 days.
- Records stay on the branch they are changed on; no commits to development from other branches.
- Todos have no claims; specs keep assignment through `spec start`.
- `task complete` commits all work; `spec complete` requires a settled branch.

## Key Files Affected

- New: `internal/checkpoint/`, `internal/converge/branch.go`, `internal/cli/onboard_staleness.go`, `internal/cli/hook_test.go`, `internal/cli/branch_test.go`, tests alongside.
- Changed: `internal/hooks/hooks.go`, `internal/git/git.go` (`InternalEnv`, `UserEmail`, `CommitAll`), `internal/converge/converge.go`, `internal/cli/{hook,compact,log,spec,task,todo,onboard,onboard_updates,root,sync}.go`, `internal/project/{patches,project}.go`, `internal/work/{todo,changes,log}.go`, `internal/importer/importer.go`, `internal/release/release.go`, `internal/agentsmd/{instructions,install}.md`, README, `docs/design.md`, `docs/status.md`, `.mem/structure.md`.
- Tags: v2026.10.08.1 and v0.7.6.

## Errors and Barriers

- A test command's `cd` into a scratch repository failed (colour codes from `ls` in the path), so `git commit` ran in the mem repository and made two local commits; they were removed with `git reset` before anything was pushed.
- A first `--fork-point` test passed with the feature disabled because Git skips patches already upstream; the replacement test (a dropped commit) fails without it.
- The auto-mode classifier blocked a test that sourced the install guide's cleanup (`rm -rf "$tmp"`) through `sh -c`; it was rerun without that line.
- After the release, Go's HTTP client kept receiving the previous "latest" redirect from GitHub for a few minutes while curl received the new one; the update succeeded on retry.
