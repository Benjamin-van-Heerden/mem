"""
Documentation utilities for indexing, searching, and managing technical docs.

Documents are stored as markdown files in .mem/docs/ and indexed into ChromaDB
for semantic search. AI-generated summaries are stored in .mem/docs/summaries/.
"""

import hashlib
import json
import os
import tomllib
from pathlib import Path
from typing import TYPE_CHECKING

from env_settings import ENV_SETTINGS

if TYPE_CHECKING:
    import chromadb
    from agno.knowledge.document import Document


def _read_config() -> dict:
    """Read local config file. Simplified version to avoid circular imports."""
    config_file = ENV_SETTINGS.config_file
    if not config_file.exists():
        return {}
    try:
        with open(config_file, "rb") as f:
            return tomllib.load(f)
    except Exception:
        return {}


def _get_docs_dir() -> Path:
    """Get the docs directory path."""
    return ENV_SETTINGS.mem_dir / "docs"


def _get_summaries_dir() -> Path:
    """Get the summaries directory path."""
    return _get_docs_dir() / "summaries"


def _get_data_dir() -> Path:
    """Get the data directory path (gitignored)."""
    return _get_docs_dir() / "data"


def _get_core_docs_dir() -> Path:
    """Get the core docs directory path."""
    return _get_docs_dir() / "core"


def _get_chroma_dir() -> Path:
    """Get the ChromaDB storage directory."""
    return _get_data_dir() / "chroma"


def _get_hashes_file() -> Path:
    """Get the path to the document hashes file."""
    return _get_data_dir() / ".doc_hashes.json"


def ensure_docs_dirs() -> None:
    """Ensure all docs directories exist."""
    _get_docs_dir().mkdir(parents=True, exist_ok=True)
    _get_summaries_dir().mkdir(parents=True, exist_ok=True)
    _get_data_dir().mkdir(parents=True, exist_ok=True)
    _get_chroma_dir().mkdir(parents=True, exist_ok=True)
    _get_core_docs_dir().mkdir(parents=True, exist_ok=True)


def get_doc_slug(file_path: Path) -> str:
    """Extract slug from document file path (filename without .md extension)."""
    return file_path.stem


def compute_file_hash(file_path: Path) -> str:
    """Compute SHA256 hash of file contents."""
    content = file_path.read_text()
    return hashlib.sha256(content.encode()).hexdigest()


def load_doc_hashes() -> dict[str, str]:
    """Load stored document hashes from JSON file."""
    hashes_file = _get_hashes_file()
    if not hashes_file.exists():
        return {}
    try:
        return json.loads(hashes_file.read_text())
    except (json.JSONDecodeError, OSError):
        return {}


def save_doc_hashes(hashes: dict[str, str]) -> None:
    """Save document hashes to JSON file."""
    ensure_docs_dirs()
    _get_hashes_file().write_text(json.dumps(hashes, indent=2))


def list_doc_files() -> list[Path]:
    """List all markdown files in docs directory (excluding core/, summaries/, and data/)."""
    docs_dir = _get_docs_dir()
    if not docs_dir.exists():
        return []

    doc_files = []
    for file_path in docs_dir.iterdir():
        if file_path.is_file() and file_path.suffix == ".md":
            doc_files.append(file_path)

    return sorted(doc_files, key=lambda p: p.name)


def list_core_doc_files() -> list[Path]:
    """List all markdown files in the core docs directory."""
    core_dir = _get_core_docs_dir()
    if not core_dir.exists():
        return []

    doc_files = []
    for file_path in core_dir.iterdir():
        if file_path.is_file() and file_path.suffix == ".md":
            doc_files.append(file_path)

    return sorted(doc_files, key=lambda p: p.name)


def get_core_doc_slug(file_path: Path) -> str:
    """Extract slug from core document file path (filename without .md extension)."""
    return file_path.stem


def get_core_doc_path(slug: str) -> Path:
    """Get path to a core document file by slug."""
    return _get_core_docs_dir() / f"{slug}.md"


