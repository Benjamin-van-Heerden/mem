# mem Implementation Plan

## Development Sequence

We will implement commands in the following order:

1. **init** - Initialize mem in current project
2. **spec** - Specification management
3. **task** - Task management
4. **subtask** - Subtask management
5. **todo** - Todo management
6. **log** - Work log management
7. **onboard** - Context building for AI agents (already implemented, may need updates)
8. **git** - Git/GitHub integration

## Phase 1: Core Primitives

### 1. Init Command (CURRENT)
- ✓ Already implemented in `src/commands/init.py`
- Review and ensure it's working correctly
- Test the migration system

### 2. Spec Commands
- `mem spec new "feature name"` - Create new spec with markdown file
- `mem spec list [--status STATUS]` - List all specs
- `mem spec show <id>` - Show spec details and tasks
- `mem spec update <id> [--status STATUS] [--assigned-to USER] [--branch BRANCH] [--issue-id ID]` - Update spec
- `mem spec complete <id>` - Mark spec complete

**New Spec Fields:**
- `assigned_to` (TEXT) - GitHub username or identifier
- `branch` (TEXT) - Git branch name
- `issue_id` (INTEGER) - GitHub issue number

### 3. Task Commands
- `mem task new "task description" [--spec SPEC_ID]` - Create new task
- `mem task list [--status STATUS] [--spec SPEC_ID]` - List all tasks
- `mem task show <id>` - Show task details
- `mem task update <id> [--status STATUS]` - Update task
- `mem task complete <id>` - Mark task complete

### 4. Subtask Commands
- `mem subtask new "subtask" --parent <id>` - Create subtask under task
- `mem subtask list --parent <id>` - List subtasks for task
- `mem subtask complete <id>` - Mark subtask complete

**Important:** A task cannot be completed until all its subtasks are complete.

### 5. Todo Commands
- `mem todo new "reminder" [--description DESC]` - Create detached todo
- `mem todo list [--status STATUS]` - List todos
- `mem todo complete <id>` - Mark todo complete
- `mem todo delete <id>` - Delete todo

### 6. Log Commands
- `mem log new` - Create new work log entry (opens editor)
- `mem log list [--limit N]` - List recent work logs
- `mem log show <id>` - Show specific work log

## Phase 2: GitHub Integration

### GitHub Issue Primitive

**Concept:**
- GitHub issues relate to specs (1:1 or many:1 relationship)
- A spec can exist independently without a GitHub issue
- If a spec has an `issue_id`, it's linked to a GitHub issue
- Only work on specs that are:
  - Assigned to you (`assigned_to` matches your GitHub username)
  - OR have no associated issue (independent specs)

### Git Commands

#### `mem git sync`
Synchronize with GitHub:
- Fetch issues from GitHub repository
- Create/update specs based on assigned issues
- Update issue status based on spec completion
- Sync issue comments with work logs
- Handle issue assignments

#### `mem git activate <spec_id_or_name>`
Activate a spec for development:
- Verify spec is assigned to you (or has no issue)
- Create or switch to the branch specified in spec
- If no branch exists, derive branch name from spec title
- Update spec status to 'in_progress'
- Display spec context for AI agent

Branch naming convention:
- Format: `feature/{issue_id}-{slug}` or `feature/{spec_id}-{slug}`
- Example: `feature/42-add-user-authentication`

#### `mem git status`
Show git/spec integration status:
- Current branch
- Associated spec (if any)
- Spec progress
- Uncommitted changes
- Sync status with GitHub

### Database Schema Updates

```sql
-- Add to specs table (migration needed)
ALTER TABLE specs ADD COLUMN assigned_to TEXT;
ALTER TABLE specs ADD COLUMN branch TEXT;
ALTER TABLE specs ADD COLUMN issue_id INTEGER;

-- New table for GitHub sync metadata
CREATE TABLE github_sync (
    id INTEGER PRIMARY KEY,
    last_synced_at TIMESTAMP,
    repository TEXT,
    sync_status TEXT -- 'success', 'failed', 'partial'
);
```

### Configuration Updates

Add to `.mem/config.toml`:

```toml
[github]
enabled = false
repository = "owner/repo"  # e.g., "username/project"
token_env = "GITHUB_TOKEN"  # Environment variable containing PAT
default_assignee = "username"

[git]
branch_prefix = "feature"  # or "feat", "task", etc.
auto_create_branch = true
require_assignment = true  # Only allow work on assigned specs
```

### GitHub API Integration

Required functionality:
- Authenticate with GitHub API (using Personal Access Token)
- Fetch issues (filtered by assignee)
- Create issues from specs
- Update issue status
- Add comments to issues
- Handle labels and milestones

Use libraries:
- `PyGithub` or `ghapi` for GitHub API
- `gitpython` for local git operations

### Workflow Example

```bash
# AI agent starts work session
mem onboard

# Sync with GitHub to fetch assigned issues
mem git sync

# See what specs are available
mem spec list --assigned-to me

# Activate a spec (sets up git environment)
mem git activate 5

# Work on tasks...
mem task update 12 --status in_progress

# Complete work
mem task complete 12
mem spec complete 5

# Sync back to GitHub (updates issue status)
mem git sync
```

## Implementation Notes

### Priority
1. Get core commands working first (spec, task, subtask, todo, log)
2. Ensure onboard command works with new data
3. Add GitHub integration as enhancement
4. Git commands build on top of GitHub integration

### Design Principles
- Keep it simple: Don't over-engineer the GitHub integration
- Fail gracefully: GitHub sync should be optional
- Clear error messages: Help users understand assignment requirements
- Idempotent operations: Safe to run `mem git sync` multiple times
- Offline-first: Should work without GitHub connection

### Testing Strategy
- Test each command individually
- Test with and without GitHub integration
- Test permission/assignment checks
- Test branch creation/switching
- Mock GitHub API for tests

## Future Enhancements (Post-MVP)

- TUI viewer with Textual
- Time tracking
- Task dependencies
- Multi-project views
- Export to different formats
- Automated work log generation from git history
- PR integration (link specs to pull requests)
- Slack/Discord notifications
- Template system for common spec types