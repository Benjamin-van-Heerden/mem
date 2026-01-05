# mem - Project Status Report

## What is mem?

**mem** is a command-line utility for managing project context, specifications, tasks, and work logs in AI-assisted development workflows. It solves the problem of maintaining consistent context across projects without duplicating documentation like `.cursorrules`, `AGENTS.md`, and workflow templates.

The project uses a **file-first, git-native** architecture where all data is stored as markdown files with YAML frontmatter. Git is the persistence layer, ensuring human readability, natural version control, and seamless GitHub integration.

## Core Philosophy

- **Project-agnostic**: Same commands work across Python, Node.js, Rust, or any project
- **Context-first**: Designed for AI agents to quickly understand project state
- **Git-native**: Uses branches to track active specs; git history for audit trail
- **File-based**: No database; markdown files are the source of truth
- **History tracking**: Every interaction is logged and tracked

---

## The Five Primitives

### 1. Specs
High-level feature specifications or project goals.
- Stored in `.mem/specs/{slug}/spec.md`
- Status: `todo`, `merge_ready`, `completed`, `abandoned`
- Linked to GitHub issues
- Contain goals, technical approach, success criteria

### 2. Tasks
Concrete work items linked to specs.
- Stored in `.mem/specs/{spec_slug}/tasks/{order}_{slug}.md`
- Status: `todo` or `completed`
- Can have embedded subtasks
- Contains description and completion notes

### 3. Subtasks
Granular breakdown of tasks.
- Embedded in task frontmatter (not separate files)
- Status: `todo` or `completed`
- Task cannot be completed until all subtasks are done

### 4. Todos
Standalone reminders not tied to specs/tasks.
- Quick notes with optional GitHub issue linking
- Independent lifecycle

### 5. Work Logs
Historical records of work sessions.
- Stored in `.mem/logs/{date}_{slug}.md`
- Capture accomplishments, blockers, next steps
- Linked to active specs for context tracing

---

## Key Commands

### Initialization
```bash
mem init                    # Initialize with GitHub integration
```
Creates `.mem/` directory, validates GitHub auth, sets up branches and labels.

### Specifications
```bash
mem spec new "feature"      # Create new spec
mem spec list               # List specs
mem spec show [slug]        # Show details and tasks
mem spec activate <slug>    # Switch to spec's branch
mem spec deactivate         # Switch back to dev
mem spec complete <slug>    # Create PR, mark merge_ready
mem spec abandon <slug>     # Move to abandoned, close issue
```

### Tasks & Subtasks
```bash
mem task new <title> <desc> # Create task
mem task complete <title>   # Complete task
mem subtask new <title>     # Add subtask to task
mem subtask complete <title># Complete subtask
```

### Context & Sync
```bash
mem onboard                 # Build context for AI agent
mem sync [--dry-run]        # Bidirectional GitHub sync
```

---

## Architecture

### Directory Structure
```
.mem/
├── config.toml              # Project configuration
├── user_mappings.toml       # Git user → GitHub username
├── specs/
│   ├── {slug}/
│   │   ├── spec.md          # Spec with frontmatter
│   │   └── tasks/
│   │       └── 01_{slug}.md # Tasks with embedded subtasks
│   ├── completed/           # Completed specs
│   └── abandoned/           # Abandoned specs
├── todos/                   # Standalone todos
└── logs/                    # Work session logs
```

### Data Models
All files use YAML frontmatter + markdown body:

```yaml
---
title: "Feature Name"
status: "todo"
subtasks:
  - title: "Subtask 1"
    status: "completed"
  - title: "Subtask 2"
    status: "todo"
created_at: "2025-01-05T10:00:00"
updated_at: "2025-01-05T10:00:00"
---
Markdown body content here...
```

### Branch-Based State
- **Active spec** is determined by current git branch
- No explicit "active" status field
- `mem spec activate` creates/switches to the spec's branch
- Git is the source of truth

---

## GitHub Integration

### Two-Way Sync
- **Outbound**: Local specs → GitHub issues with `mem-spec` label
- **Inbound**: GitHub issues → Local specs
- **Conflict detection**: Content hashes track changes on both sides

### Labels
- `mem-spec` - Marks issue as a spec
- `mem-status:todo` - Not started
- `mem-status:merge-ready` - Ready for review
- `mem-status:completed` - Done
- `mem-status:abandoned` - Abandoned

### Prerequisites
```bash
# Required
gh cli        # GitHub CLI for issue deletion
git           # Version control
GITHUB_TOKEN  # Personal access token (repo, read:user scopes)
```

