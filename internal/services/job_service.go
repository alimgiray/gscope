package services

import (
	"github.com/alimgiray/gscope/internal/models"
	"github.com/alimgiray/gscope/internal/repositories"
)

// JobService handles job-related business logic
type JobService struct {
	jobRepo *repositories.JobRepository
}

// NewJobService creates a new job service
func NewJobService(jobRepo *repositories.JobRepository) *JobService {
	return &JobService{
		jobRepo: jobRepo,
	}
}

// CreateCloneAndCommitJobs creates both clone and commit jobs with dependency
func (s *JobService) CreateCloneAndCommitJobs(projectID string, projectRepositoryID string) error {
	// Create clone job first
	cloneJob := models.NewJob(projectID, models.JobTypeClone)
	cloneJob.ProjectRepositoryID = &projectRepositoryID

	if err := s.jobRepo.Create(cloneJob); err != nil {
		return err
	}

	// Create commit job that depends on clone job
	commitJob := models.NewJob(projectID, models.JobTypeCommit)
	commitJob.ProjectRepositoryID = &projectRepositoryID
	commitJob.DependsOn = &cloneJob.ID

	if err := s.jobRepo.Create(commitJob); err != nil {
		return err
	}

	return nil
}

// CreatePullRequestAndStatsJobs creates both pull_request and stats jobs with dependency
func (s *JobService) CreatePullRequestAndStatsJobs(projectID string, projectRepositoryID string) error {
	// Create pull_request job first
	pullRequestJob := models.NewJob(projectID, models.JobTypePullRequest)
	pullRequestJob.ProjectRepositoryID = &projectRepositoryID

	if err := s.jobRepo.Create(pullRequestJob); err != nil {
		return err
	}

	// Create stats job that depends on pull_request job
	statsJob := models.NewJob(projectID, models.JobTypeStats)
	statsJob.ProjectRepositoryID = &projectRepositoryID
	statsJob.DependsOn = &pullRequestJob.ID

	if err := s.jobRepo.Create(statsJob); err != nil {
		return err
	}

	return nil
}

// GetProjectJobs retrieves all jobs for a project
func (s *JobService) GetProjectJobs(projectID string) ([]*models.Job, error) {
	return s.jobRepo.GetByProjectID(projectID)
}

// GetProjectRepositoryJobs retrieves all jobs for a project repository
func (s *JobService) GetProjectRepositoryJobs(projectRepositoryID string) ([]*models.Job, error) {
	return s.jobRepo.GetByProjectRepositoryID(projectRepositoryID)
}
