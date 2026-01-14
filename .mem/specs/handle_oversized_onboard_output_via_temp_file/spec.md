---
title: Handle oversized onboard output via temp file
status: todo
assigned_to: Benjamin-van-Heerden
issue_id: 51
issue_url: https://github.com/Benjamin-van-Heerden/mem/issues/51
branch: dev-benjamin_van_heerden-handle_oversized_onboard_output_via_temp_file
pr_url: null
created_at: '2026-01-14T14:11:16.815267'
updated_at: '2026-01-14T14:21:22.144398'
completed_at: null
last_synced_at: '2026-01-14T14:21:16.356144'
local_content_hash: 6a86b8561a5f49ffb77365164168520188072f177ea0a384a6bbfbb18069a566
remote_content_hash: 6a86b8561a5f49ffb77365164168520188072f177ea0a384a6bbfbb18069a566
---
## Overview

Some editors/agents automatically truncate large command output. `mem onboard` is specifically used to deliver context to an agent, so truncation defeats the purpose and can cause missed instructions/specs.

To make this robust, `mem onboard` should write the **project context** (specs/tasks/work logs/git info/etc.) to an OS temp file **by default**, and only print a small, non-truncated “how to proceed” message to stdout.

The stdout message must use an imperative instruction (focus on what to do, not why). Example: “READ THIS ENTIRE FILE”.

## Goals

- Prevent onboard context loss due to output truncation by common agents/editors.
- Make the default `mem onboard` behavior safe for agents and terminals regardless of context size.
- Keep stdout output short and actionable (never prints the long project context).
- Use an OS-managed temp location so files are cleaned up over time by the OS.

## Technical Approach

### High-level behavior

1. Generate onboard content as usual, but split it into:
   - **Minimal stdout section**: short “About mem” + key commands + a pointer to the temp file.
   - **Project context section**: everything that tends to be large (project info, important files, docs, specs, worktree info, work logs, todos, next steps, agent instruction, sync failure block, etc.)
2. Always write the **project context section** to a temp file.
3. Always print only the minimal stdout section, including:
   - The exact path to the temp file
   - A command to read it, e.g. `cat <path>`
   - A strong imperative instruction such as: `READ THIS ENTIRE FILE.`
4. Add a `--stdout` flag that forces printing the full project context to stdout instead of writing it to a temp file.
5. If writing the temp file fails (and `--stdout` is not set):
   - Print an explicit error message and the intended temp file path (if available)
   - Exit non-zero

### Temp file location and naming

- Use the OS temp directory (via Python’s temp file utilities).
- Filename should be recognizable and unique, e.g. `mem_onboard_<timestamp>_<random>.md`.
- File extension should be `.md` to preserve formatting.
- The file should contain the full project context in the same formatting that would have been printed previously.

### Cleanup policy

- Do not implement explicit cleanup. Rely on OS temp cleanup policies.
- If desired later, add an optional flag to write to a stable location for debugging.

### CLI / UX details

- Stdout should remain small (aim < 5 KB).
- The pointer message should be highly actionable and copy-pasteable.
- Flags:
  - `--stdout`: print the full project context to stdout (expert mode; may be truncated by some environments).
  - (Optional, later) `--path <file>`: write to a specific file path instead of OS temp.

### Errors and edge cases

- If writing the temp file fails:
  - Print an explicit error message and instructions.
  - Exit with non-zero status.
  - Optionally include a minimal fallback summary (not the full context) to avoid truncation.
- Ensure Windows compatibility (temp paths + file permissions).

## Success Criteria

- Running `mem onboard` always prints a short message with the temp file path and an explicit instruction to read the entire file.
- The temp file contains the full project context with no truncation.
- The stdout output contains only minimal mem info (and never dumps the full project context).
- Works across macOS/Linux/Windows.

## Notes

- OS temp files are generally cleaned up automatically, which helps avoid `.mem/tmp/` growing without bounds.
- The “READ THIS ENTIRE FILE” instruction must be repeated at the very top of the temp file content as well.
