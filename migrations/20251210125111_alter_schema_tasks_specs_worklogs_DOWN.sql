-- Migration Rollback: alter_schema_tasks_specs_worklogs
-- Description: Reverse changes

BEGIN TRANSACTION;

ALTER TABLE work_logs RENAME TO work_logs_old;
CREATE TABLE work_logs (
    id INTEGER PRIMARY KEY,
    file_path TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
INSERT INTO work_logs (id, file_path, created_at)
SELECT id, file_path, created_at FROM work_logs_old;
DROP TABLE work_logs_old;
CREATE INDEX idx_work_logs_created_at ON work_logs(created_at);

ALTER TABLE specs RENAME TO specs_old;
CREATE TABLE specs (
    id INTEGER PRIMARY KEY,
    title TEXT NOT NULL,
    file_path TEXT NOT NULL,
    status TEXT DEFAULT 'active',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    completed_at TIMESTAMP
);
INSERT INTO specs (id, title, file_path, status, created_at, updated_at, completed_at)
SELECT id, title, file_path, status, created_at, updated_at, completed_at FROM specs_old;
DROP TABLE specs_old;
CREATE INDEX idx_specs_updated_at ON specs(updated_at);

ALTER TABLE tasks RENAME TO tasks_old;
CREATE TABLE tasks (
    id INTEGER PRIMARY KEY,
    spec_id INTEGER,
    parent_id INTEGER,
    title TEXT NOT NULL,
    file_path TEXT NOT NULL DEFAULT '',
    status TEXT DEFAULT 'todo',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    completed_at TIMESTAMP,
    FOREIGN KEY (spec_id) REFERENCES specs(id) ON DELETE CASCADE,
    FOREIGN KEY (parent_id) REFERENCES tasks(id) ON DELETE CASCADE
);
INSERT INTO tasks (id, spec_id, parent_id, title, file_path, status, created_at, updated_at, completed_at)
SELECT id, spec_id, parent_id, title, '', status, created_at, updated_at, completed_at FROM tasks_old;
DROP TABLE tasks_old;
CREATE INDEX idx_tasks_updated_at ON tasks(updated_at);
CREATE INDEX idx_tasks_spec_id ON tasks(spec_id);
CREATE INDEX idx_tasks_parent_id ON tasks(parent_id);

COMMIT;