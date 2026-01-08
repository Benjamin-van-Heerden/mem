"""
Init command - Initialize mem in a project with GitHub integration
"""

import os
import shutil
import subprocess
from pathlib import Path

import typer
from typing_extensions import Annotated

from env_settings import ENV_SETTINGS
from src.utils.github.api import ensure_label, ensure_status_labels
from src.utils.github.client import get_authenticated_user, get_github_client
from src.utils.github.exceptions import GitHubError
from src.utils.github.git_ops import ensure_branches_exist, switch_to_branch
from src.utils.github.repo import get_git_user_info, get_repo_from_git
from src.utils.spec_template import (
    ensure_global_config_exists,
    generate_github_issue_template,
)


def check_prerequisites() -> list[str]:
    """
    Check that all prerequisites are met for mem to work.

    Returns a list of error messages. Empty list means all checks passed.
    """
    errors = []

    # Check for gh CLI
    if not shutil.which("gh"):
        errors.append(
            "GitHub CLI (gh) is not installed.\n"
            "  Install it from: https://cli.github.com/\n"
            "  macOS: brew install gh\n"
            "  Ubuntu: sudo apt install gh"
        )

    # Check for GITHUB_TOKEN
    if not os.getenv("GITHUB_TOKEN"):
        errors.append(
            "GITHUB_TOKEN environment variable is not set.\n"
            "  1. Create a token at: https://github.com/settings/tokens\n"
            "  2. Required scopes: repo, read:org\n"
            "  3. Set it: export GITHUB_TOKEN=your_token_here\n"
            "  4. Or add to .env file in project root"
        )

    # Check for git
    if not shutil.which("git"):
        errors.append(
            "Git is not installed.\n  Install it from: https://git-scm.com/downloads"
        )

    return errors


# Create a Typer app for the init command
app = typer.Typer()


def _get_template_path() -> Path:
    """Get path to the config template."""
    return Path(__file__).parent.parent / "templates" / "config.toml"


def _get_agents_template_path() -> Path:
    """Get path to the AGENTS.md template."""
    return Path(__file__).parent.parent / "templates" / "AGENTS.md"


def configure_merge_settings(project_root: Path):
    """Configure git merge settings.

    Sets merge.ff to false to prevent fast-forward merges,
    ensuring pre-merge-commit hook always triggers.
    """
    try:
        subprocess.run(
            ["git", "config", "merge.ff", "false"],
            cwd=project_root,
            check=True,
            capture_output=True,
        )
        typer.echo("✓ Configured merge.ff=false (ensures merge hooks trigger)")
    except subprocess.CalledProcessError as e:
        typer.echo(f"⚠️  Warning: Could not set merge.ff config: {e}", err=True)


def create_pre_merge_commit_hook(project_root: Path, quiet: bool = False):
    """Create pre-merge-commit hook to enforce branch merge rules.

    Rules enforced:
    - Anything can merge into dev
    - Only dev and hotfix/* can merge into test
    - Only test can merge into main
    """
    git_hooks_dir = project_root / ".git" / "hooks"
    hook_file = git_hooks_dir / "pre-merge-commit"

    hook_content = """#!/bin/bash
# mem: Git merge rules enforcement
# Rules:
#   - Anything can merge into dev
#   - Only dev and hotfix/* can merge into test
#   - Only test can merge into main

TARGET_BRANCH=$(git rev-parse --abbrev-ref HEAD)
# Use name-rev to get branch name from MERGE_HEAD (rev-parse returns literal "MERGE_HEAD")
SOURCE_BRANCH=$(git name-rev --name-only MERGE_HEAD 2>/dev/null | sed 's|remotes/origin/||')

# If we can't determine source branch, allow (might be a commit merge)
if [ -z "$SOURCE_BRANCH" ]; then
    exit 0
fi

case "$TARGET_BRANCH" in
    dev)
        # Anything can merge into dev
        exit 0
        ;;
    test)
        # Only dev and hotfix/* can merge into test
        if [ "$SOURCE_BRANCH" = "dev" ] || [[ "$SOURCE_BRANCH" == hotfix/* ]]; then
            exit 0
        fi
        echo "ERROR: Cannot merge '$SOURCE_BRANCH' into 'test'"
        echo "Only 'dev' and 'hotfix/*' branches can merge into 'test'"
        exit 1
        ;;
    main)
        # Only test can merge into main
        if [ "$SOURCE_BRANCH" = "test" ]; then
            exit 0
        fi
        echo "ERROR: Cannot merge '$SOURCE_BRANCH' into 'main'"
        echo "Only 'test' branch can merge into 'main'"
        exit 1
        ;;
    *)
        # Other branches: allow
        exit 0
        ;;
esac
"""

    if not git_hooks_dir.exists():
        if not quiet:
            typer.echo("⚠️  Warning: .git/hooks directory not found", err=True)
        return

    hook_file.write_text(hook_content)
    hook_file.chmod(0o755)
    if not quiet:
        typer.echo("✓ Created pre-merge-commit hook for branch rules")


