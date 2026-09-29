---
title: Update docs and structure for the sync catch-up
status: completed
created_at: "2026-09-29T14:05:30+02:00"
updated_at: "2026-09-29T14:16:06+02:00"
completed_at: "2026-09-29T14:16:06+02:00"
---

Update .mem/structure.md (mem sync data flow, hook compact, .claude/settings.json in External Interfaces), docs/design.md and docs/status.md (implemented vs planned), README.md if it describes mem sync or hooks, and the mem sync wording in internal/agentsmd/instructions.md only if its one line about mem sync needs to change (keep persistent context minimal). Delete the todos mem_sync_reports_what_teammates_changed, decide_on_mem_template_sync and consider_a_post_compaction_hook and commit the removal with this work.

## Completion Notes

Updated .mem/structure.md (claude package, sync/compaction data flow, RecordChanges, ClaudeConfig, External Interfaces, test list), docs/design.md (mem sync as catch-up, the After compaction section, [claude] in the config example), docs/status.md (implemented sync and compaction hook; bash assumption under limitations) and README.md (mem sync and the hook, init adds .claude/settings.json). The one mem sync line in instructions.md is still accurate and was left alone. Deleted the three todos this spec closes.
