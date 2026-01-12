# Work Log - Docker Local Dev Script Improvements

## Spec File: `agent_rules/spec/s_20260107_benjamin_van_heerden__model_costs_to_database.md`

## Overarching Goals

Improve the `docker-run-local.sh` script for local Docker development, fixing several issues discovered during testing:
1. Database connectivity between containers
2. PostgreSQL volume compatibility with pg18
3. Hot reloading for API development
4. Agno table initialization before migrations
5. Foreground process with graceful shutdown

## What Was Accomplished

### Fixed Container-to-Container Database Connectivity

The containers were failing to connect to PostgreSQL because `DB_HOST` in `.env` was set to `host.docker.internal`. Added explicit `-e DB_HOST=$DB_CONTAINER` override when starting containers so they use the container name for inter-container networking:

```bash
docker run -d \
    --name wcu-ai-api \
    --network $NETWORK_NAME \
    --env-file .env \
    -e DB_HOST=$DB_CONTAINER \
    ...
```

### Fixed PostgreSQL 18 Volume Mount

pg18 Docker images use a different data directory structure. Changed volume mount from `/var/lib/postgresql/data` to `/var/lib/postgresql`:

```bash
-v wcu-ai-postgres-data:/var/lib/postgresql \
```

### Added Hot Reloading for API Development

Mounted local source files into the API container so uvicorn's hot reload works:

```bash
-v "$SCRIPT_DIR/src:/app/src" \
-v "$SCRIPT_DIR/env_settings.py:/app/env_settings.py" \
```

### Fixed Agno Table Initialization

Migrations were failing because `ai_sessions` table didn't exist. The `get_agno_postgres_db()` call wasn't creating tables - needed to explicitly call `_create_all_tables()`:

```python
def initialize_database():
    logger.info("Initializing Agno database tables...")
    db = get_agno_postgres_db()
    db._create_all_tables()
    logger.info("✓ Agno tables initialized")

    logger.info("Running database migrations...")
    MigrationTool().run_up()
```

### Made Script Run in Foreground with Graceful Shutdown

Changed the script to stay in foreground after starting containers, with Ctrl+C triggering cleanup:

```bash
trap 'echo ""; echo "Stopping containers..."; docker rm -f wcu-ai-api wcu-ai-task-runner $DB_CONTAINER 2>/dev/null; echo "Stopped."; exit 0' INT TERM

while true; do sleep 1; done
```

### Renamed and Added Log Commands

- Renamed `logs-task` to `logs-task-runner` for clarity
- Added `logs-database` command for PostgreSQL logs

## Key Files Affected

| File | Change |
|------|--------|
| `docker-run-local.sh` | Fixed DB_HOST override, pg18 volume path, added source mounts for hot reload, foreground process with trap, renamed/added log commands |
| `src/task_runner/main.py` | Added `db._create_all_tables()` call before running migrations |
| `README.md` | Updated documentation for new script behavior |

## What Comes Next

The model costs migration spec is complete. The Docker local development workflow is now fully functional:

- `./docker-run-local.sh start` - Starts all services, stays in foreground
- `./docker-run-local.sh stop` - Manual cleanup if needed
- `./docker-run-local.sh logs-api` - View API logs
- `./docker-run-local.sh logs-task-runner` - View task runner logs
- `./docker-run-local.sh logs-database` - View PostgreSQL logs

Hot reloading works for the API. Changes to `src/` are reflected immediately without rebuilding the image.
