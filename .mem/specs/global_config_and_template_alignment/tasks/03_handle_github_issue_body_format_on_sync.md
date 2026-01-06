---
title: Handle GitHub issue body format on sync
status: completed
subtasks:
- title: Create spec-template.md format in global config
  status: completed
- title: Update mem init to generate .github/ISSUE_TEMPLATE/mem-spec.yml from global
    template
  status: completed
- title: Update spec creation to use global template for body
  status: completed
- title: 'Normalize heading levels when syncing from GitHub (### to ##)'
  status: completed
created_at: '2026-01-06T13:43:29.474825'
updated_at: '2026-01-06T14:13:11.356321'
completed_at: '2026-01-06T14:13:11.356311'
---
Ensure specs created from GitHub issues match the same format as locally created specs

## Completion Notes

Implemented single source of truth spec template: ~/.config/mem/templates/spec.md is used for both local spec creation and GitHub issue template (mem-spec.md). mem init now ensures global config exists and generates the GitHub template from it.