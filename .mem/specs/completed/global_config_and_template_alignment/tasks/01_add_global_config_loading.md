---
title: Add global config loading
status: completed
subtasks:
- title: Rename generic_templates_location to global_config_dir with default ~/.config/mem
  status: completed
- title: Update config.toml template and .mem/config.toml
  status: completed
- title: Update load_generic_templates to use global_config_dir/templates/
  status: completed
- title: Implement global + local config merging
  status: completed
created_at: '2026-01-06T13:43:28.920829'
updated_at: '2026-01-06T13:49:41.514725'
completed_at: '2026-01-06T13:49:41.514718'
---
Update config reading to load ~/.config/mem/config.toml first, then merge with local .mem/config.toml

## Completion Notes

Implemented global config dir with default ~/.config/mem, config merging, and updated templates