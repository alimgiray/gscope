package models

import (
	"time"

	"github.com/google/uuid"
)

// JobType represents the type of job
type JobType string

const (
	JobTypeClone       JobType = "clone"
	JobTypeCommit      JobType = "commit"
	JobTypePullRequest JobType = "pull_request"
	JobTypeStats       JobType = "stats"
)

// JobStatus represents the status of a job
type JobStatus string

const (
	JobStatusPending    JobStatus = "pending"
	JobStatusInProgress JobStatus = "in-progress"
	JobStatusCompleted  JobStatus = "completed"
	JobStatusFailed     JobStatus = "failed"
)

// Job represents a background job
type Job struct {
	ID                  string     `json:"id"`
	ProjectID           string     `json:"project_id"`
	ProjectRepositoryID *string    `json:"project_repository_id"`
	JobType             JobType    `json:"job_type"`
	Status              JobStatus  `json:"status"`
	ErrorMessage        *string    `json:"error_message"`
	DependsOn           *string    `json:"depends_on"`
	StartedAt           *time.Time `json:"started_at"`
	CompletedAt         *time.Time `json:"completed_at"`
	WorkerID            *string    `json:"worker_id"`
	CreatedAt           time.Time  `json:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at"`
}

// NewJob creates a new Job with a generated UUID
func NewJob(projectID string, jobType JobType) *Job {
	now := time.Now()
	return &Job{
		ID:        uuid.New().String(),
		ProjectID: projectID,
		JobType:   jobType,
		Status:    JobStatusPending,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// IsPending checks if the job is pending
func (j *Job) IsPending() bool {
	return j.Status == JobStatusPending
}

// IsInProgress checks if the job is in progress
func (j *Job) IsInProgress() bool {
	return j.Status == JobStatusInProgress
}

// IsCompleted checks if the job is completed
func (j *Job) IsCompleted() bool {
	return j.Status == JobStatusCompleted
}

// MarkStarted marks the job as started
func (j *Job) MarkStarted() {
	now := time.Now()
	j.Status = JobStatusInProgress
	j.StartedAt = &now
}

// MarkCompleted marks the job as completed
func (j *Job) MarkCompleted() {
	now := time.Now()
	j.Status = JobStatusCompleted
	j.CompletedAt = &now
}

// MarkFailed marks the job as failed
func (j *Job) MarkFailed() {
	now := time.Now()
	j.Status = JobStatusFailed
	j.CompletedAt = &now
}

// SetError sets an error message for the job
func (j *Job) SetError(message string) {
	j.ErrorMessage = &message
}

// IsFailed checks if the job is failed
func (j *Job) IsFailed() bool {
	return j.Status == JobStatusFailed
}
