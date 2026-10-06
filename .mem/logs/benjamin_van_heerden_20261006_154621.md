---
created_at: "2026-10-06T15:46:21+02:00"
user: benjamin_van_heerden
---

# Work Log - studii-tan migration, promotion force-push fix, opt-in release notes, mem v0.7.2 and v0.7.3

## Overarching Goals

Migrate a second Python-harness project before migrating everything: studii-tan, chosen for an earlier harness version and three contributors. Prove it on a testbed, fix what it surfaces in mem, release, then migrate the real repository.

## What Was Accomplished

### studii-tan findings

- `dev` held the harness (moved there 2026-05-25); `main` had 19 commits not on `dev` (Cursor agents and Marco, PR #100 SAQA HEI ingest, 2026-09-15), `test` had one merge commit (504e312, PR #98) not on `dev`, and `fundi-main` is a long-lived Fundi variant branch. `main` and `fundi-main` still carry an older prototype's `.mem/` and `AGENTS.md`.
- Testbed at `~/Documents/Studii/studii-tan-memtest` (bare clone as remote, working clone, a Marco clone), replayed by scratchpad scripts `studii_reset.sh`/`studii_migrate.sh`: merge `origin/main` and `origin/test` into `dev`, import, remove harness, tidy `.gitignore`, commit, push, onboard, promote staging and production, teammate session (todo claim, spec complete, log commit), sync and compaction digest. Deleted afterwards.

### mem fixes (ee93010, ee94ea9; released as v0.7.2, date tag v2026.10.06.1)

- `internal/release/release.go`: divergence lists merge commits (`commits(..., merges bool)`); before, a stage branch whose only extra commit was a merge produced an empty `Diverged` list and `Execute` overwrote it with `--force-with-lease` (reproduced on the testbed, which dropped 504e312). The promotion push is now a plain atomic push, so it only fast-forwards. Test `TestPromotionStopsWhenStageHasOnlyAMergeOutsideDevelopment` fails without the fix.
- `release.Status` gained `StagingOutside`/`ProductionOutside` (commits development lacks); `renderReleases` in `internal/cli/onboard.go` prints a ⚠️ line with the merge command.
- `internal/importer/importer.go`: `userMappings` keys are lowercased and `user()` matches case-insensitively; logs map `username` through it and their files are renamed to the mapped user (`marcobooysep_…` → `marco_booyse_…`). `retiredDocs` leaves out `coding_general.md`/`coding_testing.md` (removed by the harness's own 2026-06-26 update) with a note.
- `internal/cli/import.go`: todo review is its own step, independent of 🔜 WHAT COMES NEXT; open specs are checked after `mem onboard`, because `mem spec complete` commits and pushes.

### Opt-in release notes (abf558e; released as v0.7.3, date tag v2026.10.06.2)

- `project.ReleaseConfig.Notes` (`[release] notes`, default off); `mem init --release-notes`. Without notes, `mem promote production` releases at once with `release.GeneratedMessage` as the tag message; `production_pr` projects open the pull request at once with it. With notes, the draft/refine/`--confirm` flow is unchanged.
- mem's own `.mem/config.toml` sets `notes = true`. Instructions (`internal/agentsmd/instructions.md` Releases), README, `docs/design.md`, `docs/status.md` and the structure doc updated. Tests: `TestProductionReleaseWithoutNotesTagsAGeneratedSummary`; existing notes and pull request tests set `notes = true` through a `releaseConfig` helper.

### todo list shows claimed todos (this session's last commit)

- `mem todo list` always lists open and claimed todos with the claimer; `--all` and `work.OpenTodos` removed. Test `internal/cli/todo_test.go`.

### Real studii-tan migration (pushed to origin/dev: 8081150, 31bb4b5)

- Merged `origin/main` and `origin/test` into `dev`, `mem import agent-core` with v0.7.3: 1 open and 37 archived specs, 25 todos, 111 logs (users benjamin_van_heerden, marco_booyse, jacques_fourie), 14 memories, runnables for README and `src/components`.
- Todo review against the code (two Explore agents): deleted 16 (11 done, 4 folded into others, `update_dokploy_server`); added `fully_paid_subscriptions` and `billing_log_cleanup`; folded padding items into `standardize_mobile_padding_and_spacing_on_public_pages` and the email logic into the task runner todo. 11 todos remain.
- Removed `.agent_core/`, `CLAUDE.md` link, the 7 SAQA report files and `somefile.txt`; `.gitignore` ignores only `.claude/settings.local.json` among Claude files plus `/.mem/local/`; `.claude/settings.json` holds the compaction hook.
- `mem spec complete policy_notification_email` (work was in `src/emails/AdminPolicyNotificationEmail.tsx`).
- A disposable clone of `main` showed mem v0.7.3's onboard stops with "no upgrade path from project schema 0" on the older prototype's config and changes nothing.

## Decisions

- studii-tan: merge `main` and `test` into `dev` before importing; leave `fundi-main` alone with no compatibility work for it (its future is the user's call); the user talks to Marco and Jacques about how they work.
- Release notes are off by default; mem keeps them on.
- Promotion pushes are never forced.
- studii-tan is to be rewritten within about a week, so the remaining todos were kept light.

## Key Files Affected

- `internal/release/release.go`, `release_test.go`; `internal/cli/onboard.go`.
- `internal/importer/importer.go`, `importer_test.go`; `internal/cli/import.go`.
- `internal/project/project.go`; `internal/cli/promote.go`, `promote_test.go`, `promote_pr_test.go`, `init.go`.
- `internal/cli/todo.go`, `todo_test.go`; `internal/work/todo.go`.
- `internal/agentsmd/instructions.md`, `README.md`, `docs/design.md`, `docs/status.md`, `.mem/structure.md`, `.mem/config.toml`.
- Tags: v2026.10.06.1 and v0.7.2, v2026.10.06.2 and v0.7.3.

## Errors and Barriers

- The global Git config has `merge.ff=only`, so merges need `--no-ff`.
- Copying studii-tan's `.env` into the testbed was blocked by the permission classifier; the testbed did not need it.
- Promotion, the release itself and the todo deletions were blocked until the user gave explicit approval naming the action; indirect approval ("I trust your judgement", "this is fine") was not enough.
- Judging a repository idle from `dev` alone was wrong: studii-tan's activity was on `main` and `fundi-main`.
