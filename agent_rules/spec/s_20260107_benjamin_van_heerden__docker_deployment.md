# Docker Deployment for API Server and Task Runner

`%% Status: Completed %%`

## Description

Create Docker deployment configuration for the WeconnectU AI Server. The system consists of two services that need separate Docker images with different entrypoints:

1. **API Server** (`src/api/main.py`) - FastAPI server that can be horizontally scaled behind a load balancer
2. **Task Runner** (`src/task_runner/main.py`) - Background worker that must run as a singleton instance

Both services share the same codebase and dependencies, so we use a single base Dockerfile with different entrypoints. The database (PostgreSQL with pgvector) will be provisioned separately via AWS RDS.

### Key Requirements

- Use `uv` for Python package management (fast, reliable)
- Python 3.12 base image
- System dependencies: `poppler-utils` (PDF processing), `libreoffice-writer` (DOCX conversion)
- Multi-stage build to reduce final image size
- Health check endpoint for API server (`/health`)
- Graceful shutdown handling for task runner (SIGTERM)
- No dev dependencies in production image
- Database migrations run on task runner startup (singleton ensures no race conditions)

### Database Initialization Sequence

The task runner is responsible for database initialization on startup:

1. **Agno tables first** - Call `get_agno_postgres_db()` which creates:
   - `ai_metrics`
   - `ai_sessions`
   - `ai_memories`
   - `ai_versions`

2. **Custom migrations second** - Run `MigrationTool().run_up()` which:
   - Creates `ai_migrations` tracking table
   - Applies all pending migrations (e.g., `ai_server_tasks`, `parsed_documents`)

This order is important because some migrations may reference Agno tables.

### Deployment Architecture

```
                    ┌─────────────────┐
                    │  Load Balancer  │
                    └────────┬────────┘
                             │
            ┌────────────────┼────────────────┐
            │                │                │
      ┌─────▼─────┐    ┌─────▼─────┐    ┌─────▼─────┐
      │ API Pod 1 │    │ API Pod 2 │    │ API Pod N │
      └───────────┘    └───────────┘    └───────────┘
                             │
                    ┌────────▼────────┐
                    │   AWS RDS       │
                    │  (PostgreSQL)   │
                    └────────┬────────┘
                             │
                    ┌────────▼────────┐
                    │  Task Runner    │
                    │   (Singleton)   │
                    │ [runs migrations│
                    │   on startup]   │
                    └─────────────────┘
```

## Tasks

### Task 1: Add Database Initialization to Task Runner Startup

- [x] Modify `src/task_runner/main.py` to initialize Agno tables on startup
- [x] Run custom migrations after Agno initialization
- [x] Test that startup sequence works correctly

#### Implementation Details

Update `async_main()` in `src/task_runner/main.py` to call:

1. `get_agno_postgres_db()` - triggers Agno table creation (cached, so safe to call multiple times)
2. `MigrationTool().run_up()` - applies pending migrations

```python
from src.utils.agno.db import get_agno_postgres_db
from src.utils.migration_tool import MigrationTool

async def async_main():
    logger.info("Task runner starting...")
    
    # Initialize database tables
    logger.info("Initializing Agno database tables...")
    get_agno_postgres_db()
    
    logger.info("Running database migrations...")
    MigrationTool().run_up()
    
    # ... rest of startup
```

This ensures the singleton task runner handles all DB initialization before API servers start accepting traffic.

> Relevant existing files: [src/task_runner/main.py, src/utils/agno/db.py, src/utils/migration_tool.py]
> New files: []

## Completion Report and Documentation
Added `initialize_database()` function to `src/task_runner/main.py` that calls `get_agno_postgres_db()` and `MigrationTool().run_up()` before starting the worker.

---

### Task 2: Create Base Dockerfile

- [x] Create multi-stage Dockerfile with `uv` for dependency management
- [x] Install system dependencies (poppler-utils, libreoffice-writer)
- [x] Copy only production dependencies (no dev group)
- [x] Set up proper Python path and working directory

#### Implementation Details

Create `Dockerfile` in project root using multi-stage build:

**Stage 1 (builder)**: Install `uv`, copy dependency files, create virtual environment
**Stage 2 (runtime)**: Copy venv and source code, install system deps, set entrypoint

Key considerations:
- Use `python:3.12-slim-bookworm` as base (Debian-based for apt packages)
- Install `uv` via the official installer script
- Use `uv sync --frozen --no-dev` to install only production dependencies
- Use `libreoffice-writer` instead of full `libreoffice` (smaller, sufficient for DOCX)

