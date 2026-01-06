---
title: Ensure proper deactivation when spec completes
status: completed
subtasks: []
created_at: '2026-01-06T14:56:42.151343'
updated_at: '2026-01-06T15:01:02.490804'
completed_at: '2026-01-06T15:01:02.490797'
---
When a spec is completed, ensure we properly return to dev branch:

1. After PR creation and status update, check current branch
2. If on the spec's feature branch, switch back to dev
3. Verify the switch was successful
4. Update any active spec markers/state

Current behavior analysis needed:
- Check what spec complete currently does with branches
- Check if there's an 'active spec' marker beyond just branch name
- Ensure deactivate logic is called or inlined

Files to modify:
- src/commands/spec.py: Verify/fix branch switching in complete()
- src/utils/specs.py: Check for any active spec state management

Success criteria:
- After mem spec complete, user is on dev branch
- git status shows clean state (or only unrelated changes)
- No dangling feature branch references

## Completion Notes

Verified that deactivation is already properly implemented. Active spec is determined by git branch (get_active_spec checks current branch). When spec complete switches to dev branch, get_active_spec returns None. No additional state markers need to be cleared.