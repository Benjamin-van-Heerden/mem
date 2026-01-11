# Working with mem

This project uses **mem** for context management and version control in AI-assisted development.

## First Action

At the start of every session, run:

```bash
mem onboard
```

This gives you everything you need: project info, coding guidelines, active specs, tasks, and recent work logs.

## Core Workflow

### 1. Work on a Spec

Specs are feature specifications linked to GitHub issues. The active spec is determined by your current git branch *and* git worktree.
To do implementation work on a spec, we must be in the appropriate worktree directory. If there is no active spec, the assumption is that we are doing planning work, creating new specs, tasks etc.

### 2. Complete Tasks as You Go

Mark tasks complete immediately after finishing them. Do not batch completions.

```bash
mem task complete "task title"
```

### 3. Create a Work Log Before Completing

At the end of a session and as a requirement before completing a spec, create a work log documenting what was done:

```bash
mem log
```

### 4. Completing a Spec

This creates a PR and marks the spec as merge-ready:

```bash
mem spec complete <slug> "commit message for git"
```

## Key Commands

| Command | Purpose |
|---------|---------|
| `mem onboard` | Get full project context |
| `mem spec list` | List all specs |
| `mem spec activate <slug>` | Start working on a spec |
| `mem spec show` | View active spec details |
| `mem task new "title" "desc"` | Create a task |
| `mem task complete "title"` | Mark task done |
| `mem log` | Create work log for session |
| `mem spec complete <slug> "msg"` | Create PR, finish spec |
| `mem sync` | Sync with GitHub |

## Expectations

1. **Run `mem onboard` first** - Always start sessions this way
2. **Complete tasks immediately** - Mark done as soon as finished, not in batches
3. **One task at a time** - Focus on the current task before moving on
4. **Document before completing** - Run `mem log` before `mem spec complete`
5. **Stay on the active spec** - Don't mix work across multiple specs

## File Structure

All mem data lives in `.mem/`:

```
.mem/
├── config.toml           # Project configuration
├── specs/                # Feature specifications
│   └── {slug}/
│       ├── spec.md       # Spec details
│       └── tasks/        # Tasks for this spec
└── logs/                 # Work session logs
```

## Need More Context?

Run `mem onboard` again. It always shows the current state.

## Notes:

- There is ALMOST NEVER a need to `cd` into the project directory, YOUR SHELL IS ALREADY LOCATED AT THE ROOT OF THE PROJECT DIRECTORY.  
- Do not enter plan mode, `mem` is the only tool you need for planning
- Do not use external task lists or task management tools. Use `mem task` instead
- When working with tests. Stop after test runs to ask for instruction. DO NOT RUN TESTS IN A LOOP.

----------------------------------------------------------------------
🛑 IMPORTANT INFORMATION
----------------------------------------------------------------------

- Remember, we are using `mem` to develop `mem` itself. Should changes to the commands be made, e.g. new commands registered or existing commands modified, you will need to test it through `uv run python main.py ...` - where `uv run python main.py` is the equivalent of `mem` (mem is installed in ~/utils/mem). In general, while developing this project it will always be safer to substitute `mem` with `uv run python main.py` to avoid any unintended consequences.
