---
title: Add the folder mode to project loading
status: todo
created_at: "2026-11-01T14:52:34+02:00"
updated_at: "2026-11-01T14:52:34+02:00"
---

internal/project: add mode = "folder" to Config, a Folder() helper, validation rejecting [git], [release] and [templates] in a folder config, and WriteConfig leaving out an empty [git]. Load walks up to the nearest .mem/config.toml and returns a folder project without touching Git; anything else goes through git.Toplevel as today. User takes the slug of the OS account's full name in folder mode, falling back to the account name. Add one shared refusal error and use it in log commit, sync, promote, deploy, import, template, spec, task and the Git hooks. Test: a folder project loads with a failing git on PATH, and each refused command exits non-zero with the same line and changes nothing. Git projects stay unchanged.
