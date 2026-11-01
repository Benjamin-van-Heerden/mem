---
created_at: "2026-11-01T20:13:19+02:00"
user: benjamin_van_heerden
commits: ["4232273", 494212e]
---

# Work Log - mem precommit and v0.13.0

## Goal

Give every Git project a local check command, make promotion use it, and ship v0.13.0 without releases waiting on CI.

## What Was Accomplished

- `mem precommit` runs `[precommit] commands` from `.mem/config.toml` from the repository root, in order, and stops at the first failure, printing that command's output. With none set, it tells the agent how to add them. It refuses in folder projects.
- `mem promote staging` and `mem deploy` run the same checks and refuse when they fail. `release.yml` no longer waits for CI.
- This repository's checks are `go vet ./...` and `go test ./...`. The full suite takes 29 seconds locally, and about 1 second when the test cache holds.
- `TestSyncReportsANewerReleaseWithoutInstallingIt` clears `MEM_NO_UPDATE`, which otherwise failed the checks for anyone with the variable set.
- v0.13.0 was released through `mem promote production --version v0.13.0 --confirm`, using `dist/mem-dev`. The Release workflow took about 2 minutes and published 7 assets, and installed mem is v0.13.0. CI and Windows both passed on `494212e`.

## Decisions

- **No per-project todos for precommit.** The managed block tells agents to run `mem precommit` before each commit, and its output tells them to add checks when none are set, so each project picks this up at its next onboard. The user asked about a cross-project todo mechanism and chose to ship first; none was built.
- **Checks run at promotion to staging, not at production.** Production moves to exactly what staging previewed.

## Key Files Affected

- `internal/cli/precommit.go`, `internal/cli/precommit_test.go`, `internal/cli/promote.go`, `internal/cli/deploy.go`, `internal/cli/root.go`, `internal/cli/sync_test.go`
- `internal/project/project.go`, `internal/agentsmd/instructions.md`
- `.github/workflows/release.yml`, `.mem/config.toml`, `.mem/structure.md`, `AGENTS.md`, `README.md`, `docs/design.md`, `docs/status.md`

## Notes

- `4232273` reached `dev` with `mem precommit` failing, because `MEM_NO_UPDATE=1` from the shell leaked into a test and a pipe hid the exit code. The check itself stopped it reaching staging.
