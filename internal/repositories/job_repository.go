package repositories

import (
	"database/sql"
	"sync"
	"time"

	"github.com/alimgiray/gscope/internal/models"
)

// JobRepository handles database operations for jobs
type JobRepository struct {
	db *sql.DB
	mu sync.RWMutex
}

// NewJobRepository creates a new JobRepository
func NewJobRepository(db *sql.DB) *JobRepository {
	return &JobRepository{db: db}
}

// Create creates a new job
func (r *JobRepository) Create(job *models.Job) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	query := `
		INSERT INTO jobs (id, project_id, project_repository_id, job_type, status, error_message, started_at, completed_at, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := r.db.Exec(query,
		job.ID,
		job.ProjectID,
		job.ProjectRepositoryID,
		job.JobType,
		job.Status,
		job.ErrorMessage,
		job.StartedAt,
		job.CompletedAt,
		job.CreatedAt,
		job.UpdatedAt,
	)
	return err
}

// GetByID retrieves a job by ID
func (r *JobRepository) GetByID(id string) (*models.Job, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	query := `
		SELECT id, project_id, project_repository_id, job_type, status, error_message, started_at, completed_at, created_at, updated_at
		FROM jobs WHERE id = ?
	`

	job := &models.Job{}
	err := r.db.QueryRow(query, id).Scan(
		&job.ID,
		&job.ProjectID,
		&job.ProjectRepositoryID,
		&job.JobType,
		&job.Status,
		&job.ErrorMessage,
		&job.StartedAt,
		&job.CompletedAt,
		&job.CreatedAt,
		&job.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return job, nil
}

// GetByProjectID retrieves all jobs for a project
func (r *JobRepository) GetByProjectID(projectID string) ([]*models.Job, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	query := `
		SELECT id, project_id, project_repository_id, job_type, status, error_message, started_at, completed_at, created_at, updated_at
		FROM jobs 
		WHERE project_id = ?
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(query, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var jobs []*models.Job
	for rows.Next() {
		job := &models.Job{}
		err := rows.Scan(
			&job.ID,
			&job.ProjectID,
			&job.ProjectRepositoryID,
			&job.JobType,
			&job.Status,
			&job.ErrorMessage,
			&job.StartedAt,
			&job.CompletedAt,
			&job.CreatedAt,
			&job.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		jobs = append(jobs, job)
	}

	return jobs, nil
}

// GetPendingJobs retrieves all pending jobs
func (r *JobRepository) GetPendingJobs() ([]*models.Job, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	query := `
		SELECT id, project_id, project_repository_id, job_type, status, error_message, started_at, completed_at, created_at, updated_at
		FROM jobs 
		WHERE status = ?
		ORDER BY created_at ASC
	`

	rows, err := r.db.Query(query, models.JobStatusPending)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var jobs []*models.Job
	for rows.Next() {
		job := &models.Job{}
		err := rows.Scan(
			&job.ID,
			&job.ProjectID,
			&job.ProjectRepositoryID,
			&job.JobType,
			&job.Status,
			&job.ErrorMessage,
			&job.StartedAt,
			&job.CompletedAt,
			&job.CreatedAt,
			&job.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		jobs = append(jobs, job)
	}

	return jobs, nil
}

// GetByProjectRepositoryID retrieves all jobs for a specific project repository
func (r *JobRepository) GetByProjectRepositoryID(projectRepositoryID string) ([]*models.Job, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	query := `
		SELECT id, project_id, project_repository_id, job_type, status, error_message, started_at, completed_at, created_at, updated_at
		FROM jobs 
		WHERE project_repository_id = ?
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(query, projectRepositoryID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var jobs []*models.Job
	for rows.Next() {
		job := &models.Job{}
		err := rows.Scan(
			&job.ID,
			&job.ProjectID,
			&job.ProjectRepositoryID,
			&job.JobType,
			&job.Status,
			&job.ErrorMessage,
			&job.StartedAt,
			&job.CompletedAt,
			&job.CreatedAt,
			&job.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		jobs = append(jobs, job)
	}

	return jobs, nil
}

// GetNextPendingJob retrieves the next pending job of a specific type (FIFO)
// This method is thread-safe and marks the job as in-progress
func (r *JobRepository) GetNextPendingJob(jobType models.JobType) (*models.Job, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Use a transaction to ensure atomicity
	tx, err := r.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// Get the oldest pending job of the specified type
	query := `
		SELECT id, project_id, project_repository_id, job_type, status, error_message, started_at, completed_at, created_at, updated_at
		FROM jobs 
		WHERE status = ? AND job_type = ?
		ORDER BY created_at ASC
		LIMIT 1
	`

	job := &models.Job{}
	err = tx.QueryRow(query, models.JobStatusPending, jobType).Scan(
		&job.ID,
		&job.ProjectID,
		&job.ProjectRepositoryID,
		&job.JobType,
		&job.Status,
		&job.ErrorMessage,
		&job.StartedAt,
		&job.CompletedAt,
		&job.CreatedAt,
		&job.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // No pending jobs found
		}
		return nil, err
	}

	// Mark the job as in-progress
	job.MarkStarted()
	updateQuery := `
		UPDATE jobs 
		SET status = ?, started_at = ?, updated_at = ?
		WHERE id = ?
	`

	_, err = tx.Exec(updateQuery, job.Status, job.StartedAt, time.Now(), job.ID)
	if err != nil {
		return nil, err
	}

	// Commit the transaction
	if err = tx.Commit(); err != nil {
		return nil, err
	}

	return job, nil
}

// Update updates a job
func (r *JobRepository) Update(job *models.Job) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	query := `
		UPDATE jobs 
		SET project_id = ?, project_repository_id = ?, job_type = ?, status = ?, error_message = ?, 
		    started_at = ?, completed_at = ?, updated_at = ?
		WHERE id = ?
	`

	_, err := r.db.Exec(query,
		job.ProjectID,
		job.ProjectRepositoryID,
		job.JobType,
		job.Status,
		job.ErrorMessage,
		job.StartedAt,
		job.CompletedAt,
		time.Now(),
		job.ID,
	)
	return err
}

// Delete deletes a job by ID
func (r *JobRepository) Delete(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	query := `DELETE FROM jobs WHERE id = ?`
	_, err := r.db.Exec(query, id)
	return err
}

// DeleteByProjectID deletes all jobs for a project
func (r *JobRepository) DeleteByProjectID(projectID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	query := `DELETE FROM jobs WHERE project_id = ?`
	_, err := r.db.Exec(query, projectID)
	return err
}
