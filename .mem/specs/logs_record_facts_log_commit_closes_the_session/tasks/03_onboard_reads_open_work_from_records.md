---
title: Onboard reads open work from records
status: completed
created_at: "2026-09-28T09:02:02+02:00"
updated_at: "2026-09-28T09:11:21+02:00"
completed_at: "2026-09-28T09:11:21+02:00"
---

In internal/cli/onboard.go: show the current user's latest log in full and list other logs from the last 14 days by date, user and title (at most 10) with mem log show; show todo age and claimer in OPEN TODOS; change the summary step to: summarize from open specs, open todos and release status; logs are background, not a list of open work. Extend an existing cli test or add one for the log selection and the todo age.

## Completion Notes

Onboard's WORK LOGS section shows the user's latest log in full (or the latest from anyone when the user has none), lists other logs from the last 14 days (at most 10) by date, user, heading and name with mem log show, and says logs are background while open work is in specs and todos. OPEN TODOS gains AGE and CLAIMED BY columns. The summary step now works from open specs, open todos and release status and tells the agent not to present an old log's plans as open work. Verified by TestOnboardShowsTheLatestLogInFullAndListsRecentOnes (latest in full, older and teammate logs listed, a 30-day-old log left out, todo age 12 days, new instruction). go vet clean, cli tests pass.
