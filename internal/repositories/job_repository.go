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
		INSERT INTO jobs (id, project_id, project_repository_id, job_type, status, error_message, depends_on, started_at, completed_at, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := r.db.Exec(query,
		job.ID,
		job.ProjectID,
		job.ProjectRepositoryID,
		job.JobType,
		job.Status,
		job.ErrorMessage,
		job.DependsOn,
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
		SELECT id, project_id, project_repository_id, job_type, status, error_message, depends_on, started_at, completed_at, worker_id, created_at, updated_at
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
		&job.DependsOn,
		&job.StartedAt,
		&job.CompletedAt,
		&job.WorkerID,
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
		SELECT id, project_id, project_repository_id, job_type, status, error_message, depends_on, started_at, completed_at, worker_id, created_at, updated_at
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
			&job.DependsOn,
			&job.StartedAt,
			&job.CompletedAt,
			&job.WorkerID,
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
		SELECT id, project_id, project_repository_id, job_type, status, error_message, depends_on, started_at, completed_at, worker_id, created_at, updated_at
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
			&job.DependsOn,
			&job.StartedAt,
			&job.CompletedAt,
			&job.WorkerID,
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
		SELECT id, project_id, project_repository_id, job_type, status, error_message, depends_on, started_at, completed_at, worker_id, created_at, updated_at
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
			&job.DependsOn,
			&job.StartedAt,
			&job.CompletedAt,
			&job.WorkerID,
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

// GetNextPendingJob retrieves the next pending or in-progress job of a specific type (FIFO)
// This method is thread-safe and marks the job as in-progress if it was pending
func (r *JobRepository) GetNextPendingJob(jobType models.JobType, workerID string) (*models.Job, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Use a transaction to ensure atomicity
	tx, err := r.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// First, try to find and claim a pending job atomically
	// Use a more aggressive approach to prevent race conditions
	claimQuery := `
		UPDATE jobs 
		SET status = ?, started_at = ?, updated_at = ?, worker_id = ?
		WHERE id = (
			SELECT j.id
			FROM jobs j
			LEFT JOIN jobs dep ON j.depends_on = dep.id
			WHERE j.status = ? AND j.job_type = ?
			AND (j.depends_on IS NULL OR dep.status IN (?, ?))
			ORDER BY j.created_at ASC
			LIMIT 1
		)
	`

	result, err := tx.Exec(claimQuery,
		models.JobStatusInProgress, time.Now(), time.Now(), workerID,
		models.JobStatusPending, jobType, models.JobStatusCompleted, models.JobStatusFailed,
	)

	if err != nil {
		return nil, err
	}

	// Check if we actually updated a row
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return nil, err
	}

	if rowsAffected > 0 {
		// We successfully claimed a pending job, now get its details
		// Make sure we only get the job we just claimed by filtering by worker_id
		query := `
			SELECT j.id, j.project_id, j.project_repository_id, j.job_type, j.status, j.error_message, 
			       j.depends_on, j.started_at, j.completed_at, j.worker_id, j.created_at, j.updated_at
			FROM jobs j
			WHERE j.status = ? AND j.job_type = ? AND j.worker_id = ?
			ORDER BY j.created_at ASC
			LIMIT 1
		`

		job := &models.Job{}
		err = tx.QueryRow(query, models.JobStatusInProgress, jobType, workerID).Scan(
			&job.ID,
			&job.ProjectID,
			&job.ProjectRepositoryID,
			&job.JobType,
			&job.Status,
			&job.ErrorMessage,
			&job.DependsOn,
			&job.StartedAt,
			&job.CompletedAt,
			&job.WorkerID,
			&job.CreatedAt,
			&job.UpdatedAt,
		)

		if err != nil {
			return nil, err
		}

		// Commit the transaction
		if err = tx.Commit(); err != nil {
			return nil, err
		}

		return job, nil
	}

	// No pending jobs found, try to find in-progress jobs that belong to this worker
	query := `
		SELECT j.id, j.project_id, j.project_repository_id, j.job_type, j.status, j.error_message, 
		       j.depends_on, j.started_at, j.completed_at, j.worker_id, j.created_at, j.updated_at
		FROM jobs j
		LEFT JOIN jobs dep ON j.depends_on = dep.id
		WHERE j.status = ? AND j.job_type = ? AND j.worker_id = ?
		AND (j.depends_on IS NULL OR dep.status IN (?, ?))
		ORDER BY j.created_at ASC
		LIMIT 1
	`

	job := &models.Job{}
	err = tx.QueryRow(query, models.JobStatusInProgress, jobType, workerID, models.JobStatusCompleted, models.JobStatusFailed).Scan(
		&job.ID,
		&job.ProjectID,
		&job.ProjectRepositoryID,
		&job.JobType,
		&job.Status,
		&job.ErrorMessage,
		&job.DependsOn,
		&job.StartedAt,
		&job.CompletedAt,
		&job.WorkerID,
		&job.CreatedAt,
		&job.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			// Commit the transaction even if no jobs found
			if err = tx.Commit(); err != nil {
				return nil, err
			}
			return nil, nil // No pending or in-progress jobs found
		}
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
		    depends_on = ?, started_at = ?, completed_at = ?, worker_id = ?, updated_at = ?
		WHERE id = ?
	`

	_, err := r.db.Exec(query,
		job.ProjectID,
		job.ProjectRepositoryID,
		job.JobType,
		job.Status,
		job.ErrorMessage,
		job.DependsOn,
		job.StartedAt,
		job.CompletedAt,
		job.WorkerID,
		job.UpdatedAt,
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

// GetJobsByDependency retrieves all jobs that depend on a specific job
func (r *JobRepository) GetJobsByDependency(dependsOnJobID string) ([]*models.Job, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	query := `
		SELECT id, project_id, project_repository_id, job_type, status, error_message, depends_on, started_at, completed_at, worker_id, created_at, updated_at
		FROM jobs 
		WHERE depends_on = ?
		ORDER BY created_at ASC
	`

	rows, err := r.db.Query(query, dependsOnJobID)
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
			&job.DependsOn,
			&job.StartedAt,
			&job.CompletedAt,
			&job.WorkerID,
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

// GetPendingJobsWithDependencies retrieves pending jobs that have no dependencies or whose dependencies are in final state
func (r *JobRepository) GetPendingJobsWithDependencies() ([]*models.Job, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	query := `
		SELECT j.id, j.project_id, j.project_repository_id, j.job_type, j.status, j.error_message, 
		       j.depends_on, j.started_at, j.completed_at, j.worker_id, j.created_at, j.updated_at
		FROM jobs j
		LEFT JOIN jobs dep ON j.depends_on = dep.id
		WHERE j.status = ? 
		AND (j.depends_on IS NULL OR dep.status IN (?, ?))
		ORDER BY j.created_at ASC
	`

	rows, err := r.db.Query(query, models.JobStatusPending, models.JobStatusCompleted, models.JobStatusFailed)
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
			&job.DependsOn,
			&job.StartedAt,
			&job.CompletedAt,
			&job.WorkerID,
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
