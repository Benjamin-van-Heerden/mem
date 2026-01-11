---
title: Create doc summarizer AI agent
status: todo
created_at: '2026-01-11T12:41:32.916128'
updated_at: '2026-01-11T12:41:32.916128'
completed_at: null
---
Create src/utils/ai/doc_summarizer.py:
- Use same pattern as log_parser.py
- Model: OpenRouter google/gemini-3-flash-preview
- Agent instructions: Generate concise summary of technical documentation
- Output should be plain text (not structured), ~200-400 words
- Include: what the doc covers, core concepts, getting started tips if applicable
- Function: summarize_document(content: str) -> str | None