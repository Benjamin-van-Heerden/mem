# mem

A command-line utility for managing project context, specifications, tasks, and work logs in agentic coding workflows.

## Vision

`mem` solves a critical problem in AI-assisted development: maintaining consistent context, tracking work history, and managing project specifications across multiple projects. Instead of duplicating `.cursorrules`, `AGENTS.md`, and workflow documentation across every project, `mem` provides a centralized, project-agnostic system that works consistently everywhere.

## Core Concepts

### Primitives

`mem` is built around five core primitives:

1. **Specs** - High-level feature specifications or project goals
2. **Tasks** - Concrete work items linked to specs
3. **Subtasks** - Granular breakdown of tasks (tasks with a `parent_id`)
4. **Todos** - Detached reminders not tied to any spec or task
5. **Work Logs** - Historical records of interactions, progress, blockers, and suggestions

### Key Design Principles

- **Project-agnostic**: Works the same way in any project
- **Project-specific provisions**: Extensible for language/framework-specific rules (e.g., Python patterns)
- **Context-first**: Designed for AI agents to quickly understand project state
- **History tracking**: Accurate record of what was done, why, and how
- **File + Database hybrid**: Markdown files for human editing, SQLite for querying
- **Caller vs. mem directory awareness**: Clear distinction between project paths and `mem` internal paths

## Installation

```bash
# TODO: Installation instructions
pip install mem-cli
```

## Quick Start

### Initialize a Project

```bash
# In your project root
mem init
```

This creates a `.mem/` directory containing:
- `mem.db` - SQLite database
- `config.toml` - Project-specific configuration
- `specs/` - Directory for specification markdown files
- `tasks/` - Directory for task/subtask markdown files
- `logs/` - Directory for work log files

### Onboard to a Project

```bash
mem onboard
```

The `onboard` command constructs initial context by:
- Reading project-specific configuration
- Identifying important files
- Loading active specs and their tasks
- Providing AI agents with everything they need to start working

## Commands

### Initialization

```bash
mem init                    # Initialize mem in current directory
mem config edit             # Edit project-specific configuration
```

### Specs

```bash
mem spec new "feature name"           # Create new spec with markdown file
mem spec list                         # List all specs
mem spec show <spec-id>               # Show spec details and tasks
mem spec update <spec-id>             # Update spec status/metadata
mem spec complete <spec-id>           # Mark spec complete (prompts for completion context)
```

When you create a spec:
1. Entry created in SQLite database
2. Markdown file created at `.mem/specs/{date}_feature_name.md`
3. Prompt displayed: "Edit the spec file with goals and create tasks with `mem task new`"

### Tasks

```bash
mem task new "task description"            # Create task (prompts for spec link)
mem task new "task" --spec <spec-id>       # Create task linked to spec
mem task list                              # List all tasks
mem task show <task-id>                    # Show task details
mem task update <task-id>                  # Update task status/metadata
mem task complete <task-id>                # Mark complete (prompts for implementation details)
```

Tasks follow the same pattern as specs:
1. Database entry created
2. Markdown file created at `.mem/tasks/{date}_task_description.md`
3. AI prompted to fill in implementation details

### Subtasks

```bash
mem subtask new "subtask" --parent <task-id>    # Create subtask under a task
mem subtask list --parent <task-id>             # List subtasks for a task
mem subtask complete <subtask-id>               # Mark subtask complete
```

**Important**: A task cannot be marked complete until all its subtasks are complete.

### Todos

```bash
mem todo new "reminder"             # Create detached todo
mem todo list                       # List all todos
mem todo complete <todo-id>         # Mark todo complete
mem todo delete <todo-id>           # Delete todo
```

### Work Logs

```bash
mem log new                         # Create new work log entry
mem log today                       # Show today's work log
mem log list                        # List recent work logs
mem log show <log-id>               # Show specific work log
```

Work logs capture:
- What was worked on
- What was accomplished
- Errors or blockers encountered
- Suggestions for next steps

### Context Building

```bash
mem onboard                         # Build initial context for AI agent
mem context refresh                 # Refresh context cache
mem context show                    # Display current context
```

## Configuration

### Project Configuration (`.mem/config.toml`)

```toml
[project]
name = "My Project"
description = "What this project is about"
type = "python"  # or "typescript", "rust", etc.

[context]
# Files to always include in onboard
important_files = [
    "README.md",
    "pyproject.toml",
    "src/main.py"
]

# Directories to scan for context
scan_directories = ["src/", "tests/"]

[rules.python]
# Python-specific coding patterns and rules
style = "Follow PEP 8"
type_hints = "Always use type hints"
testing = "Use pytest for testing"

[rules.typescript]
# TypeScript-specific rules (if applicable)
# ...
```

## Database Schema

```sql
-- Specs
CREATE TABLE specs (
    id INTEGER PRIMARY KEY,
    title TEXT NOT NULL,
    file_path TEXT NOT NULL,
    status TEXT DEFAULT 'active',  -- active, completed, archived
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    completed_at TIMESTAMP
);

-- Tasks (and subtasks)
CREATE TABLE tasks (
    id INTEGER PRIMARY KEY,
    spec_id INTEGER,
    parent_id INTEGER,  -- NULL for tasks, set for subtasks
    title TEXT NOT NULL,
    file_path TEXT NOT NULL,
    status TEXT DEFAULT 'todo',  -- todo, in_progress, blocked, completed
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    completed_at TIMESTAMP,
    FOREIGN KEY (spec_id) REFERENCES specs(id),
    FOREIGN KEY (parent_id) REFERENCES tasks(id)
);

-- Todos
CREATE TABLE todos (
    id INTEGER PRIMARY KEY,
    title TEXT NOT NULL,
    description TEXT,
    status TEXT DEFAULT 'open',  -- open, completed
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    completed_at TIMESTAMP
);

-- Work Logs
CREATE TABLE work_logs (
    id INTEGER PRIMARY KEY,
    file_path TEXT NOT NULL,
    summary TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

## Path Handling

`mem` distinguishes between two types of paths:

- **Caller paths**: Relative to the directory where `mem` is invoked (the project root)
- **Mem paths**: Relative to `.mem/` directory

Example:
- Caller path: `src/main.py` (in project)
- Mem path: `specs/2024-01-15_new_feature.md` (in `.mem/` directory)

Configuration files and database queries handle this distinction automatically.

## Future Enhancements

### TUI Viewer

A Textual-based terminal UI for:
- Viewing tasks by status
- Filtering specs and tasks
- Interactive task management
- Work log browsing

```bash
mem tui                     # Launch interactive viewer
```

### Advanced Features

- Task dependencies and blocking relationships
- Time tracking and effort estimates
- Export to other formats (JSON, Markdown reports)
- Git integration for automatic work log generation
- Multi-project views and cross-project task management

## Development

```bash
# Clone repository
git clone https://github.com/yourusername/mem.git
cd mem

# Install in development mode
pip install -e .

# Run tests
pytest
```

## Contributing

Contributions welcome! Please open an issue to discuss major changes.

## License

MIT