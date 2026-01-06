"""
Onboard command - Build context for AI agents

This command reads project configuration, important files, and active specs/tasks
to provide comprehensive context for AI agents working on the project.
"""

import sys
import tomllib
from pathlib import Path
from typing import Any

from env_settings import ENV_SETTINGS
from src.utils import logs, specs, tasks, todos


def ensure_mem_initialized() -> bool:
    """Check if mem is initialized in the current directory."""
    return ENV_SETTINGS.mem_dir.exists() and ENV_SETTINGS.config_file.exists()


def deep_merge(base: dict, override: dict) -> dict:
    """Deep merge two dicts. override values take precedence."""
    result = base.copy()
    for key, value in override.items():
        if key in result and isinstance(result[key], dict) and isinstance(value, dict):
            result[key] = deep_merge(result[key], value)
        else:
            result[key] = value
    return result


def read_config() -> dict:
    """Read config, merging global (~/.config/mem/config.toml) with local (.mem/config.toml).

    Local config values override global defaults.
    """
    config = {}

    # Load global config first
    if ENV_SETTINGS.global_config_file.exists():
        try:
            with open(ENV_SETTINGS.global_config_file, "rb") as f:
                config = tomllib.load(f)
        except Exception:
            pass

    # Merge local config on top
    if ENV_SETTINGS.config_file.exists():
        try:
            with open(ENV_SETTINGS.config_file, "rb") as f:
                local_config = tomllib.load(f)
                config = deep_merge(config, local_config)
        except Exception:
            pass

    return config


def read_file_safely(file_path: Path) -> str | None:
    """Read a file safely, returning None if it doesn't exist or can't be read"""
    try:
        if file_path.exists():
            return file_path.read_text()
    except Exception:
        pass
    return None


def get_global_config_dir(config: dict) -> Path:
    """Get the global config directory, with default to ~/.config/mem."""
    vars_config = config.get("vars", {})
    global_dir = vars_config.get("global_config_dir", "")

    if global_dir:
        return Path(global_dir).expanduser()
    return ENV_SETTINGS.global_config_dir


def load_generic_templates(config: dict) -> dict[str, str]:
    """Load generic templates from global_config_dir/templates/."""
    global_config_dir = get_global_config_dir(config)
    templates_path = global_config_dir / "templates"

    if not templates_path.exists():
        return {}

    project_config = config.get("project", {})
    template_names = project_config.get("generic_templates", [])

    result = {}
    for name in template_names:
        template_file = templates_path / f"{name}.md"
        if template_file.exists():
            result[name] = template_file.read_text()

    return result


def format_spec_detail(spec: dict[str, Any]) -> str:
    """Format a single spec with full details."""
    output = []
    output.append(f"Title: {spec['title']}")
    output.append(f"Status: {spec['status']}")
    output.append(f"Branch: {spec.get('branch') or 'N/A'}")

    # Show spec body
    body = spec.get("body", "").strip()
    if body:
        output.append("\n### Details:\n")
        output.append(body)

    # List tasks with subtasks
    task_list = tasks.list_tasks(spec["slug"])
    if task_list:
        output.append("\n### Tasks:")
        for task in task_list:
            status_icon = "[x]" if task["status"] == "completed" else "[ ]"
            output.append(f"  {status_icon} {task['title']}")

            # Show subtasks (embedded in frontmatter)
            subtask_list = task.get("subtasks", [])
            for subtask in subtask_list:
                sub_icon = "[x]" if subtask["status"] == "completed" else "[ ]"
                output.append(f"      {sub_icon} {subtask['title']}")

    return "\n".join(output)


def format_spec_summary(spec: dict[str, Any]) -> str:
    """Format a spec as a brief summary."""
    output = []
    output.append(f"\n## {spec['slug']} ({spec['status']})")
    output.append(f"Title: {spec['title']}")

    # Brief preview of body
    body = spec.get("body", "").strip()
    if body:
        preview = body[:300]
        if len(body) > 300:
            preview += "..."
        output.append(f"\n{preview}")

    # Task summary
    task_list = tasks.list_tasks(spec["slug"])
    if task_list:
        completed = sum(1 for t in task_list if t["status"] == "completed")
        output.append(f"\nTasks: {completed}/{len(task_list)} completed")

    return "\n".join(output)


