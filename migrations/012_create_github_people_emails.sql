-- Migration: Create github_people_emails table
-- Date: 2024-12-19

-- Create github_people_emails table
CREATE TABLE IF NOT EXISTS github_people_emails (
    id TEXT PRIMARY KEY,
    project_id TEXT NOT NULL,
    github_person_id TEXT NOT NULL,
    person_id TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (project_id) REFERENCES projects (id) ON DELETE CASCADE,
    FOREIGN KEY (github_person_id) REFERENCES github_people (id) ON DELETE CASCADE,
    FOREIGN KEY (person_id) REFERENCES people (id) ON DELETE CASCADE,
    UNIQUE(project_id, person_id)  -- Each person can only be associated with one GitHub person per project
);

-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS  idx_github_people_emails_project_id ON github_people_emails(project_id);
CREATE INDEX IF NOT EXISTS idx_github_people_emails_github_person_id ON github_people_emails(github_person_id);
CREATE INDEX IF NOT EXISTS idx_github_people_emails_person_id ON github_people_emails(person_id);

-- Create trigger to update updated_at timestamp
CREATE TRIGGER IF NOT EXISTS update_github_people_emails_updated_at
    AFTER UPDATE ON github_people_emails
    FOR EACH ROW
BEGIN
    UPDATE github_people_emails SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END; 