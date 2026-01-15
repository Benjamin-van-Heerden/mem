"""
Main config loader and drift detection for mem.

This module provides:
- Loading `.mem/config.toml` into a validated Pydantic model (`MemLocalConfig`)
- Non-fatal validation reporting (so onboard can warn without breaking)
- Unknown-key drift detection (extra keys that are not part of the supported schema)

TOML parsing is performed with `tomllib`:
    with open(path, "rb") as f:
        raw = tomllib.load(f)

Then validation:
    cfg = MemLocalConfig.model_validate(raw)

Unknown keys are detected by diffing the raw TOML dict keys against the model's
schema (recursively), without requiring `extra="forbid"`.
"""

from __future__ import annotations

import tomllib
from dataclasses import dataclass
from pathlib import Path
from typing import Any, get_args, get_origin

from pydantic import BaseModel, ValidationError

from src.config.models import MemLocalConfig


@dataclass(frozen=True)
class LocalConfigLoadResult:
    """
    Result of loading the local config.

    - `raw`: dict parsed directly from TOML (may include unknown keys)
    - `config`: validated config model when validation succeeds, else None
    - `validation_error`: Pydantic ValidationError when validation fails, else None
    """

    raw: dict[str, Any]
    config: MemLocalConfig | None
    validation_error: ValidationError | None


def read_local_toml(path: Path) -> dict[str, Any]:
    """Read TOML from `path` and return a dict; return {} on missing/unreadable."""
    try:
        if not path.exists():
            return {}
        with open(path, "rb") as f:
            data = tomllib.load(f)
        if isinstance(data, dict):
            # tomllib returns dict[str, Any] at top-level for TOML documents
            return data
        return {}
    except Exception:
        return {}


def load_and_validate_local_config(path: Path) -> LocalConfigLoadResult:
    """
    Load and validate `.mem/config.toml`.

    This function is intentionally non-throwing for file/parse errors. For
    validation errors, it returns the `ValidationError` so callers can decide
    how to display warnings/errors.
    """
    raw = read_local_toml(path)
    if not raw:
        return LocalConfigLoadResult(raw=raw, config=None, validation_error=None)

    try:
        cfg = MemLocalConfig.model_validate(raw)
        return LocalConfigLoadResult(raw=raw, config=cfg, validation_error=None)
    except ValidationError as e:
        return LocalConfigLoadResult(raw=raw, config=None, validation_error=e)


def _unwrap_optional(annotation: Any) -> Any:
    """
    If `annotation` is Optional[T] / T|None, return T; otherwise return unchanged.
    """
    origin = get_origin(annotation)
    if origin is None:
        return annotation

    # Optional[T] becomes Union[T, NoneType]
    if origin is list or origin is dict:
        return annotation

    if origin is type | None:  # pragma: no cover (defensive; not expected in practice)
        return annotation

    if origin is None:
        return annotation

    if (
        origin is getattr(__import__("typing"), "Union", None)
        or str(origin) == "typing.Union"
    ):
        args = [a for a in get_args(annotation) if a is not type(None)]  # noqa: E721
        if len(args) == 1:
            return args[0]
    return annotation


def _is_basemodel_type(t: Any) -> bool:
    try:
        return isinstance(t, type) and issubclass(t, BaseModel)
    except Exception:
        return False


def _list_item_model_type(annotation: Any) -> type[BaseModel] | None:
    """
    If annotation is list[SomeBaseModel] (possibly optional), return SomeBaseModel.
    Otherwise return None.
    """
    annotation = _unwrap_optional(annotation)
    origin = get_origin(annotation)
    if origin is not list:
        return None
    args = get_args(annotation)
    if len(args) != 1:
        return None
    item_t = _unwrap_optional(args[0])
    if _is_basemodel_type(item_t):
        return item_t
    return None


def _nested_model_type(annotation: Any) -> type[BaseModel] | None:
    """
    If annotation is SomeBaseModel (possibly optional), return it, else None.
    """
    annotation = _unwrap_optional(annotation)
    if _is_basemodel_type(annotation):
        return annotation
    return None


def find_unknown_key_paths(
    raw: Any, model: type[BaseModel], prefix: str = ""
) -> list[str]:
    """
    Recursively find unknown keys in `raw` compared to `model`.

    Returns a list of "paths" describing unknown keys, e.g.:
      - "project.weird_key"
      - "worktree.symlink_magic"
      - "files[0].extra_field"

    Notes:
    - Only dict keys are considered for unknown-key drift.
    - For list fields whose item type is another BaseModel, each list element
      that is a dict is recursed into.
    """
    if not isinstance(raw, dict):
        return []

    allowed = set(model.model_fields.keys())
    unknown_here = [f"{prefix}{k}" for k in raw.keys() if k not in allowed]

    unknown_nested: list[str] = []

    # Recurse into known keys that correspond to nested models or list[model]
    for field_name, field_info in model.model_fields.items():
        if field_name not in raw:
            continue

        value = raw.get(field_name)
        annotation = field_info.annotation

        nested_model = _nested_model_type(annotation)
        if nested_model is not None:
            next_prefix = f"{prefix}{field_name}."
            unknown_nested.extend(
                find_unknown_key_paths(value, nested_model, next_prefix)
            )
            continue

        list_item_model = _list_item_model_type(annotation)
        if list_item_model is not None:
            if isinstance(value, list):
                for idx, item in enumerate(value):
                    if isinstance(item, dict):
                        next_prefix = f"{prefix}{field_name}[{idx}]."
                        unknown_nested.extend(
                            find_unknown_key_paths(item, list_item_model, next_prefix)
                        )
            continue

    return unknown_here + unknown_nested


def has_unknown_key_drift(raw: dict[str, Any]) -> bool:
    """Return True if `raw` contains keys not represented by `MemLocalConfig`."""
    return len(find_unknown_key_paths(raw, MemLocalConfig)) > 0


def summarize_validation_error(err: ValidationError, max_lines: int = 6) -> str:
    """
    Produce a concise, human-readable summary for a ValidationError.

    Intended for short warnings in commands like `mem onboard`.
    """
    lines: list[str] = []
    for e in err.errors():
        loc = ".".join(str(p) for p in e.get("loc", []))
        msg = e.get("msg", "Invalid value")
        if loc:
            lines.append(f"{loc}: {msg}")
        else:
            lines.append(msg)

        if len(lines) >= max_lines:
            break

    if len(err.errors()) > max_lines:
        lines.append(f"... ({len(err.errors()) - max_lines} more)")
    return "\n".join(lines)
