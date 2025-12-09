-- Migration: create_initial_schema
-- Description: Create initial tables for specs, tasks, todos, and work_logs

-- Specs
CREATE TABLE specs (
    id INTEGER PRIMARY KEY,
    title TEXT NOT NULL,
    file_path TEXT NOT NULL,
    status TEXT DEFAULT 'active', -- active, completed, archived
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    completed_at TIMESTAMP
);
CREATE INDEX idx_specs_updated_at ON specs(updated_at);

-- Tasks (and subtasks)
CREATE TABLE tasks (
    id INTEGER PRIMARY KEY,
    spec_id INTEGER,
    parent_id INTEGER, -- NULL for tasks, set for subtasks
    title TEXT NOT NULL,
    file_path TEXT NOT NULL,
    status TEXT DEFAULT 'todo', -- todo, in_progress, blocked, completed
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    completed_at TIMESTAMP,
    FOREIGN KEY (spec_id) REFERENCES specs(id) ON DELETE CASCADE,
    FOREIGN KEY (parent_id) REFERENCES tasks(id) ON DELETE CASCADE
);
CREATE INDEX idx_tasks_updated_at ON tasks(updated_at);
CREATE INDEX idx_tasks_spec_id ON tasks(spec_id);
CREATE INDEX idx_tasks_parent_id ON tasks(parent_id);

-- Todos
CREATE TABLE todos (
    id INTEGER PRIMARY KEY,
    title TEXT NOT NULL,
    description TEXT,
    status TEXT DEFAULT 'open', -- open, completed
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    completed_at TIMESTAMP
);
CREATE INDEX idx_todos_status ON todos(status);

-- Work Logs
CREATE TABLE work_logs (
    id INTEGER PRIMARY KEY,
    file_path TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_work_logs_created_at ON work_logs(created_at);
