package workers

import (
	"context"
	"log"
	"time"

	"github.com/alimgiray/gscope/internal/models"
	"github.com/alimgiray/gscope/internal/repositories"
	"github.com/alimgiray/gscope/internal/services"
)

// StatsWorker handles stats jobs
type StatsWorker struct {
	*BaseWorker
	jobRepo               *repositories.JobRepository
	peopleStatsService    *services.PeopleStatisticsService
	projectRepositoryRepo *repositories.ProjectRepositoryRepository
}

// NewStatsWorker creates a new stats worker
func NewStatsWorker(
	workerID string,
	jobRepo *repositories.JobRepository,
	peopleStatsService *services.PeopleStatisticsService,
	projectRepositoryRepo *repositories.ProjectRepositoryRepository,
) *StatsWorker {
	return &StatsWorker{
		BaseWorker:            NewBaseWorker(workerID, models.JobTypeStats),
		jobRepo:               jobRepo,
		peopleStatsService:    peopleStatsService,
		projectRepositoryRepo: projectRepositoryRepo,
	}
}

// Start begins the stats worker process
func (w *StatsWorker) Start(ctx context.Context) error {
	w.Running = true
	log.Printf("Stats worker %s started", w.WorkerID)

	for {
		select {
		case <-ctx.Done():
			log.Printf("Stats worker %s stopping due to context cancellation", w.WorkerID)
			return ctx.Err()
		case <-w.StopChan:
			log.Printf("Stats worker %s stopping", w.WorkerID)
			return nil
		default:
			// Try to get a pending stats job
			job, err := w.jobRepo.GetNextPendingJob(models.JobTypeStats, w.WorkerID)
			if err != nil {
				log.Printf("Stats worker %s error getting job: %v", w.WorkerID, err)
				time.Sleep(5 * time.Second)
				continue
			}

			if job == nil {
				// No jobs available, sleep and try again
				time.Sleep(10 * time.Second)
				continue
			}

			// Process the stats job
			w.processStatsJob(ctx, job)
		}
	}
}

// processStatsJob handles the actual stats job processing
func (w *StatsWorker) processStatsJob(ctx context.Context, job *models.Job) {
	log.Printf("Stats worker %s processing job %s", w.WorkerID, job.ID)

	// Mark job as started
	job.MarkStarted()
	if err := w.jobRepo.Update(job); err != nil {
		log.Printf("Stats worker %s error updating job %s: %v", w.WorkerID, job.ID, err)
		return
	}

	// Process the stats job
	if err := w.ProcessJob(ctx, job); err != nil {
		log.Printf("Stats worker %s error processing job %s: %v", w.WorkerID, job.ID, err)
		job.SetError(err.Error())
		job.MarkFailed()
		if err := w.jobRepo.Update(job); err != nil {
			log.Printf("Stats worker %s error marking job %s as failed: %v", w.WorkerID, job.ID, err)
		}
		return
	}

	// Mark job as completed
	job.MarkCompleted()
	if err := w.jobRepo.Update(job); err != nil {
		log.Printf("Stats worker %s error completing job %s: %v", w.WorkerID, job.ID, err)
		return
	}

	log.Printf("Stats worker %s completed job %s", w.WorkerID, job.ID)
}

// ProcessJob processes a stats job
func (w *StatsWorker) ProcessJob(ctx context.Context, job *models.Job) error {
	log.Printf("Processing stats job for project: %s", job.ProjectID)

	// Check if this is a repository-specific job
	if job.ProjectRepositoryID != nil {
		// Process only the specific repository
		projectRepo, err := w.projectRepositoryRepo.GetByID(*job.ProjectRepositoryID)
		if err != nil {
			return err
		}

		if !projectRepo.IsTracked {
			return nil // Skip untracked repositories
		}

		// Clear existing statistics for this repository before recalculating
		log.Printf("Clearing existing statistics for repository: %s", projectRepo.ID)
		if err := w.peopleStatsService.DeleteStatisticsByRepository(projectRepo.ID); err != nil {
			log.Printf("Warning: failed to clear existing statistics for repository %s: %v", projectRepo.ID, err)
			// Continue anyway - we'll recalculate from scratch
		}

		log.Printf("Calculating statistics for repository: %s", projectRepo.ID)
		err = w.peopleStatsService.CalculateStatisticsForRepository(job.ProjectID, projectRepo.ID, projectRepo.GithubRepoID)
		if err != nil {
			return err
		}

		// Mark repository as analyzed after successful stats calculation
		now := time.Now()
		return w.projectRepositoryRepo.UpdateLastAnalyzed(projectRepo.ID, &now)
	} else {
		// Process all tracked repositories in the project
		projectRepos, err := w.projectRepositoryRepo.GetByProjectID(job.ProjectID)
		if err != nil {
			return err
		}

		for _, projectRepo := range projectRepos {
			if !projectRepo.IsTracked {
				continue
			}

			// Clear existing statistics for this repository before recalculating
			log.Printf("Clearing existing statistics for repository: %s", projectRepo.ID)
			if err := w.peopleStatsService.DeleteStatisticsByRepository(projectRepo.ID); err != nil {
				log.Printf("Warning: failed to clear existing statistics for repository %s: %v", projectRepo.ID, err)
				// Continue anyway - we'll recalculate from scratch
			}

			log.Printf("Calculating statistics for repository: %s", projectRepo.ID)
			if err := w.peopleStatsService.CalculateStatisticsForRepository(job.ProjectID, projectRepo.ID, projectRepo.GithubRepoID); err != nil {
				log.Printf("Error calculating statistics for repository %s: %v", projectRepo.ID, err)
				continue
			}

			// Mark repository as analyzed after successful stats calculation
			now := time.Now()
			if err := w.projectRepositoryRepo.UpdateLastAnalyzed(projectRepo.ID, &now); err != nil {
				log.Printf("Error updating analysis status for repository %s: %v", projectRepo.ID, err)
			}
		}

		return nil
	}
}
