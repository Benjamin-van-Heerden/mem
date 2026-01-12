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
from src.commands.init import create_pre_merge_commit_hook
from src.utils import docs, logs, specs, tasks, todos, worktrees
from src.utils.specs import ensure_on_dev_branch


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


def filter_readme_sections(content: str) -> str:
    """Filter out Installation and Prerequisites sections from README content."""
    sections_to_remove = ["installation", "prerequisites"]
    lines = content.split("\n")
    result_lines = []
    skip_until_next_section = False
    current_heading_level = 0

    for line in lines:
        stripped = line.strip()

        # Check if this is a heading
        if stripped.startswith("#"):
            # Count heading level
            heading_level = 0
            for char in stripped:
                if char == "#":
                    heading_level += 1
                else:
                    break

            # Get heading text (lowercase for comparison)
            heading_text = stripped.lstrip("#").strip().lower()

            # Check if this heading should be skipped
            if heading_text in sections_to_remove:
                skip_until_next_section = True
                current_heading_level = heading_level
                continue

            # If we were skipping and hit a heading at same or higher level, stop skipping
            if skip_until_next_section and heading_level <= current_heading_level:
                skip_until_next_section = False

        if not skip_until_next_section:
            result_lines.append(line)

    return "\n".join(result_lines)


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

    # List tasks
    task_list = tasks.list_tasks(spec["slug"])
    if task_list:
        output.append("\n### Tasks:")
        for task in task_list:
            status_icon = "[x]" if task["status"] == "completed" else "[ ]"
            output.append(f"  {status_icon} {task['title']}")

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
            steps.append('Add tasks to spec: mem task new "title" "description"')
    else:
        steps.append('Create a new spec: mem spec new "feature name"')
        steps.append("Or assign an existing spec: mem spec assign <slug>")

    steps.append("Create a work log for this session: mem log")

    output: list[str] = []
    for i, step in enumerate(steps, 1):
        output.append(f"{i}. {step}")

    return "\n".join(output)


class SyncFailure:
    """Represents a sync failure that needs user attention."""

    def __init__(self, error_type: str, message: str):
        self.error_type = error_type
        self.message = message


def run_sync_quietly() -> SyncFailure | None:
    """Run sync, returning failure info if it fails.

    Returns:
        SyncFailure if sync failed, None if successful.
    """
    import typer

    from src.commands.sync import sync

    try:
        sync(dry_run=False, no_git=False, no_cleanup=False)
        return None
    except typer.Exit as e:
        if e.exit_code != 0:
            return SyncFailure(
                error_type="sync_failed",
                message="Sync failed. Check the error messages above.",
            )
        return None
    except Exception as e:
        return SyncFailure(
            error_type="unexpected_error",
            message=f"Unexpected error during sync: {e}",
        )


def format_work_log_entry(log: dict[str, Any]) -> str:
    """Format a work log entry for display with visual separation."""
    output = []
    date = log.get("date", "Unknown date")
    username = log.get("username", "")
    spec_slug = log.get("spec_slug", "")

    # Top border
    output.append("┌" + "─" * 68 + "┐")

    # Header line
    header = f"  {date}"
    if username:
        header += f" ({username})"
    if spec_slug:
        header += f" - spec: {spec_slug}"
    output.append(header)

    output.append("├" + "─" * 68 + "┤")

    body = log.get("body", "").strip()
    if body:
        output.append(body)

    # Bottom border
    output.append("└" + "─" * 68 + "┘")

    return "\n".join(output)