def format_next_steps(active_spec: dict | None, branch_name: str) -> str:
    """Generate actionable next steps based on current state."""
    steps = []

    if active_spec:
        task_list = tasks.list_tasks(active_spec["slug"])

        # Find incomplete tasks
        pending = [t for t in task_list if t["status"] != "completed"]
        if pending:
            steps.append(f"Continue working on: {pending[0]['title']}")
        elif task_list:
            steps.append(
                f'All tasks completed! Run: mem spec complete {active_spec["slug"]} "commit message"'
            )
        else:
            steps.append(
                f'Add tasks to spec: mem task new "title" "description" --spec {active_spec["slug"]}'
            )
    else:
        if branch_name in ("dev", "main", "master"):
            steps.append("Activate a spec to start working: mem spec activate <slug>")
            steps.append('Or create a new spec: mem spec new "feature name"')
        else:
            steps.append(f"You're on branch '{branch_name}' with no associated spec.")
            steps.append("Consider creating a spec for this work or switching to dev.")

    steps.append("Create a work log for this session: mem log")

    output: list[str] = []
    for i, step in enumerate(steps, 1):
        output.append(f"{i}. {step}")

    return "\n".join(output)


def run_sync_quietly():
    """Run sync in background, fail gracefully."""
    try:
        from src.commands.sync import sync

        sync()
    except Exception:
        pass


def format_work_log_entry(log: dict[str, Any]) -> str:
    """Format a work log entry for display."""
    output = []
    date = log.get("date", "Unknown date")
    username = log.get("username", "")
    spec_slug = log.get("spec_slug", "")

    header = f"**{date}**"
    if username:
        header += f" ({username})"
    if spec_slug:
        header += f" - spec: {spec_slug}"
    output.append(header)

    body = log.get("body", "").strip()
    if body:
        # Show first 500 chars of body
        preview = body[:500]
        if len(body) > 500:
            preview += "..."
        output.append(preview)

    return "\n".join(output)


