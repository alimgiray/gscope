CREATE TABLE IF NOT EXISTS excluded_extensions (
    id TEXT PRIMARY KEY,
    project_id TEXT NOT NULL,
    extension TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE
);

-- Create index on project_id for faster queries
CREATE INDEX IF NOT EXISTS idx_excluded_extensions_project_id 
ON excluded_extensions(project_id);

-- Create unique constraint to prevent duplicate extensions per project
CREATE UNIQUE INDEX IF NOT EXISTS idx_excluded_extensions_project_extension_unique 
ON excluded_extensions(project_id, extension); 