---
title: Add the folder mode to project loading
status: completed
created_at: "2026-11-01T14:52:34+02:00"
updated_at: "2026-11-01T15:00:53+02:00"
completed_at: "2026-11-01T15:00:53+02:00"
---

internal/project: add mode = "folder" to Config, a Folder() helper, validation rejecting [git], [release] and [templates] in a folder config, and WriteConfig leaving out an empty [git]. Load walks up to the nearest .mem/config.toml and returns a folder project without touching Git; anything else goes through git.Toplevel as today. Config gets a top-level user, which a folder config must have, and User returns its slug in folder mode. Add one shared refusal error and use it in log commit, sync, promote, deploy, import, template, spec, task and the Git hooks. Test: a folder project loads with a failing git on PATH, and each refused command exits non-zero with the same line and changes nothing. Git projects stay unchanged.

## Completion Notes

project.Config has mode and user; Load finds a folder project by walking up to its config without running Git, and validate rejects a config that mixes modes. User reads the config in folder mode. A root PersistentPreRunE refuses the Git-only commands with project.ErrFolder. Verified by TestFolderProjectRefusesCommandsThatNeedGit, TestFolderProjectAttributesWorkToItsConfiguredUser and the Load tests, run with a git that fails the test if called; go test on project, agentsmd, checkpoint, structure and cli -run 'Onboard|Init|Sync|Log|Single|Hook', and go vet, pass.
