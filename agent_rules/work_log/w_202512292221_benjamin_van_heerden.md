# Work Log - MySQL to PostgreSQL Migration and Offenses Agent Setup

## Overarching Goals

1. Migrate the database layer from MySQL to PostgreSQL for the AI Server
2. Begin implementing the Offenses Agent - a new agent responsible for handling community rule violations and offense-related queries

## What Was Accomplished

### PostgreSQL Migration

Completed a full migration from MySQL to PostgreSQL 18.1:

**Docker Compose (`dev.docker-compose.yaml`):**
- Changed image from `mysql:8.0` to `postgres:18.1`
- Updated environment variables for PostgreSQL (`POSTGRES_USER`, `POSTGRES_PASSWORD`, `POSTGRES_DB`)
- Changed volume mount from `/var/lib/postgresql/data` to `/var/lib/postgresql` (required for PostgreSQL 18+ directory structure)
- Updated healthcheck to use `pg_isready`

**Database Connection (`src/utils/db/connection.py`):**
- Replaced `mysql.connector` with `psycopg`
- Changed `mysql_db_url()` to `postgres_db_url()` with SQLAlchemy-compatible format
- Updated `DBCursorCtx` to use psycopg's sync and async connections
- Fixed row factory to use `dict_row` function (not `DictRow` class)

**Agno Database (`src/utils/agno/db.py`):**
- Changed from `MySQLDb` to `PostgresDb`
- Renamed `get_agno_mysql_db()` to `get_agno_postgres_db()`
- Updated parameter from `db_schema` to `schema`

**Migration Tool (`src/utils/migration_tool.py`):**
- Updated template example to use `BIGSERIAL` instead of `BIGINT AUTO_INCREMENT`
- Added `# type: ignore[arg-type]` for dynamic SQL execution (psycopg requires `LiteralString` at compile time)

**Migration Files:**
- `20251201185017_create_ai_server_tasks_table_UP.sql`: Converted to PostgreSQL syntax
  - `BIGINT AUTO_INCREMENT` → `BIGSERIAL`
  - `JSON` → `JSONB`
  - `ENUM(...)` → `CHECK` constraint
  - Separate `CREATE INDEX` statements
- `20251229155552_add_ai_sessions_indexes_UP.sql`: Already PostgreSQL-compatible

**Updated References (6 files):**
- `src/agents/core.py`
- `src/teams/core.py`
- `src/api/routes/agents/ops.py`
- `src/api/routes/teams/ops.py`
- `src/api/routes/utils/agno_sessions.py`
- `.env.example` (changed default port to 5432)

### Insomnia API Collection

Created `insomnia_wcu_api.json` - a complete Insomnia export with:

**Folders:**
- Auth (Login, Validate Token)
- Communities (List Reseller Communities)
- Meetings (List Meetings, Get Agenda, Get Participants)
- Offenses (11 endpoints for offense management)

**Features:**
- Environment variables for `host`, `jwt`, `reseller_uid`, `community_uid`
- Pre-configured Bearer auth on all authenticated requests
- Added `User-Agent` header to Login request (required by WCU API)

### Offenses Module (In Progress)

Started implementing the offenses module following existing patterns:

**Created `src/utils/models/wcu_offense.py`:**
```python
class Offense(BaseModel):
    offense_id: int
    community_id: int
    unit_id: int | None = None
    category1: str | None = None
    rule1: str | None = None
    offense_type: int | None = None
    offense_type_name: str | None = None
    fine_amount: str | None = None
    # ... (30+ fields from API response)

class OffenseListMeta(BaseModel):
    count: int
    page: int
    total_count: int

class OffenseListResponse(BaseModel):
    offenses: list[Offense]
    meta: OffenseListMeta
```

**Created `src/utils/weconnectu/offenses.py`:**
- `wcu_list_offenses_reseller()` - List offenses with pagination/sorting
- `wcu_list_offenses_community()` - List offenses for a community
- `wcu_get_offense()` - Get single offense details

## Key Files Affected

| File | Change |
|------|--------|
| `dev.docker-compose.yaml` | MySQL → PostgreSQL 18.1, updated volume mount path |
| `src/utils/db/connection.py` | mysql.connector → psycopg, dict_row factory |
| `src/utils/agno/db.py` | MySQLDb → PostgresDb |
| `src/utils/migration_tool.py` | Updated template, added type ignores |
| `migrations/20251201185017_create_ai_server_tasks_table_UP.sql` | PostgreSQL syntax |
| `src/agents/core.py` | get_agno_postgres_db import |
| `src/teams/core.py` | get_agno_postgres_db import |
| `src/api/routes/agents/ops.py` | get_agno_postgres_db import |
| `src/api/routes/teams/ops.py` | get_agno_postgres_db import |
| `src/api/routes/utils/agno_sessions.py` | get_agno_postgres_db import |
| `.env.example` | DB_PORT=5432 |
| `insomnia_wcu_api.json` | NEW - Complete API collection |
| `src/utils/models/wcu_offense.py` | NEW - Offense Pydantic models |
| `src/utils/weconnectu/offenses.py` | NEW - Offense API functions |

## What Comes Next

### Immediate Next Steps (Offenses Agent)

1. **Complete offenses.py functions** - Test and implement remaining endpoints:
   - `wcu_list_offenses_unit()` - Offenses per unit
   - `wcu_get_my_offenses()` - Current user's offenses
   - `wcu_get_offense_setups()` - Offense rules/configuration for a community

2. **Create the Offenses Agent** following the `BaseAgent` pattern:
   - `src/agents/offenses/agent.py` - OffensesAgent class
   - `src/agents/offenses/tools.py` - Tool factory functions
   - Register in `src/agents/config.py`

3. **Key questions for the agent**:
   - "Does X constitute an offense?"
   - "What are the rules for this community?"
   - "List recent offenses for unit Y"
   - "What's the fine for parking violations?"

### Testing Workflow

Use Insomnia to test each endpoint and provide the response shape:
1. Run request in Insomnia
2. Share response structure
3. Update models/functions accordingly

The `Offense` model is based on `list_offenses_reseller` response - may need adjustment for other endpoints.
