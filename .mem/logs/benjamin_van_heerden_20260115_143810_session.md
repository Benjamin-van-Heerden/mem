---
created_at: '2026-01-15T14:38:10.527568'
username: benjamin_van_heerden
spec_slug: config_drift_detection_mem_patch_config
---
# Work Log - Pydantic config model + onboard drift warning

## Overarching Goals

Harden config handling so `.mem/config.toml` reflects supported behavior and drift is detectable without mutating user files. Establish a single source of truth for config shape (Pydantic model) and ensure `mem onboard` warns clearly when config is out of date or contains unsupported keys.

## What Was Accomplished

### Config model + loader (single source of truth)
- Added a proper Pydantic schema for the project-local config and a shared loader/validation entry point.
- Implemented unknown-key drift detection by comparing raw TOML dict keys (recursively) against the model schema, without using strict “forbid extras”.

### Onboard drift detection and config access refactor
- Refactored `mem onboard` to use the validated config model for all config access (project name/description, important files, etc.) instead of dict-style `.get(...)`.
- Added a concise, report-only warning when drift is detected:
  - unknown keys exist, and/or
  - validation fails for known keys (prints a bounded validation summary).

### Remove local overrides / vars section and deprecate global config.toml merge behavior
- Removed `[vars]` from `.mem/config.toml` and from the default config template so users can’t override core defaults (e.g., GitHub token env var name).
- Stopped merging any global `config.toml` into onboard config resolution.
- Ensured the global config directory remains supported but non-overridable (fixed default `~/.config/mem`).

### Additional call site cleanup
- Updated worktree symlink creation to use the shared config loader/model instead of parsing TOML directly.
- Updated docs collection naming to prefer the validated config model.

## Key Files Affected

- `src/config/models.py`: New Pydantic models for `.mem/config.toml` (no user-configurable `[vars]`).
- `src/config/main_config.py`: Shared TOML loader + validation + unknown-key drift detection utilities.
- `src/commands/onboard.py`: Drift warning + refactor to use validated config model; removed global config.toml merge and local override behavior.
- `src/commands/spec.py`: Uses shared config loader/model for `worktree.symlink_paths`.
- `src/utils/docs.py`: Uses shared config loader/model for project name (collection naming).
- `.mem/config.toml`: Removed `[vars]` section.
- `src/templates/config.toml`: Removed `[vars]` section.

## Errors and Barriers

- `mem onboard` triggers an internal sync step that fails when there are uncommitted changes (rebase cannot proceed). This is unrelated to config drift detection, but it makes ad-hoc manual verification noisier because sync errors are printed before the main output.

## What Comes Next

- Implement `mem patch config`:
  - remove unknown keys by default (including any reintroduced `[vars]`),
  - add missing supported keys with defaults,
  - preserve user-set values for known keys,
  - ensure idempotency.
- Add tests covering:
  - onboard drift warning without mutation,
  - patch dry-run does not modify files,
  - patch applies + is idempotent.
- Finish removing any remaining references to global config.toml merging patterns (if any exist outside onboard), while keeping the global templates dir fixed at `~/.config/mem`.
