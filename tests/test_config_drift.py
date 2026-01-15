"""
Tests for config drift detection and patching.
"""

import tempfile
from pathlib import Path

from src.config.main_config import (
    find_unknown_key_paths,
    generate_default_config_toml,
    has_unknown_key_drift,
    load_and_validate_local_config,
)
from src.config.models import MemLocalConfig


class TestDriftDetection:
    """Tests for drift detection logic."""

    def test_clean_config_no_drift(self):
        """A valid config with no unknown keys should not trigger drift."""
        config = """
[project]
name = "test"
description = "test desc"
generic_templates = ["python"]

[[files]]
path = "README.md"

[worktree]
symlink_paths = []
"""
        with tempfile.NamedTemporaryFile(mode="w", suffix=".toml", delete=False) as f:
            f.write(config)
            path = Path(f.name)

        try:
            result = load_and_validate_local_config(path)
            assert result.config is not None
            assert result.validation_error is None
            assert not has_unknown_key_drift(result.raw)
            assert find_unknown_key_paths(result.raw, MemLocalConfig) == []
        finally:
            path.unlink()

    def test_unknown_top_level_section(self):
        """Unknown top-level sections should be detected."""
        config = """
[project]
name = "test"
description = "test desc"

[vars]
github_token_env = "GITHUB_TOKEN"

[unknown_section]
foo = "bar"
"""
        with tempfile.NamedTemporaryFile(mode="w", suffix=".toml", delete=False) as f:
            f.write(config)
            path = Path(f.name)

        try:
            result = load_and_validate_local_config(path)
            assert result.config is not None  # Still validates (extra="ignore")
            assert has_unknown_key_drift(result.raw)
            unknown = find_unknown_key_paths(result.raw, MemLocalConfig)
            assert "vars" in unknown
            assert "unknown_section" in unknown
        finally:
            path.unlink()

    def test_unknown_nested_key(self):
        """Unknown keys nested in known sections should be detected."""
        config = """
[project]
name = "test"
description = "test desc"
unknown_project_key = "bad"
"""
        with tempfile.NamedTemporaryFile(mode="w", suffix=".toml", delete=False) as f:
            f.write(config)
            path = Path(f.name)

        try:
            result = load_and_validate_local_config(path)
            assert has_unknown_key_drift(result.raw)
            unknown = find_unknown_key_paths(result.raw, MemLocalConfig)
            assert "project.unknown_project_key" in unknown
        finally:
            path.unlink()

    def test_unknown_key_in_list_item(self):
        """Unknown keys in list items (like files) should be detected."""
        config = """
[project]
name = "test"
description = "test desc"

[[files]]
path = "README.md"
extra_field = true
"""
        with tempfile.NamedTemporaryFile(mode="w", suffix=".toml", delete=False) as f:
            f.write(config)
            path = Path(f.name)

        try:
            result = load_and_validate_local_config(path)
            assert has_unknown_key_drift(result.raw)
            unknown = find_unknown_key_paths(result.raw, MemLocalConfig)
            assert "files[0].extra_field" in unknown
        finally:
            path.unlink()

    def test_validation_error_missing_required(self):
        """Missing required fields should cause validation error."""
        config = """
[project]
name = "test"
# missing description
"""
        with tempfile.NamedTemporaryFile(mode="w", suffix=".toml", delete=False) as f:
            f.write(config)
            path = Path(f.name)

        try:
            result = load_and_validate_local_config(path)
            assert result.config is None
            assert result.validation_error is not None
        finally:
            path.unlink()


class TestGenerateConfig:
    """Tests for config generation."""

    def test_generated_config_validates(self):
        """Generated config should pass validation."""
        config_str = generate_default_config_toml(project_name="test-project")

        with tempfile.NamedTemporaryFile(mode="w", suffix=".toml", delete=False) as f:
            f.write(config_str)
            path = Path(f.name)

        try:
            result = load_and_validate_local_config(path)
            assert result.config is not None
            assert result.validation_error is None
            assert not has_unknown_key_drift(result.raw)
        finally:
            path.unlink()

    def test_generated_config_preserves_values(self):
        """Generated config should use provided values."""
        config_str = generate_default_config_toml(
            project_name="my-project",
            project_description="My description",
            generic_templates=["rust", "go"],
            important_files=[{"path": "main.rs", "description": "Entry point"}],
            symlink_paths=[".cache"],
        )

        with tempfile.NamedTemporaryFile(mode="w", suffix=".toml", delete=False) as f:
            f.write(config_str)
            path = Path(f.name)

        try:
            result = load_and_validate_local_config(path)
            assert result.config is not None
            assert result.config.project.name == "my-project"
            assert result.config.project.description == "My description\n"
            assert result.config.project.generic_templates == ["rust", "go"]
            assert len(result.config.files) == 1
            assert result.config.files[0].path == "main.rs"
            assert result.config.worktree.symlink_paths == [".cache"]
        finally:
            path.unlink()


class TestPatchConfig:
    """Tests for the patch config command logic."""

    def test_patch_removes_unknown_keys(self):
        """Patching should remove unknown keys."""
        from src.commands.patch import _extract_known_values

        raw = {
            "project": {
                "name": "test",
                "description": "desc",
                "unknown_key": "bad",
            },
            "vars": {"foo": "bar"},
            "files": [],
        }

        known = _extract_known_values(raw)
        assert "project.name" in known
        assert "project.description" in known
        assert known["project.name"] == "test"
        # unknown_key should not be in known values
        assert "project.unknown_key" not in known
        assert "vars" not in known

    def test_patch_preserves_files(self):
        """Patching should preserve valid file entries."""
        from src.commands.patch import _extract_known_values, _filter_valid_files

        raw = {
            "project": {"name": "test", "description": "desc"},
            "files": [
                {"path": "README.md", "description": "Readme"},
                {"path": "main.py", "extra": "ignored"},
            ],
        }

        known = _extract_known_values(raw)
        files = _filter_valid_files(known.get("files", []))

        assert len(files) == 2
        assert files[0]["path"] == "README.md"
        assert files[0]["description"] == "Readme"
        assert files[1]["path"] == "main.py"
        assert "extra" not in files[1]

    def test_patch_idempotent(self):
        """Running patch twice should produce identical output."""
        config_str = generate_default_config_toml(project_name="test")

        with tempfile.NamedTemporaryFile(mode="w", suffix=".toml", delete=False) as f:
            f.write(config_str)
            path = Path(f.name)

        try:
            # Load and check - should have no drift
            result = load_and_validate_local_config(path)
            assert result.config is not None
            assert not has_unknown_key_drift(result.raw)

            # Generate again with same values
            config_str2 = generate_default_config_toml(
                project_name=result.config.project.name,
                project_description=result.config.project.description.strip(),
                generic_templates=result.config.project.generic_templates,
                important_files=[
                    {"path": f.path, "description": f.description}  # type: ignore
                    for f in result.config.files
                ],
                symlink_paths=result.config.worktree.symlink_paths,
            )

            # Write second version
            path.write_text(config_str2)

            # Should still have no drift
            result2 = load_and_validate_local_config(path)
            assert result2.config is not None
            assert not has_unknown_key_drift(result2.raw)
        finally:
            path.unlink()
