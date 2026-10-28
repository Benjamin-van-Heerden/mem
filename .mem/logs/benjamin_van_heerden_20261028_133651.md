---
created_at: "2026-10-28T13:36:51+02:00"
user: benjamin_van_heerden
---

# Work Log - mem v0.11.0 released: single-branch mode

## Goal

Release the single-branch mode (6038e42), so the LearnHarness fork can convert to it.

## What Was Accomplished

- CI was green on ubuntu, macOS and Windows for dev at 4b16d68, and for the other session's work log that was pushed in between (987d1ee).
- `mem promote staging` moved `test` from 8d73226 to 4b16d68, with 5 commits.
- `mem promote production` drafted the notes, and they were rewritten as a user-facing summary of single-branch mode. The user confirmed them, and `mem promote production --confirm` tagged `v2026.10.28.1` on `main` at 4b16d68.
- `v0.11.0` was tagged on the same commit with the date tag's notes (`--cleanup=whitespace`) and pushed. The Release workflow (run 40680849719) succeeded with 7 assets.
- `mem update` moved the installed mem from v0.10.0 to v0.11.0 (`mem version`: v0.11.0, 4b16d68, schema 2).
- The LearnHarness handover note (`side-session-notes/local_707259bb…/mem-single-branch-mode.md`) now says the mode shipped in v0.11.0, and to check `mem version` before converting.

## Decisions

- **Minor version bump (v0.11.0):** the release adds a feature and changes nothing for staged projects.

## Key Files Affected

- Tags `v2026.10.28.1` and `v0.11.0` on `main` (4b16d68); `test` and `main` fast-forwarded to 4b16d68.
- Outside the repository: the LearnHarness handover note.

## Notes

- **The `/releases/latest` redirect lagged behind the new release.** For a few minutes after the workflow finished, `mem update` reported "mem v0.10.0 is the latest release", although `gh release view v0.11.0` showed it published and curl already saw the new redirect. A retry a few minutes later updated. Nothing in mem was wrong.
- **`go` and `python3` are asdf shims.** In the scratchpad, which has no `.tool-versions`, they exit with code 126. A Go check program ran fine with `go run` from the repository.
