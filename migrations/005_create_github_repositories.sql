-- Create github_repositories table
CREATE TABLE IF NOT EXISTS github_repositories (
    id TEXT PRIMARY KEY,
    github_id INTEGER UNIQUE NOT NULL,
    name TEXT NOT NULL,
    full_name TEXT NOT NULL,
    description TEXT,
    url TEXT NOT NULL,
    clone_url TEXT NOT NULL,
    language TEXT,
    stars INTEGER DEFAULT 0,
    forks INTEGER DEFAULT 0,
    private BOOLEAN DEFAULT FALSE,
    default_branch TEXT,
    local_path TEXT,
    is_cloned BOOLEAN DEFAULT FALSE,
    last_cloned DATETIME,
    github_created_at DATETIME,
    github_updated_at DATETIME,
    github_pushed_at DATETIME,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_github_repositories_github_id ON github_repositories(github_id);
CREATE INDEX IF NOT EXISTS idx_github_repositories_full_name ON github_repositories(full_name);
CREATE INDEX IF NOT EXISTS idx_github_repositories_language ON github_repositories(language);
CREATE INDEX IF NOT EXISTS idx_github_repositories_private ON github_repositories(private);
CREATE INDEX IF NOT EXISTS idx_github_repositories_is_cloned ON github_repositories(is_cloned);
CREATE INDEX IF NOT EXISTS idx_github_repositories_created_at ON github_repositories(created_at);

-- Create trigger to update updated_at timestamp
CREATE TRIGGER IF NOT EXISTS update_github_repositories_updated_at
    AFTER UPDATE ON github_repositories
    FOR EACH ROW
BEGIN
    UPDATE github_repositories SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END; 