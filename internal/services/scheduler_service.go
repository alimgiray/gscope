package services

import (
	"log"
	"time"

	"github.com/alimgiray/gscope/internal/models"
	"github.com/alimgiray/gscope/internal/repositories"
)

type SchedulerService struct {
	projectUpdateSettingsRepo *repositories.ProjectUpdateSettingsRepository
	jobRepo                   *repositories.JobRepository
	githubRepoService         *GitHubRepositoryService
}

func NewSchedulerService(
	projectUpdateSettingsRepo *repositories.ProjectUpdateSettingsRepository,
	jobRepo *repositories.JobRepository,
	githubRepoService *GitHubRepositoryService,
) *SchedulerService {
	return &SchedulerService{
		projectUpdateSettingsRepo: projectUpdateSettingsRepo,
		jobRepo:                   jobRepo,
		githubRepoService:         githubRepoService,
	}
}

// StartScheduler starts the automatic update scheduler
func (s *SchedulerService) StartScheduler() {
	go func() {
		for {
			now := time.Now()
			currentHour := now.Hour()

			// Get all enabled project update settings
			settings, err := s.projectUpdateSettingsRepo.GetAllEnabled()
			if err != nil {
				log.Printf("Error getting project update settings: %v", err)
				time.Sleep(1 * time.Hour)
				continue
			}

			// Check each project that should be updated at the current hour
			for _, setting := range settings {
				if setting.Hour == currentHour {
					log.Printf("Scheduling automatic update for project %s at hour %d", setting.ProjectID, setting.Hour)
					if err := s.scheduleProjectUpdate(setting.ProjectID); err != nil {
						log.Printf("Error scheduling update for project %s: %v", setting.ProjectID, err)
					}
				}
			}

			// Sleep until the next hour
			nextHour := now.Add(1 * time.Hour)
			nextHour = time.Date(nextHour.Year(), nextHour.Month(), nextHour.Day(), nextHour.Hour(), 0, 0, 0, nextHour.Location())
			sleepDuration := nextHour.Sub(now)
			time.Sleep(sleepDuration)
		}
	}()
}

// scheduleProjectUpdate schedules the "Update All" jobs for a project
func (s *SchedulerService) scheduleProjectUpdate(projectID string) error {
	// Get all tracked repositories for this project
	repositories, err := s.githubRepoService.GetProjectRepositories(projectID)
	if err != nil {
		return err
	}

	// Filter to only tracked repositories
	var trackedRepos []*models.ProjectRepository
	for _, repo := range repositories {
		if repo.IsTracked {
			trackedRepos = append(trackedRepos, repo)
		}
	}

	if len(trackedRepos) == 0 {
		log.Printf("No tracked repositories found for project %s", projectID)
		return nil
	}

	// Create jobs in the correct order with dependencies
	// Each repository gets its own chain: clone -> commit -> pull_request -> stats

	// Step 1: Create clone jobs for all tracked repositories
	for _, repo := range trackedRepos {
		cloneJob := models.NewJob(projectID, models.JobTypeClone)
		cloneJob.ProjectRepositoryID = &repo.ID
		if err := s.jobRepo.Create(cloneJob); err != nil {
			log.Printf("Failed to create clone job for repository %s: %v", repo.ID, err)
			continue
		}
		log.Printf("Created automatic clone job %s for repository %s", cloneJob.ID, repo.ID)

		// Step 2: Create commit job that depends on this clone job
		commitJob := models.NewJob(projectID, models.JobTypeCommit)
		commitJob.ProjectRepositoryID = &repo.ID
		commitJob.DependsOn = &cloneJob.ID
		if err := s.jobRepo.Create(commitJob); err != nil {
			log.Printf("Failed to create commit job for repository %s: %v", repo.ID, err)
			continue
		}
		log.Printf("Created automatic commit job %s for repository %s (depends on %s)", commitJob.ID, repo.ID, cloneJob.ID)

		// Step 3: Create pull request job that depends on this commit job
		pullRequestJob := models.NewJob(projectID, models.JobTypePullRequest)
		pullRequestJob.ProjectRepositoryID = &repo.ID
		pullRequestJob.DependsOn = &commitJob.ID
		if err := s.jobRepo.Create(pullRequestJob); err != nil {
			log.Printf("Failed to create pull request job for repository %s: %v", repo.ID, err)
			continue
		}
		log.Printf("Created automatic pull request job %s for repository %s (depends on %s)", pullRequestJob.ID, repo.ID, commitJob.ID)

		// Step 4: Create stats job that depends on this pull request job
		statsJob := models.NewJob(projectID, models.JobTypeStats)
		statsJob.ProjectRepositoryID = &repo.ID
		statsJob.DependsOn = &pullRequestJob.ID
		if err := s.jobRepo.Create(statsJob); err != nil {
			log.Printf("Failed to create stats job for repository %s: %v", repo.ID, err)
			continue
		}
		log.Printf("Created automatic stats job %s for repository %s (depends on %s)", statsJob.ID, repo.ID, pullRequestJob.ID)
	}

	return nil
}
