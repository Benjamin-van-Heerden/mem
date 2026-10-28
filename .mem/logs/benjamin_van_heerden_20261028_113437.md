---
created_at: "2026-10-28T11:34:37+02:00"
user: benjamin_van_heerden
commits: ["6038e42"]
---

# Work Log - Single-branch mode: projects without staging and production

## Goal

Give mem a mode for projects with a development branch (and optionally `[git] base`) but no staging or production branches, for the user's LearnHarness fork. There, a pushed `v*` tag would fire upstream's inherited `release.yml`, and nothing deploys from branches.

## What Was Accomplished

### Design (decided with the user, one question at a time)

The LearnHarness session (local_707259bb) answered the branch-model questions first:

- upstream releases through `v*` tags on `main`;
- the fork keeps `main` as the upstream mirror (`base`) and works only on `dev`;
- the fork must never push tags.

### Config and guards (`internal/project`)

- **`GitConfig.SingleBranch`**: `toml:"single_branch,omitempty"`. `staging` and `production` are now `omitempty`; staged configs marshal unchanged.
- **`Load` → `Config.validate`** refuses `single_branch = true` alongside `staging`/`production` or `[release] production_pr`/`notes`.
- **`GitConfig.Guarded(branch)`** names what mem never commits to or pushes: staging and production, or base in single-branch mode. It is false for `""`, which fixes the detached-HEAD trap where empty stage names would match.
- **`Stage()`** has no stages in this mode.

### Releases and commands

- **`release.SingleBranchError`** is returned by `promote` and `deploy` (before any fetch or push) and by `release.Prepare`.
- **Onboard** skips `renderReleases`, and the agent instruction summarizes "the open specs and open todos".
- **`onboard_staleness`** excludes guarded branches via `g.Guarded`.
- **`init --single-branch`**:
  - `singleBranchConfig` takes the current branch as development. It refuses `--development`, `--staging`, `--production`, `--production-pr` and `--release-notes`, and refuses a detached HEAD.
  - `ensureSingleBranch` makes the empty first commit on it in a repo without commits, and publishes it if the remote lacks it.
  - Output: `Branches: dev (single branch; remote origin)`.

### converge and hooks

- **`converge.Report`**: `Staging`/`Production` replaced by `Guarded` and `SingleBranch`.
  - `Push` never pushes a guarded branch.
  - `feature()` excludes guarded branches.
  - On base, the nudge reads "You are on main, the base branch, which only takes in released code…".
- **`hooks.PrePush`/`PreCommit`** check `Guarded`, with base-branch wording in single-branch mode. `purpose(p)` words the foreign-hook message.

### AGENTS.md variants (`internal/agentsmd`)

- **`instructions.md` is a `text/template`.** `{{rule}}` numbers the Ground Rules (a counter swapped in per render via `Clone().Funcs`), and `{{if .Single}}` selects the passages.
- **`Install`, `ReplaceLegacy` and `Refresh`** take `single bool`. The importer passes `false`; onboard passes `p.Config.Git.SingleBranch`.
- **The staged rendering is byte-identical** to the old `instructions.md` (checked with `cmp` against a copy).
- **The single variant** rewords Rule 3 around base, drops Rule 4 (rules renumber to 5), drops "release status", "or a promotion" and "before a release", and replaces the Releases section with one sentence.

### Tests and verification

- **New tests:**
  - `project`: validation, and the `Guarded` table, including the detached HEAD.
  - `agentsmd`: both variants, and `Refresh` switching modes.
  - `hooks`: base guarded only in single-branch mode.
  - `converge`: base neither rebased nor pushed.
  - `cli/single_branch_test.go`: init, config, no stage branches, AGENTS.md, onboard without RELEASES, promote/deploy refused with no branch moved and no tag pushed, and init flag conflicts and detached HEAD.
- **Runs:** the affected packages and the related cli tests pass, and `go vet ./internal/...` is clean.
- **End to end with `dist/mem-dev`** in a disposable repo: a staged project converted by hand (config edit, delete `test`/`release`) got the single-branch AGENTS.md at onboard, no RELEASES, refused promote and deploy, and no tags.

### Docs

- `docs/design.md`: config example, a "Single-branch projects" paragraph, Hooks, and the managed-block paragraph.
- `docs/status.md`, `.mem/structure.md`.
- Handover note for LearnHarness: `side-session-notes/local_707259bb…/mem-single-branch-mode.md`, with the exact config and conversion steps.

Committed and pushed as 6038e42. Not released.

## Decisions

- **Explicit key `[git] single_branch = true`, never inferred** from missing staging or production keys. The user required the mode to be explicitly enabled at initialization, not a default.
- **`mem promote` and `mem deploy` refuse with one line**, rather than aliasing to sync or tagging dev.
- **Single-branch projects get their own managed block**, rendered from the same source, so the two variants cannot drift.
- **Onboard shows nothing in place of RELEASES.** The Base line already covers where dev stands.
- **The base branch is guarded** in single-branch mode: the user wanted whatever keeps iterating on LearnHarness simple, and a commit on the mirror would break `gh repo sync`'s fast-forward. Staged projects keep base unguarded, unchanged.
- **The guard has no fallback when the base checkout has no `.mem/config.toml`.** The user saw no need: nothing in their workflow commits on `main`.
- **`init --single-branch` takes the current branch as development and creates nothing else.** There is no `--base` flag; base stays a hand edit.
- **No conversion command.** LearnHarness is converted by hand.

## Key Files Affected

- `internal/cli/single_branch_test.go` (85 lines)
- `internal/cli/init.go` (70 lines)
- `internal/agentsmd/agentsmd_test.go` (55 lines)
- `internal/converge/converge.go` (54 lines)
- `internal/project/project_test.go` (50 lines)
- `internal/hooks/hooks.go` (46 lines)
- `internal/project/project.go` (43 lines)
- `internal/agentsmd/agentsmd.go` (39 lines)
- `internal/hooks/hooks_test.go` (30 lines)
- `internal/agentsmd/instructions.md` (27 lines)
- `internal/converge/branch_test.go` (26 lines)
- `internal/cli/onboard.go` (10 lines)
- `internal/release/release.go` (8 lines)
- `docs/design.md` (7 lines)
- `internal/converge/branch.go` (6 lines)
- `internal/cli/deploy.go` (3 lines)
- `internal/cli/promote.go` (3 lines)
- `internal/cli/onboard_staleness.go` (2 lines)
- `internal/cli/onboard_updates.go` (2 lines)
- `internal/importer/importer.go` (2 lines)
- `internal/templates/sync_test.go` (2 lines)
- `docs/status.md` (1 lines)

- `.mem/structure.md`: project, agentsmd, converge, hooks and promotion entries.
- Outside the repository: the handover note in the LearnHarness session's notes folder.

## Notes

- **The first end-to-end attempt silently tested a staged project.** A config edit in a long `&&` chain didn't apply, and `python3` in the scratch directory exited 126. `sed` worked. Run disposable checks step by step.
- **`mem hook pre-commit`/`pre-push` load the config from the working tree.** On an upstream-mirror base branch there is none, so the hooks allow everything there (as outside any mem project). `git push origin main` from `dev` is refused.
- **`gofmt -l` flags `internal/checkpoint/checkpoint.go`**, untouched by this work.
