# Migrate Model Costs from File to Database

`%% Status: Completed %%`

## Description

The current model cost system stores pricing data in a local JSON file (`.ai_costs.json`). This worked when both API and Task Runner ran on the same machine, but with containerized deployment the two services are isolated - the Task Runner updates the file, but the API containers can't access it.

### Current Architecture (Broken)

```
Task Runner Container          API Container(s)
┌─────────────────────┐       ┌─────────────────────┐
│ CostDataUpdateTask  │       │  API Server         │
│   writes to:        │       │   reads from:       │
│   .ai_costs.json    │       │   .ai_costs.json    │ ← FILE DOESN'T EXIST
└─────────────────────┘       └─────────────────────┘
```

### Target Architecture

```
Task Runner Container          API Container(s)
┌─────────────────────┐       ┌─────────────────────┐
│ CostDataUpdateTask  │       │  API Server         │
│   writes to:        │       │   reads from:       │
│   PostgreSQL ───────┼───────┼── PostgreSQL        │
└─────────────────────┘       └─────────────────────┘
                    ↓
              ┌──────────┐
              │ AWS RDS  │
              │ ai_costs │
              └──────────┘
```

### Data Structure

Current JSON structure to migrate:
```json
{
  "last_updated": "2026-01-07T10:00:00Z",
  "usd_to_zar": 18.45,
  "providers": {
    "openrouter": [
      {"model_id": "gpt-4", "input_cost_per_1m_tokens": 30.0, "output_cost_per_1m_tokens": 60.0},
      ...
    ],
    "anthropic": [...]
  }
}
```

Proposed database schema:
- `ai_cost_metadata` - single row with `last_updated` and `usd_to_zar`
- `ai_model_costs` - one row per model with provider, model_id, input/output costs

## Tasks

### Task 1: Create Database Migration

- [ ] Create migration for `ai_cost_metadata` table
- [ ] Create migration for `ai_model_costs` table
- [ ] Add appropriate indexes

#### Implementation Details

Create migration `YYYYMMDDHHMMSS_create_ai_costs_tables_UP.sql`:

```sql
CREATE TABLE ai_cost_metadata (
    id INTEGER PRIMARY KEY DEFAULT 1 CHECK (id = 1),
    last_updated TIMESTAMP WITH TIME ZONE NOT NULL,
    usd_to_zar DECIMAL(10, 4) NOT NULL
);

CREATE TABLE ai_model_costs (
    id BIGSERIAL PRIMARY KEY,
    provider VARCHAR(50) NOT NULL,
    model_id VARCHAR(255) NOT NULL,
    input_cost_per_1m_tokens DECIMAL(20, 10) NOT NULL,
    output_cost_per_1m_tokens DECIMAL(20, 10) NOT NULL,
    UNIQUE(provider, model_id)
);

CREATE INDEX idx_ai_model_costs_provider ON ai_model_costs(provider);
```

The `ai_cost_metadata` table uses `id = 1` constraint to ensure only one row exists (singleton pattern).

> Relevant existing files: [src/utils/migration_tool.py]
> New files: [migrations/YYYYMMDDHHMMSS_create_ai_costs_tables_UP.sql, migrations/YYYYMMDDHHMMSS_create_ai_costs_tables_DOWN.sql]

## Completion Report and Documentation
To be completed on task finalization

---

### Task 2: Update provider_costs.py for Database Access

- [ ] Replace `read_cost_data()` to query database instead of file
- [ ] Replace `write_cost_data()` to upsert database instead of file
- [ ] Keep TTL cache for read performance

#### Implementation Details

Update `src/utils/provider_costs.py`:

```python
from src.utils.db.connection import DBCursorCtx
from src.utils.models.model_costs import CostData, ModelCost
from src.utils.ttl_cache import ttl_cache


@ttl_cache(seconds=3600)
def read_cost_data() -> CostData:
    """Read cost data from database."""
    with DBCursorCtx() as cur:
        # Get metadata
        cur.execute("SELECT last_updated, usd_to_zar FROM ai_cost_metadata WHERE id = 1")
        row = cur.fetchone()
        if not row:
            raise ValueError("No cost data found in database")
        
        # Get model costs grouped by provider
        cur.execute("""
            SELECT provider, model_id, input_cost_per_1m_tokens, output_cost_per_1m_tokens 
            FROM ai_model_costs ORDER BY provider, model_id
        """)
        
        providers: dict[str, list[ModelCost]] = {}
        for row in cur.fetchall():
            provider = row["provider"]
            if provider not in providers:
                providers[provider] = []
            providers[provider].append(ModelCost(
                model_id=row["model_id"],
                input_cost_per_1m_tokens=float(row["input_cost_per_1m_tokens"]),
                output_cost_per_1m_tokens=float(row["output_cost_per_1m_tokens"]),
            ))
        
        return CostData(
            last_updated=metadata["last_updated"],
            usd_to_zar=float(metadata["usd_to_zar"]),
            providers=providers,
        )


def write_cost_data(cost_data: CostData) -> None:
    """Write cost data to database using upsert."""
    with DBCursorCtx() as cur:
        # Upsert metadata
        cur.execute("""
            INSERT INTO ai_cost_metadata (id, last_updated, usd_to_zar)
            VALUES (1, %(last_updated)s, %(usd_to_zar)s)
            ON CONFLICT (id) DO UPDATE SET
                last_updated = EXCLUDED.last_updated,
                usd_to_zar = EXCLUDED.usd_to_zar
        """, {"last_updated": cost_data.last_updated, "usd_to_zar": cost_data.usd_to_zar})
        
        # Clear and insert model costs (simpler than individual upserts)
        cur.execute("DELETE FROM ai_model_costs")
        for provider, models in cost_data.providers.items():
            for model in models:
                cur.execute("""
                    INSERT INTO ai_model_costs (provider, model_id, input_cost_per_1m_tokens, output_cost_per_1m_tokens)
                    VALUES (%(provider)s, %(model_id)s, %(input)s, %(output)s)
                """, {
                    "provider": provider,
                    "model_id": model.model_id,
                    "input": model.input_cost_per_1m_tokens,
                    "output": model.output_cost_per_1m_tokens,
                })
```

