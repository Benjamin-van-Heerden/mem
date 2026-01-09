---
created_at: '2026-01-09T16:25:26.722836'
username: benjamin_van_heerden
spec_slug: config_important_infos_field
---
# Work Log - Add important_infos config field

## Overarching Goals

Add an optional `important_infos` field to the project config that displays custom information at the bottom of `mem onboard` output, allowing users to include project-specific notes and reminders for AI agents.

## What Was Accomplished

### Added important_infos field to config template

Updated `src/templates/config.toml` to document the new optional field with a commented example:

```toml
# Important information to display at the bottom of onboard output
# Use this for project-specific notes, reminders, or context that should
# always be visible to AI agents (e.g., deployment notes, key contacts, etc.)
# important_infos = """
# - Remember to run tests before committing
# - API keys are stored in .env.local
# """
```

### Updated onboard command to display the section

Added code to `src/commands/onboard.py` to render an "IMPORTANT INFORMATION" section at the bottom of onboard output when the field is set:

```python
# Important infos section (if configured)
important_infos = project.get("important_infos", "").strip()
if important_infos:
    output.append("")
    output.append("-" * 70)
    output.append("IMPORTANT INFORMATION")
    output.append("-" * 70)
    output.append("")
    output.append(important_infos)
```

### Verified backwards compatibility

Tested that onboard output is unchanged when the field is not set.

## Key Files Affected

- `src/templates/config.toml` - Added documented `important_infos` field with example
- `src/commands/onboard.py` - Added rendering of IMPORTANT INFORMATION section at bottom of output

## What Comes Next

Spec is complete and ready for PR creation via `mem spec complete`.
