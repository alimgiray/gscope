-- Create commits table
CREATE TABLE IF NOT EXISTS commits (
    id TEXT PRIMARY KEY,
    github_repository_id TEXT NOT NULL,
    commit_sha TEXT UNIQUE NOT NULL,
    message TEXT NOT NULL,
    author_name TEXT NOT NULL,
    author_email TEXT,
    commit_date DATETIME NOT NULL,
    is_merge_commit BOOLEAN DEFAULT FALSE,
    merge_commit_sha TEXT,
    additions INTEGER DEFAULT 0,
    deletions INTEGER DEFAULT 0,
    changes INTEGER DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (github_repository_id) REFERENCES github_repositories (id)
);

-- Create commit_files table
CREATE TABLE IF NOT EXISTS commit_files (
    id TEXT PRIMARY KEY,
    commit_id TEXT NOT NULL,
    filename TEXT NOT NULL,
    status TEXT NOT NULL, -- "added", "modified", "removed", "renamed"
    additions INTEGER DEFAULT 0,
    deletions INTEGER DEFAULT 0,
    changes INTEGER DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (commit_id) REFERENCES commits (id)
);

-- Create people table
CREATE TABLE IF NOT EXISTS people (
    id TEXT PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    primary_email VARCHAR(255) NOT NULL UNIQUE,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_commits_github_repository_id ON commits(github_repository_id);
CREATE INDEX IF NOT EXISTS idx_commits_commit_date ON commits(commit_date);
CREATE INDEX IF NOT EXISTS idx_commits_author_email ON commits(author_email);
CREATE INDEX IF NOT EXISTS idx_commit_files_commit_id ON commit_files(commit_id);
CREATE INDEX IF NOT EXISTS idx_commit_files_filename ON commit_files(filename);
CREATE INDEX IF NOT EXISTS idx_people_primary_email ON people(primary_email);

-- Create trigger to update updated_at for people table
CREATE TRIGGER IF NOT EXISTS update_people_updated_at
    AFTER UPDATE ON people
    FOR EACH ROW
BEGIN
    UPDATE people SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END; 