"""
Docs command - Manage technical documentation
"""

from typing import Optional

import typer
from typing_extensions import Annotated

from src.utils import docs
from src.utils.ai.doc_summarizer import summarize_document

app = typer.Typer(help="Manage technical documentation")


@app.command()
def index():
    """
    Index documents into the vector store.

    Scans .mem/docs/ for markdown files and:
    - Indexes new documents (not yet in the hash file)
    - Re-indexes changed documents (hash mismatch)
    - Removes orphaned entries (document deleted)
    - Generates AI summaries for new/changed documents
    """
    all_present, missing = docs.check_docs_env_vars()
    if not all_present:
        typer.echo(f"❌ Missing required environment variables: {', '.join(missing)}")
        typer.echo("\nSet these variables to use document indexing.")
        raise typer.Exit(code=1)

    docs.ensure_docs_dirs()

    new_slugs, changed_slugs, deleted_slugs = docs.get_docs_needing_index()

    if not new_slugs and not changed_slugs and not deleted_slugs:
        typer.echo("✅ All documents are up to date. Nothing to index.")
        return

    hashes = docs.load_doc_hashes()

    if deleted_slugs:
        typer.echo(f"\n🗑️  Removing {len(deleted_slugs)} deleted document(s)...")
        for slug in deleted_slugs:
            typer.echo(f"  - {slug}")
            try:
                docs.delete_doc_from_index(slug)
            except Exception as e:
                typer.echo(f"    ⚠️  Warning: Could not remove from index: {e}")
            if slug in hashes:
                del hashes[slug]
            summary_path = docs.get_summary_path(slug)
            if summary_path.exists():
                summary_path.unlink()

    to_index = new_slugs + changed_slugs
    if to_index:
        typer.echo(f"\n📚 Indexing {len(to_index)} document(s)...")
        for slug in to_index:
            action = "new" if slug in new_slugs else "changed"
            typer.echo(f"  - {slug} ({action})")

            doc_path = docs.get_doc_path(slug)
            content = doc_path.read_text()

            try:
                chunk_count = docs.index_document(slug, content)
                typer.echo(f"    ✅ Indexed {chunk_count} chunk(s)")
            except Exception as e:
                typer.echo(f"    ❌ Failed to index: {e}")
                continue

            typer.echo(f"    🤖 Generating summary...")
            try:
                summary = summarize_document(content, slug)
                if summary:
                    docs.write_summary(slug, summary)
                    typer.echo(f"    ✅ Summary generated")
                else:
                    typer.echo(f"    ⚠️  Could not generate summary")
            except Exception as e:
                typer.echo(f"    ⚠️  Summary generation failed: {e}")

            hashes[slug] = docs.compute_file_hash(doc_path)

    docs.save_doc_hashes(hashes)

    typer.echo(f"\n✅ Indexing complete.")
    typer.echo(
        f"   New: {len(new_slugs)}, Updated: {len(changed_slugs)}, Removed: {len(deleted_slugs)}"
    )


@app.command("list")
def list_docs():
    """
    List all documents with their index status.

    Shows core documents (always included in onboard) and indexed documents
    (require indexing, summaries shown in onboard).
    """
    core_doc_files = docs.list_core_doc_files()
    indexed_docs = docs.get_indexed_docs()

    if not core_doc_files and not indexed_docs:
        typer.echo("No documents found in .mem/docs/")
        typer.echo("\n💡 Add markdown files to:")
        typer.echo("   - .mem/docs/core/ for core docs (auto-included in full)")
        typer.echo("   - .mem/docs/ for indexed docs (run 'mem docs index')")
        return

    if core_doc_files:
        typer.echo("\n📚 CORE DOCUMENTS (auto-included in onboard):\n")
        typer.echo(f"{'Slug':<30} {'Type':<10}")
        typer.echo("=" * 40)

        for file_path in core_doc_files:
            slug = docs.get_core_doc_slug(file_path)
            typer.echo(f"{slug:<30} {'core':<10}")

        typer.echo(f"\n📊 Total: {len(core_doc_files)} core document(s)")

    if indexed_docs:
        typer.echo("\n📖 INDEXED DOCUMENTS (summaries in onboard):\n")
        typer.echo(f"{'Slug':<30} {'Indexed':<10} {'Summary':<10}")
        typer.echo("=" * 50)

        for doc in indexed_docs:
            slug = doc["slug"]
            indexed = "✅" if doc["indexed"] else "❌"
            has_summary = "✅" if doc["has_summary"] else "❌"
            typer.echo(f"{slug:<30} {indexed:<10} {has_summary:<10}")

        typer.echo(f"\n📊 Total: {len(indexed_docs)} indexed document(s)")

        unindexed = [d for d in indexed_docs if not d["indexed"]]
        if unindexed:
            typer.echo(
                f"\n⚠️  {len(unindexed)} document(s) need indexing. Run 'mem docs index'"
            )