> Relevant existing files: [src/utils/provider_costs.py, src/utils/db/connection.py]
> New files: []

## Completion Report and Documentation
To be completed on task finalization

---

### Task 3: Update CostDataUpdateTask

- [ ] Update `_should_update()` to check database staleness instead of file age
- [ ] Remove `FileNotFoundError` handling (use `ValueError` for missing data)
- [ ] Keep staleness threshold (4 hours default, configurable via `MODEL_COST_STALENESS_SECONDS`)

#### Implementation Details

The task already uses `read_cost_data()` and `write_cost_data()` - minimal changes needed since we're updating the underlying implementation.

Update `_should_update()`:
```python
def _should_update(self) -> bool:
    """Check if cost data needs updating based on staleness threshold"""
    try:
        cost_data = read_cost_data()
        now = datetime.now(timezone.utc)
        age_seconds = (now - cost_data.last_updated).total_seconds()
        return age_seconds > ENV_SETTINGS.model_cost_staleness_seconds
    except ValueError:
        # No data in database yet
        return True
```

Keep `MODEL_COST_STALENESS_SECONDS` in env_settings (rename from `MODEL_COST_FILE_AGE_MAX_SECONDS`).

> Relevant existing files: [src/task_runner/tasks/periodic_tasks/cost_data_update_task.py]
> New files: []

## Completion Report and Documentation
To be completed on task finalization

---

### Task 4: Clean Up Environment Variables and Files

- [ ] Remove `MODEL_COST_FILE_PATH` from env_settings.py
- [ ] Rename `MODEL_COST_FILE_AGE_MAX_SECONDS` to `MODEL_COST_STALENESS_SECONDS`
- [ ] Update .env.example
- [ ] Delete .ai_costs.json
- [ ] Add .ai_costs.json to .gitignore (in case of local dev artifacts)

#### Implementation Details

Changes to `env_settings.py`:
- Remove `model_cost_file_path: str`
- Rename `model_cost_file_age_max_seconds` → `model_cost_staleness_seconds`
- Keep `providers_in_use` as it's still needed

> Relevant existing files: [env_settings.py, .env.example, .gitignore]
> New files: []

## Completion Report and Documentation
To be completed on task finalization

---

### Task 5: Create Local Dev Docker Script

- [ ] Create `_deployment/run-local.sh` script
- [ ] Support `start`, `stop`, `logs` commands
- [ ] Load environment from `.env`
- [ ] Start both containers with proper networking

#### Implementation Details

Create `_deployment/run-local.sh`:

```bash
#!/bin/bash
set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"
NETWORK_NAME="wcu-ai-network"

cd "$PROJECT_DIR"

start() {
    if [ ! -f .env ]; then
        echo "Error: .env file not found"
        exit 1
    fi

    echo "Building images..."
    docker build -f _deployment/API.Dockerfile -t wcu-ai-api .
    docker build -f _deployment/TaskRunner.Dockerfile -t wcu-ai-task-runner .

    docker network create $NETWORK_NAME 2>/dev/null || true

    # Stop existing containers
    docker rm -f wcu-ai-api wcu-ai-task-runner 2>/dev/null || true

    echo "Starting task runner..."
    docker run -d \
        --name wcu-ai-task-runner \
        --network $NETWORK_NAME \
        --env-file .env \
        wcu-ai-task-runner

    sleep 3

    echo "Starting API server..."
    docker run -d \
        --name wcu-ai-api \
        --network $NETWORK_NAME \
        --env-file .env \
        -p 8000:8000 \
        wcu-ai-api

    echo ""
    echo "Services started!"
    echo "API: http://localhost:8000"
    echo "Logs: $0 logs"
}

stop() {
    echo "Stopping containers..."
    docker rm -f wcu-ai-api wcu-ai-task-runner 2>/dev/null || true
    echo "Stopped."
}

logs() {
    docker logs -f wcu-ai-api
}

case "${1:-start}" in
    start) start ;;
    stop) stop ;;
    logs) logs ;;
    *) echo "Usage: $0 {start|stop|logs}"; exit 1 ;;
esac
```

Usage:
- `./run-local.sh` or `./run-local.sh start` - Build and start containers
- `./run-local.sh stop` - Stop and remove containers
- `./run-local.sh logs` - Follow API logs

> Relevant existing files: []
> New files: [_deployment/run-local.sh]

## Completion Report and Documentation
To be completed on task finalization

---

# Final Review

To be completed on spec finalization
