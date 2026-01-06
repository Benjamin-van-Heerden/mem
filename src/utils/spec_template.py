"""
Spec template utilities.

Handles loading spec templates from global config and generating
GitHub issue templates from the markdown spec template.

The spec.md template is the single source of truth - it's used for:
1. Local spec creation (body content)
2. GitHub issue template (copied with frontmatter added)
"""

from pathlib import Path

from env_settings import ENV_SETTINGS

# Default spec template content (used if no global template exists)
DEFAULT_SPEC_TEMPLATE = """\
## Overview

{Describe the feature or change}

## Goals

- {Goal 1}
- {Goal 2}

## Technical Approach

{How to implement this}

## Success Criteria

- {Criterion 1}
- {Criterion 2}

## Notes

{Additional context}
"""


def get_global_spec_template_path() -> Path:
    """Get path to global spec template."""
    return ENV_SETTINGS.global_config_dir / "templates" / "spec.md"


def get_local_spec_template_path() -> Path:
    """Get path to local (bundled) spec template."""
    return Path(__file__).parent.parent / "templates" / "spec.md"


def ensure_global_config_exists() -> None:
    """Ensure ~/.config/mem/ and default templates exist."""
    global_dir = ENV_SETTINGS.global_config_dir
    templates_dir = global_dir / "templates"

    # Create directories
    templates_dir.mkdir(parents=True, exist_ok=True)

    # Create default spec template if it doesn't exist
    spec_template_path = templates_dir / "spec.md"
    if not spec_template_path.exists():
        spec_template_path.write_text(DEFAULT_SPEC_TEMPLATE)


def load_spec_template() -> str:
    """Load spec template, preferring global over local."""
    global_path = get_global_spec_template_path()
    if global_path.exists():
        return global_path.read_text()
    return get_local_spec_template_path().read_text()


def generate_github_issue_template(template: str | None = None) -> str:
    """Generate GitHub issue template markdown from spec template.

    Adds YAML frontmatter for GitHub issue template configuration.
    If no template provided, loads from global or local spec.md.
    """
    if template is None:
        template = load_spec_template()

    frontmatter = """\
---
name: mem Specification
about: Create a new specification
title: '[Spec]: '
labels: mem-spec
---

"""
    return frontmatter + template
