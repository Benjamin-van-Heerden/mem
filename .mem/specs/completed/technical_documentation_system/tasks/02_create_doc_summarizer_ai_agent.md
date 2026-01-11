---
title: Create doc summarizer AI agent
status: completed
created_at: '2026-01-11T12:41:32.916128'
updated_at: '2026-01-11T13:41:10.590597'
completed_at: '2026-01-11T13:41:10.590590'
---
Create src/utils/ai/doc_summarizer.py:
- Use same pattern as log_parser.py
- Model: OpenRouter google/gemini-3-flash-preview
- Agent instructions: Generate concise summary of technical documentation
- Output should be plain text (not structured), ~200-400 words
- Include: what the doc covers, core concepts, getting started tips if applicable
- Function: summarize_document(content: str) -> str | None

## Completion Notes

Created src/utils/ai/doc_summarizer.py with DocSummary Pydantic model and summarize_document function using agno Agent with OpenRouter gemini-2.5-flash

## Completion Notes

Created src/utils/ai/doc_summarizer.py with simplified agent using OpenRouter gemini-2.5-flash, returns plain text summary without structured output