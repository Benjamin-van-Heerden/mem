---
title: Documentation and managed instructions
status: completed
created_at: "2026-09-25T14:00:22+02:00"
updated_at: "2026-09-28T07:56:30+02:00"
completed_at: "2026-09-28T07:56:30+02:00"
---

Add one short entry point to internal/agentsmd/instructions.md (when the user wants a memory or skill in other projects: mem template promote). Add a Templates section to docs/design.md, move Templates to Implemented in docs/status.md, document --template in README.md, and update .mem/structure.md for internal/templates and the new commands.

## Completion Notes

instructions.md gains one entry point (mem template promote) and no longer says memories sit at the end of AGENTS.md; design.md has a Templates section, the layout additions and the implemented self-update under Updates; status.md lists templates and self-update as implemented and keeps template content as planned; README documents --template, template_source, templates in How it works, templates.lock and .agents/skills, and automatic updates; .mem/structure.md describes internal/templates, internal/selfupdate, the new commands, the onboard order and the new external interfaces. go vet ./... clean.
