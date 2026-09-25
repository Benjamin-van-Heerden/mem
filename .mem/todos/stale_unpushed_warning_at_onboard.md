---
title: Stale unpushed warning at onboard
status: open
created_at: "2026-09-25T12:25:58+02:00"
---

Onboard prints the ⚠️ 'local commits not pushed' nudge under SHARED CODEBASE, then its UPDATES step commits and pushes the mem project files, which pushes those commits too. The agent is then instructed to report a warning that no longer holds. Seen on the first onboard of the mem repository (2026-09-25). Re-evaluate the nudge after the update push, or run updates before reporting.
