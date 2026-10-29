---
created_at: "2026-10-29T10:36:13+02:00"
user: benjamin_van_heerden
---

# Work Log - mem v0.12.1 released: onboard leaves staging and production; local stage branches follow releases

## Goal

Release the onboard fix for fresh clones on production (649e4a4) and the change that moves the local staging and production branches along with releases (f87cf77).

## What Was Accomplished

- **CI:** passed on ubuntu, macOS and Windows for dev at fc85e62.
- **Staging:** `mem promote staging` moved `test` to fc85e62.
- **Production:** `mem promote production` drafted notes for `v2026.10.29.2`. They were rewritten as a summary of the two changes, the user confirmed them, and `mem promote production --confirm` tagged `v2026.10.29.2` on `main` at fc85e62.
- **Semver tag:** `v0.12.1` was tagged on the same commit with the date tag's notes (`--cleanup=whitespace`) and pushed. The Release workflow (run 40843047944) succeeded with 7 assets.
- **Local install:** `mem update` moved it from v0.12.0 to v0.12.1 (fc85e62, schema 2) on the first try.
- **Local branches:** this release ran on v0.12.0, which predates the local-branch change, so `main` and `test` were fast-forwarded by hand (`git fetch . refs/remotes/origin/<b>:refs/heads/<b>`). dev, main and test now match origin.
- **ev-reporting:** the session (local_d6fb92b7) was sent a summary of the fix when it reached dev. Projects pick up v0.12.1 at their next onboard.

## Decisions

- **Patch version (v0.12.1):** a bug fix plus a small behaviour change around releases, with no new commands or formats.

## Key Files Affected

- Tags `v2026.10.29.2` and `v0.12.1` on `main` (fc85e62); `test` and `main` fast-forwarded to fc85e62 on the remote and locally.
- Outside this repository: the installed mem at `~/.local/bin/mem` is v0.12.1.

## Notes

- From this release on, `mem promote` fast-forwards this checkout's local `main` and `test` itself.
- nickfols17's ev-reporting checkout picks up v0.12.1 at its next onboard. Until then, or if it is still on production from before, it needs `git switch development`.
