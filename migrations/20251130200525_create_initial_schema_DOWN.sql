-- Migration Rollback: create_initial_schema
-- Description: Rollback for create_initial_schema migration

-- Drop indexes first
DROP INDEX IF EXISTS idx_work_logs_created_at;
DROP INDEX IF EXISTS idx_todos_status;
DROP INDEX IF EXISTS idx_tasks_parent_id;
DROP INDEX IF EXISTS idx_tasks_spec_id;
DROP INDEX IF EXISTS idx_tasks_updated_at;
DROP INDEX IF EXISTS idx_specs_updated_at;

-- Then drop tables
DROP TABLE IF EXISTS work_logs;
DROP TABLE IF EXISTS todos;
DROP TABLE IF EXISTS tasks;
DROP TABLE IF EXISTS specs;
