"""
Document summarizer agent for generating concise summaries of technical documentation.
"""

from agno.agent import Agent
from agno.models.openrouter import OpenRouter

model = OpenRouter("google/gemini-3-flash-preview", max_tokens=2048)

doc_summarizer_agent = Agent(
    model=model,
    name="Document Summarizer Agent",
    instructions=[
        """You are an expert at summarizing technical documentation for developers.

Given a technical document, write a brief summary (200-400 words) that:
1. Explains what the document is about in 1-2 sentences
2. Covers the quickstart or core ideas - what does someone need to know to get started?

Keep it practical and concise. No headers or fancy formatting - just clear prose.
"""
    ],
)


def summarize_document(content: str, doc_name: str) -> str | None:
    """Generate a summary for a technical document.

    Args:
        content: The full document content
        doc_name: The document name/slug for context

    Returns:
        The summary text if successful, None if generation fails.
    """
    prompt = f"Summarize this technical document titled '{doc_name}':\n\n{content}"

    try:
        response = doc_summarizer_agent.run(prompt)
        return response.content
    except Exception:
        return None
