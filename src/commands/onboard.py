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


def read_config() -> dict:
    """Read the .mem/config.toml file"""
    if not ENV_SETTINGS.config_file.exists():
        return {}

    try:
        with open(ENV_SETTINGS.config_file, "rb") as f:
            return tomllib.load(f)
    except Exception:
        return {}


def read_file_safely(file_path: Path) -> str | None:
    """Read a file safely, returning None if it doesn't exist or can't be read"""
    try:
        if file_path.exists():
            return file_path.read_text()
    except Exception:
        pass
    return None


def load_generic_templates(config: dict) -> dict[str, str]:
    """Load generic templates from configured location."""
    vars_config = config.get("vars", {})
    templates_location = vars_config.get("generic_templates_location", "")

    if not templates_location:
        return {}

    templates_path = Path(templates_location).expanduser()
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
        output.append(body[:3000])
        if len(body) > 3000:
            output.append("... (truncated)")

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

    # Header
    output.append("=" * 70)
    output.append("PROJECT CONTEXT")
    output.append("=" * 70)
    output.append("")

    # Project info
    project = config.get("project", {})
    output.append(f"**Project:** {project.get('name', 'Unknown')}")
    if project.get("description"):
        desc = project["description"].strip()
        output.append(f"**Description:** {desc}")
    output.append(f"**Branch:** {branch_name}")
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
                output.append(content[:3000])
                if len(content) > 3000:
                    output.append("... (truncated)")
        output.append("")

    # Spec context
    output.append("-" * 70)
    if active_spec:
        output.append(f"ACTIVE SPEC: {active_spec['title']}")
        output.append("-" * 70)
        output.append(format_spec_detail(active_spec))

        # Show related work logs
        recent_logs = logs.list_logs(limit=3)
        spec_logs = [
            log for log in recent_logs if log.get("spec_slug") == active_spec["slug"]
        ]
        if spec_logs:
            output.append("\n### Related Work Logs:")
            for log in spec_logs:
                body_preview = log.get("body", "")[:200]
                output.append(f"  - {log.get('date', 'N/A')}: {body_preview}...")
    else:
        output.append("AVAILABLE SPECS")
        output.append("-" * 70)
        todo_specs = specs.list_specs(status="todo")
        if todo_specs:
            for spec in todo_specs:
                output.append(format_spec_summary(spec))
        else:
            output.append('No specs available. Create one with: mem spec new "title"')
    output.append("")

    # Open todos
    try:
        open_todos = todos.list_todos(status="open")
        if open_todos:
            output.append("-" * 70)
            output.append("OPEN TODOS")
            output.append("-" * 70)
            for todo in open_todos[:5]:
                output.append(f"- {todo['title']}")
            output.append("")
    except Exception:
        pass

    # Next steps
    output.append("-" * 70)
    output.append("SUGGESTED NEXT STEPS")
    output.append("-" * 70)
    output.append(format_next_steps(active_spec, branch_name))

    print("\n".join(output))
