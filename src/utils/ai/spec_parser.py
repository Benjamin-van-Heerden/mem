"""
Spec parser agent for converting old agent_rules specs to mem format.
"""

import json

from agno.agent import Agent
from agno.models.openrouter import OpenRouter

from src.utils.ai.models import ParsedSpec

model = OpenRouter("google/gemini-3-flash-preview", max_tokens=8192)

spec_parser_agent = Agent(
    model=model,
    name="Spec Parser Agent",
    instructions=[
        """You are an expert at parsing semi-structured markdown spec files and converting them to a structured format.

Given an old spec file, extract:
1. title: The spec title (from the first heading)
2. body: A cleaned markdown body with ## Overview, ## Goals, ## Technical Approach sections
3. tasks: A list of tasks with their titles, and descriptions

When converting the body:
- Use the Description section as the Overview
- Extract goals from the content if present
- Keep technical details in Technical Approach
- Remove any status markers like "%% Status: ... %%"
- Remove task sections from the body (they go in the tasks array)

When extracting tasks:
- Look for "### Task" or "### Task N:" sections (e.g. "### Task 1: Create Database Migration")
- The task title is the text after "Task:" or "Task N:" (e.g. "Create Database Migration")
- Include everything under the task heading in the description (checkbox items, implementation details, etc.)
"""
    ],
    output_schema=ParsedSpec,
)


def parse_spec(content: str) -> ParsedSpec | None:
    """Parse an old spec file content using the AI agent.

    Returns ParsedSpec if successful, None if parsing fails.
    """
    response = spec_parser_agent.run(f"Parse this spec file:\n\n{content}")

    if isinstance(response.content, ParsedSpec):
        return response.content

    if isinstance(response.content, str):
        try:
            data = json.loads(response.content)
            return ParsedSpec(**data)
        except (json.JSONDecodeError, ValueError):
            return None

    return None
