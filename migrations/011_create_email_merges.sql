-- Migration: Create email merges table
-- Date: 2024-12-19

-- Create email_merges table
CREATE TABLE IF NOT EXISTS email_merges (
    id TEXT PRIMARY KEY,
    project_id TEXT NOT NULL,
    source_email TEXT NOT NULL,  -- The email being merged (will be hidden)
    target_email TEXT NOT NULL,  -- The email to merge into (will be shown)
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (project_id) REFERENCES projects (id) ON DELETE CASCADE,
    UNIQUE(project_id, source_email)  -- Each source email can only be merged once per project
);

-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_email_merges_project_id ON email_merges(project_id);
CREATE INDEX IF NOT EXISTS idx_email_merges_source_email ON email_merges(source_email);
CREATE INDEX IF NOT EXISTS idx_email_merges_target_email ON email_merges(target_email);

-- Create trigger to update updated_at timestamp
CREATE TRIGGER IF NOT EXISTS update_email_merges_updated_at
    AFTER UPDATE ON email_merges
    FOR EACH ROW
BEGIN
    UPDATE email_merges SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END; 