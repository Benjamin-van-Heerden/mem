"""
Tests for the docs functionality.

Tests are split into:
- Unit tests for utility functions (no external APIs)
- Integration tests for indexing/search (require VOYAGE_AI_API_KEY)
"""

import json
import os
import tempfile
from pathlib import Path
from unittest.mock import patch

import pytest


@pytest.fixture
def temp_mem_dir(monkeypatch):
    """Create a temporary .mem directory structure for testing."""
    with tempfile.TemporaryDirectory() as tmpdir:
        tmpdir = Path(tmpdir)
        mem_dir = tmpdir / ".mem"
        mem_dir.mkdir()

        docs_dir = mem_dir / "docs"
        docs_dir.mkdir()
        (docs_dir / "summaries").mkdir()
        (docs_dir / "data").mkdir()

        config_file = mem_dir / "config.toml"
        config_file.write_text("""
[project]
name = "test_project"
""")

        from env_settings import EnvSettings

        mock_settings = EnvSettings()

        with patch.object(mock_settings, "caller_dir", tmpdir):
            with patch.object(mock_settings, "mem_dir", mem_dir):
                with patch.object(mock_settings, "config_file", config_file):
                    with patch("src.utils.docs.ENV_SETTINGS", mock_settings):
                        yield tmpdir


