---
created_at: "2026-10-14T15:16:37+02:00"
user: benjamin_van_heerden
---

# Work Log - mem v0.8.1 released

## Goal

Release the CLAUDE.md cleanup so teammates in migrated repos get it on their next onboard.

## What Was Accomplished

### Release v0.8.1

- CI on `dev` was green on ubuntu, macOS and Windows for `7873e5e` and `e28b8a4`.
- `mem promote staging` moved `origin/test` to `e28b8a4`. `mem promote production --confirm` moved `origin/main` there and tagged `v2026.10.14.1` with the refined notes.
- `v0.8.1` was created as an annotated tag on `origin/main`, with the notes from the date tag (`git tag -a v0.8.1 origin/main --cleanup=whitespace -F <notes>`), and pushed.
- The Release workflow (run 38567307084) succeeded and published 7 assets: darwin, linux and windows binaries for amd64 and arm64, plus `checksums.txt`.
- The installed mem here was updated to v0.8.1 with `mem update`.

The release notes cover the CLAUDE.md cleanup in onboard and import, the onboard instruction for ⚠️ items under UPDATES, import removing every old harness block, and session titles that name only the work.

### Auto-update confirmed for teammates

`autoUpdate` in `internal/cli/update.go` runs at every onboard unless `--offline` or `MEM_NO_UPDATE=1` is set. It installs a newer release and reruns the same onboard with it, so the CLAUDE.md cleanup happens on the first onboard after the update. `buildinfo` reads the module version for `go install …@latest` builds, so those builds update too. `dev` builds never update themselves.

## Decisions

- Patch version (v0.8.1): every change since v0.8.0 is a fix to migration leftovers or an instruction, with no new commands or config.

## Key Files Affected

No source files changed. The release moved `origin/test` and `origin/main` and added the tags `v2026.10.14.1` and `v0.8.1`.

## Notes

The first `mem update` right after the Release workflow finished still reported v0.8.0 as the latest. GitHub's `/releases/latest` redirect lagged the newly published release by a few seconds; a retry installed v0.8.1.
