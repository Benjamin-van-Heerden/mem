# Work Log - Model Cost Scraping Periodic Task Implementation

## Overarching Goals

Implement a periodic background task to automatically scrape and cache model token pricing data from external sources. The task fetches USD/ZAR exchange rates and model costs from multiple AI providers, storing them in a local JSON file for use throughout the application. This provides up-to-date pricing information without requiring manual updates or hitting external APIs on every request.

## What Was Accomplished

### 1. Created Cost Data Models

**Created `src/utils/models/model_costs.py`:**
- `ModelCost` - Pydantic model for individual model pricing
  - `model_id` - Model identifier (e.g., 'gpt-4')
  - `input_cost_per_1m_tokens` - Input token cost in USD
  - `output_cost_per_1m_tokens` - Output token cost in USD
- `CostData` - Complete cost data structure
  - `last_updated` - Timestamp for staleness checking
  - `usd_to_zar` - Exchange rate
  - `providers` - Dict mapping provider names to their model costs

### 2. Created Provider Cost Utilities

**Created `src/utils/provider_costs.py`:**
- `read_cost_data()` - Reads and validates cost data from configured JSON file
  - Uses `CostData.model_validate_json()` for type safety
  - Raises `FileNotFoundError` if file doesn't exist
  - Raises `ValueError` if JSON is invalid or doesn't match schema
- `write_cost_data(cost_data)` - Writes validated cost data to configured JSON file
  - Uses `cost_data.model_dump_json(indent=2)` for pretty formatting
  - Atomic write operation

Both functions use `ENV_SETTINGS.model_cost_file_path` internally, removing the need to pass file paths around.

### 3. Implemented Cost Data Update Task

**Created `src/task_runner/tasks/cost_data_update_task.py`:**

**Task Configuration:**
- Runs every 3600 seconds (1 hour) - hardcoded interval
- Uses 💰 emoji for log messages
- Integrates with `BasePeriodicTask` abstraction

**Update Logic:**
- Only updates if file is older than configured threshold (default: 4 hours)
- Backs up existing data before updating
- Restores backup if update fails
- Validates data structure before writing

**Data Fetching:**
- `_fetch_exchange_rate()` - Uses yfinance to get USD/ZAR rate from `USDZAR=X` ticker
- `_fetch_model_costs_for_provider()` - Scrapes models.dev API for provider pricing
  - Fetches all models for configured providers
  - Filters models that have cost data
  - Extracts input/output token costs per 1M tokens

**Error Handling:**
- Backup/restore pattern prevents data loss
- Per-provider error handling prevents one failure from stopping entire update
- Comprehensive logging at each step

### 4. Environment Configuration

**Updated `env_settings.py`:**
- `model_cost_file_age_max_seconds` - Max file age before update (default: 14400s = 4 hours)
- `model_cost_file_path` - Path to cost data JSON file (default: `.ai_costs.json`)
- `providers_in_use` - List of provider names to fetch (e.g., `["openrouter", "anthropic"]`)

**Updated `.env`:**
```
MODEL_COST_FILE_AGE_MAX_SECONDS=14400
MODEL_COST_FILE_PATH=".ai_costs.json"
PROVIDERS_IN_USE='["openrouter","anthropic"]'
```

### 5. Task Registration

**Updated `src/task_runner/config.py`:**
- Added `CostDataUpdateTask` to `PERIODIC_TASKS` list
- Task will automatically start when task runner boots

### 6. Dependencies

Dependencies were already properly installed via uv:
- `httpx>=0.28.1` - Async HTTP client for API requests
- `yfinance>=0.2.66` - Yahoo Finance API for exchange rates

## Key Files Affected

**Created:**
- `src/utils/models/model_costs.py` (27 lines) - Pydantic models for cost data
- `src/utils/provider_costs.py` (43 lines) - Read/write utilities with Pydantic validation
- `src/task_runner/tasks/cost_data_update_task.py` (217 lines) - Complete periodic task implementation

**Modified:**
- `env_settings.py` - Added cost data configuration settings
- `src/task_runner/config.py` - Registered CostDataUpdateTask
- `.env` - Added cost data environment variables

## What Comes Next

### Immediate Next Steps

1. **WCU API Integration Migration:**
   - Migrate WeconnectU API helper functions from legacy codebase
   - Authentication utilities for JWT token handling
   - API client functions for meeting data retrieval
   - Reusable patterns for authenticated API calls

2. **Testing Infrastructure:**
   - **End-to-end task runner tests** - Spin up task runner instance, submit tasks via API, poll for completion
   - **Database task tests** - Submit via POST /tasks, verify processing, check success/error states
   - **Periodic task tests** - Verify tasks execute on schedule, handle errors gracefully
   - **API integration tests** - FastAPI TestClient + task runner interplay
   - **Connection tests** - Database connectivity, API authentication flows

3. **Test Environment Setup:**
   - Configure test database (Docker container or similar)
   - Mock external API dependencies (models.dev, yfinance)
   - Test fixtures for task payloads and expected results
   - Cleanup strategies for test data

### Architecture Notes

The cost data update task demonstrates the clean separation of concerns in the periodic task system:
- **Models** - Type-safe data structures with validation
- **Utilities** - Reusable read/write functions with consistent interface
- **Task** - Business logic only, no infrastructure concerns
- **Configuration** - Environment-driven behavior

This pattern will be applied to future tasks and can serve as a reference implementation.

Infrastructure Foundations spec remains **COMPLETED** with production-ready periodic and database task systems.