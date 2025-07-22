package services

import (
	"fmt"

	"github.com/alimgiray/gscope/internal/models"
	"github.com/alimgiray/gscope/internal/repositories"
)

// JobService handles job creation and management
type JobService struct {
	jobRepo               *repositories.JobRepository
	projectRepositoryRepo *repositories.ProjectRepositoryRepository
}

// NewJobService creates a new job service
func NewJobService(jobRepo *repositories.JobRepository, projectRepositoryRepo *repositories.ProjectRepositoryRepository) *JobService {
	return &JobService{
		jobRepo:               jobRepo,
		projectRepositoryRepo: projectRepositoryRepo,
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

// CreateAnalyzeJobs creates the three analysis jobs in order: commits, pull_requests, stats
func (s *JobService) CreateAnalyzeJobs(projectID string, projectRepositoryID string) error {
	// Check if there's already an active analyze job for this repository
	hasActive, err := s.HasActiveJob(projectRepositoryID, models.JobTypeCommit)
	if err != nil {
		return fmt.Errorf("failed to check existing jobs: %w", err)
	}
	if hasActive {
		return fmt.Errorf("an analysis job is already in progress or pending for this repository")
	}

	hasActive, err = s.HasActiveJob(projectRepositoryID, models.JobTypePullRequest)
	if err != nil {
		return fmt.Errorf("failed to check existing jobs: %w", err)
	}
	if hasActive {
		return fmt.Errorf("an analysis job is already in progress or pending for this repository")
	}

	hasActive, err = s.HasActiveJob(projectRepositoryID, models.JobTypeStats)
	if err != nil {
		return fmt.Errorf("failed to check existing jobs: %w", err)
	}
	if hasActive {
		return fmt.Errorf("an analysis job is already in progress or pending for this repository")
	}

	// Create jobs in order: commits, pull_requests, stats
	jobTypes := []models.JobType{models.JobTypeCommit, models.JobTypePullRequest, models.JobTypeStats}

	for _, jobType := range jobTypes {
		job := models.NewJob(projectID, jobType)
		job.ProjectRepositoryID = &projectRepositoryID

		if err := s.jobRepo.Create(job); err != nil {
			return fmt.Errorf("failed to create %s job: %w", jobType, err)
		}
	}

	return nil
}

// CreateCloneJobsForAllTrackedRepositories creates clone jobs for all tracked repositories in a project
func (s *JobService) CreateCloneJobsForAllTrackedRepositories(projectID string) (int, error) {
	// Get all project repositories for this project
	projectRepos, err := s.projectRepositoryRepo.GetByProjectID(projectID)
	if err != nil {
		return 0, fmt.Errorf("failed to get project repositories: %w", err)
	}

	createdCount := 0
	for _, projectRepo := range projectRepos {
		// Only create jobs for tracked repositories
		if !projectRepo.IsTracked {
			continue
		}

		// Check if there's already an active clone job for this repository
		hasActive, err := s.HasActiveJob(projectRepo.ID, models.JobTypeClone)
		if err != nil {
			continue // Skip this repo if we can't check its status
		}

		if hasActive {
			continue // Skip if already has an active job
		}

		// Create clone job for this repository
		job := models.NewJob(projectID, models.JobTypeClone)
		job.ProjectRepositoryID = &projectRepo.ID

		if err := s.jobRepo.Create(job); err != nil {
			continue // Skip if we can't create the job
		}

		createdCount++
	}

	return createdCount, nil
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
