---
title: Slim the managed AGENTS.md block
created_at: "2026-10-25T15:33:39+02:00"
---

mem's managed block in AGENTS.md is 16.3 KB (6.9 KB workflow, 9.4 KB general principles and testing philosophy), half of Codex's default 32 KiB AGENTS.md limit before any project memories; shannon's memories add 17 KB. Tighten internal/agentsmd/instructions.md, or move the general principles to something loaded on demand, so projects with many memories stay well within every agent's instruction budget. Discuss what must stay always-on with the user first.
