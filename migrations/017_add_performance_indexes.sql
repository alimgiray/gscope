-- Migration 017: Add performance indexes for stats jobs
-- Date: 2025-01-30

-- Indexes for commits table to speed up date and email filtering
CREATE INDEX IF NOT EXISTS idx_commits_repository_date ON commits(github_repository_id, commit_date);
CREATE INDEX IF NOT EXISTS idx_commits_author_email ON commits(author_email);
CREATE INDEX IF NOT EXISTS idx_commits_repository_email_date ON commits(github_repository_id, author_email, commit_date);

-- Indexes for pull_requests table
CREATE INDEX IF NOT EXISTS idx_pull_requests_repository_created ON pull_requests(repository_id, github_created_at);
CREATE INDEX IF NOT EXISTS idx_pull_requests_user ON pull_requests(user);

-- Indexes for pr_reviews table
CREATE INDEX IF NOT EXISTS idx_pr_reviews_repository_created ON pr_reviews(repository_id, github_created_at);
CREATE INDEX IF NOT EXISTS idx_pr_reviews_reviewer ON pr_reviews(reviewer_login);

-- Indexes for commit_files table
CREATE INDEX IF NOT EXISTS idx_commit_files_commit_id ON commit_files(commit_id);

-- Indexes for people_statistics table
CREATE INDEX IF NOT EXISTS idx_people_statistics_repository_date ON people_statistics(repository_id, stat_date);
CREATE INDEX IF NOT EXISTS idx_people_statistics_project_date ON people_statistics(project_id, stat_date);
CREATE INDEX IF NOT EXISTS idx_people_statistics_person_date ON people_statistics(github_person_id, stat_date);

-- Indexes for email_merges table
CREATE INDEX IF NOT EXISTS idx_email_merges_project ON email_merges(project_id);
CREATE INDEX IF NOT EXISTS idx_email_merges_source ON email_merges(source_email);

-- Indexes for github_people_emails table
CREATE INDEX IF NOT EXISTS idx_github_people_emails_project ON github_people_emails(project_id);
CREATE INDEX IF NOT EXISTS idx_github_people_emails_person ON github_people_emails(github_person_id); 