class TestDocsUtilities:
    """Unit tests for docs utility functions that don't need external APIs."""

    def test_get_doc_slug(self):
        """Test extracting slug from file path."""
        from src.utils.docs import get_doc_slug

        assert get_doc_slug(Path("/some/path/my_guide.md")) == "my_guide"
        assert get_doc_slug(Path("test.md")) == "test"
        assert get_doc_slug(Path("/a/b/c/api_reference.md")) == "api_reference"

    def test_compute_file_hash(self, tmp_path):
        """Test file hash computation."""
        from src.utils.docs import compute_file_hash

        test_file = tmp_path / "test.md"
        test_file.write_text("Hello, world!")

        hash1 = compute_file_hash(test_file)
        assert len(hash1) == 64  # SHA256 hex digest length

        hash2 = compute_file_hash(test_file)
        assert hash1 == hash2

        test_file.write_text("Different content")
        hash3 = compute_file_hash(test_file)
        assert hash1 != hash3

    def test_load_save_doc_hashes(self, tmp_path):
        """Test loading and saving document hashes."""
        from src.utils.docs import _get_hashes_file, load_doc_hashes, save_doc_hashes

        with patch("src.utils.docs._get_data_dir", return_value=tmp_path):
            with patch(
                "src.utils.docs._get_hashes_file",
                return_value=tmp_path / ".doc_hashes.json",
            ):
                with patch("src.utils.docs.ensure_docs_dirs"):
                    hashes = load_doc_hashes()
                    assert hashes == {}

                    test_hashes = {"doc1": "abc123", "doc2": "def456"}
                    save_doc_hashes(test_hashes)

                    loaded = load_doc_hashes()
                    assert loaded == test_hashes

    def test_list_doc_files(self, tmp_path):
        """Test listing document files."""
        from src.utils.docs import list_doc_files

        docs_dir = tmp_path / "docs"
        docs_dir.mkdir()
        (docs_dir / "guide1.md").write_text("# Guide 1")
        (docs_dir / "guide2.md").write_text("# Guide 2")
        (docs_dir / "not_a_doc.txt").write_text("ignored")
        (docs_dir / "summaries").mkdir()
        (docs_dir / "data").mkdir()

        with patch("src.utils.docs._get_docs_dir", return_value=docs_dir):
            files = list_doc_files()

            assert len(files) == 2
            slugs = [f.stem for f in files]
            assert "guide1" in slugs
            assert "guide2" in slugs

    def test_read_doc(self, tmp_path):
        """Test reading document content."""
        from src.utils.docs import read_doc

        docs_dir = tmp_path / "docs"
        docs_dir.mkdir()
        (docs_dir / "test_doc.md").write_text("# Test Document\n\nContent here.")

        with patch("src.utils.docs._get_docs_dir", return_value=docs_dir):
            content = read_doc("test_doc")
            assert content == "# Test Document\n\nContent here."

            missing = read_doc("nonexistent")
            assert missing is None

    def test_read_write_summary(self, tmp_path):
        """Test reading and writing summaries."""
        from src.utils.docs import read_summary, write_summary

        docs_dir = tmp_path / "docs"
        summaries_dir = docs_dir / "summaries"
        summaries_dir.mkdir(parents=True)

        with patch("src.utils.docs._get_docs_dir", return_value=docs_dir):
            with patch("src.utils.docs._get_summaries_dir", return_value=summaries_dir):
                with patch("src.utils.docs.ensure_docs_dirs"):
                    assert read_summary("test") is None

                    write_summary("test", "This is a test summary.")

                    content = read_summary("test")
                    assert content == "This is a test summary."

    def test_get_docs_needing_index(self, tmp_path):
        """Test determining which docs need indexing."""
        from src.utils.docs import compute_file_hash, get_docs_needing_index

        docs_dir = tmp_path / "docs"
        docs_dir.mkdir()
        data_dir = docs_dir / "data"
        data_dir.mkdir()

        doc1 = docs_dir / "new_doc.md"
        doc1.write_text("# New Document")

        doc2 = docs_dir / "existing_doc.md"
        doc2.write_text("# Existing Document")

        hashes = {
            "existing_doc": compute_file_hash(doc2),
            "deleted_doc": "oldhash123",
        }
        (data_dir / ".doc_hashes.json").write_text(json.dumps(hashes))

        with patch("src.utils.docs._get_docs_dir", return_value=docs_dir):
            with patch("src.utils.docs._get_data_dir", return_value=data_dir):
                with patch(
                    "src.utils.docs._get_hashes_file",
                    return_value=data_dir / ".doc_hashes.json",
                ):
                    new, changed, deleted = get_docs_needing_index()

                    assert "new_doc" in new
                    assert "existing_doc" not in new
                    assert "existing_doc" not in changed
                    assert "deleted_doc" in deleted

    def test_get_docs_needing_index_changed(self, tmp_path):
        """Test detecting changed documents."""
        from src.utils.docs import get_docs_needing_index

        docs_dir = tmp_path / "docs"
        docs_dir.mkdir()
        data_dir = docs_dir / "data"
        data_dir.mkdir()

        doc = docs_dir / "changed_doc.md"
        doc.write_text("# Changed Document - Updated")

        hashes = {"changed_doc": "old_hash_that_doesnt_match"}
        (data_dir / ".doc_hashes.json").write_text(json.dumps(hashes))

        with patch("src.utils.docs._get_docs_dir", return_value=docs_dir):
            with patch("src.utils.docs._get_data_dir", return_value=data_dir):
                with patch(
                    "src.utils.docs._get_hashes_file",
                    return_value=data_dir / ".doc_hashes.json",
                ):
                    new, changed, deleted = get_docs_needing_index()

                    assert "changed_doc" in changed
                    assert "changed_doc" not in new

    def test_get_indexed_docs(self, tmp_path):
        """Test getting indexed docs list."""
        from src.utils.docs import get_indexed_docs

        docs_dir = tmp_path / "docs"
        docs_dir.mkdir()
        summaries_dir = docs_dir / "summaries"
        summaries_dir.mkdir()
        data_dir = docs_dir / "data"
        data_dir.mkdir()

        (docs_dir / "indexed_doc.md").write_text("# Indexed")
        (docs_dir / "unindexed_doc.md").write_text("# Unindexed")
        (summaries_dir / "indexed_doc_summary.md").write_text("Summary")

        hashes = {"indexed_doc": "somehash"}
        (data_dir / ".doc_hashes.json").write_text(json.dumps(hashes))

        with patch("src.utils.docs._get_docs_dir", return_value=docs_dir):
            with patch("src.utils.docs._get_summaries_dir", return_value=summaries_dir):
                with patch("src.utils.docs._get_data_dir", return_value=data_dir):
                    with patch(
                        "src.utils.docs._get_hashes_file",
                        return_value=data_dir / ".doc_hashes.json",
                    ):
                        docs = get_indexed_docs()

                        assert len(docs) == 2

                        indexed = next(d for d in docs if d["slug"] == "indexed_doc")
                        assert indexed["indexed"] is True
                        assert indexed["has_summary"] is True

                        unindexed = next(
                            d for d in docs if d["slug"] == "unindexed_doc"
                        )
                        assert unindexed["indexed"] is False
                        assert unindexed["has_summary"] is False

    def test_check_docs_env_vars(self, monkeypatch):
        """Test environment variable checking."""
        from src.utils.docs import check_docs_env_vars

        monkeypatch.delenv("VOYAGE_AI_API_KEY", raising=False)
        monkeypatch.delenv("OPENROUTER_API_KEY", raising=False)

        ok, missing = check_docs_env_vars()
        assert ok is False
        assert "VOYAGE_AI_API_KEY" in missing
        assert "OPENROUTER_API_KEY" in missing

        monkeypatch.setenv("VOYAGE_AI_API_KEY", "test_key")
        monkeypatch.setenv("OPENROUTER_API_KEY", "test_key")

        ok, missing = check_docs_env_vars()
        assert ok is True
        assert missing == []


