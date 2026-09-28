---
title: init --template and mem template use/list
status: completed
created_at: "2026-09-25T14:00:22+02:00"
updated_at: "2026-09-28T07:51:13+02:00"
completed_at: "2026-09-28T07:51:13+02:00"
---

Add repeatable --template and --template-source to mem init (fail with a clear instruction when no source is given or configured, before writing any files), and a mem template command group with use <name> and list. use adds to [templates] use and syncs; list shows library templates with descriptions and this project's item states. Output follows the existing heading/section/instruction style.

## Completion Notes

init takes repeatable --template and --template-source; it resolves the library (flag, then template_source in ~/.config/mem/config.toml) and validates the templates before writing anything, then syncs and lists the template items in its output and commit instruction. New mem template group (internal/cli/template.go) with use <template> (adds to [templates] use, syncs, tells the agent what to commit) and list (library templates with descriptions and a USED HERE column, plus each project item's state via templates.Status, which shares its classification with Sync). Verified end to end with dist/mem-dev against a local bare library in a scratch repo: init with base+nextjs-web installed memories, the skill and its .claude link and the lock; list showed in sync/local edits; use on an initialized project without a source failed with the configuration instruction, and succeeded with --template-source. go vet clean, templates tests pass.
