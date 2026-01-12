"""
Pydantic models for AI agent structured outputs.
"""

from pydantic import BaseModel, Field


class ParsedTask(BaseModel):
    title: str = Field(
        ..., description="The task title (e.g. 'Create Database Migration')"
    )
    description: str = Field(
        "",
        description="The full task description including checkbox items and implementation details",
    )


class ParsedSpec(BaseModel):
    title: str = Field(..., description="The spec title from the first heading")
    body: str = Field(
        ...,
        description="Cleaned markdown body with ## Overview, ## Goals, ## Technical Approach sections",
    )
    tasks: list[ParsedTask] = Field(
        default_factory=list, description="List of tasks extracted from the spec"
    )


class ParsedLog(BaseModel):
    title: str = Field(..., description="Short descriptive title for this work session")
    spec_file: str | None = Field(None, description="The spec file path if mentioned")
    body: str = Field(..., description="Cleaned work log content")
