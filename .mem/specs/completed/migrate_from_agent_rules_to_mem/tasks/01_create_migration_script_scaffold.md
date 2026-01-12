---
title: Create migration script scaffold
status: completed
subtasks: []
created_at: '2026-01-10T15:12:19.202114'
updated_at: '2026-01-10T16:46:59.307086'
completed_at: '2026-01-10T16:46:59.307077'
---
Create scripts/migrate_agent_rules.py with: CLI argument parsing (target_project_dir, --dry-run), discovery logic to find spec files in agent_rules/spec/ and work log files in agent_rules/work_log/, validation that agent_rules/ exists, and basic progress reporting. The script should use argparse and print what it finds.

## Completion Notes

Created scripts/migrate_agent_rules.py with CLI, discovery, and main flow