def create_agents_files(project_root: Path):
    """Create AGENTS.md and CLAUDE.md symlink in project root."""
    agents_file = project_root / "AGENTS.md"
    claude_file = project_root / "CLAUDE.md"

    if not agents_file.exists():
        template_path = _get_agents_template_path()
        if template_path.exists():
            shutil.copy(template_path, agents_file)
            typer.echo("✓ Created AGENTS.md")
        else:
            typer.echo("⚠️  Warning: AGENTS.md template not found", err=True)
            return

    if not claude_file.exists():
        claude_file.symlink_to("AGENTS.md")
        typer.echo("✓ Created CLAUDE.md symlink -> AGENTS.md")
    elif claude_file.is_symlink():
        typer.echo("✓ CLAUDE.md symlink already exists")
    else:
        typer.echo("⚠️  Warning: CLAUDE.md exists but is not a symlink", err=True)


def create_config_with_discovery(repo_name: str):
    """Create config.toml from template with discovered values."""
    template_path = _get_template_path()
    template = template_path.read_text()

    config_content = template.replace("{project_name}", repo_name)
    config_content = config_content.replace(
        "{project_description}", "Add your project description here."
    )

    with open(ENV_SETTINGS.config_file, "w") as f:
        f.write(config_content)

    typer.echo(f"Created config file: {ENV_SETTINGS.config_file_stripped}")


def create_user_mappings(github_username: str, git_name: str, git_email: str):
    """Create user_mappings.toml with initial user"""
    mappings_file = ENV_SETTINGS.mem_dir / "user_mappings.toml"

    content = f"""# GitHub username to Git user mappings
# Each section is a GitHub username with name and email for commits

[{github_username}]
name = "{git_name}"
email = "{git_email}"
"""

    mappings_file.write_text(content)
    typer.echo("✓ Created user mappings: .mem/user_mappings.toml")


