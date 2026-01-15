"""
Pydantic models for validating and working with mem's TOML configuration.

These models represent the supported schema of `.mem/config.toml` and are intended
to become the single source of truth for:
- validation (types + required fields)
- defaults
- drift detection (unknown keys)
- patching (canonicalization)

TOML parsing is performed with `tomllib`, then validated via:
`MemLocalConfig.model_validate(data)`
"""

from __future__ import annotations

from typing import Any

from pydantic import BaseModel, ConfigDict, Field


class MemVarsConfig(BaseModel):
    """Reserved for future use; not user-configurable via `.mem/config.toml`."""

    model_config = ConfigDict(extra="ignore")


class MemProjectConfig(BaseModel):
    """Project metadata and onboarding context configuration."""

    model_config = ConfigDict(extra="ignore")

    name: str = Field(
        ...,
        description="Project name (displayed in onboard context).",
    )

    description: str = Field(
        ...,
        description="Project description displayed in onboard output to provide context.",
    )

    generic_templates: list[str] = Field(
        default_factory=list,
        description=(
            "Template slugs to load from global_config_dir/templates/ (e.g. ['python', 'general'])."
        ),
    )


class MemImportantFile(BaseModel):
    """An important file to include in `mem onboard` output."""

    model_config = ConfigDict(extra="ignore")

    path: str = Field(
        ...,
        description="Path to the file relative to the project root.",
    )

    description: str | None = Field(
        default=None,
        description="Optional description that explains the file's purpose.",
    )


class MemWorktreeConfig(BaseModel):
    """Worktree behavior configuration."""

    model_config = ConfigDict(extra="ignore")

    symlink_paths: list[str] = Field(
        default_factory=list,
        description=(
            "Paths (relative to repo root) to symlink into created worktrees instead of copying."
        ),
    )


class MemLocalConfig(BaseModel):
    """
    Model for `.mem/config.toml` (the project-local config).

    Notes:
    - Unknown keys are ignored by default so we can *detect* them as drift without
      failing validation. Drift detection should separately compare raw keys
      against this model's schema.
    - Rendering/patch operations should preserve user values for known keys and
      (optionally) remove unknown keys when requested.
    """

    model_config = ConfigDict(extra="ignore")

    project: MemProjectConfig = Field(
        ...,
        description="Project metadata used by mem onboard and other commands.",
    )

    files: list[MemImportantFile] = Field(
        default_factory=list,
        description="Important files included in onboard output.",
    )

    worktree: MemWorktreeConfig = Field(
        default_factory=MemWorktreeConfig,
        description="Worktree behavior settings.",
    )


def get_model_field_names(model: type[BaseModel]) -> set[str]:
    """
    Return the set of supported keys for a model at the current level.

    This is a helper for drift detection logic that compares raw TOML dict keys
    with the model's fields (recursion handled elsewhere).
    """
    return set(model.model_fields.keys())


def is_known_freeform_mapping_type(annotation: Any) -> bool:
    """
    Return True if a field annotation implies an open-ended mapping.

    This is intentionally conservative; mem config currently aims to be explicit
    rather than allowing arbitrary nested dicts. Drift removal should not delete
    content that a model explicitly declares as freeform.

    Currently unused, but kept for future expansion where certain sections may
    allow arbitrary key/value pairs.
    """
    # Placeholder: no explicit freeform config sections yet.
    _ = annotation
    return False
