---
created_at: "2026-10-26T11:04:09+02:00"
user: benjamin_van_heerden
commits: [8bec910]
---

# Work Log - AGENTS.md size warning at memory set and onboard

## Goal

Codex reads only the first 32 KiB of AGENTS.md by default (`project_doc_max_bytes`), and project memories come after mem's block, so they are what a cut drops. Shannon's AGENTS.md is already 33,456 bytes. The aim was for mem to warn when the file goes over that limit and have the agent settle it with the user.

## What Was Accomplished

### `agentsSizeWarning` (8bec910)

- `internal/cli/memory.go` has `agentsSizeLimit = 32 << 10` and `agentsSizeWarning(text)`. Over the limit, it returns one ⚠️ line with the size in KiB. The line says that some agent runtimes, such as Codex, read only the first 32 KiB and silently drop the project memories and instructions at the end. It tells the agent to propose which memories to merge, shorten or retire (`mem memory list`) and change them only as the user decides, and mentions `project_doc_max_bytes` in `.codex/config.toml` as the option for Codex alone.
- `mem memory set` prints the line after the saved memory and adds "Stop here: tell the user about the ⚠️ above and settle it with them before you continue with other work." to its agent instruction.
- Onboard reads AGENTS.md after `syncTemplates` and appends the line to the 🔄 UPDATES lines, so the existing "Tell the user about each ⚠️ item under 🔄 UPDATES" step picks it up. Checking there covers every way the file grows: memory set, template memories, import, refreshes of mem's block and hand edits.
- The warning appears on every check while the file is over the limit, not only on the write that crosses it.

### Docs

`docs/design.md` (Memory in the vocabulary), `docs/status.md` (onboard) and `.mem/structure.md` (`memory.go` in the cli module list).

### Verification

- `TestAnAgentsMdOverCodexsLimitIsFlaggedAtMemorySetAndOnboard` (new `internal/cli/memory_test.go`): no warning for a small memory, a warning and the stop instruction for one that crosses 32 KiB, a warning again on a later small memory while over, the onboard warning after the teammate's push, and none after `memory remove`.
- `go test ./internal/cli -run 'TestAnAgentsMdOverCodexsLimit|TestOnboard'` and `go vet ./internal/cli` passed.
- `dist/mem-dev memory set` in a scratch repository printed the warning at 35.5 KiB.
- CI passed on ubuntu, macos and windows.

## Decisions

- **Warn at memory set and at onboard (user's choice).** Memory set alone would miss template memories, imports and hand edits.
- **One threshold, at exactly 32 KiB, with no "getting close" level.**
- **The agent only proposes.** Memories are the user's conventions, so the agent stops and offers merges, trims or retirements rather than making them.

## Key Files Affected

- `internal/cli/memory_test.go` (new)
- `internal/cli/memory.go`
- `internal/cli/onboard.go`
- `docs/design.md`, `docs/status.md`, `.mem/structure.md`

## Notes

- Whether Codex counts the user's global `~/.codex/AGENTS.md` against the same 32 KiB was not verified, so the warning does not claim it.
- mem's own block is about 16 KB, so trimming memories frees only so much. The todo `slim_managed_agents_md_block` is still open.
