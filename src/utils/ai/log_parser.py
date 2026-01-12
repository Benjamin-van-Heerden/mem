"""
Log parser agent for converting old agent_rules work logs to mem format.
"""

import json

from agno.agent import Agent
from agno.models.openrouter import OpenRouter

from src.utils.ai.models import ParsedLog

model = OpenRouter("google/gemini-3-flash-preview", max_tokens=8192)

log_parser_agent = Agent(
    model=model,
    name="Log Parser Agent",
    instructions=[
        """You are an expert at parsing work log files and cleaning them up.

Given an old work log file, extract and clean the content.

When cleaning the body:
- Keep the main sections: Overarching Goals, What Was Accomplished, Key Files Affected, What Comes Next
- Remove the "## Spec File:" section (we extract it separately)
- Clean up any formatting issues
- Keep the content concise and well-structured
"""
    ],
    output_schema=ParsedLog,
)


def parse_log(content: str) -> ParsedLog | None:
    """Parse an old work log file content using the AI agent.

    Returns ParsedLog if successful, None if parsing fails.
    """
    response = log_parser_agent.run(f"Parse this work log file:\n\n{content}")

    if isinstance(response.content, ParsedLog):
        return response.content

    if isinstance(response.content, str):
        try:
            data = json.loads(response.content)
            return ParsedLog(**data)
        except (json.JSONDecodeError, ValueError):
            return None

    return None
