---
title: Install setup at init
status: todo
created_at: "2026-09-30T11:56:30+02:00"
updated_at: "2026-09-30T11:56:30+02:00"
---

internal/cli/init.go: after templates.Sync, when lib.Setup(templateNames) is non-empty write .mem/setup.md, list it under 📄 FILES ('.mem/setup.md  one-time setup from the template; onboard walks the agent through it') and include it in the paths the instruction tells the agent to commit (templateNote / synced.Paths). mem template use (internal/cli/template.go) never installs setup; when the added template has one, print: '<name> has a setup for new projects (<name>/setup.md in the library); this project did not run it. Its steps can guide adding what is missing.' mem template list marks templates with a setup. Tests in init_test.go and template_test.go.