def read_core_doc(slug: str) -> str | None:
    """Read core document content by slug. Returns None if not found."""
    doc_path = get_core_doc_path(slug)
    if not doc_path.exists():
        return None
    return doc_path.read_text()


def get_doc_path(slug: str) -> Path:
    """Get path to a document file by slug."""
    return _get_docs_dir() / f"{slug}.md"


def get_summary_path(slug: str) -> Path:
    """Get path to a document's summary file."""
    return _get_summaries_dir() / f"{slug}_summary.md"


def read_doc(slug: str) -> str | None:
    """Read document content by slug. Returns None if not found."""
    doc_path = get_doc_path(slug)
    if not doc_path.exists():
        return None
    return doc_path.read_text()


def read_summary(slug: str) -> str | None:
    """Read summary content by slug. Returns None if not found."""
    summary_path = get_summary_path(slug)
    if not summary_path.exists():
        return None
    return summary_path.read_text()


def write_summary(slug: str, content: str) -> None:
    """Write summary content for a document."""
    ensure_docs_dirs()
    summary_path = get_summary_path(slug)
    summary_path.write_text(content)


def delete_doc(slug: str) -> bool:
    """Delete a document and its associated files.

    Removes:
    - The document file (.mem/docs/{slug}.md)
    - The summary file (.mem/docs/summaries/{slug}_summary.md)
    - All chunks from ChromaDB
    - The hash entry from .doc_hashes.json

    Returns True if document existed and was deleted, False otherwise.
    """
    doc_path = get_doc_path(slug)
    if not doc_path.exists():
        return False

    doc_path.unlink()

    summary_path = get_summary_path(slug)
    if summary_path.exists():
        summary_path.unlink()

    try:
        delete_doc_from_index(slug)
    except Exception:
        pass

    hashes = load_doc_hashes()
    if slug in hashes:
        del hashes[slug]
        save_doc_hashes(hashes)

    return True


def _get_collection_name() -> str:
    """Get ChromaDB collection name based on project name."""
    config = _read_config()
    project = config.get("project", {})
    project_name = project.get("name", "default")
    safe_name = "".join(c if c.isalnum() or c == "_" else "_" for c in project_name)
    return f"{safe_name}_docs"


def _get_embedding_function():
    """Create VoyageAI embedding function."""
    from chromadb.utils.embedding_functions import VoyageAIEmbeddingFunction

    api_key = os.getenv("VOYAGE_AI_API_KEY")
    if not api_key:
        raise ValueError("VOYAGE_AI_API_KEY environment variable is required")
    return VoyageAIEmbeddingFunction(
        api_key=api_key,
        model_name="voyage-3-large",
    )


def get_chroma_client() -> "chromadb.ClientAPI":
    """Get ChromaDB persistent client."""
    import chromadb

    ensure_docs_dirs()
    return chromadb.PersistentClient(path=str(_get_chroma_dir()))


def get_collection() -> "chromadb.Collection":
    """Get or create the docs ChromaDB collection."""
    client = get_chroma_client()
    embedding_fn = _get_embedding_function()
    return client.get_or_create_collection(
        name=_get_collection_name(),
        embedding_function=embedding_fn,  # type: ignore
    )


def chunk_document(slug: str, content: str) -> list["Document"]:
    """Chunk a document using MarkdownChunking.

    Returns list of Document objects with metadata.
    """
    from agno.knowledge.chunking.markdown import MarkdownChunking
    from agno.knowledge.document import Document

    doc = Document(
        content=content,
        id=slug,
        name=slug,
        meta_data={"doc_slug": slug},
    )

    chunker = MarkdownChunking(
        chunk_size=5000,
        overlap=200,
        split_on_headings=2,
    )

    return chunker.chunk(doc)


