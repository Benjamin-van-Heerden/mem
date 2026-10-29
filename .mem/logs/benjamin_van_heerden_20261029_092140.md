---
created_at: "2026-10-29T09:21:40+02:00"
user: benjamin_van_heerden
---

# Work Log - mem v0.12.0 released: template change notes; React Compiler note published

## Goal

Release template change notes, then publish the first note, the nextjs-web React Compiler change, so existing projects receive it.

## What Was Accomplished

- **CI:** passed on ubuntu, macOS and Windows for dev at 47ca1db, the spec's work log commit, and earlier for f1370b1 and 4b14570.
- **Staging:** `mem promote staging` moved `test` to 47ca1db (16 commits since v0.11.0).
- **Production:** `mem promote production` drafted the notes. They were rewritten as a user-facing summary of:
  - template change notes;
  - convention todos during a pending setup;
  - convention todos for tooling the project never adopted.

  The user confirmed them, and `mem promote production --confirm` tagged `v2026.10.29.1` on `main` at 47ca1db.
- **Semver tag:** `v0.12.0` was tagged on the same commit with the date tag's notes (`--cleanup=whitespace`, `#` headings kept) and pushed. The Release workflow (run 40834679357) succeeded with 7 assets.
- **Local install:** `mem update` first reported v0.11.0 as the latest; the `/releases/latest` redirect lagged again. A retry a minute later moved the installed mem to v0.12.0 (47ca1db, schema 2).
- **Change note:** mem-templates eb333eb was pushed to `main`, with `nextjs-web/changes/2026-10-29-react-compiler.md` and the README's Change notes section. Every nextjs-web project gets the "Apply the nextjs-web template change: React Compiler through the Rust port in Turbopack" todo at its next onboard; onboard updates mem to v0.12.0 before it syncs templates. Todo `push_react_compiler_change_note` was deleted.

## Decisions

- **Minor version bump (v0.12.0):** the release adds a feature and changes nothing for existing locks or projects without notes.

## Key Files Affected

- Tags `v2026.10.29.1` and `v0.12.0` on `main` (47ca1db); `test` and `main` fast-forwarded to 47ca1db.
- `.mem/todos/push_react_compiler_change_note.md`: deleted.
- Outside this repository: mem-templates eb333eb pushed; the installed mem at `~/.local/bin/mem` is v0.12.0.

## Notes

- **`git rev-parse --short` with several revisions fails** with "Needed a single revision" in this git version. `git log -1 --format=%h <ref>` per ref works.
