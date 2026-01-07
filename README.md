# mem

A command-line utility for managing project context, specifications, tasks, and work logs in AI-assisted development workflows.

## What is mem?

`mem` solves a critical problem in AI-assisted development: maintaining consistent context across projects without duplicating documentation like `.cursorrules`, `AGENTS.md`, and workflow templates.

The project uses a **file-first, git-native** architecture where all data is stored as markdown files with YAML frontmatter. Git is the persistence layer, ensuring human readability, natural version control, and seamless GitHub integration.

## Core Philosophy

- **Project-agnostic**: Same commands work across Python, Node.js, Rust, or any project
- **Context-first**: Designed for AI agents to quickly understand project state
- **Git-native**: Uses branches to track active specs; git history for audit trail
- **File-based**: No database; markdown files are the source of truth
- **History tracking**: Every interaction is logged and tracked

## Installation

```bash
# Clone the repository
git clone https://github.com/Benjamin-van-Heerden/mem.git
cd mem

# Install with uv
uv sync
```

### Prerequisites

- Python 3.12+
- Git
- GitHub CLI (`gh`)
- `GITHUB_TOKEN` environment variable (with `repo` and `read:user` scopes)

## Quick Start

```bash
# Initialize mem in your project
mem init

# Create a new spec
mem spec new "User Authentication"

# Activate the spec (creates and switches to feature branch)
mem spec activate user_authentication

# Create tasks
mem task new "Set up OAuth" "Configure OAuth providers"
mem subtask new "Register app" --task "Set up OAuth"

# Complete work
mem subtask complete "Register app" --task "Set up OAuth"
mem task complete "Set up OAuth" "OAuth working with Google and GitHub"

# Complete the spec (creates PR, marks as merge_ready)
mem spec complete user_authentication "Implemented OAuth"

# Sync with GitHub
mem sync

# Get context for AI agent
mem onboard
```

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
- Stored in `.mem/logs/{username}_{YYYYMMDD}_{HHMMSS}_session.md`
- Capture accomplishments, blockers, next steps
- Linked to active specs for context tracing

## Commands

### Initialization

```bash
mem init                    # Initialize with GitHub integration
```

Creates `.mem/` directory, validates GitHub auth, sets up branches and labels.

### Specifications

```bash
mem spec new "feature"      # Create new spec
mem spec list               # List specs (optionally filter by status)
mem spec show [slug]        # Show details and tasks
mem spec activate <slug>    # Switch to spec's branch
mem spec deactivate         # Switch back to dev
mem spec complete <slug>    # Create PR, mark merge_ready
mem spec abandon <slug>     # Move to abandoned, close issue
```

### Tasks & Subtasks

```bash
mem task new <title> <desc> [--spec <slug>]  # Create task
mem task complete <title> [--spec <slug>]    # Complete task

mem subtask new <title> --task <task>        # Add subtask to task
mem subtask complete <title> --task <task>   # Complete subtask
```

### Context & Sync

```bash
mem onboard                 # Build context for AI agent
mem sync [--dry-run]        # Bidirectional GitHub sync
```

### Work Logs

```bash
mem log                     # Create or update today's work log
```

## Directory Structure

```
.mem/
├── config.toml              # Project configuration
├── user_mappings.toml       # Git user -> GitHub username
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

## File Format

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

## GitHub Integration

### Two-Way Sync
- **Outbound**: Local specs create GitHub issues with `mem-spec` label
- **Inbound**: GitHub issues sync back to local specs
- **Conflict detection**: Content hashes track changes on both sides

### Labels
- `mem-spec` - Marks issue as a spec
- `mem-status:todo` - Not started
- `mem-status:merge-ready` - Ready for review
- `mem-status:completed` - Done
- `mem-status:abandoned` - Abandoned

## Branch-Based Workflow

Active spec is determined by current git branch, not a stored status field:

```bash
# On dev branch - no active spec
mem spec activate my_feature
# Now on dev-username-my_feature branch - spec is active

mem spec deactivate
# Back on dev branch
```

When you complete a spec:
1. Status updated to `merge_ready`
2. All changes committed and pushed
3. Pull Request created targeting `dev`
4. Switches back to `dev` branch

## Configuration

### `.mem/config.toml`

```toml
[vars]
# Environment variable containing GitHub token
github_token_env = "GITHUB_TOKEN"

[project]
name = "My Project"
description = "What this project is about"

# Generic templates to load from global config (~/.config/mem/templates/)
generic_templates = ["python", "general"]

# Important files to include in onboard context
[[files]]
path = "README.md"
description = "Project overview and setup instructions"

[[files]]
path = "src/main.py"
description = "Application entry point"
```

## Development

```bash
# Install dependencies
uv sync

# Run tests
uv run python -m pytest tests/ -v

# Run a specific test file
uv run python -m pytest tests/test_spec_complete.py -v
```

## Technology Stack

- **Language**: Python 3.12+
- **CLI Framework**: Typer
- **Data Validation**: Pydantic
- **File Format**: Markdown + YAML frontmatter
- **Git Integration**: GitPython
- **GitHub API**: PyGithub
- **Configuration**: TOML

## License

Do what you want
