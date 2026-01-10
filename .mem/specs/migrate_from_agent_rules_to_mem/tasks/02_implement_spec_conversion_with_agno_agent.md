---
title: Implement spec conversion with Agno agent
status: todo
subtasks: []
created_at: '2026-01-10T15:12:27.019096'
updated_at: '2026-01-10T15:12:27.019096'
completed_at: null
---
Create an Agno agent using OpenRouter (anthropic/claude-sonnet-4) that parses old spec files and extracts title, body, and tasks as JSON. For each spec: (1) run through agent to get structured data, (2) create .mem/specs/completed/<slug>/spec.md with proper SpecFrontmatter (status=completed, timestamps from filename), (3) create task files in tasks/ subdirectory. Use src/utils/markdown.py for slugify and write_md_file, and src/models.py for frontmatter creation.