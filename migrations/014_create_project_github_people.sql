-- Migration: 014_create_project_github_people.sql
-- Date: 2025-01-28
-- Description: Create table to track which GitHub people are associated with which projects

CREATE TABLE IF NOT EXISTS project_github_people (
    id TEXT PRIMARY KEY,
    project_id TEXT NOT NULL,
    github_person_id TEXT NOT NULL,
    source_type TEXT NOT NULL, -- "pull_request", "contributor", "commit_author"
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    is_deleted INTEGER DEFAULT 0,
    FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE,
    FOREIGN KEY (github_person_id) REFERENCES github_people(id) ON DELETE CASCADE,
    UNIQUE(project_id, github_person_id)
);

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_project_github_people_project_id ON project_github_people(project_id);
CREATE INDEX IF NOT EXISTS idx_project_github_people_github_person_id ON project_github_people(github_person_id);
CREATE INDEX IF NOT EXISTS idx_project_github_people_source_type ON project_github_people(source_type);
CREATE INDEX IF NOT EXISTS idx_project_github_people_is_deleted ON project_github_people(is_deleted);

-- Trigger to update updated_at timestamp
CREATE TRIGGER IF NOT EXISTS update_project_github_people_updated_at
    AFTER UPDATE ON project_github_people
    FOR EACH ROW
BEGIN
    UPDATE project_github_people SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END; 
