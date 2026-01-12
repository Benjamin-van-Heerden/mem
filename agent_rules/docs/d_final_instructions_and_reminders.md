# Final Instructions and Reminders

This document outlines the core development principles, patterns, and technical guidelines for this project.

## Python Package Structure and Imports

### No `__init__.py` Files
The project follows a simplified approach without `__init__.py` files in directories. This keeps the structure clean and avoids complex import hierarchies.

### Import Patterns
When importing from modules in subdirectories, use explicit imports (from the root of the project):

```python
from src.example.module import example_function
from src.example.models import ExampleModel
```

**NEVER import libraries inside functions unless there is an extremely specific use case for it!**

All imports should be at the top of the file, following this order:
1. Standard library imports (asyncio, logging, etc.)
2. Third-party library imports (cv2, numpy, httpx, etc.)
3. Local project imports (src.* modules)

```python
# CORRECT - imports at top of file
import asyncio
import logging
from typing import Tuple

import cv2
import httpx
import numpy as np

from src.example.models import ExampleModel

# WRONG - importing inside function
async def download_file(url: str) -> bytes:
    import httpx  # DON'T DO THIS!
    ...
```

### Module Organization
- **models.py**: Pydantic models and enums
- **database.py**: Database operations and queries
- Functional modules for specific tasks (image_processing.py, validation.py, etc.)

## Database Migration System

The project uses a custom migration system for AI-specific database changes, managed through `scripts/migrate.py`.

### Migration Commands

**Generate new migration:**
```bash
python scripts/migrate.py --generate "create_new_table"
```
- Creates timestamped UP/DOWN migration files
- Files follow format: `YYYYMMDDHHMMSS_migration_name_UP.sql` and `_DOWN.sql`
- Templates include comments and examples

**Apply pending migrations:**
```bash
python scripts/migrate.py --up
```
- Runs all pending migrations in chronological order
- Updates `ai_server_migrations` tracking table
- Provides clear success/failure feedback

**Rollback migrations:**
```bash
python scripts/migrate.py --rollback -n 2
```
- Rolls back the last N migrations using DOWN files
- Removes records from tracking table
- Supports partial rollbacks

**Check migration status:**
```bash
python scripts/migrate.py --status
```
- Shows applied vs pending migrations
- Useful for deployment verification

For more information, refer to `./scripts/migrate.py`.

## Code Style and Standards

### Error Handling
- Use specific exceptions with clear messages
- Log errors appropriately using `logging.getLogger(__name__)`
- Provide meaningful error context for debugging

### Type Hints
- Use comprehensive type hints for all function signatures
- Import types from `typing` module when needed
- Use Pydantic models for structured data validation

### Async Patterns
- Use `async`/`await` for database operations and external API calls
- Avoid blocking operations in async contexts
- Use proper async context managers (`async with`)
- For network requests, use `httpx`

### Model Design
- Separate technical pipeline status from business assessment results
- Use enums for categorical data with clear, descriptive values
- Include comprehensive field descriptions in Pydantic models
- Use appropriate field validation (ge, le, etc.)

## Database Patterns

### Connection Management
Use the async database cursor context manager:
```python
from src.utils.connections import DBCursorCtx

async with DBCursorCtx() as cur:
    await cur.execute("SELECT * FROM table WHERE id = %(id)s", {"id": record_id})
    result = await cur.fetchone()
```

### Query Patterns
- Use parameterized queries to prevent SQL injection
- Add appropriate indexes for performance
- Use JOINs efficiently with proper foreign key relationships

### Manual SQL Execution
When the migration system fails or you need to run SQL directly for debugging:

```python
# Use uv run python -c for quick SQL execution
uv run python -c "
import asyncio
import sys
sys.path.insert(0, 'src')
from src.utils.connections import DBCursorCtx

async def run_sql():
    async with DBCursorCtx() as cur:
        try:
            await cur.execute('DESCRIBE table_name')
            results = await cur.fetchall()
            for row in results:
                print(row)
        except Exception as e:
            print(f'Error: {e}')

asyncio.run(run_sql())
"
```

