---
title: Staleness at onboard
status: completed
created_at: "2026-10-08T13:07:22+02:00"
updated_at: "2026-10-08T14:32:58+02:00"
completed_at: "2026-10-08T14:32:58+02:00"
---

Onboard flags: open todos claimed more than 30 days ago (in the todo table, with an instruction to check each against the code and the user); active specs with no completed task and no record change for 14 days; remote branches other than development/staging/production not merged into development with no commit for 14 days, listed with age and last author (merge, delete or keep, with the user); release status adds the age of the oldest commit on development not on staging. Tests for each flag. Update docs and the structure doc.

## Completion Notes

onboard_staleness.go: writeStaleness prints a ⏳ CHECK THESE section after the todos for todos claimed over 30 days ago, active specs whose directory's last commit is over 14 days old, and remote branches not merged into origin/<dev> (for-each-ref --no-merged) with no commit for 14 days, excluding development, staging, production and promotion/ branches; the onboard instruction then asks to go through each with the user. release.CurrentStatus adds OldestUnreleased (relative date of the first development commit not on staging), shown in onboard's release status. Verified: TestOnboardFlagsRecordsAndBranchesThatStoppedMoving (backdated claim, spec commit and pushed branch flagged, fresh claim not flagged, unreleased age shown, instruction step present); full internal/cli and internal/release suites; go vet. Docs: design, status, structure doc.
