-- Migration: Create pull_requests, pr_reviews, and github_people tables
-- Date: 2024-12-19

CREATE TABLE IF NOT EXISTS pull_requests (
    id TEXT PRIMARY KEY,
    repository_id TEXT NOT NULL,
    github_pr_number INTEGER NOT NULL,
    github_pr_id INTEGER UNIQUE NOT NULL,
    title TEXT NOT NULL,
    body TEXT,
    state TEXT NOT NULL, -- "open", "closed"
    merged_at DATETIME,
    merge_commit_sha TEXT,
    closed_at DATETIME,
    user TEXT, -- JSON object with user information
    requested_reviewers TEXT, -- JSON array of reviewer objects
    requested_teams TEXT, -- JSON array of team objects
    draft BOOLEAN DEFAULT FALSE,
    github_created_at DATETIME,
    github_updated_at DATETIME,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (repository_id) REFERENCES github_repositories (id),
    UNIQUE(repository_id, github_pr_number)
);

CREATE TABLE IF NOT EXISTS pr_reviews (
    id TEXT PRIMARY KEY,
    repository_id TEXT NOT NULL,
    pull_request_id TEXT NOT NULL,
    github_review_id INTEGER UNIQUE NOT NULL,
    reviewer_id INTEGER NOT NULL,
    reviewer_login TEXT NOT NULL,
    reviewer_type TEXT,
    reviewer_avatar_url TEXT,
    body TEXT,
    body_html TEXT,
    body_text TEXT,
    state TEXT NOT NULL, -- "APPROVED", "CHANGES_REQUESTED", "COMMENTED", "DISMISSED"
    author_association TEXT,
    submitted_at DATETIME,
    commit_id TEXT NOT NULL,
    html_url TEXT,
    pull_request_url TEXT,
    url TEXT,
    github_created_at DATETIME,
    github_updated_at DATETIME,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (repository_id) REFERENCES github_repositories (id),
    FOREIGN KEY (pull_request_id) REFERENCES pull_requests (id)
);

CREATE TABLE IF NOT EXISTS github_people (
    id TEXT PRIMARY KEY,
    github_user_id INTEGER UNIQUE NOT NULL,
    username TEXT NOT NULL,
    display_name TEXT,
    avatar_url TEXT,
    profile_url TEXT,
    type TEXT, -- "User", "Bot", "Organization"
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Create triggers to automatically update the updated_at column
CREATE TRIGGER IF NOT EXISTS update_pull_requests_updated_at
    AFTER UPDATE ON pull_requests
    FOR EACH ROW
BEGIN
    UPDATE pull_requests SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END;

CREATE TRIGGER IF NOT EXISTS update_pr_reviews_updated_at
    AFTER UPDATE ON pr_reviews
    FOR EACH ROW
BEGIN
    UPDATE pr_reviews SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END;

CREATE TRIGGER IF NOT EXISTS update_github_people_updated_at
    AFTER UPDATE ON github_people
    FOR EACH ROW
BEGIN
    UPDATE github_people SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END; 