# Work Log - Docker Deployment and Model Costs Migration Spec

## Spec Files
- `agent_rules/spec/s_20260107_benjamin_van_heerden__docker_deployment.md` (Completed)
- `agent_rules/spec/s_20260107_benjamin_van_heerden__model_costs_to_database.md` (Draft)

## Overarching Goals

1. Create Docker deployment configuration for the WeconnectU AI Server with separate images for API and Task Runner services
2. Identify and plan fix for model costs file-sharing issue between isolated containers

## What Was Accomplished

### Docker Deployment (Completed)

Created separate Dockerfiles for independent service deployment:

**`_deployment/API.Dockerfile`** (525MB):
- Multi-stage build with uv for dependency management
- Python 3.12-slim-bookworm base
- No system dependencies (lightweight)
- Exposes port 8000

**`_deployment/TaskRunner.Dockerfile`** (912MB):
- Same base as API
- Includes `libreoffice-writer` and `poppler-utils` for document processing
- Copies migrations directory
- Runs database initialization on startup

### Task Runner Database Initialization

Added `initialize_database()` to `src/task_runner/main.py`:
```python
def initialize_database():
    """Initialize database tables and run migrations"""
    logger.info("Initializing Agno database tables...")
    get_agno_postgres_db()
    logger.info("✓ Agno tables initialized")

    logger.info("Running database migrations...")
    MigrationTool().run_up()
```

This ensures Agno tables are created before custom migrations run.

### Updated deployment.md

Rewrote deployment documentation to reflect:
- Two separate Dockerfiles in `_deployment/`
- Build commands for each image
- Architecture diagram
- AWS Secrets Manager for env vars (no docker-compose)

### Model Costs Migration Spec (Draft)

Identified issue: `.ai_costs.json` file can't be shared between isolated containers. Created spec with 5 tasks:

1. Create database migration for `ai_cost_metadata` and `ai_model_costs` tables
2. Update `provider_costs.py` to read/write from database
3. Update `CostDataUpdateTask` for database-backed staleness checks
4. Clean up env vars (remove `MODEL_COST_FILE_PATH`, rename `MODEL_COST_FILE_AGE_MAX_SECONDS`)
5. Create `_deployment/run-local.sh` script with start/stop/logs commands

## Key Files Affected

| File | Change |
|------|--------|
| `_deployment/API.Dockerfile` | NEW - Slim API image |
| `_deployment/TaskRunner.Dockerfile` | NEW - Task runner with system deps |
| `.dockerignore` | NEW - Excludes tests, frontend, docs from build context |
| `src/task_runner/main.py` | Added `initialize_database()` for Agno + migrations |
| `deployment.md` | Updated with Docker build instructions |
| `agent_rules/spec/s_20260107_benjamin_van_heerden__docker_deployment.md` | NEW - Completed spec |
| `agent_rules/spec/s_20260107_benjamin_van_heerden__model_costs_to_database.md` | NEW - Draft spec |

Removed:
- `Dockerfile` (replaced by separate Dockerfiles)
- `docker-compose.yaml` (not needed for AWS deployment)

## What Comes Next

The model costs migration spec is ready for implementation:

**Spec file:** `agent_rules/spec/s_20260107_benjamin_van_heerden__model_costs_to_database.md`

**Tasks to implement:**
1. Task 1: Create database migration for cost tables
2. Task 2: Update `provider_costs.py` for database access
3. Task 3: Update `CostDataUpdateTask` 
4. Task 4: Clean up environment variables and delete `.ai_costs.json`
5. Task 5: Create `_deployment/run-local.sh` script

After completing the spec, the Docker deployment will be fully functional with both services able to share cost data via the database.
