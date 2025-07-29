-- Create project_repositories table
CREATE TABLE IF NOT EXISTS project_repositories (
    id TEXT PRIMARY KEY,
    project_id TEXT NOT NULL,
    github_repo_id TEXT NOT NULL,
    is_analyzed BOOLEAN DEFAULT FALSE,
    last_analyzed DATETIME,
    is_tracked BOOLEAN DEFAULT FALSE,
    last_fetched DATETIME,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    deleted_at DATETIME,
    FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE,
    FOREIGN KEY (github_repo_id) REFERENCES github_repositories(id) ON DELETE CASCADE
);

-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_project_repositories_project_id ON project_repositories(project_id);
CREATE INDEX IF NOT EXISTS idx_project_repositories_github_repo_id ON project_repositories(github_repo_id);
CREATE INDEX IF NOT EXISTS idx_project_repositories_is_analyzed ON project_repositories(is_analyzed);
CREATE INDEX IF NOT EXISTS idx_project_repositories_deleted_at ON project_repositories(deleted_at);
CREATE INDEX IF NOT EXISTS idx_project_repositories_created_at ON project_repositories(created_at);
CREATE INDEX IF NOT EXISTS idx_project_repositories_last_fetched ON project_repositories(last_fetched); 

-- Create unique constraint to prevent duplicate project-repo associations
CREATE UNIQUE INDEX IF NOT EXISTS idx_project_repositories_unique 
    ON project_repositories(project_id, github_repo_id) 
    WHERE deleted_at IS NULL;

-- Create trigger to update updated_at timestamp
CREATE TRIGGER IF NOT EXISTS update_project_repositories_updated_at
    AFTER UPDATE ON project_repositories
    FOR EACH ROW
BEGIN
    UPDATE project_repositories SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END; 