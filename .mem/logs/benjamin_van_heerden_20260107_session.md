---
date: '2026-01-07'
username: benjamin_van_heerden
spec_slug: onboard_and_workflow_refinements
---
# Work Log - Onboard and workflow refinements spec setup

## Overarching Goals

Set up a new spec to address improvements to `mem onboard` output and fix workflow ordering issues, particularly around status labels being set at the wrong time in the workflow.

## What Was Accomplished

### Created new spec: onboard_and_workflow_refinements

Created spec to address 7 items:
1. Display full work log content in onboard (no truncation)
2. Require at least one work log before spec completion
3. Remove installation/prerequisites from onboard output
4. Add clear file separators in IMPORTANT FILES and CODING GUIDELINES sections
5. Make sync hard (not dry-run) during onboard
6. Fix status label ordering - `mem spec complete` should set 'merge_ready', not `mem merge`
7. Update spec new output with task instructions

### Updated spec new command output

Added task creation instructions to `mem spec new` output:
- Step 5: `mem task new "title" "detailed description with implementation notes if necessary" --spec <slug>`
- Note about not adding tasks in spec body
- Note about simpler syntax when spec is active

```python
typer.echo(f'  5. Add tasks: mem task new "title" "detailed description with implementation notes if necessary" --spec {slug}')
typer.echo("Note: Do not add tasks in the spec body - use 'mem task new' instead.")
typer.echo('Note: If the spec is active, add tasks like: mem task new "title" "detailed description with implementation notes if necessary"')
```

## Key Files Affected

- `.mem/specs/onboard_and_workflow_refinements/spec.md` - new spec created
- `.mem/specs/onboard_and_workflow_refinements/tasks/*.md` - 7 tasks created
- `src/commands/spec.py` - updated spec new command output with task instructions

## What Comes Next

Activate the spec and work through the 7 tasks:
- `mem spec activate onboard_and_workflow_refinements`

Task 7 (update spec new output) is already complete and can be marked done immediately.

Key implementation focus will be on task 6 (status label ordering) - need to move the 'merge_ready' label update from the sync command (called during merge) to the `mem spec complete` command.