---

## Technology Stack

- **Language**: Python 3.12+
- **CLI Framework**: Typer
- **Data Validation**: Pydantic
- **File Format**: Markdown + YAML frontmatter
- **Git Integration**: GitPython
- **GitHub API**: PyGithub
- **Configuration**: TOML, .env

---

## Testing

Tests use a dedicated `mem-test` repository that gets nuked and recreated at the start of each test session. This keeps the main repo clean.

```bash
# Run tests
uv run pytest tests/ -v

# Tests use pytest-xdist for parallel execution
# File locking ensures only one worker creates the test repo
```

### Test Coverage
- GitHub API integration
- Git branch operations
- Markdown parsing/writing
- Sync conflict detection
- Command execution
- Spec/task/subtask workflows

---

## Recent Changes

1. **Simplified TaskStatus**: Removed `in_progress`; now just `todo` → `completed`
2. **Embedded subtasks**: Subtasks live in task frontmatter, not separate files
3. **Branch-based activation**: Active spec determined by git branch
4. **Test repo isolation**: Tests use dedicated `mem-test` repo
5. **PR filtering in sync**: PRs no longer incorrectly picked up as todos
6. **Prerequisite checks**: `mem init` validates gh CLI, GITHUB_TOKEN, git

---

## Typical Workflow

```bash
# 1. Initialize
mem init

# 2. Create and activate spec
mem spec new "User Authentication"
mem spec activate user_authentication

# 3. Create tasks with subtasks
mem task new "OAuth setup" "Configure OAuth providers"
mem subtask new "Register app" --task "OAuth setup"
mem subtask new "Test flow" --task "OAuth setup"

# 4. Work and track progress
mem subtask complete "Register app" --task "OAuth setup"
mem subtask complete "Test flow" --task "OAuth setup"
mem task complete "OAuth setup" "OAuth working with Google and GitHub"

# 5. Complete spec (creates PR)
mem spec complete user_authentication "Implemented OAuth"

# 6. Sync after PR merge
mem sync  # Moves spec to completed/

# 7. Get context for AI agent
mem onboard
```

---

## What's Next

The project is functional with all core features working:
- Spec/task/subtask management
- GitHub bidirectional sync
- Branch-based activation
- AI context generation via onboard
- Work logging

Potential future enhancements:
- TUI viewer (Textual-based)
- Task dependencies
- Time tracking
- Multi-project views
- Git commit → work log integration

---

## File Inventory

### Source Code
```
main.py                      # CLI entry point
env_settings.py              # Global configuration
src/
├── models.py                # Pydantic frontmatter models
├── commands/
│   ├── init.py              # Initialization
│   ├── spec.py              # Spec management
│   ├── task.py              # Task management
│   ├── subtask.py           # Subtask management
│   ├── sync.py              # GitHub sync
│   ├── onboard.py           # AI context builder
│   └── log.py               # Work logs
├── utils/
│   ├── markdown.py          # YAML frontmatter parsing
│   ├── specs.py             # Spec CRUD
│   ├── tasks.py             # Task/subtask CRUD
│   ├── logs.py              # Log operations
│   ├── todos.py             # Todo operations
│   └── github/
│       ├── client.py        # GitHub authentication
│       ├── api.py           # Issue/PR/label operations
│       ├── repo.py          # Repository operations
│       ├── git_ops.py       # Branch/push operations
│       └── exceptions.py    # Error types
└── templates/
    ├── config.toml          # Config template
    ├── spec.md              # Spec template
    └── task.md              # Task template
```

### Tests
```
tests/
├── conftest.py              # Fixtures, test repo setup
├── test_github_api.py       # GitHub API tests
├── test_github_sync.py      # Sync tests
├── test_init.py             # Init command tests
├── test_logs_username.py    # Log tests
├── test_sandbox.py          # Environment tests
├── test_spec_abandon.py     # Abandon workflow tests
├── test_spec_activate.py    # Activation tests
├── test_spec_subdirectories.py  # File organization tests
├── test_sync_pr_detection.py    # PR merge detection tests
└── test_task_integrity.py   # Task/subtask tests
```

---

## Summary

**mem** is a mature, well-tested project management tool for AI-assisted development. It successfully bridges local markdown files with GitHub issues while maintaining git as the source of truth. The file-first approach ensures human readability and natural version control, while the comprehensive `onboard` command makes it easy for AI agents to understand project context.

All 52 tests pass. The project is ready for production use.
