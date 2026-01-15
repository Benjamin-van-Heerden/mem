---
title: Config drift detection + mem patch config
status: merge_ready
assigned_to: Benjamin-van-Heerden
issue_id: 53
issue_url: https://github.com/Benjamin-van-Heerden/mem/issues/53
branch: dev-benjamin_van_heerden-config_drift_detection_mem_patch_config
pr_url: https://github.com/Benjamin-van-Heerden/mem/pull/54
created_at: '2026-01-15T12:53:28.081799'
updated_at: '2026-01-15T15:27:19.056356'
completed_at: null
last_synced_at: '2026-01-15T12:54:33.240126'
local_content_hash: 979662159e2d163970ad65e95beccc8a4953d9f8a96e25bd03b1f8ac5ab61caa
remote_content_hash: 979662159e2d163970ad65e95beccc8a4953d9f8a96e25bd03b1f8ac5ab61caa
---
## Overview

`mem` relies on `.mem/config.toml` (and templates used to create it) to reflect current supported functionality. Today, the config template can drift from real behavior (e.g. symlink-related config not represented), and users have no guided way to correct drift safely.

This spec introduces:
1. A template update so new projects start with an accurate `.mem/config.toml`.
2. Drift detection in `mem onboard` (Option A: detect + report only; do not mutate).
3. A new `mem patch config` command to apply a safe, reversible sync/patch to `.mem/config.toml` (generic `mem patch` surface, with `config` as the first patch target).

## Goals

- Align the `.mem/config.toml` template(s) with the actual config schema/behavior that `mem` supports today, including symlink-related config where applicable.
- Make `mem onboard` detect configuration drift and guide the user to resolve it without modifying any files by default.
- Introduce `mem patch config` that can update `.mem/config.toml` to the current canonical shape while preserving user-defined values.
- Keep unknown/unrecognized keys intact by default (to avoid destroying user extensions or future-forward config).
- Add focused automated tests that cover drift detection and patch application to prevent regression.

## Technical Approach

### 1) Define “canonical config” and drift rules
- Establish a canonical representation of the config using the existing config parsing model/struct used by the CLI (the “global config” path mentioned in the todo).
- Canonical representation should include:
  - required sections/keys with defaults,
  - optional sections/keys (when absent, not an error),
  - any known deprecated/renamed keys (migration mapping if needed).

Drift detection compares:
- Parsed user config (including raw TOML where needed to preserve unknown keys/comments as best-effort),
- Against canonical schema + defaults for the current version of `mem`.

Drift categories:
- Missing keys/sections that `mem` expects (add with defaults)
- Deprecated keys that should be removed or migrated (suggest migration)
- Keys with invalid types/values (report clearly; do not auto-fix unless unambiguous)
- Unknown keys (preserve, do not remove unless explicitly requested by a flag)

### 2) Update config template(s)
- Locate the template used by `mem init` (or whichever command creates `.mem/config.toml`) and update it to match the current canonical config.
- Ensure symlink directories/files configuration is present if `mem` currently supports it; if it was removed, remove it from the template and document the replacement.
- Add/adjust inline descriptions so a fresh `.mem/config.toml` is self-explanatory.

### 3) `mem onboard` drift detection (report-only)
In `mem onboard` execution:
- Parse `.mem/config.toml`.
- If drift is detected:
  - Print a concise warning section stating drift exists.
  - Provide a suggested action: run `mem patch config` (and optionally `--dry-run` support in that command).
  - Optionally print a short summary of proposed changes (counts + key names), but avoid printing huge diffs in normal onboard output.

Constraints:
- `mem onboard` must never modify `.mem/config.toml` in this spec.
- The output must be agent-agnostic and consistent with current onboard output size management.

### 4) New `mem patch` command (generic) with `config` target
Introduce a new top-level command group `mem patch` with at least:
- `mem patch config`:
  - Reads `.mem/config.toml`
  - Computes a patch that brings it up to date with canonical config
  - Applies the patch to disk

Recommended flags:
- `--dry-run`: show the patch without applying
- `--format {unified|summary}` (optional): control output verbosity
- `--prune-unknown` (optional, default false): remove unknown keys (off by default)

Implementation notes:
- Preserve existing values for known keys.
- Add missing keys with sensible defaults.
- For deprecated keys: if safe to migrate, migrate; otherwise, leave and warn.
- Preserve unknown keys by default.
- When writing TOML back, aim for stable formatting (consistent key ordering for known sections). If comment preservation is difficult, prefer correctness and stability over perfect comment retention, but do not drop unknown keys.

### 5) Tests
Add dedicated tests covering:
- Drift detection identifies missing canonical keys and produces a non-empty patch suggestion.
- `mem onboard` reports drift and exits successfully without mutating `.mem/config.toml`.
- `mem patch config --dry-run` does not modify files and prints expected output.
- `mem patch config` applies changes:
  - adds missing keys,
  - preserves user-set known values,
  - preserves unknown keys by default,
  - idempotent: running twice results in no further changes.

## Success Criteria

- A newly initialized project’s `.mem/config.toml` includes all currently supported config options (including symlink-related config if supported) and no stale/deprecated options.
- When `.mem/config.toml` is out of date, `mem onboard` clearly reports drift and instructs the user to run `mem patch config`, without changing any files.
- `mem patch config` can bring a drifting `.mem/config.toml` into alignment while preserving user values and retaining unknown keys by default.
- Patch application is idempotent (second run results in “no changes”).
- All new tests pass and cover the core drift + patch behaviors.

## Notes

- Default behavior should be safe and conservative:
  - onboard detects + advises;
  - patch command applies changes with explicit user intent.
- Unknown keys should be preserved unless an explicit prune flag is provided.
- If comment preservation is not feasible with the chosen TOML tooling, prioritize preserving semantics (values and unknown keys) and stable output formatting.