def onboard():
    """Build and display project context for AI agents."""

    if not ensure_mem_initialized():
        print("mem is not initialized. Run 'mem init' first.", file=sys.stderr)
        sys.exit(1)

    # Ensure git hooks are in place (silent)
    create_pre_merge_commit_hook(ENV_SETTINGS.caller_dir, quiet=True)

    # Auto-switch to dev if on main or test
    switched, switch_msg = ensure_on_dev_branch()
    if switched and switch_msg:
        print(f"⚠️  {switch_msg}", file=sys.stderr)

    # 1. Run sync and capture any failures
    print("Syncing with GitHub...", file=sys.stderr)
    sync_failure = run_sync_quietly()
    if sync_failure:
        print(f"Sync failed: {sync_failure.message}", file=sys.stderr)
    else:
        print("Sync complete.", file=sys.stderr)

    print("", file=sys.stderr)  # Blank line before main output

    # 2. Load config
    config = read_config()

    # 3. Get branch state
    branch_name, active_spec, warning = specs.get_branch_status()

    # 4. Build output
    output = []

    # Header with mem explanation
    output.append("=" * 70)
    output.append("📋 PROJECT CONTEXT (generated by mem)")
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
    output.append("- **Work Logs**: Session records of what was done and what's next")
    output.append("")
    output.append("**Key commands:**")
    output.append('- `mem spec new "title"` - Create a new spec')
    output.append(
        "- `mem spec assign <slug>` - Assign spec to yourself and create worktree"
    )
    output.append(
        '- `mem task new "title" "description"` - Create a task for active spec'
    )
    output.append('- `mem task complete "title"` - Mark task done')
    output.append(
        '- `mem spec complete <slug> "commit message"` - Create PR, mark spec merge_ready'
    )
    output.append("- `mem merge` - Merge completed PRs and clean up worktrees")
    output.append("- `mem log` - Create/update work log for the session")
    output.append("- `mem sync` - Bidirectional sync with GitHub issues")
    output.append("")
    output.append("**Branch merge rules:**")
    output.append("- anything → dev (feature branches merge here)")
    output.append("- dev or hotfix/* → test")
    output.append("- test → main")
    output.append("")

    # Project info
    output.append("-" * 70)
    output.append("📁 PROJECT INFO")
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
        output.append("📝 CODING GUIDELINES")
        output.append("-" * 70)
        for name, content in generic_templates.items():
            output.append(f"\n## {name}\n")
            output.append(content.strip())
            output.append("\n" + "=" * 40 + "\n")
        output.append("")

    # Important files
    files_config = config.get("files", [])
    if files_config:
        output.append("-" * 70)
        output.append("📄 IMPORTANT FILES")
        output.append("-" * 70)
        for file_entry in files_config:
            path = file_entry.get("path")
            desc = file_entry.get("description", "")
            content = read_file_safely(Path(path))
            if content:
                # Filter out Installation/Prerequisites from README files
                if path.lower().endswith("readme.md"):
                    content = filter_readme_sections(content)
                output.append(f"\n## {path}")
                if desc:
                    output.append(f"*{desc}*\n")
                output.append(content)
                output.append("\n" + "=" * 40 + "\n")
        output.append("")

    # Core documentation section (always included in full)
    try:
        core_doc_files = docs.list_core_doc_files()
        if core_doc_files:
            output.append("-" * 70)
            output.append("📚 CORE DOCUMENTATION")
            output.append("-" * 70)
            output.append("")

            for file_path in core_doc_files:
                slug = docs.get_core_doc_slug(file_path)
                content = docs.read_core_doc(slug)
                if content:
                    output.append(f"### {slug}")
                    output.append("")
                    output.append(content.strip())
                    output.append("")
    except Exception:
        pass

    # Technical documentation section
    try:
        indexed_docs = docs.get_indexed_docs()
        if indexed_docs:
            output.append("-" * 70)
            output.append("📖 TECHNICAL DOCUMENTATION")
            output.append("-" * 70)
            output.append("")

            indexed = [d for d in indexed_docs if d["indexed"]]
            unindexed = [d for d in indexed_docs if not d["indexed"]]

            if indexed:
                for doc_info in indexed:
                    slug = doc_info["slug"]
                    output.append(f"### {slug}")
                    summary = docs.read_summary(slug)
                    if summary:
                        output.append(summary.strip())
                    else:
                        output.append("*(No summary available)*")
                    output.append("")

            if unindexed:
                unindexed_names = ", ".join(d["slug"] for d in unindexed)
                output.append(
                    f"⚠️ Unindexed docs found: {unindexed_names}. Run `mem docs index` to index."
                )
                output.append("")
    except Exception:
        pass

    # Spec context
    output.append("-" * 70)
    if active_spec:
        output.append(f"📋 ACTIVE SPEC: {active_spec['title']}")
        output.append("-" * 70)
        output.append("")
        output.append(
            "You are currently working on this spec. Complete its tasks, then run"
        )
        output.append(
            f'`mem spec complete {active_spec["slug"]} "detailed commit message"` to create a PR.'
        )
        output.append("")

        # Show files changed in this spec branch
        diff_stat = specs.get_branch_diff_stat()
        if diff_stat:
            output.append("### Files modified in this spec (vs dev):")
            output.append("```")
            output.append(diff_stat)
            output.append("```")
            output.append("")

        output.append(format_spec_detail(active_spec))
    else:
        output.append("📋 AVAILABLE SPECS")
        output.append("-" * 70)
        output.append("")
        output.append("No spec is currently active. You are in the main repo.")
        output.append("")

        # Show active worktrees
        main_repo_path = ENV_SETTINGS.caller_dir
        all_worktrees = worktrees.list_worktrees(main_repo_path)
        spec_worktrees = [wt for wt in all_worktrees if not wt.is_main]
        if spec_worktrees:
            output.append("### 📂 Active worktrees:")
            output.append("Each worktree is an isolated workspace for a spec.")
            for wt in spec_worktrees:
                slug = wt.path.name
                output.append(f"  - {slug}: {wt.path}")
            output.append("")
            output.append(
                "To work on a spec, open a terminal in its worktree directory."
            )
            output.append("")

        todo_specs = specs.list_specs(status="todo")
        merge_ready_specs = specs.list_specs(status="merge_ready")

        if merge_ready_specs:
            output.append("### ✅ Specs ready to merge:")
            for spec in merge_ready_specs:
                output.append(f"  - {spec['slug']}: {spec['title']}")
                if spec.get("pr_url"):
                    output.append(f"    🔗 PR: {spec['pr_url']}")
            output.append("")
            output.append("💡 Run `mem merge` to merge these PRs.")
            output.append("")

        if todo_specs:
            output.append("### 📝 Specs to work on:")
            for spec in todo_specs:
                output.append(format_spec_summary(spec))
        else:
            if not merge_ready_specs:
                output.append(
                    '💡 No specs available. Create one with: mem spec new "title"'
                )
                output.append("")

                # Show recently completed specs for context
                completed_specs = specs.list_specs(status="completed")
                if completed_specs:
                    output.append("### ✅ Recently completed specs:")
                    output.append("These were the last completed specs for context:")
                    output.append("")
                    for spec in completed_specs[:2]:
                        output.append(f"**{spec['slug']}**: {spec['title']}")
                        body = spec.get("body", "").strip()
                        if body:
                            # Show truncated body (first 500 chars)
                            if len(body) > 500:
                                body = body[:500] + "..."
                            output.append(body)
                        output.append("")
    output.append("")

    # Work Logs section - show recent logs prominently
    output.append("-" * 70)
    output.append("📝 RECENT WORK LOGS")
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
        if active_spec:
            # Show ALL logs for the active spec (no limit)
            recent_logs = logs.list_logs(limit=100, spec_slug=active_spec["slug"])
        else:
            recent_logs = logs.list_logs(limit=3)
            # Ensure diversity: if all 3 logs are from the same spec,
            # show only 2 from that spec and find 1 from a different spec
            if len(recent_logs) == 3:
                spec_slugs = [log.get("spec_slug") for log in recent_logs]
                if spec_slugs[0] and spec_slugs[0] == spec_slugs[1] == spec_slugs[2]:
                    # All 3 from same spec - find a log from a different spec
                    all_logs = logs.list_logs(limit=10)
                    different_spec_log = None
                    for log in all_logs:
                        if log.get("spec_slug") != spec_slugs[0]:
                            different_spec_log = log
                            break
                    if different_spec_log:
                        # Keep 2 from the dominant spec, add 1 from different spec
                        recent_logs = recent_logs[:2] + [different_spec_log]
        if recent_logs:
            # Reverse to chronological order (oldest first, newest last)
            for log in reversed(recent_logs):
                output.append(format_work_log_entry(log))
                output.append("")
        else:
            if active_spec:
                output.append(
                    f"No work logs for spec '{active_spec['slug']}' yet. Create one with: mem log"
                )
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
            output.append("📌 OPEN TODOS")
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

    # Agent workflow hints
    if active_spec:
        output.append("-" * 70)
        output.append("💡 AGENT WORKFLOW HINTS")
        output.append("-" * 70)
        output.append("")
        output.append("Working with tasks:")
        output.append(
            '  - Create task: mem task new "title" "detailed description with implementation notes if necessary"'
        )
        output.append(
            '  - Complete task: mem task complete "title" "notes about what was done"'
        )
        output.append("  - List tasks: mem task list")
        output.append("")
        output.append("Important workflow rules:")
        output.append(
            "  - Complete ONE task at a time, then STOP and await further instructions"
        )
        output.append(
            "  - **IMPORTANT**: Mark each task complete AS SOON AS you finish it!"
        )
        output.append("  - Do not batch task completions - complete them one at a time")
        output.append(
            '  - When all tasks are done, run: mem spec complete <slug> "detailed commit message"'
        )
        output.append("")

    # Next steps
    output.append("-" * 70)
    output.append("👉 SUGGESTED NEXT STEPS")
    output.append("-" * 70)
    output.append("")
    output.append(format_next_steps(active_spec, branch_name))
    output.append("")
    output.append("Remember to create a work log at the end of your session: mem log")

    # Agent halt instruction
    output.append("")
    output.append("-" * 70)
    output.append("[AGENT INSTRUCTION]")
    output.append("-" * 70)
    output.append("Your next response must:")
    output.append(
        "1. Briefly summarize the current state (active spec, pending tasks, etc.)"
    )
    output.append("2. Ask the user how they would like to proceed")
    output.append("Do NOT call any tools. Do NOT start working on tasks yet.")
    output.append("Wait for explicit user instruction before taking any action.")

    # CRITICAL: Show sync failure warning at the very end so it's the last thing seen
    if sync_failure:
        output.append("")
        output.append("")
        output.append("!" * 70)
        output.append("!" * 70)
        output.append("🚨🚨🚨 SYNC FAILED - FIX THIS BEFORE DOING ANYTHING ELSE 🚨🚨🚨")
        output.append("!" * 70)
        output.append("!" * 70)
        output.append("")
        output.append("The sync/rebase operation failed. This means your branch is")
        output.append("OUT OF SYNC with origin/dev and needs manual intervention.")
        output.append("")
        output.append("DO NOT proceed with any work until this is resolved!")
        output.append("")
        output.append("To fix this:")
        output.append("  1. git fetch origin")
        output.append("  2. git rebase origin/dev")
        output.append("  3. Resolve any conflicts that arise")
        output.append("  4. git rebase --continue")
        output.append("  5. git push --force-with-lease")
        output.append("  6. Run 'mem onboard' again to verify")
        output.append("")
        output.append("If the rebase is too complex, you can also:")
        output.append("  - git rebase --abort  (to undo the rebase attempt)")
        output.append("  - Ask for help resolving the conflicts")
        output.append("")
        output.append("!" * 70)
        output.append("!" * 70)

    print("\n".join(output))
