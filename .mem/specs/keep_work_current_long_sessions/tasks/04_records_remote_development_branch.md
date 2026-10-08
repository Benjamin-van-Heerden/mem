---
title: Say when records on a feature branch reach development
status: completed
created_at: "2026-10-08T13:07:22+02:00"
updated_at: "2026-10-08T14:25:04+02:00"
completed_at: "2026-10-08T14:25:04+02:00"
---

On a branch other than development, record commands that commit something (publish and commitRecord: claims, spec start, task and spec completion, project files; and mem log commit) add one line saying teammates see the change once the branch merges into development. Records stay on the branch; nothing is committed to another branch. Test: claiming a todo on a feature branch commits it there and prints the line; on development no such line. Update docs and the structure doc.

## Completion Notes

Reduced with the user from committing records to development to a notice: records stay on the branch they are changed on. branchNotice (root.go) returns, on a branch other than development, 'You are on <b>, not <dev>: teammates see this record change once <b> merges into <dev>.'; publish, commitRecord and log commit print it after committing. Spec and task text updated to the reduced scope. Verified: TestRecordChangesOnAFeatureBranchStayThereAndSaySo (claim on dev has no notice; claim on a feature branch prints it, is committed on the branch and not on dev); related cli tests and go vet pass. Docs: design (Records), structure doc.
