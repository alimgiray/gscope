-- Create jobs table
CREATE TABLE IF NOT EXISTS jobs (
    id TEXT PRIMARY KEY,
    project_id TEXT NOT NULL,
    project_repository_id TEXT,
    job_type TEXT NOT NULL CHECK (job_type IN ('clone', 'commit', 'pull_request', 'stats')),
    status TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'in-progress', 'completed', 'failed')),
    error_message TEXT,
    depends_on TEXT,
    started_at DATETIME,
    completed_at DATETIME,
    worker_id TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE,
    FOREIGN KEY (project_repository_id) REFERENCES project_repositories(id) ON DELETE CASCADE
);

-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_jobs_project_id ON jobs(project_id);
CREATE INDEX IF NOT EXISTS idx_jobs_job_type ON jobs(job_type);
CREATE INDEX IF NOT EXISTS idx_jobs_status ON jobs(status);
CREATE INDEX IF NOT EXISTS idx_jobs_created_at ON jobs(created_at);
CREATE INDEX IF NOT EXISTS idx_jobs_status_created_at ON jobs(status, created_at);
CREATE INDEX IF NOT EXISTS idx_jobs_depends_on ON jobs(depends_on); 
CREATE INDEX IF NOT EXISTS idx_jobs_worker_id ON jobs(worker_id); 

-- Create trigger to update updated_at timestamp
CREATE TRIGGER IF NOT EXISTS update_jobs_updated_at
    AFTER UPDATE ON jobs
    FOR EACH ROW
BEGIN
    UPDATE jobs SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END; 