class TestDocsChunking:
    """Tests for document chunking."""

    def test_chunk_document(self):
        """Test chunking a markdown document."""
        from src.utils.docs import chunk_document

        content = """# Main Title

Introduction paragraph.

## Section One

Content for section one.

## Section Two

Content for section two.
"""
        chunks = chunk_document("test_doc", content)

        assert len(chunks) > 0
        for chunk in chunks:
            assert chunk.content
            assert chunk.meta_data.get("doc_slug") == "test_doc"


class TestDocsDelete:
    """Tests for document deletion."""

    def test_delete_doc(self, tmp_path):
        """Test deleting a document and its associated files."""
        from src.utils.docs import delete_doc

        docs_dir = tmp_path / "docs"
        docs_dir.mkdir()
        summaries_dir = docs_dir / "summaries"
        summaries_dir.mkdir()
        data_dir = docs_dir / "data"
        data_dir.mkdir()

        doc_file = docs_dir / "to_delete.md"
        doc_file.write_text("# To Delete")
        summary_file = summaries_dir / "to_delete_summary.md"
        summary_file.write_text("Summary")

        hashes = {"to_delete": "somehash"}
        hashes_file = data_dir / ".doc_hashes.json"
        hashes_file.write_text(json.dumps(hashes))

        with patch("src.utils.docs._get_docs_dir", return_value=docs_dir):
            with patch("src.utils.docs._get_summaries_dir", return_value=summaries_dir):
                with patch("src.utils.docs._get_data_dir", return_value=data_dir):
                    with patch(
                        "src.utils.docs._get_hashes_file", return_value=hashes_file
                    ):
                        with patch("src.utils.docs.ensure_docs_dirs"):
                            with patch(
                                "src.utils.docs.delete_doc_from_index", return_value=0
                            ):
                                result = delete_doc("to_delete")

                                assert result is True
                                assert not doc_file.exists()
                                assert not summary_file.exists()

                                loaded_hashes = json.loads(hashes_file.read_text())
                                assert "to_delete" not in loaded_hashes

    def test_delete_nonexistent_doc(self, tmp_path):
        """Test deleting a document that doesn't exist."""
        from src.utils.docs import delete_doc

        docs_dir = tmp_path / "docs"
        docs_dir.mkdir()

        with patch("src.utils.docs._get_docs_dir", return_value=docs_dir):
            result = delete_doc("nonexistent")
            assert result is False


@pytest.mark.skipif(
    not os.getenv("VOYAGE_AI_API_KEY"), reason="VOYAGE_AI_API_KEY not set"
)
class TestDocsIndexingIntegration:
    """Integration tests that require VOYAGE_AI_API_KEY."""

    def test_index_and_search_document(self, tmp_path, monkeypatch):
        """Test indexing a document and searching it."""
        from src.utils.docs import (
            delete_doc_from_index,
            get_chroma_client,
            index_document,
            search_docs,
        )

        docs_dir = tmp_path / "docs"
        docs_dir.mkdir()
        data_dir = docs_dir / "data"
        data_dir.mkdir()
        chroma_dir = data_dir / "chroma"
        chroma_dir.mkdir()

        config_file = tmp_path / ".mem" / "config.toml"
        config_file.parent.mkdir(parents=True)
        config_file.write_text('[project]\nname = "test_integration"')

        with patch("src.utils.docs._get_docs_dir", return_value=docs_dir):
            with patch("src.utils.docs._get_data_dir", return_value=data_dir):
                with patch("src.utils.docs._get_chroma_dir", return_value=chroma_dir):
                    with patch(
                        "src.utils.docs._read_config",
                        return_value={"project": {"name": "test_integration"}},
                    ):
                        with patch("src.utils.docs.ensure_docs_dirs"):
                            content = """# Python Guide

This guide covers Python programming basics.

## Variables

Python variables are dynamically typed.

## Functions

Define functions using the def keyword.
"""
                            chunk_count = index_document("python_guide", content)
                            assert chunk_count > 0

                            results = search_docs("Python variables", n_results=3)
                            assert len(results) > 0
                            assert any(
                                "variable" in r["content"].lower() for r in results
                            )

                            deleted = delete_doc_from_index("python_guide")
                            assert deleted > 0