def index_document(slug: str, content: str) -> int:
    """Index a document into ChromaDB.

    Chunks the document and upserts all chunks.
    Returns the number of chunks indexed.
    """
    chunks = chunk_document(slug, content)
    if not chunks:
        return 0

    collection = get_collection()

    ids = []
    documents = []
    metadatas = []

    for i, chunk in enumerate(chunks):
        chunk_id = f"{slug}_{i}"
        ids.append(chunk_id)
        documents.append(chunk.content)
        metadata = {
            "doc_slug": slug,
            "chunk_index": i,
        }
        if chunk.name:
            metadata["heading"] = chunk.name
        metadatas.append(metadata)

    collection.upsert(
        ids=ids,
        documents=documents,
        metadatas=metadatas,
    )

    return len(chunks)


def delete_doc_from_index(slug: str) -> int:
    """Delete all chunks for a document from ChromaDB.

    Returns the number of chunks deleted.
    """
    collection = get_collection()

    results = collection.get(
        where={"doc_slug": slug},
        include=[],
    )

    if not results["ids"]:
        return 0

    collection.delete(ids=results["ids"])
    return len(results["ids"])


def search_docs(
    query: str,
    doc_slug: str | None = None,
    n_results: int = 10,
) -> list[dict]:
    """Search documents using semantic similarity.

    Args:
        query: Search query string
        doc_slug: Optional filter to search within a specific document
        n_results: Maximum number of results to return

    Returns:
        List of dicts with keys: content, doc_slug, chunk_index, heading, distance
    """
    collection = get_collection()

    where_filter = None
    if doc_slug:
        where_filter = {"doc_slug": doc_slug}

    results = collection.query(
        query_texts=[query],
        n_results=n_results,
        where=where_filter,  # type: ignore
        include=["documents", "metadatas", "distances"],
    )

    search_results = []
    if results["ids"] and results["ids"][0]:
        for i, chunk_id in enumerate(results["ids"][0]):
            result = {
                "content": results["documents"][0][i] if results["documents"] else "",
                "doc_slug": results["metadatas"][0][i].get("doc_slug", "")
                if results["metadatas"]
                else "",
                "chunk_index": results["metadatas"][0][i].get("chunk_index", 0)
                if results["metadatas"]
                else 0,
                "heading": results["metadatas"][0][i].get("heading", "")
                if results["metadatas"]
                else "",
                "distance": results["distances"][0][i] if results["distances"] else 0.0,
            }
            search_results.append(result)

    return search_results


def get_docs_needing_index() -> tuple[list[str], list[str], list[str]]:
    """Determine which documents need indexing, updating, or removal.

    Returns:
        (new_slugs, changed_slugs, deleted_slugs)
        - new_slugs: Documents that exist but have never been indexed
        - changed_slugs: Documents that have changed since last index
        - deleted_slugs: Documents that were indexed but no longer exist
    """
    stored_hashes = load_doc_hashes()
    current_files = list_doc_files()

    current_slugs = {get_doc_slug(f) for f in current_files}
    stored_slugs = set(stored_hashes.keys())

    new_slugs = []
    changed_slugs = []
    deleted_slugs = list(stored_slugs - current_slugs)

    for file_path in current_files:
        slug = get_doc_slug(file_path)
        current_hash = compute_file_hash(file_path)

        if slug not in stored_hashes:
            new_slugs.append(slug)
        elif stored_hashes[slug] != current_hash:
            changed_slugs.append(slug)

    return new_slugs, changed_slugs, deleted_slugs


def get_indexed_docs() -> list[dict]:
    """Get list of all documents with their index status.

    Returns list of dicts with keys: slug, path, indexed, has_summary
    """
    doc_files = list_doc_files()
    stored_hashes = load_doc_hashes()

    docs = []
    for file_path in doc_files:
        slug = get_doc_slug(file_path)
        docs.append(
            {
                "slug": slug,
                "path": file_path,
                "indexed": slug in stored_hashes,
                "has_summary": get_summary_path(slug).exists(),
            }
        )

    return docs


def check_docs_env_vars() -> tuple[bool, list[str]]:
    """Check if required environment variables are set.

    Returns:
        (all_present, missing_vars)
    """
    required = ["VOYAGE_AI_API_KEY", "OPENROUTER_API_KEY"]
    missing = [var for var in required if not os.getenv(var)]
    return len(missing) == 0, missing
