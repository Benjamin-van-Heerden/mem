---
created_at: "2026-10-14T14:41:42+02:00"
user: benjamin_van_heerden
commits: [dc8e056, c3260dd, 7873e5e]
---

# Work Log - Migration leftovers: CLAUDE.md copies, extra harness blocks, session titles

## Goal

Fix what the Agent Core to mem migrations left behind in other checkouts. A teammate onboarding a freshly migrated repo found a CLAUDE.md that still described the old harness and contradicted the mem instructions in AGENTS.md.

## What Was Accomplished

### Onboard and import delete the old harness's CLAUDE.md (7873e5e)

The old harness's `ensure_claude_file` made CLAUDE.md a symlink to AGENTS.md. On Windows it copied the file instead (`old/coding/setup.py`, `old/legal/setup.py`). Such a copy is a plain, usually untracked file, so the migration commit never removed it. Claude Code loads CLAUDE.md alongside AGENTS.md, so the old instructions conflicted with mem in every checkout that still had one. Previously, `mem onboard` never looked at CLAUDE.md. Import only removed a link, and only in the importing checkout, and noted a separate file there.

New `internal/cli/claude_md.go`:
- `removeClaudeLink` and `claudeLinkRemoved` moved here from `root.go`. `init` still uses them alone.
- `settleClaudeFile(ctx, root) (line, tracked, err)` handles three cases:
  - A link to AGENTS.md is removed.
  - A file holding a harness block (`agentsmd.HasLegacy`, the existing `legacyBlock` regexp) is deleted. The output says to disregard CLAUDE.md instructions loaded at session start.
  - Any other CLAUDE.md is left in place with a ⚠️ line: move what is still needed into AGENTS.md and delete it.
  - `tracked` comes from `git ls-files`, so onboard can publish the deletion.
- `applyUpdates` calls it and adds `CLAUDE.md` to the published paths when it was tracked. `mem import` calls it too, without committing.
- The importer's old note is removed. It wrongly said Claude Code reads CLAUDE.md "instead of" AGENTS.md.
- The onboard instruction gains a step: "Tell the user about each ⚠️ item under 🔄 UPDATES and settle it with them" (`contextState.updateWarnings`). Before this, ⚠️ lines under UPDATES (tracked `.mem/local`, a failed publish) had no matching instruction.

Test `TestOnboardDeletesTheOldHarnessCLAUDECopyAndFlagsItsOwnFile` covers an untracked copy, a committed copy (deleted, committed and pushed) and a CLAUDE.md with its own content (kept and flagged). Focused cli, importer and agentsmd tests pass, and `go vet` is clean.

### Earlier in the day

- c3260dd: import drops every old harness block from AGENTS.md. Later harness versions added `<AGENT_CORE>` without removing `<core_instructions>`, and ev-ledger's AGENTS.md kept the old one. `extraLegacyBlock` removes the rest after the first is replaced.
- dc8e056: the session title instruction says to name only the work, not the project. An agent had copied a project prefix from an old log title.

## Decisions

- Onboard deletes a harness CLAUDE.md copy without asking (the user's call). It holds only harness boilerplate that AGENTS.md has replaced, and the harness always overwrote CLAUDE.md wholesale, so a file with a harness block is never hand-written.
- A CLAUDE.md without a harness block is never deleted, only flagged at each onboard until it is resolved. Claude Code reads AGENTS.md natively, so mem projects need no CLAUDE.md at all.

## Key Files Affected

- `internal/cli/claude_md.go` (61 lines): new; CLAUDE.md link removal and `settleClaudeFile`
- `internal/cli/onboard_test.go` (38 lines): CLAUDE.md onboard test
- `internal/cli/root.go` (19 lines): CLAUDE.md helpers moved out
- `internal/agentsmd/agentsmd.go` (15 lines): `HasLegacy`; every harness block dropped on import (c3260dd)
- `internal/agentsmd/agentsmd_test.go` (11 lines): several harness blocks test (c3260dd)
- `internal/cli/onboard_updates.go` (10 lines): calls `settleClaudeFile` and publishes a tracked removal
- `internal/cli/import.go` (6 lines): uses `settleClaudeFile`
- `internal/cli/onboard.go` (5 lines): instruction for ⚠️ under UPDATES
- `README.md` (4 lines): onboard and import describe the CLAUDE.md cleanup
- `internal/importer/importer.go` (3 lines): incorrect CLAUDE.md note removed
- `internal/cli/session_title.go` (2 lines), `internal/cli/session_title_test.go` (2 lines): dc8e056
- `.mem/structure.md`: `claude_md.go` and the onboard flow

## Notes

The fix reaches other checkouts only after the next mem release. Until then, someone whose CLAUDE.md still holds the old harness instructions can delete it by hand.
