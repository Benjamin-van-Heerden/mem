"""
Tests for spec subdirectory functionality:
- Specs can be moved to completed/ and abandoned/ subdirectories
- list_specs respects status filtering for subdirectories
- get_spec finds specs in all locations
"""

import pytest

from src.utils import specs, tasks


@pytest.fixture
def initialized_mem(setup_test_env, monkeypatch):
    """Initialize mem directory structure and return the repo path."""
    repo_path = setup_test_env
    monkeypatch.chdir(repo_path)

    # Create .mem directory structure
    (repo_path / ".mem").mkdir(exist_ok=True)
    (repo_path / ".mem" / "specs").mkdir(exist_ok=True)
    (repo_path / ".mem" / "specs" / "completed").mkdir(exist_ok=True)
    (repo_path / ".mem" / "specs" / "abandoned").mkdir(exist_ok=True)
    (repo_path / ".mem" / "todos").mkdir(exist_ok=True)
    (repo_path / ".mem" / "logs").mkdir(exist_ok=True)

    return repo_path


def test_move_spec_to_completed(initialized_mem):
    """Test moving a spec to the completed subdirectory."""
    # Create a spec
    specs.create_spec("Completable Feature")
    spec_slug = "completable_feature"

    # Verify it exists in root
    spec = specs.get_spec(spec_slug)
    assert spec is not None
    assert spec["status"] == "todo"

    # Verify it appears in list_specs (default)
    all_specs = specs.list_specs()
    assert any(s["slug"] == spec_slug for s in all_specs)

    # Move to completed
    new_path = specs.move_spec_to_completed(spec_slug)
    assert "completed" in str(new_path)

    # Verify spec can still be found by get_spec
    spec = specs.get_spec(spec_slug)
    assert spec is not None
    assert spec["status"] == "completed"
    assert spec["completed_at"] is not None

    # Verify it no longer appears in default list_specs
    all_specs = specs.list_specs()
    assert not any(s["slug"] == spec_slug for s in all_specs)

    # Verify it appears when filtering by completed status
    completed_specs = specs.list_specs(status="completed")
    assert any(s["slug"] == spec_slug for s in completed_specs)


def test_move_spec_to_abandoned(initialized_mem):
    """Test moving a spec to the abandoned subdirectory."""
    # Create a spec
    specs.create_spec("Abandonable Feature")
    spec_slug = "abandonable_feature"

    # Verify it exists
    spec = specs.get_spec(spec_slug)
    assert spec is not None
    assert spec["status"] == "todo"

    # Move to abandoned
    new_path = specs.move_spec_to_abandoned(spec_slug)
    assert "abandoned" in str(new_path)

    # Verify spec can still be found
    spec = specs.get_spec(spec_slug)
    assert spec is not None
    assert spec["status"] == "abandoned"

    # Verify it no longer appears in default list_specs
    all_specs = specs.list_specs()
    assert not any(s["slug"] == spec_slug for s in all_specs)

    # Verify it appears when filtering by abandoned status
    abandoned_specs = specs.list_specs(status="abandoned")
    assert any(s["slug"] == spec_slug for s in abandoned_specs)


def test_move_spec_with_tasks(initialized_mem):
    """Test that tasks are preserved when moving a spec."""
    # Create spec with tasks (new API requires description)
    specs.create_spec("Spec With Tasks")
    spec_slug = "spec_with_tasks"

    tasks.create_task(spec_slug, "Task One", "Description of task one")
    tasks.create_task(spec_slug, "Task Two", "Description of task two")

    # Verify tasks exist
    task_list = tasks.list_tasks(spec_slug)
    assert len(task_list) == 2

    task_one = tasks.get_task(spec_slug, "01_task_one.md")
    assert task_one is not None

    # Move to completed
    specs.move_spec_to_completed(spec_slug)

    # Verify tasks still exist after move
    task_list = tasks.list_tasks(spec_slug)
    assert len(task_list) == 2

    task_one = tasks.get_task(spec_slug, "01_task_one.md")
    assert task_one is not None


def test_move_nonexistent_spec_fails(initialized_mem):
    """Test that moving a non-existent spec raises an error."""
    with pytest.raises(ValueError, match="not found"):
        specs.move_spec_to_completed("nonexistent_spec")

    with pytest.raises(ValueError, match="not found"):
        specs.move_spec_to_abandoned("nonexistent_spec")


