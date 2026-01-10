# Work Log - Add Database-Only Mode to Docker Script

## Overarching Goals

Add a `db` command to `docker-run-local.sh` to start only PostgreSQL for running tests locally, and ensure all 101 tests pass.

## What Was Accomplished

### Added `db` Command to docker-run-local.sh

Added a new `db` command that starts only the PostgreSQL container, allowing local Python processes (tests, API server, task runner) to connect via `localhost:5432`:

```bash
db() {
    docker network create $NETWORK_NAME 2>/dev/null || true

    if ! docker ps --format '{{.Names}}' | grep -q "^${DB_CONTAINER}$"; then
        echo "Starting PostgreSQL..."
        docker run -d \
            --name $DB_CONTAINER \
            --network $NETWORK_NAME \
            -e POSTGRES_USER=root \
            -e POSTGRES_PASSWORD=password \
            -e POSTGRES_DB=database \
            -p 5432:5432 \
            -v wcu-ai-postgres-data:/var/lib/postgresql \
            pgvector/pgvector:pg18-trixie
        # ...
    fi
    # Runs in foreground with Ctrl+C cleanup
}
```

### Workflow Clarification

The script now supports three main workflows:

1. **Full stack development**: `./docker-run-local.sh start`
   - Starts PostgreSQL, API (with hot reload), and Task Runner in Docker
   - API connects to DB via container name override (`-e DB_HOST=wcu-ai-postgres`)

2. **Local development/testing**: `./docker-run-local.sh db`
   - Starts only PostgreSQL
   - Local processes use `DB_HOST=localhost` from `.env`
   - Tests spawn their own task runner subprocess

3. **Cleanup**: `./docker-run-local.sh stop`

### Fixed DB_HOST Configuration

The `.env` file should use `DB_HOST=localhost` for local development. The `docker-run-local.sh` script overrides this to the container name when running in Docker, so both scenarios work:

- Local: `.env` has `DB_HOST=localhost`, connects directly
- Docker: Script passes `-e DB_HOST=wcu-ai-postgres`, overriding `.env`

### Updated README

Added documentation for running tests:

```bash
# In one terminal, start the database
./docker-run-local.sh db

# In another terminal, run tests
uv run pytest tests/
```

## Key Files Affected

| File | Change |
|------|--------|
| `docker-run-local.sh` | Added `db` command for database-only mode |
| `README.md` | Added test running documentation |

## What Comes Next

All 101 tests pass. The local development workflow is now complete:

- `./docker-run-local.sh start` for full stack Docker development
- `./docker-run-local.sh db` for local Python development and testing
- Tests can run with `uv run pytest tests/` while `db` is running

Remember to use `DB_HOST=localhost` in `.env` for local development.
