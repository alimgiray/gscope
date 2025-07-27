-- Migration: Update pr_reviews table - remove unwanted columns
-- Date: 2024-12-19

-- Drop the existing table and recreate it with the new schema
DROP TABLE IF EXISTS pr_reviews;

CREATE TABLE pr_reviews (
    id TEXT PRIMARY KEY,
    repository_id TEXT NOT NULL,
    pull_request_id TEXT NOT NULL,
    github_review_id INTEGER UNIQUE NOT NULL,
    reviewer_id INTEGER NOT NULL,
    reviewer_login TEXT NOT NULL,
    body TEXT,
    state TEXT NOT NULL, -- "APPROVED", "CHANGES_REQUESTED", "COMMENTED", "DISMISSED"
    author_association TEXT,
    submitted_at DATETIME,
    commit_id TEXT NOT NULL,
    html_url TEXT,
    github_created_at DATETIME,
    github_updated_at DATETIME,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (repository_id) REFERENCES github_repositories (id),
    FOREIGN KEY (pull_request_id) REFERENCES pull_requests (id)
);

-- Recreate the trigger
CREATE TRIGGER IF NOT EXISTS update_pr_reviews_updated_at
    AFTER UPDATE ON pr_reviews
    FOR EACH ROW
BEGIN
    UPDATE pr_reviews SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END; 