def onboard():
    """Build and display project context for AI agents."""

    if not ensure_mem_initialized():
        print("mem is not initialized. Run 'mem init' first.", file=sys.stderr)
        sys.exit(1)

    # 1. Run sync (optional, fail silently)
    print("Syncing with GitHub...", file=sys.stderr)
    try:
        run_sync_quietly()
        print("Sync complete.", file=sys.stderr)
    except Exception as e:
        print(f"Sync skipped: {e}", file=sys.stderr)

    print("", file=sys.stderr)  # Blank line before main output

    # 2. Load config
    config = read_config()

    # 3. Get branch state
    branch_name, active_spec, warning = specs.get_branch_status()

    # 4. Build output
    output = []

    # Header with mem explanation
    output.append("=" * 70)
    output.append("PROJECT CONTEXT (generated by mem)")
    output.append("=" * 70)
    output.append("")
    output.append("## About mem")
    output.append("")
    output.append(
        "mem is a CLI tool for managing project context in AI-assisted development."
    )
    output.append(
        "It uses a file-first, git-native architecture where all data is stored as"
    )
    output.append("markdown files with YAML frontmatter in the .mem/ directory.")
    output.append("")
    output.append("**Core concepts:**")
    output.append(
        "- **Specs**: High-level feature specifications (linked to GitHub issues)"
    )
    output.append("- **Tasks**: Concrete work items within a spec")
    output.append("- **Subtasks**: Granular breakdown embedded in task frontmatter")
    output.append("- **Work Logs**: Session records of what was done and what's next")
    output.append("")
    output.append("**Key commands:**")
    output.append("- `mem spec activate <slug>` - Switch to a spec's feature branch")
    output.append('- `mem task new "title" "desc"` - Create a task for active spec')
    output.append('- `mem task complete "title"` - Mark task done')
    output.append(
        '- `mem spec complete <slug> "msg"` - Create PR, mark spec merge_ready'
    )
    output.append("- `mem merge` - Merge completed PRs and clean up branches")
    output.append("- `mem log` - Create/update work log for the session")
    output.append("- `mem sync` - Bidirectional sync with GitHub issues")
    output.append("")

    # Project info
    output.append("-" * 70)
    output.append("PROJECT INFO")
    output.append("-" * 70)
    project = config.get("project", {})
    output.append(f"**Project:** {project.get('name', 'Unknown')}")
    if project.get("description"):
        desc = project["description"].strip()
        output.append(f"**Description:** {desc}")
    output.append(f"**Current Branch:** {branch_name}")
    output.append("")

    # Warning if applicable
    if warning:
        output.append(f"WARNING: {warning}")
        output.append("")

    # Generic templates (coding guidelines)
    generic_templates = load_generic_templates(config)
    if generic_templates:
        output.append("-" * 70)
        output.append("CODING GUIDELINES")
        output.append("-" * 70)
        for name, content in generic_templates.items():
            output.append(f"\n## {name}\n")
            output.append(content.strip())
        output.append("")

    # Important files
    files_config = config.get("files", [])
    if files_config:
        output.append("-" * 70)
        output.append("IMPORTANT FILES")
        output.append("-" * 70)
        for file_entry in files_config:
            path = file_entry.get("path")
            desc = file_entry.get("description", "")
            content = read_file_safely(Path(path))
            if content:
                output.append(f"\n## {path}")
                if desc:
                    output.append(f"*{desc}*\n")
                output.append(content)
        output.append("")

    # Spec context
    output.append("-" * 70)
    if active_spec:
        output.append(f"ACTIVE SPEC: {active_spec['title']}")
        output.append("-" * 70)
        output.append("")
        output.append(
            "You are currently working on this spec. Complete its tasks, then run"
        )
        output.append(
            f'`mem spec complete {active_spec["slug"]} "commit message"` to create a PR.'
        )
        output.append("")
        output.append(format_spec_detail(active_spec))
    else:
        output.append("AVAILABLE SPECS")
        output.append("-" * 70)
        output.append("")
        output.append("No spec is currently active. You are on the dev branch.")
        output.append(
            "Activate a spec with `mem spec activate <slug>` to start working on it."
        )
        output.append("")
        todo_specs = specs.list_specs(status="todo")
        merge_ready_specs = specs.list_specs(status="merge_ready")

        if merge_ready_specs:
            output.append("### Specs ready to merge:")
            for spec in merge_ready_specs:
                output.append(f"  - {spec['slug']}: {spec['title']}")
                if spec.get("pr_url"):
                    output.append(f"    PR: {spec['pr_url']}")
            output.append("")
            output.append("Run `mem merge` to merge these PRs.")
            output.append("")

        if todo_specs:
            output.append("### Specs to work on:")
            for spec in todo_specs:
                output.append(format_spec_summary(spec))
        else:
            if not merge_ready_specs:
                output.append(
                    'No specs available. Create one with: mem spec new "title"'
                )
    output.append("")

    # Work Logs section - show recent logs prominently
    output.append("-" * 70)
    output.append("RECENT WORK LOGS")
    output.append("-" * 70)
    output.append("")
    output.append(
        "Work logs capture what was accomplished in each session, blockers encountered,"
    )
    output.append(
        "and suggested next steps. Review these to understand recent progress."
    )
    output.append("")

    try:
        recent_logs = logs.list_logs(limit=5)
        if recent_logs:
            for log in recent_logs:
                output.append(format_work_log_entry(log))
                output.append("")
        else:
            output.append("No work logs yet. Create one with: mem log")
            output.append("")
    except Exception:
        output.append("Could not load work logs.")
        output.append("")

    # Open todos
    try:
        open_todos = todos.list_todos(status="open")
        if open_todos:
            output.append("-" * 70)
            output.append("OPEN TODOS")
            output.append("-" * 70)
            output.append("")
            output.append("Standalone reminders not tied to any spec:")
            for todo in open_todos[:5]:
                output.append(f"- {todo['title']}")
            if len(open_todos) > 5:
                output.append(f"  ... and {len(open_todos) - 5} more")
            output.append("")
    except Exception:
        pass

    # Next steps
    output.append("-" * 70)
    output.append("SUGGESTED NEXT STEPS")
    output.append("-" * 70)
    output.append("")
    output.append(format_next_steps(active_spec, branch_name))
    output.append("")
    output.append("Remember to create a work log at the end of your session: mem log")

    print("\n".join(output))
