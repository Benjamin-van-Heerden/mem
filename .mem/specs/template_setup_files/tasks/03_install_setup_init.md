---
title: Install setup at init
status: completed
created_at: "2026-09-30T11:56:30+02:00"
updated_at: "2026-09-30T11:59:34+02:00"
completed_at: "2026-09-30T11:59:34+02:00"
---

internal/cli/init.go: after templates.Sync, when lib.Setup(templateNames) is non-empty write .mem/setup.md, list it under 📄 FILES ('.mem/setup.md  one-time setup from the template; onboard walks the agent through it') and include it in the paths the instruction tells the agent to commit (templateNote / synced.Paths). mem template use (internal/cli/template.go) never installs setup; when the added template has one, print: '<name> has a setup for new projects (<name>/setup.md in the library); this project did not run it. Its steps can guide adding what is missing.' mem template list marks templates with a setup. Tests in init_test.go and template_test.go.

## Completion Notes

mem init writes lib.Setup(templateNames) to .mem/setup.md when non-empty, lists it under FILES and among the paths to commit, and its step-3 instruction says onboard presents the setup first. mem template use never installs it and prints '<name> has a setup for new projects …' when the template has one; template list has a SETUP column. TestInitInstallsTheTemplateSetupAndTemplateUseDoesNot covers init, list and use; related template and branch tests and go vet pass.
