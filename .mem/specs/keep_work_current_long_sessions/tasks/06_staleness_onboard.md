---
title: Staleness at onboard
status: todo
created_at: "2026-10-08T13:07:22+02:00"
updated_at: "2026-10-08T13:07:22+02:00"
---

Onboard flags: open todos claimed more than 30 days ago (in the todo table, with an instruction to check each against the code and the user); active specs with no completed task and no record change for 14 days; remote branches other than development/staging/production not merged into development with no commit for 14 days, listed with age and last author (merge, delete or keep, with the user); release status adds the age of the oldest commit on development not on staging. Tests for each flag. Update docs and the structure doc.
