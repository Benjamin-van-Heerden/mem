---
title: Let task and spec completion commit, sync and push
status: open
created_at: "2026-09-30T13:24:23+02:00"
---

Found while dogfooding spec template_setup_files. mem task complete asks the agent to commit the task's changes together with the record, which is ambiguous in order (natural rhythm is code commit then record) and awkward when the work lives in another repo; its drift nudges only read local state and print after the instruction, so 9 commits piled up unpushed. spec start commits and pushes, but spec complete only asks the agent to; mem sync never pushes. Proposal agreed in discussion: the agent commits the task's code first, then mem task complete commits the record itself ('Complete task <slug>'), runs the mem sync catch-up (fetch, fast-forward or safe rebase, push) and reports incoming changes before the instruction; with other uncommitted changes it commits only the record and warns. Same for spec complete. Update spec start/task instructions and internal/agentsmd/instructions.md accordingly. Decide whether mem sync should push unpushed commits.
