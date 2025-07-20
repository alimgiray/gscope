CREATE TABLE IF NOT EXISTS score_settings (
    id TEXT PRIMARY KEY,
    project_id TEXT NOT NULL,
    additions INTEGER DEFAULT 1,
    deletions INTEGER DEFAULT 3,
    commits INTEGER DEFAULT 10,
    pull_requests INTEGER DEFAULT 20,
    comments INTEGER DEFAULT 100,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE
);

-- Create unique constraint to ensure one score_settings per project
CREATE UNIQUE INDEX IF NOT EXISTS idx_score_settings_project_unique 
ON score_settings(project_id);

-- Create index on project_id for faster queries
CREATE INDEX IF NOT EXISTS idx_score_settings_project_id 
ON score_settings(project_id); 

-- Create trigger to update updated_at timestamp
CREATE TRIGGER IF NOT EXISTS trig_score_settings_updated_at 
AFTER UPDATE ON score_settings
BEGIN
    UPDATE score_settings SET updated_at = CURRENT_TIMESTAMP 
    WHERE id = NEW.id;
END;
