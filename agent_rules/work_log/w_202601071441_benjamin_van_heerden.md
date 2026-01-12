# Work Log - Model Costs Migration and Docker Local Dev Script

## Spec File: `agent_rules/spec/s_20260107_benjamin_van_heerden__model_costs_to_database.md`

## Overarching Goals

Migrate the model cost system from file-based storage (`.ai_costs.json`) to PostgreSQL database tables, enabling cost data sharing between isolated Docker containers (API and Task Runner). Additionally, create a unified local development script for running the full stack with Docker.

## What Was Accomplished

### Database Migration

Created migration `20260107124834_create_ai_costs_tables` with:
- `ai_cost_metadata` table - singleton row with `last_updated` and `usd_to_zar`
- `ai_model_costs` table - one row per provider/model with pricing data
- Index on `provider` column for efficient lookups

### Updated provider_costs.py

Replaced file-based read/write with database operations:
- `read_cost_data()` - queries both tables, returns `CostData` object
- `write_cost_data()` - upserts metadata, batch inserts model costs using `executemany`

### Updated CostDataUpdateTask

- Changed `FileNotFoundError` handling to `ValueError` 
- Renamed setting reference from `model_cost_file_age_max_seconds` to `model_cost_staleness_seconds`
- Updated log messages to reflect database storage

### Environment Variable Cleanup

- Removed `MODEL_COST_FILE_PATH` from `env_settings.py`
- Renamed `MODEL_COST_FILE_AGE_MAX_SECONDS` to `MODEL_COST_STALENESS_SECONDS`
- Deleted `.ai_costs.json` file
- Added `.ai_costs.json` to `.gitignore`

### Docker Local Development Script

Created `docker-run-local.sh` in project root with commands:
- `start` - Creates network, starts PostgreSQL (if needed), builds images, starts task runner and API
- `stop` - Stops app containers, keeps database running
- `stop-all` - Stops everything including database
- `logs-api` / `logs-task` - Follow container logs

PostgreSQL container uses named volume `wcu-ai-postgres-data` for data persistence.

### .env File Format Fix

Fixed `.env` and `.env.example` to work with Docker's `--env-file`:
- Removed all quotes around values (Docker passes them literally)
- Changed `DB_HOST=host.docker.internal` for Docker-to-host connectivity
- Consistent structure between both files

### Documentation

- Moved deployment docs to `_deployment/README.md`
- Updated architecture diagram to correctly show API and Task Runner both connecting to PostgreSQL
- Documented `docker-run-local.sh` usage
- Removed `dev.docker-compose.yaml` dependency (PostgreSQL now managed by script)

## Key Files Affected

| File | Change |
|------|--------|
| `migrations/20260107124834_create_ai_costs_tables_UP.sql` | NEW - Creates cost tables |
| `migrations/20260107124834_create_ai_costs_tables_DOWN.sql` | NEW - Drops cost tables |
| `src/utils/provider_costs.py` | Replaced file I/O with database queries |
| `src/task_runner/tasks/periodic_tasks/cost_data_update_task.py` | Updated exception handling and setting name |
| `env_settings.py` | Removed `model_cost_file_path`, renamed staleness setting |
| `.env` | Removed quotes, updated for Docker compatibility |
| `.env.example` | Removed quotes, cleaned up structure |
| `.gitignore` | Added `.ai_costs.json` |
| `docker-run-local.sh` | NEW - Unified local dev script |
| `_deployment/README.md` | NEW - Moved deployment docs here |
| `deployment.md` | Updated with local Docker instructions |

Deleted:
- `.ai_costs.json`
- `_deployment/run-local.sh` (moved to project root as `docker-run-local.sh`)

## What Comes Next

The model costs migration spec is now complete. All 5 tasks have been implemented:

1. Database migration - Done
2. provider_costs.py update - Done
3. CostDataUpdateTask update - Done
4. Environment cleanup - Done
5. Local dev script - Done (enhanced beyond original spec)

The spec file should be updated to mark status as `Completed`.

Optional follow-ups:
- Delete `dev.docker-compose.yaml` (no longer needed)
- Test the full Docker stack end-to-end with `./docker-run-local.sh start`
