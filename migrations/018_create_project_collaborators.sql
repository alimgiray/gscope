-- Migration: 018_create_project_collaborators.sql
-- Date: 2024-12-19
-- Description: Create project_collaborators table for managing project access

CREATE TABLE IF NOT EXISTS project_collaborators (
    id TEXT PRIMARY KEY,
    project_id TEXT NOT NULL,
    user_id TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    UNIQUE(project_id, user_id)
);

-- Create index for faster lookups
CREATE INDEX IF NOT EXISTS idx_project_collaborators_project_id ON project_collaborators(project_id);
CREATE INDEX IF NOT EXISTS idx_project_collaborators_user_id ON project_collaborators(user_id); 