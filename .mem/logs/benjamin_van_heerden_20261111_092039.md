---
created_at: "2026-11-11T09:20:39+02:00"
user: benjamin_van_heerden
commits: [783ff10, c98c2bc]
---

# Work Log - mem v0.14.0 released: Codex hooks; declined approvals are asked again

## Goal

Settle what happens when a user declines mem's Codex hooks, correct the docs to match, and release the Codex hooks as v0.14.0.

## What Was Accomplished

### Declining Codex hooks

- Tested in an interactive Codex 0.164 on a scratch repository with only mem's two hooks. The user declined them. `~/.codex/config.toml` gained only the folder's `[projects]` trust entry and no `[hooks.state]` record.
- At the next start Codex showed "Hooks need review — 2 hooks are new or changed" with three choices: review hooks, trust all and continue, or continue without trusting. So a declined hook is asked about again at every start until it is approved.
- Codex's hook state records also have an `enabled` field. Turning a hook off inside the review screen was not tested.
- `783ff10` corrected the onboard line, the `codex.go` comment, `docs/design.md` and `docs/status.md`. They had said Codex skips unapproved hooks silently, which holds only for `codex exec`. `c98c2bc` documented that declining stores nothing and Codex asks again at the next start.
- The scratch repository's trust records were removed from the user's `~/.codex/config.toml`, with backups in the session scratchpad.

### Release

- `mem promote staging` ran the checks and moved `test` to `c98c2bc`. CI and Windows on `test` both passed.
- `mem promote production --version v0.14.0 --confirm` used notes that described the Codex hooks. The Release workflow published 7 assets, and `mem update` installed v0.14.0.

## Decisions

- **No onboard check for unapproved Codex hooks.** The user pointed out that Codex already asks, and the test showed that declining is asked about again at every start. An onboard check could only test whether a record exists, because Codex's `trusted_hash` could not be reproduced from the entry (320 candidate serialisations tried), so it would add little.
- **mem never writes Codex approval records**, since that would skip the user's review.

## Key Files Affected

- `internal/cli/onboard_updates.go`, `internal/agenthooks/codex.go`: wording only.
- `docs/design.md`, `docs/status.md`, `README.md`

## Notes

- Agent shell commands under Codex get `CODEX_THREAD_ID`, `CODEX_SESSION_ID`, `CODEX_VERSION` and `CODEX_SANDBOX` (`seatbelt`, with `CODEX_SANDBOX_NETWORK_DISABLED=1` in read-only and workspace-write sessions), but hooks get none of these.
- `-c check_for_update_on_startup=false` keeps Codex's update prompt out of scripted interactive runs.
