---
title: End-to-end run with real Neon, Vercel and GitHub
status: todo
created_at: "2026-09-30T13:26:00+02:00"
updated_at: "2026-09-30T13:26:00+02:00"
---

Ask the user first: this creates a throwaway GitHub repository, Neon project and Vercel project. The user approved (2026-09-30) creating throwaway resources and asked for a representative test: create the empty GitHub repository and clone it, then spawn a subagent with only the repository path and a vague request ("I want to start a new web app here with mem's nextjs-web template") and no knowledge of this spec. The subagent drives mem init and .mem/setup.md itself; this session plays the user, answering its questions and (you) steps via SendMessage, without steering. Record every place it stalls or goes wrong, and fix setup.md, the skills or mem accordingly. Record each created resource in the scratchpad teardown list, and tear everything down afterwards (GitHub repo, Neon project, Vercel project, local clones). Then ask the user to review and push mem-templates.
