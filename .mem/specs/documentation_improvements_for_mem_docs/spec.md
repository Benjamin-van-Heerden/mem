---
title: Documentation improvements for mem docs
status: todo
assigned_to: null
issue_id: 47
issue_url: https://github.com/Benjamin-van-Heerden/mem/issues/47
branch: null
pr_url: null
created_at: '2026-01-13T18:06:57.195381'
updated_at: '2026-01-13T18:08:21.367842'
completed_at: null
last_synced_at: '2026-01-13T18:08:21.366833'
local_content_hash: 0f18187d35b62e5815de7eb631cc1432c891bc688deaf0c44806f30e0dacdda1
remote_content_hash: 0f18187d35b62e5815de7eb631cc1432c891bc688deaf0c44806f30e0dacdda1
---
## Overview

Improve the `mem docs` feature discoverability and make the indexing process more efficient by:
1. Adding usage hints about `mem docs` commands in the onboard output
2. Skipping summary generation during `mem docs index` if a summary already exists
3. Ensuring `.mem/docs/data` is added to `.gitignore` during `mem init`

## Goals

- Make users aware of the `mem docs` search functionality through `mem onboard`
- Avoid redundant AI summary generation when re-indexing documents that already have summaries
- Ensure the vector DB data directory is properly gitignored in new projects

## Technical Approach

### 1. Add `mem docs` usage hints in onboard command

In `src/commands/onboard.py`, add a new section after the "About mem" section that describes the `mem docs` commands:

```python
output.append("**Document search:**")
output.append("- `mem docs search \"query\"` - Semantic search across indexed documentation")
output.append("- `mem docs list` - List all documents and their index status")
output.append("- `mem docs index` - Index new or changed documents")
output.append("")
```

This should be added to the "Key commands" block in the About mem section (around line 290-310).

### 2. Skip summary generation if summary already exists

In `src/commands/docs.py`, in the `index()` command (around line 75), check if a summary already exists before generating a new one:

```python
# Only generate summary if it doesn't exist
summary_path = docs.get_summary_path(slug)
if not summary_path.exists():
    typer.echo("    🤖 Generating summary...")
    # ... existing summary generation code ...
else:
    typer.echo("    ✅ Summary already exists (skipped)")
```

### 3. Ensure .mem/docs/data in gitignore

The `.gitignore` handling in `src/commands/init.py` (lines 320-341) already adds `.mem/docs/data/` to the gitignore. This is already implemented - verify it works correctly.

## Success Criteria

- Running `mem onboard` shows usage hints for `mem docs search`, `mem docs list`, and `mem docs index` in the "About mem" section
- Running `mem docs index` on a document with an existing summary skips the summary generation step and shows "Summary already exists (skipped)"
- The `.mem/docs/data/` entry is properly added to `.gitignore` during `mem init` (already working)

## Notes

- The third item (gitignore) is already implemented in init.py. The task for this is just to verify it works.
- The summary skip logic should only apply to documents that already have summaries - new documents or documents without summaries should still generate summaries.
