---
created_at: "2026-11-01T15:26:39+02:00"
user: benjamin_van_heerden
commits: [409ba9e]
---

# Work Log - Folder projects have no user

## Goal

Remove the user from folder mode. The user pointed out it serves no purpose there: it only tells people apart, and a folder project is one person's.

## What Was Accomplished

- `project.Config.User`, its validation, `init --user` and the OS account lookup (`accountName`) are gone. `project.User` returns an empty name in a folder project.
- `work.NewLog` names a log by its time alone when there is no user, and `LogMeta.User` is `omitempty`, so folder logs have no `user:` line.
- `work.Log.By(user)` counts every log as the user's when the user is empty. `LatestLog`, `recentLogs` and onboard's log headings use it, so logs moved across by hand count without any field, whatever they are named.
- `checkpoint.FolderSince` and `FolderCover` take no user. `log list`, `log show`, `log new` and onboard's log tables print no user or spec in a folder project, and onboard prints no `User:` line.
- `TestFolderLogsNameNoOneAndAllCountAsYours` covers a log moved in by hand beside a new one. README, `docs/design.md`, `.mem/structure.md` and the todo `move_work_finance_folder_mode` are updated.
- `go test ./internal/cli/ -run 'Folder'`, `-run 'Onboard|Init|Sync|Log|Single|Hook|Compact'`, `go test` on project, work, checkpoint, structure and agentsmd, and `go vet` pass.

## Decisions

- **Folder projects record no identity at all.** The previous log's decision to keep the name in the config is replaced: a shared folder is not what these projects are for, and the field can be added back if that changes.

## Key Files Affected

- `internal/project/project.go`, `project_test.go`, `internal/work/log.go`, `internal/checkpoint/folder.go`
- `internal/cli/folder.go`, `folder_test.go`, `init.go`, `log.go`, `onboard.go`, `onboard_context.go`, `hook.go`, `compact.go`
- `README.md`, `docs/design.md`, `.mem/structure.md`, `.mem/todos/move_work_finance_folder_mode.md`
