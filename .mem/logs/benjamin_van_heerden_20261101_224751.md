---
created_at: "2026-11-01T22:47:51+02:00"
user: benjamin_van_heerden
---

# Work Log - work moved to folder mode, on v0.13.0

## Goal

Put folder mode to real use by moving the `work` folder onto it, and record what is left in this repository.

## What Was Accomplished

- `work` now lives at iCloud Drive/work as a folder project running mem v0.13.0. `~/Documents/work` and the GitHub repository Benjamin-van-Heerden/work are untouched and hold the state from before the move. Its own logs under `.mem/logs/` cover the detail: skills, billing scripts, the invoice ledger and the October invoices.
- Added the todo `consider_replacing_mem_precommit_checks`: running checks before every commit may cost too much, and checks before promotion may be enough.
- The todo `move_work_finance_folder_mode` now says finance moves into `work` rather than on its own. The consolidation is tracked as a todo in the `work` folder.

## Decisions

- **Finance joins `work` as one folder project.** Invoices need the bank and payment records to show what is unpaid.

## Key Files Affected

- `.mem/todos/consider_replacing_mem_precommit_checks.md`, `.mem/todos/move_work_finance_folder_mode.md`

## Notes

- Folder mode held up in use: onboard, runnables, memories, todos and logs all worked from iCloud Drive, and `mem init --folder` needed no changes.