def test_move_already_moved_spec_fails(initialized_mem):
    """Test that moving a spec twice raises an error."""
    specs.create_spec("Double Move Spec")
    spec_slug = "double_move_spec"

    # Move to completed
    specs.move_spec_to_completed(spec_slug)

    # Try to move again - should fail (already in completed)
    with pytest.raises(ValueError, match="already exists"):
        specs.move_spec_to_completed(spec_slug)


def test_list_specs_excludes_subdirectories_by_default(initialized_mem):
    """Test that list_specs with no status excludes completed/abandoned."""
    # Create specs in different states
    specs.create_spec("Active Spec")
    specs.create_spec("Completed Spec")
    specs.create_spec("Abandoned Spec")

    # Move some to subdirectories
    specs.move_spec_to_completed("completed_spec")
    specs.move_spec_to_abandoned("abandoned_spec")

    # Default list should only show active_spec
    all_specs = specs.list_specs()
    slugs = [s["slug"] for s in all_specs]

    assert "active_spec" in slugs
    assert "completed_spec" not in slugs
    assert "abandoned_spec" not in slugs


def test_list_specs_with_status_filter(initialized_mem):
    """Test that list_specs correctly filters by status."""
    # Create specs with different statuses
    # Note: "active" and "inactive" statuses no longer exist
    # Status is now: todo, merge_ready, completed, abandoned
    specs.create_spec("Todo Spec")

    specs.create_spec("Merge Ready Spec")
    specs.update_spec_status("merge_ready_spec", "merge_ready")

    # Filter by todo
    todo_specs = specs.list_specs(status="todo")
    assert any(s["slug"] == "todo_spec" for s in todo_specs)
    assert not any(s["slug"] == "merge_ready_spec" for s in todo_specs)

    # Filter by merge_ready
    merge_ready_specs = specs.list_specs(status="merge_ready")
    assert any(s["slug"] == "merge_ready_spec" for s in merge_ready_specs)
    assert not any(s["slug"] == "todo_spec" for s in merge_ready_specs)


def test_get_spec_finds_spec_in_completed(initialized_mem):
    """Test that get_spec can find specs in the completed subdirectory."""
    specs.create_spec("Find Me Completed")
    spec_slug = "find_me_completed"

    # Move to completed
    specs.move_spec_to_completed(spec_slug)

    # get_spec should still find it
    spec = specs.get_spec(spec_slug)
    assert spec is not None
    assert spec["slug"] == spec_slug
    assert spec["status"] == "completed"


def test_get_spec_finds_spec_in_abandoned(initialized_mem):
    """Test that get_spec can find specs in the abandoned subdirectory."""
    specs.create_spec("Find Me Abandoned")
    spec_slug = "find_me_abandoned"

    # Move to abandoned
    specs.move_spec_to_abandoned(spec_slug)

    # get_spec should still find it
    spec = specs.get_spec(spec_slug)
    assert spec is not None
    assert spec["slug"] == spec_slug
    assert spec["status"] == "abandoned"


def test_get_spec_file_path_works_for_moved_specs(initialized_mem):
    """Test that get_spec_file_path returns correct path for moved specs."""
    specs.create_spec("Path Test Spec")
    spec_slug = "path_test_spec"

    # Get path before move
    path_before = specs.get_spec_file_path(spec_slug)
    assert path_before.exists()
    assert "completed" not in str(path_before)

    # Move to completed
    specs.move_spec_to_completed(spec_slug)

    # Get path after move
    path_after = specs.get_spec_file_path(spec_slug)
    assert path_after.exists()
    assert "completed" in str(path_after)


def test_update_spec_works_for_moved_specs(initialized_mem):
    """Test that update_spec works for specs in subdirectories."""
    specs.create_spec("Update After Move")
    spec_slug = "update_after_move"

    # Move to abandoned
    specs.move_spec_to_abandoned(spec_slug)

    # Update should still work
    specs.update_spec(spec_slug, assigned_to="test_user")

    spec = specs.get_spec(spec_slug)
    assert spec is not None
    assert spec["assigned_to"] == "test_user"


def test_delete_spec_works_for_moved_specs(initialized_mem):
    """Test that delete_spec works for specs in subdirectories."""
    specs.create_spec("Delete After Move")
    spec_slug = "delete_after_move"

    # Move to completed
    specs.move_spec_to_completed(spec_slug)

    # Verify it exists
    assert specs.get_spec(spec_slug) is not None

    # Delete it
    specs.delete_spec(spec_slug)

    # Verify it's gone
    assert specs.get_spec(spec_slug) is None
