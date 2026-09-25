---
title: Literal mem tag anywhere in AGENTS.md breaks onboard
status: open
created_at: "2026-09-25T12:25:58+02:00"
---

agentsmd.span counts every occurrence of the tag text, so a user section that mentions the managed block's tag in backticks makes onboard fail with 'AGENTS.md must contain exactly one <mem> ... </mem> block'. Match the tags only as whole lines, as the block writes them.
