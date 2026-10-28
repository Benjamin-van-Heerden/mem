---
title: Staged mode treats a checked-out base branch as a feature branch
created_at: "2026-10-28T11:34:48+02:00"
---

In a staged project with [git] base set, checking out the base branch (e.g. main as an upstream mirror) makes converge.feature() true, so mem sync could rebase it onto origin/<development> (and, when the base's remote commits outside development are all the user's, force-push it). Single-branch mode excludes the base via GitConfig.Guarded; staged mode was left unchanged on purpose. Decide with the user whether staged projects should also never treat base as a feature branch (exclude it in feature()), and add a converge test.