```dockerfile
# Example structure
FROM python:3.12-slim-bookworm AS builder
# Install uv, copy pyproject.toml, uv.lock, sync dependencies

FROM python:3.12-slim-bookworm AS runtime
# Copy venv from builder, install system deps, copy source
```

> Relevant existing files: [pyproject.toml, uv.lock, .python-version]
> New files: [Dockerfile]

## Completion Report and Documentation
Created multi-stage `Dockerfile` with builder stage (uv sync) and runtime stage. Final image size: 912MB (includes LibreOffice).

---

### Task 3: Create Docker Compose for Production

- [x] Create `docker-compose.yaml` for production deployment
- [x] Define `api` service with appropriate scaling hints
- [x] Define `task-runner` service as singleton (replicas: 1)
- [x] Configure health checks
- [x] Document environment variable requirements

#### Implementation Details

Create `docker-compose.yaml` with two services using the same image but different commands:

```yaml
services:
  api:
    build: .
    command: ["python", "-m", "src.api.main"]
    # ... ports, env, health check

  task-runner:
    build: .
    command: ["python", "-m", "src.task_runner.main"]
    deploy:
      replicas: 1  # Must be singleton
    # ... env, no ports needed
```

The API server exposes port 8000 and has a health check on `/health`.
The task runner has no exposed ports and relies on graceful shutdown via SIGTERM.

> Relevant existing files: [dev.docker-compose.yaml, deployment.md]
> New files: [docker-compose.yaml]

## Completion Report and Documentation
Created `docker-compose.yaml` with `api` and `task-runner` services. API depends on task-runner to ensure migrations run first. Health check configured for API on `/health`.

---

### Task 4: Create .dockerignore

- [x] Exclude unnecessary files from Docker context
- [x] Reduce build context size and improve build speed

#### Implementation Details

Create `.dockerignore` to exclude:
- `.git/` directory
- `__pycache__/` directories
- `.env` files (secrets should be injected at runtime)
- `tests/` directory
- `agent_rules/` documentation
- `docs/` directory
- IDE/editor files (`.vscode/`, `.idea/`)
- `frontend/` directory (not needed in backend containers)

> Relevant existing files: [.gitignore]
> New files: [.dockerignore]

## Completion Report and Documentation
Created `.dockerignore` excluding `.git/`, `__pycache__/`, `.env`, `tests/`, `agent_rules/`, `docs/`, `frontend/`, and IDE files.

---

### Task 5: Update deployment.md

- [x] Add Docker build and run instructions
- [x] Document the two-service architecture
- [x] Add example docker-compose commands

#### Implementation Details

Update `deployment.md` with:
- Docker build commands
- Docker Compose usage for local testing
- Production deployment notes (API scaling, task runner singleton)
- Environment variable injection patterns

> Relevant existing files: [deployment.md]
> New files: []

## Completion Report and Documentation
Updated `deployment.md` with Docker build/run instructions, architecture diagram, service details table, and health check documentation.

---

# Final Review

## Summary

Successfully implemented Docker deployment configuration for the WeconnectU AI Server:

| File | Description |
|------|-------------|
| `src/task_runner/main.py` | Added `initialize_database()` for Agno tables + migrations on startup |
| `_deployment/API.Dockerfile` | Slim API image (no system deps) |
| `_deployment/TaskRunner.Dockerfile` | Task runner with LibreOffice + poppler |
| `.dockerignore` | Excludes unnecessary files from build context |
| `deployment.md` | Updated with Docker instructions and architecture docs |

## Key Design Decisions

1. **Separate Dockerfiles**: Each service has its own Dockerfile for independent builds and deployments
2. **API image is slim**: No LibreOffice/poppler needed, only 525MB
3. **Task runner handles migrations**: Singleton ensures no race conditions during DB initialization
4. **Agno tables first**: `get_agno_postgres_db()` creates Agno tables before custom migrations run
5. **AWS Secrets Manager**: Environment variables injected at runtime, no docker-compose needed
6. **libreoffice-writer only**: Smaller than full LibreOffice suite, sufficient for DOCX conversion

## Build Stats

| Image | Size |
|-------|------|
| `wcu-ai-api` | 525MB |
| `wcu-ai-task-runner` | 912MB |

## Testing

- Both Docker builds verified successfully
- Images run with correct Python path and dependencies
