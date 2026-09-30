---
created_at: "2026-10-01T01:32:19+02:00"
user: benjamin_van_heerden
---

# Work Log - Structure doc refresh, .DS_Store cleanup, v0.6.1 and its revert in v0.6.2

## Overarching Goals

Tidy up after the v0.6.0 release: stop tracking Finder files, bring `.mem/structure.md` up to date as a living document, decide how `mem log new` should prompt for structure updates, and release the result.

## What Was Accomplished

### Repository hygiene

- `.mem/.DS_Store` and `internal/.DS_Store` removed from the index (`git rm --cached`) and `.DS_Store` added to `.gitignore` (a81869f).

### Structure doc refresh (861c1d6)

- A full read of `.mem/structure.md` found statements line-count drift could not reveal: external services now include GitHub's REST API (`internal/github`, only in `production_pr` projects, with a token); the test list gained `github` and the `promote`, `promote_pr` (fake GitHub), `deploy` and `onboard_setup` CLI tests; slugs are described with the current rules (hyphenated words whole, repeated words dropped); conventions gained "normalise CRLF in text read from Git checkouts" and "check Windows CI after every push"; `Library.Setup` notes the CRLF normalisation.

### `log new` structure prompt: tried, released, reverted

- 86509c9/5bf790e made `mem log new` list the code changed since the structure doc's baseline and ask for an update after any code change, not only past the drift threshold; released as v0.6.1 (date tag v2026.10.01.2).
- The user then preferred the original behaviour. Reverted on dev (a19600c, f73b9c7), so `log new` again asks only past the threshold (more than 5 changed code files or 1000+ changed lines), and released as v0.6.2 (date tag v2026.10.01.3) through the drafted-notes flow with the user's confirmation; Release workflow succeeded with 7 assets; installed mem updated v0.6.0 → v0.6.2 (never ran v0.6.1).

## Decisions

- `mem log new` prompts for structure doc updates only once drift passes the threshold; smaller changes are left to the agent's judgement and the standing AGENTS.md instruction to keep the doc current.
- A published version is not deleted or re-tagged: an unwanted change is reverted and released as the next patch version.

## Key Files Affected

- `.gitignore` (`.DS_Store`); `.mem/.DS_Store`, `internal/.DS_Store` untracked.
- `.mem/structure.md` (external services, tests, slugs, conventions, templates).
- `internal/cli/log.go`, `internal/cli/log_test.go`, `docs/design.md`: changed for the always-on prompt, then reverted to their earlier content.
- Tags: v2026.10.01.2 and v0.6.1 (always-on prompt), v2026.10.01.3 and v0.6.2 (revert).

## Errors and Barriers

- The release was interrupted after `v0.6.1` had been tagged and pushed; its build finished and published regardless. Machines that ran `mem onboard` in that window may have taken v0.6.1 and move to v0.6.2 at their next onboard.
