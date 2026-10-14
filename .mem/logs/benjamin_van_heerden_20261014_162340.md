---
created_at: "2026-10-14T16:23:40+02:00"
user: benjamin_van_heerden
commits: [d231e2e]
---

# Work Log - mem v0.8.2 (task commits with removed paths, ignored logs); ev-ledger, ev-reporting and wingman on mem; band and replay label fixes

## Goal

Move more of Benjamin's repositories onto mem (ev-ledger, ev-reporting, and wingman, which was never on the harness), fixing what the migrations exposed in mem and in the mods, and release mem v0.8.2.

## What Was Accomplished

### mem v0.8.2

- `internal/git/git.go`: `Commit` stages only paths `present` on disk or in the index (`os.Lstat`, then `git ls-files -- <path>`). `git add --all -- <paths>` rejected a path already taken out of the index by `git rm` or `git mv` ("pathspec did not match any files"), so `mem task complete` failed for tasks that deleted or moved files; `git commit -- <paths>` records those removals itself, and a path matching nothing still fails there. Test `TestCommitTakesPathsAlreadyRemovedOrMovedInTheIndex` (git rm, git mv, an unstaged deletion, an unnamed change left alone, a typo rejected).
- `internal/cli/log.go`: `log commit` calls `ignoreRule` (`git check-ignore --verbose`) before finishing the log and stops with the rule, e.g. `.gitignore:14:logs/`. An ignored log never shows in `git status`, which read as "already committed" and reported a successful commit and push. Test `TestLogCommitRefusesALogThatGitIgnores`. `.mem/structure.md` mentions it.
- Both tests fail without their fix. CI green on ubuntu, macOS and Windows for `d231e2e`. `mem promote staging`, `mem promote production --confirm` (date tag `v2026.10.14.2`), `v0.8.2` tagged on `origin/main` with the confirmed notes (`--cleanup=whitespace`); Release run 38579824128 built 7 assets; `mem update` moved the installed mem from v0.8.1 to v0.8.2.
- Earlier today, outside this log's commits: `nameSession()` says "Name only the work, not the project" (an agent had copied a project prefix from an old log title), and the importer drops every harness block; both went out in v0.8.1 from another session.

### Migrations

- **ev-ledger** (Zero-Carbon-Charge, `development`): imported; the second harness block removed by hand; todos checked against code and logs (6 deleted as done; dead session pruning deleted after the ev-server session confirmed the 2026-07-22 fix; sensitive logging kept and reworded with ev-server's files and lines); last 11 %-style log calls converted to f-strings; pytest added as a uv dev dependency with `pythonpath = ["."]` and `--import-mode=importlib`; GitHub issues #1–#8 closed; todos for the SSH-timing-out test deploys and the DEPLOY-SMOKE-TEST row; work log written there.
- **ev-reporting** (Zero-Carbon-Charge, `development`): imported; `enabledPlugins` moved to `settings.local.json` before sharing `settings.json`; `.dockerignore` excludes `.mem`; N3 prefill spec completed (Tendani's `2dccb8e`); todos for the first real assisted start (2026-10-15), terms wording, frequent issues content, the fleet invoice backfill and RFID card `2D21EEF3`; structure doc written; work log written there.
- **wingman** (AnssumAdmin, Python/Flask, no harness): `dev` created from Marion's `codex/wingman-improvements-oct-2026` tip, `mem init` with `dev`/`test`/`main`, `CLAUDE.md` removed, `full_deploy` memory and `scripts/deploy_heroku.sh`; spec "Wingman housekeeping: uv tooling, CI and docs" completed (uv and `uv.lock`, a Tests workflow with an image job checking `/health`, old documents in `docs/history/`, new README and structure doc); its own work log covers the details.

### Mods (`~/.claude/mods`, repo `Benjamin-van-Heerden/claude-mods`)

- mission-control 0.3.1 (`186630e`): desktop sessions save the app's session list, with their own row, to `~/.claude/mission-control-sessions.json`; terminal sessions read it, so the terminal band has titles and links and leaves out archived sessions. `pillLabel` never repeats the project ("shannon · shannon").
- Replay Theater local.23 (`a477f72`): turn labels for background task notices use the quoted name in the `<summary>`; messages starting with `@"/path/file"` show the typed text, or the file name when nothing was typed; `@` inside a word (an email) is not a mention.

## Decisions

- Ghostty tab switching from the band was built and dropped before committing: Benjamin uses the desktop app almost always.
- Wingman's `main` stays at the May handover until the team agrees; the first full deploy fast-forwards it. Heroku deploys stay manual until Benjamin has credentials, then move to GitHub Actions (todo there).
- The ZCC repositories still on the harness wait until Benjamin hears how Tendani's onboarding went.

## Key Files Affected

- `internal/git/git.go`, `internal/git/git_test.go`: `Commit` and `present`, and the test.
- `internal/cli/log.go`, `internal/cli/log_test.go`: `ignoreRule`, the refusal, and the test.
- `.mem/structure.md`: the `log commit` line.
- Outside this repository: the three projects above (`~/Documents/zcc/ev-ledger`, `~/Documents/zcc/ev-reporting`, `~/Documents/Anssum/wingman`) and `~/.claude/mods/{mission-control,replay-theater}`.

## Notes

- In wingman, a bare `logs/` in `.gitignore` matched `.mem/logs/`; it is now `/logs/`. Other projects with such a rule now get the refusal instead of a silent non-commit.
- Auto mode blocked the first `git rm -r .agent_core` and `gh issue close` in ev-ledger as writes outside the session's folder; Benjamin added allow rules for both.
