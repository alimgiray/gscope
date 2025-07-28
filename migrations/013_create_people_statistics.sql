-- Migration: Create people_statistics table
-- Date: 2024-12-19

-- Create people_statistics table
CREATE TABLE IF NOT EXISTS people_statistics (
    id TEXT PRIMARY KEY,
    project_id TEXT NOT NULL,
    repository_id TEXT NOT NULL,
    github_person_id TEXT NOT NULL,
    stat_date DATE NOT NULL,
    commits INTEGER DEFAULT 0,
    additions INTEGER DEFAULT 0,
    deletions INTEGER DEFAULT 0,
    comments INTEGER DEFAULT 0,
    pull_requests INTEGER DEFAULT 0,
    score INTEGER DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE,
    FOREIGN KEY (repository_id) REFERENCES project_repositories(id) ON DELETE CASCADE,
    FOREIGN KEY (github_person_id) REFERENCES github_people(id) ON DELETE CASCADE,
    UNIQUE(project_id, repository_id, github_person_id, stat_date)
);

-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_people_statistics_project_id ON people_statistics(project_id);
CREATE INDEX IF NOT EXISTS idx_people_statistics_repository_id ON people_statistics(repository_id);
CREATE INDEX IF NOT EXISTS idx_people_statistics_github_person_id ON people_statistics(github_person_id);
CREATE INDEX IF NOT EXISTS idx_people_statistics_stat_date ON people_statistics(stat_date);
CREATE INDEX IF NOT EXISTS idx_people_statistics_project_person_date ON people_statistics(project_id, github_person_id, stat_date);
CREATE INDEX IF NOT EXISTS idx_people_statistics_repository_person_date ON people_statistics(repository_id, github_person_id, stat_date);

-- Create trigger to update updated_at timestamp
CREATE TRIGGER IF NOT EXISTS update_people_statistics_updated_at
    AFTER UPDATE ON people_statistics
    FOR EACH ROW
BEGIN
    UPDATE people_statistics SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END; 