@app.command()
def init(
    force: Annotated[
        bool,
        typer.Option("--force", "-f", help="Force reinitialization without prompting"),
    ] = False,
):
    """
    Initialize mem in the current project.

    This command:
    1. Validates GitHub authentication (requires GITHUB_TOKEN)
    2. Discovers repository from git remote
    3. Links GitHub user to local git user
    4. Creates .mem directory structure
    5. Creates config and user mappings files
    6. Ensures main/test/dev branches exist locally and remotely
    7. Switches to dev branch
    8. Creates GitHub issue template for specs
    9. Creates 'mem-spec' label on GitHub
    10. Creates status labels for sync (mem-status:*)
    """

    typer.echo("🚀 Initializing mem with GitHub integration...\n")

    # Step 0: Check prerequisites
    typer.echo("Checking prerequisites...")
    errors = check_prerequisites()
    if errors:
        typer.echo("\n❌ Prerequisites not met:\n", err=True)
        for error in errors:
            typer.echo(f"• {error}\n", err=True)
        raise typer.Exit(code=1)
    typer.echo("✓ All prerequisites met\n")

    # Step 1: Check for GitHub token
    typer.echo("Step 1/10: Validating GitHub authentication...")
    try:
        github_client = get_github_client()
        typer.echo("✓ GitHub authentication successful")
    except GitHubError as e:
        typer.echo(f"❌ {e}", err=True)
        raise typer.Exit(code=1)

    # Step 2: Get authenticated GitHub user
    typer.echo("\nStep 2/10: Discovering GitHub user...")
    try:
        github_user = get_authenticated_user(github_client)
        github_username = github_user["username"]
        typer.echo(f"✓ Authenticated as GitHub user: {github_username}")
    except GitHubError as e:
        typer.echo(f"❌ {e}", err=True)
        raise typer.Exit(code=1)

    # Step 3: Discover repository from git remote
    typer.echo("\nStep 3/10: Discovering GitHub repository from git remote...")
    try:
        repo_owner, repo_name = get_repo_from_git(ENV_SETTINGS.caller_dir)
        typer.echo(f"✓ Repository: {repo_owner}/{repo_name}")
    except GitHubError as e:
        typer.echo(f"❌ {e}", err=True)
        raise typer.Exit(code=1)

    # Step 4: Get local git user configuration
    typer.echo("\nStep 4/10: Reading local git configuration...")
    try:
        git_user = get_git_user_info(ENV_SETTINGS.caller_dir)
        git_name = git_user["name"]
        git_email = git_user["email"]
        typer.echo(f"✓ Git user: {git_name} <{git_email}>")
    except GitHubError as e:
        typer.echo(f"❌ {e}", err=True)
        raise typer.Exit(code=1)

    # Check if already initialized
    if ENV_SETTINGS.mem_dir.exists() and ENV_SETTINGS.config_file.exists():
        typer.echo("\n⚠️  mem is already initialized in this directory.")

        if not force:
            response = typer.confirm(
                "Reinitialize? This will keep existing data.",
                default=False,
            )
            if not response:
                typer.echo("Initialization cancelled.")
                raise typer.Exit(code=0)

    # Step 5: Create directory structure
    typer.echo("\nStep 5/10: Creating .mem directory structure...")
    ENV_SETTINGS.mem_dir.mkdir(exist_ok=True)
    typer.echo(f"  ✓ Created: {ENV_SETTINGS.mem_dir_stripped}")

    ENV_SETTINGS.specs_dir.mkdir(exist_ok=True)
    typer.echo(f"  ✓ Created: {ENV_SETTINGS.specs_dir_stripped}")

    ENV_SETTINGS.logs_dir.mkdir(exist_ok=True)
    typer.echo(f"  ✓ Created: {ENV_SETTINGS.logs_dir_stripped}")

    # Create todos directory
    todos_dir = ENV_SETTINGS.mem_dir / "todos"
    todos_dir.mkdir(exist_ok=True)
    typer.echo("  ✓ Created: .mem/todos/")

    # Step 6: Create config and user mappings
    typer.echo("\nStep 6/10: Creating configuration files...")

    if not ENV_SETTINGS.config_file.exists() or force:
        create_config_with_discovery(repo_name)
    else:
        typer.echo(f"Config file already exists: {ENV_SETTINGS.config_file_stripped}")

    # Create user mappings file
    mappings_file = ENV_SETTINGS.mem_dir / "user_mappings.toml"
    if not mappings_file.exists() or force:
        create_user_mappings(github_username, git_name, git_email)
    else:
        typer.echo("✓ User mappings file already exists")

    # Create AGENTS.md and CLAUDE.md symlink
    typer.echo("\nCreating agent configuration files...")
    create_agents_files(ENV_SETTINGS.caller_dir)

    # Step 7: Ensure branches exist and switch to dev
    typer.echo("\nStep 7/10: Setting up git branches...")
    try:
        ensure_branches_exist(ENV_SETTINGS.caller_dir, ["main", "test", "dev"])
        typer.echo("✓ Ensured branches exist: main, test, dev")

        switch_to_branch(ENV_SETTINGS.caller_dir, "dev")
        typer.echo("✓ Switched to 'dev' branch")
    except GitHubError as e:
        typer.echo(f"⚠️  Warning: {e}", err=True)
        typer.echo("  (You can manually create branches later)")

    # Create pre-merge-commit hook for branch rules
    create_pre_merge_commit_hook(ENV_SETTINGS.caller_dir)

    # Configure merge settings (disable fast-forward)
    configure_merge_settings(ENV_SETTINGS.caller_dir)

    # Step 8: Ensure global config and create GitHub issue template
    typer.echo("\nStep 8/10: Setting up global config and GitHub issue template...")

    # Ensure ~/.config/mem/ exists with default templates
    try:
        ensure_global_config_exists()
        typer.echo(f"✓ Ensured global config exists: {ENV_SETTINGS.global_config_dir}")
    except Exception as e:
        typer.echo(f"⚠️  Warning: Could not create global config: {e}", err=True)

    # Create GitHub issue template from spec template
    template_dir = ENV_SETTINGS.caller_dir / ".github" / "ISSUE_TEMPLATE"
    template_dir.mkdir(parents=True, exist_ok=True)
    template_file = template_dir / "mem-spec.md"

    template_content = generate_github_issue_template()
    repo = None
    try:
        template_file.write_text(template_content)
        typer.echo("✓ Created issue template: .github/ISSUE_TEMPLATE/mem-spec.md")
    except Exception as e:
        typer.echo(f"⚠️  Warning: Could not create issue template: {e}", err=True)

    # Step 9: Create GitHub label
    typer.echo("\nStep 9/10: Creating 'mem-spec' label on GitHub...")
    try:
        repo = github_client.get_repo(f"{repo_owner}/{repo_name}")
        ensure_label(
            repo,
            name="mem-spec",
            color="0E8A16",
            description="Specifications managed by mem CLI",
        )
        typer.echo("✓ Ensured 'mem-spec' label exists on GitHub")
    except Exception as e:
        typer.echo(f"⚠️  Warning: Could not create GitHub label: {e}", err=True)

    # Step 10: Create status labels for sync
    typer.echo("\nStep 10/10: Creating status labels on GitHub...")
    try:
        assert repo is not None
        ensure_status_labels(repo)
        typer.echo("✓ Created mem-status:* labels for sync")
    except Exception as e:
        typer.echo(f"⚠️  Warning: Could not create status labels: {e}", err=True)

    # Success!
    typer.echo("\n" + "=" * 60)
    typer.echo("✨ mem initialized successfully!")
    typer.echo("=" * 60)
    typer.echo("\n📊 Summary:")
    typer.echo(f"  Repository: {repo_owner}/{repo_name}")
    typer.echo(f"  GitHub User: {github_username}")
    typer.echo(f"  Git User: {git_name}")
    typer.echo("  Current Branch: dev")
    typer.echo(f"\n📁 Configuration: {ENV_SETTINGS.config_file_stripped}")
    typer.echo("\n🎯 Next steps:")
    typer.echo("  1. Run 'mem sync' to pull GitHub issues")
    typer.echo("  2. Run 'mem spec list' to see available specs")
    typer.echo("  3. Run 'mem spec activate <slug>' to start working on a spec")
    typer.echo("  4. Run 'mem onboard' to build AI agent context")


if __name__ == "__main__":
    app()
