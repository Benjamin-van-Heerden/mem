---
title: Consider replacing mem precommit with checks that run before promotion
created_at: "2026-11-01T20:15:43+02:00"
---

mem precommit is currently meant to run before every commit, which may cost a lot of time and get in the way of other work. Consider removing the command and its instruction in internal/agentsmd/instructions.md, and keeping only checks that run before promotion: mem promote staging and mem deploy already run [precommit] commands from .mem/config.toml. Decide what the setting should then be called.