**Use cases:**
- Debugging migration failures
- Checking table structure and indexes
- Verifying data integrity
- Quick database operations without creating files

## Project Architecture

### Functional Approach
- Direct route handlers instead of complex class hierarchies
- Simple function signatures with clear input/output types
- Minimal abstractions - prefer explicit over implicit
- Each route handles one specific business function

### Modular Design
- Separate concerns into focused modules
- Clear interfaces between pipeline stages
- Extensible architecture supporting new meter types and validation rules
- Robust error handling at each stage

### Performance Considerations
- Chunked processing for batch operations when applicable
- Database query optimization with proper indexing
- Memory management for large batch processing

## Working Principles

### Communication Style
- Be conversational but professional
- Think through considerations and requirements before writing code
- Planning first, then execution - we discuss the problem before implementing
- Don't be afraid to ask for help or input
- If you are unsure or need to guess about something, please ask me

### Development Workflow
- Make small edits, stop and ask if they are ok, then proceed with the next edit
- Never undo changes in files made by me - if you see code in an unexpected state, STOP and ask me what to do
- Don't write code unprompted - the conversation flow should always be planning first, then execution
- If attempting to run a command or modify a file doesn't work, STOP and ask me what to do
- Focus on one task at a time when working with specs

### Code Quality Standards
- Code should be self-explanatory - NEVER add comments unless absolutely necessary
- Never use print statements directly - always use logging via `logger = logging.getLogger(__name__)`
- Follow established patterns and conventions in the codebase
- Prioritize clarity and maintainability over cleverness

### Testing Philosophy
Testing is critical to maintaining software quality, but not all tests are created equal. Focus on testing meaningful functionality that could actually break and impact the application.

### Test Structure
- Tests live in `tests/` directory with mirrored source structure
- Focus on meaningful functionality that could realistically break
- Avoid "idiot tests" that test framework behavior or trivial logic

### What to Test
- **Business logic**: Complex algorithms, validation rules, data transformations
- **API endpoints**: Request/response handling, authentication, error cases
- **Database operations**: Query correctness, constraint validation, data integrity
- **Integration points**: External API calls, file processing, inter-service communication

### What NOT to Test
- Framework internals (FastAPI routing, Pydantic validation basics)
- Third-party library behavior (MySQL connector, OpenCV functions)
- Trivial getters/setters or simple data transformations
- Implementation details that don't affect public behavior

**Test Quality Principles:**
1. **Clarity Over Quantity** - Fewer, well-focused tests are better than many redundant ones
2. **Test Behavior, Not Implementation** - Focus on what the code does, not how it does it
3. **Meaningful Assertions** - Each test should verify something that could realistically fail
4. **Isolated Tests** - Tests should not depend on each other or external state
5. **Descriptive Names** - Test names should clearly describe what they're validating

**When in Doubt, Ask:**
- "If this test fails, would it indicate a real problem?"
- "Does this test validate critical business logic or user-facing behavior?"
- "Could this functionality realistically break in a way that matters?"

If the answer is no, skip the test and focus on more valuable testing efforts.

Remember: "Whenever I'm about to do something, I think, 'Would an idiot do that?' And if they would, I do not do that thing." - Dwight Schrute

**NEVER WRITE OR RUN TESTS UNLESS PROMPTED TO OR WE ARE EXPLICITLY WORKING ON TEST CASES.**

# Final Remarks

- **NEVER** run a server yourself
- **NEVER** run any services of any kind (including main.py, dev servers, task runners, API servers, etc.)
- **NEVER** execute commands that start, test, or interact with running services
- **NEVER** use timeout or any method to run service startup code, even briefly
- **NEVER** perform any mutating actions on the database without explicit consent
