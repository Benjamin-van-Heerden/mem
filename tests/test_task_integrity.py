"""
Tests for task integrity rules:
- Tasks cannot be completed if they have incomplete subtasks
- Subtasks are now embedded in task frontmatter
"""

import pytest
import typer

from src.commands.subtask import complete as subtask_complete
from src.commands.subtask import delete as subtask_delete
from src.commands.task import complete as task_complete
from src.commands.task import delete as task_delete
from src.utils import specs, tasks


@pytest.fixture
def initialized_mem(setup_test_env, monkeypatch):
    """Initialize mem directory structure and return the repo path."""
    repo_path = setup_test_env
    monkeypatch.chdir(repo_path)

    # Create .mem directory structure
    (repo_path / ".mem").mkdir(exist_ok=True)
    (repo_path / ".mem" / "specs").mkdir(exist_ok=True)
    (repo_path / ".mem" / "todos").mkdir(exist_ok=True)
    (repo_path / ".mem" / "logs").mkdir(exist_ok=True)

    return repo_path


def test_task_completion_guard(initialized_mem):
    """Verify tasks cannot be completed if they have incomplete subtasks."""

    # 1. Setup: Create spec, task, and subtask
    specs.create_spec("Integrity Spec")
    spec_slug = "integrity_spec"

    # New API: create_task requires title and description
    tasks.create_task(spec_slug, "Parent Task", "Description of parent task")
    task_filename = "01_parent_task.md"

    # New API: add_subtask instead of create_subtask
    tasks.add_subtask(spec_slug, task_filename, "Subtask")

    # 2. Attempt to complete parent task - should fail due to incomplete subtask
    with pytest.raises(typer.Exit) as excinfo:
        # New API: complete requires title and notes
        task_complete(
            title="Parent Task", notes="Completion notes", spec_slug=spec_slug
        )
    assert excinfo.value.exit_code == 1

    # Verify status is still todo
    task = tasks.get_task(spec_slug, task_filename)
    assert task is not None
    assert task["status"] == "todo"

    # 3. Complete subtask using new API
    subtask_complete(title="Subtask", task_title="Parent Task", spec_slug=spec_slug)

    # 4. Attempt to complete parent task - should succeed now
    task_complete(title="Parent Task", notes="Task completed", spec_slug=spec_slug)
    task = tasks.get_task(spec_slug, task_filename)
    assert task is not None
    assert task["status"] == "completed"


def test_task_deletion(initialized_mem):
    """Verify tasks can be deleted (subtasks are embedded, not separate files)."""

    specs.create_spec("Integrity Spec 2")
    spec_slug = "integrity_spec_2"

    tasks.create_task(spec_slug, "Parent Task", "Description")
    task_filename = "01_parent_task.md"

    tasks.add_subtask(spec_slug, task_filename, "Subtask")

    # Delete task - should work since subtasks are embedded
    task_delete(title="Parent Task", spec_slug=spec_slug)
    assert tasks.get_task(spec_slug, task_filename) is None


def test_subtask_deletion(initialized_mem):
    """Verify subtasks can be deleted regardless of status."""

    specs.create_spec("Integrity Spec 3")
    spec_slug = "integrity_spec_3"

    tasks.create_task(spec_slug, "Parent Task", "Description")
    task_filename = "01_parent_task.md"

    tasks.add_subtask(spec_slug, task_filename, "Sub1")
    tasks.add_subtask(spec_slug, task_filename, "Sub2")

    # Complete one subtask
    subtask_complete(title="Sub1", task_title="Parent Task", spec_slug=spec_slug)

    # Delete both (one completed, one todo)
    subtask_delete(title="Sub1", task_title="Parent Task", spec_slug=spec_slug)
    subtask_delete(title="Sub2", task_title="Parent Task", spec_slug=spec_slug)

    # Verify subtasks are gone
    subtasks = tasks.list_subtasks(spec_slug, task_filename)
    assert len(subtasks) == 0


def test_subtasks_embedded_in_frontmatter(initialized_mem):
    """Verify subtasks are stored in task frontmatter."""

    specs.create_spec("Embedded Spec")
    spec_slug = "embedded_spec"

    tasks.create_task(spec_slug, "Test Task", "Description")
    task_filename = "01_test_task.md"

    tasks.add_subtask(spec_slug, task_filename, "Subtask 1")
    tasks.add_subtask(spec_slug, task_filename, "Subtask 2")

    # Get task and check subtasks are in frontmatter
    task = tasks.get_task(spec_slug, task_filename)
    assert task is not None
    assert "subtasks" in task
    assert len(task["subtasks"]) == 2
    assert task["subtasks"][0]["title"] == "Subtask 1"
    assert task["subtasks"][1]["title"] == "Subtask 2"