@app.command()
def read(
    slug: Annotated[str, typer.Argument(help="Document slug (filename without .md)")],
):
    """
    Read and display a document's full content.
    """
    content = docs.read_doc(slug)
    if content is None:
        typer.echo(f"❌ Document '{slug}' not found.")
        typer.echo("\n💡 Available documents:")
        for doc in docs.get_indexed_docs():
            typer.echo(f"  - {doc['slug']}")
        raise typer.Exit(code=1)

    typer.echo(f"\n📄 {slug}.md\n")
    typer.echo("=" * 60)
    typer.echo(content)


@app.command()
def search(
    query: Annotated[str, typer.Argument(help="Search query")],
    doc: Annotated[
        Optional[str],
        typer.Option("--doc", "-d", help="Filter to a specific document slug"),
    ] = None,
    limit: Annotated[
        int,
        typer.Option("--limit", "-n", help="Maximum number of results"),
    ] = 5,
):
    """
    Search documents using semantic similarity.

    Returns the most relevant chunks from indexed documents.
    """
    all_present, missing = docs.check_docs_env_vars()
    if not all_present:
        typer.echo(f"❌ Missing required environment variables: {', '.join(missing)}")
        raise typer.Exit(code=1)

    try:
        results = docs.search_docs(query, doc_slug=doc, n_results=limit)
    except Exception as e:
        typer.echo(f"❌ Search failed: {e}")
        raise typer.Exit(code=1)

    if not results:
        typer.echo("No results found.")
        if doc:
            typer.echo(
                f"\n💡 Try searching without --doc filter, or check if '{doc}' is indexed."
            )
        return

    typer.echo(f'\n🔍 Search results for: "{query}"\n')

    for i, result in enumerate(results, 1):
        typer.echo(f"─── Result {i} ───")
        typer.echo(f"📄 Document: {result['doc_slug']}")
        if result.get("heading"):
            typer.echo(f"📑 Section: {result['heading']}")
        typer.echo(f"📊 Relevance: {1 - result['distance']:.2%}")
        typer.echo("")
        content_preview = result["content"][:500]
        if len(result["content"]) > 500:
            content_preview += "..."
        typer.echo(content_preview)
        typer.echo("")

    typer.echo(f"📊 Showing {len(results)} result(s)")


@app.command()
def delete(
    slug: Annotated[str, typer.Argument(help="Document slug to delete")],
    force: Annotated[
        bool,
        typer.Option("--force", "-f", help="Skip confirmation prompt"),
    ] = False,
):
    """
    Delete a document and its associated files.

    Removes the document, its summary, and all index entries.
    """
    if not docs.get_doc_path(slug).exists():
        typer.echo(f"❌ Document '{slug}' not found.")
        raise typer.Exit(code=1)

    if not force:
        typer.confirm(
            f"Delete document '{slug}' and all associated data?",
            abort=True,
        )

    if docs.delete_doc(slug):
        typer.echo(f"✅ Deleted document: {slug}")
        typer.echo("   Removed: document file, summary, and index entries")
    else:
        typer.echo(f"❌ Failed to delete document '{slug}'")
        raise typer.Exit(code=1)


if __name__ == "__main__":
    app()
