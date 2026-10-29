---
title: Document change notes and write the first note
status: completed
created_at: "2026-10-29T08:27:22+02:00"
updated_at: "2026-10-29T08:51:09+02:00"
completed_at: "2026-10-29T08:51:09+02:00"
---

Update docs/design.md (template library section: changes/ directory, note format, lock [[change]] records, when notes become todos, why init and template use record silently), docs/status.md (Templates bullet) and .mem/structure.md (internal/templates: Library.Changes, Result.Changes, lock; Templates data flow: changeTodos). Leave internal/agentsmd/instructions.md unchanged. In ../mem-templates: add a 'Change notes' section to README.md (when to write one, when not, format with an example, notes are permanent: write a new one rather than edit) and write nextjs-web/changes/2026-10-29-react-compiler.md as the spec describes (next.config.ts settings from mem-templates 4630cd8, remove babel-plugin-react-compiler if present, fix bun run lint findings such as form.watch -> useWatch, check the UI, not for webpack builds). No note for the Tailwind/partialPrefetching change. Commit mem-templates on main ('nextjs-web: …' / 'README: …', no attribution) but push it only after a mem release that includes this feature; releasing is the user's call, so if none has happened, leave the commit unpushed and tell the user.

## Completion Notes

docs/design.md: change notes in the template library layout and a bullet on notes, the lock's [[change]] records, todos, silent recording for init and template use, the setup variant and permanence. docs/status.md: the Templates bullet. .mem/structure.md: internal/templates (lock.go, changes.go, Sync's adopted templates), internal/cli (template_changes.go, the shared 📌 helpers) and the Templates data flow. Managed instructions unchanged. In mem-templates: a Layout line and a 'Change notes' README section, and nextjs-web/changes/2026-10-29-react-compiler.md (Next 16.3+, the next.config.ts settings, removing babel-plugin-react-compiler, lint fixes such as useWatch, typecheck/build, a UI check by the user, not for webpack builds); no note for the Tailwind/partialPrefetching change. Committed as mem-templates eb333eb on main and deliberately not pushed until a mem release includes change notes.
