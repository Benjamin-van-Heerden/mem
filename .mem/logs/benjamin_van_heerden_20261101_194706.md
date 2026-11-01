---
created_at: "2026-11-01T19:47:06+02:00"
user: benjamin_van_heerden
commits: [edcb7d7, 63c9d5d]
---

# Work Log - Faster tests, CI and releases

## Goal

Stop releases waiting on slow CI. The user does not accept long release times, and fast iteration is why mem is written in Go.

## What Was Accomplished

### Where the time went

- `go build` takes 1 to 3 seconds. The wait was `go test` on Windows: `internal/cli` starts 6205 git processes, one after another, and took 7 minutes there against 21 seconds on Ubuntu.
- CI on `edcb7d7` also showed a real Windows failure: `TestInitFolderSetsUpAPlainFolderAndOnboardBuildsItsContext` expected a shell runnable to execute there. The test now leaves that out on Windows.
- `9f1c08c` failed on Ubuntu and macOS because `internal/templates/sync_test.go` still passed a bool to `agentsmd.Install`. Fixed in `edcb7d7`.

### Tests

- `internal/gittest.Main` gives each test binary one HOME with Git maintenance off, replacing per-test `t.Setenv("HOME", ...)`, and every package that drives Git has a `main_test.go` calling it.
- 93 tests in cli, converge, release, templates, checkpoint and git call `t.Parallel()`. Tests that use `t.Setenv` stay serial.
- Locally `internal/cli` went from 88s to 23s, and every test passes.

### CI and releases

- `ci.yml` runs Ubuntu (vet, test, build) first, then macOS. `windows.yml` runs Windows beside it and blocks nothing. `release.yml` waits for CI on the tagged commit.
- `[release] version_tags = true` and `mem promote production --version vX.Y.Z` name a release after its version with no date tag. `mem deploy` refuses there, and the setting cannot be combined with `production_pr`. This repository sets it.

### Measured on CI

- `63c9d5d`: Test step 14s on Ubuntu, 62s on macOS, 7m50 on Windows, all passing. `internal/cli` on Windows halved to 220s.
- `f59e010`: CI took 2m25 with Windows out of it.

## Decisions

- **Windows does not block releases.** It stays slow after parallelising, so it runs beside CI and reports on its own. The user asked for the slow releases to be fixed and left the choice open.
- **Parallel tests keep real Git.** Mocking Git would remove the coverage these tests exist for.

## Key Files Affected

- `internal/gittest/gittest.go`, `internal/*/main_test.go` and the test files of cli, converge, release, templates, checkpoint and git.
- `.github/workflows/ci.yml`, `release.yml`, `windows.yml`.
- `internal/project/project.go`, `internal/release/release.go`, `internal/cli/promote.go`, `internal/cli/deploy.go`, `internal/cli/promote_test.go`: version tags.
- `.mem/config.toml`, `.mem/structure.md`, `AGENTS.md`, `README.md`, `docs/design.md`, `docs/status.md`.

## Notes

- Setting `core.fsync = none` and an empty `init.templateDir` for test Git made no difference on macOS, so both were dropped.
- About four minutes of the Windows Test step sits outside the test binaries' reported run times: `internal/cli` reported 220s of a 470s step. The cause was not found.
- Narrowing CI to `test` earlier today moved the whole CI wait from background pushes to `dev` onto release time.
