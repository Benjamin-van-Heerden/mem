---
created_at: "2026-10-07T08:41:34+02:00"
user: benjamin_van_heerden
---

# Work Log - Install guide requires a permanent PATH entry, mem v0.7.5

## Overarching Goals

Make the install guide unambiguous that mem must end up permanently on PATH; a mem available only for one session breaks the next session's `mem onboard`.

## What Was Accomplished

### Install guide (dfea452)

- `internal/agentsmd/install.md` restructured into four numbered steps: install (Go or release binary, both setting `$dir`), put mem on PATH permanently, check that a new session finds it, run `mem onboard`. A bold opening states the requirement and why.
- Step 2 for macOS/Linux appends `export PATH="<dir>:$PATH"` to `~/.zshenv` for zsh (read by every zsh, including non-interactive agent shells), `~/.bashrc` and `~/.profile` for bash, `~/.profile` otherwise, skipping files that already contain the directory, and exports it for the current session. Windows adds `$dir` to the user Path setting unless present.
- Step 3 checks from a clean environment: `env -i HOME="$HOME" TERM=dumb "${SHELL:-/bin/sh}" -ic 'command -v mem'`; Windows reads the user Path from the registry. An inherited PATH would have made the first version of this check pass with a broken profile.
- Agents tell the user what they add to the profile and why, and make the change once the user agrees; environments rebuilt per session get a suggestion to add the install to their setup script.
- `internal/agentsmd/instructions.md`: the pointer now says "follow all of it, including putting mem permanently on PATH".
- Tested steps 1–3 with zsh and bash in scratch homes under `env -i`: a fresh shell found mem, a second run of step 2 added no duplicate, and with the profile line removed step 3 found nothing. Windows steps not run.

### Release and rollout

- mem v0.7.5 (date tag v2026.10.07.1); release workflow built 7 assets; installed mem updated v0.7.4 → v0.7.5.
- studii-tan onboard committed and pushed the new guide and AGENTS.md line (99abf07 on its `dev`).

## Decisions

- A permanent PATH entry is a required install step; agents still get the user's agreement before editing a shell profile or the user PATH.
- For zsh the PATH line goes in `~/.zshenv` rather than `~/.zshrc`, so non-interactive shells find mem.

## Key Files Affected

- `internal/agentsmd/install.md`, `internal/agentsmd/instructions.md`.
- Tags: v2026.10.07.1 and v0.7.5.

## Errors and Barriers

- A test that sourced the guide's step 1 through `sh -c` was stopped by the safety check because the step removes its temporary download directory; the test was rerun from a script file with that removal stripped, leaving scratch directories in the session scratchpad.
