-- Migration: alter_schema_tasks_specs_worklogs
-- Description: 
-- - Recreate tasks table: remove file_path, add detail TEXT
-- - Add assigned_to TEXT and branch TEXT to specs
-- - Add spec_id INTEGER REFERENCES specs(id) ON DELETE SET NULL to work_logs

BEGIN TRANSACTION;

CREATE TABLE tasks_new (
    id INTEGER PRIMARY KEY,
    spec_id INTEGER,
    parent_id INTEGER,
    title TEXT NOT NULL,
    detail TEXT,
    status TEXT DEFAULT 'todo',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    completed_at TIMESTAMP,
    FOREIGN KEY (spec_id) REFERENCES specs(id) ON DELETE CASCADE,
    FOREIGN KEY (parent_id) REFERENCES tasks(id) ON DELETE CASCADE
);

INSERT INTO tasks_new (id, spec_id, parent_id, title, detail, status, created_at, updated_at, completed_at)
SELECT id, spec_id, parent_id, title, NULL, status, created_at, updated_at, completed_at FROM tasks;

DROP TABLE tasks;

ALTER TABLE tasks_new RENAME TO tasks;

CREATE INDEX idx_tasks_updated_at ON tasks(updated_at);
CREATE INDEX idx_tasks_spec_id ON tasks(spec_id);
CREATE INDEX idx_tasks_parent_id ON tasks(parent_id);

ALTER TABLE specs ADD COLUMN assigned_to TEXT;
ALTER TABLE specs ADD COLUMN branch TEXT;

ALTER TABLE work_logs ADD COLUMN spec_id INTEGER REFERENCES specs(id) ON DELETE SET NULL;

COMMIT;