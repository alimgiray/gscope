package services

import (
	"fmt"

	"github.com/alimgiray/gscope/internal/models"
	"github.com/alimgiray/gscope/internal/repositories"
)

// JobService handles job creation and management
type JobService struct {
	jobRepo *repositories.JobRepository
}

// NewJobService creates a new job service
func NewJobService(jobRepo *repositories.JobRepository) *JobService {
	return &JobService{
		jobRepo: jobRepo,
	}
}

// CreateCloneJob creates a new clone job for a specific project repository
func (s *JobService) CreateCloneJob(projectID string, projectRepositoryID string) (*models.Job, error) {
	// Check if there's already an active clone job for this repository
	hasActive, err := s.HasActiveJob(projectRepositoryID, models.JobTypeClone)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing jobs: %w", err)
	}

	if hasActive {
		return nil, fmt.Errorf("a clone job is already in progress or pending for this repository")
	}

	job := models.NewJob(projectID, models.JobTypeClone)
	job.ProjectRepositoryID = &projectRepositoryID

	if err := s.jobRepo.Create(job); err != nil {
		return nil, err
	}

	return job, nil
}

// HasActiveJob checks if there's already a pending or in-progress job of the specified type for a repository
func (s *JobService) HasActiveJob(projectRepositoryID string, jobType models.JobType) (bool, error) {
	existingJobs, err := s.jobRepo.GetByProjectRepositoryID(projectRepositoryID)
	if err != nil {
		return false, fmt.Errorf("failed to check existing jobs: %w", err)
	}

	for _, existingJob := range existingJobs {
		if existingJob.JobType == jobType &&
			(existingJob.Status == models.JobStatusPending || existingJob.Status == models.JobStatusInProgress) {
			return true, nil
		}
	}

	return false, nil
}

// GetJobsByProject retrieves all jobs for a project
func (s *JobService) GetJobsByProject(projectID string) ([]*models.Job, error) {
	return s.jobRepo.GetByProjectID(projectID)
}

// GetJobsByProjectRepository retrieves all jobs for a specific project repository
func (s *JobService) GetJobsByProjectRepository(projectRepositoryID string) ([]*models.Job, error) {
	return s.jobRepo.GetByProjectRepositoryID(projectRepositoryID)
}

// GetJobByID retrieves a job by ID
func (s *JobService) GetJobByID(jobID string) (*models.Job, error) {
	return s.jobRepo.GetByID(jobID)
}
