package workers

import (
	"context"
	"time"

	"github.com/alimgiray/gscope/internal/models"
	"github.com/alimgiray/gscope/internal/repositories"
	"github.com/alimgiray/gscope/internal/services"
	"github.com/alimgiray/gscope/pkg/logger"
	"github.com/sirupsen/logrus"
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
	logger.WithField("worker_id", w.WorkerID).Info("Stats worker started")

	for {
		select {
		case <-ctx.Done():
			logger.WithField("worker_id", w.WorkerID).Info("Stats worker stopping due to context cancellation")
			return ctx.Err()
		case <-w.StopChan:
			logger.WithField("worker_id", w.WorkerID).Info("Stats worker stopping")
			return nil
		default:
			// Try to get a pending stats job
			job, err := w.jobRepo.GetNextPendingJob(models.JobTypeStats, w.WorkerID)
			if err != nil {
				logger.WithField("worker_id", w.WorkerID).WithError(err).Error("Stats worker error getting job")
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
	logger.WithFields(logrus.Fields{
		"worker_id": w.WorkerID,
		"job_id":    job.ID,
	}).Info("Stats worker processing job")

	// Mark job as started
	job.MarkStarted()
	if err := w.jobRepo.Update(job); err != nil {
		logger.WithFields(logrus.Fields{
			"worker_id": w.WorkerID,
			"job_id":    job.ID,
		}).WithError(err).Error("Stats worker error updating job")
		return
	}

	// Process the stats job
	if err := w.ProcessJob(ctx, job); err != nil {
		logger.WithFields(logrus.Fields{
			"worker_id": w.WorkerID,
			"job_id":    job.ID,
		}).WithError(err).Error("Stats worker error processing job")
		job.SetError(err.Error())
		job.MarkFailed()
		if err := w.jobRepo.Update(job); err != nil {
			logger.WithFields(logrus.Fields{
				"worker_id": w.WorkerID,
				"job_id":    job.ID,
			}).WithError(err).Error("Stats worker error marking job as failed")
		}
		return
	}

	// Mark job as completed
	job.MarkCompleted()
	if err := w.jobRepo.Update(job); err != nil {
		logger.WithFields(logrus.Fields{
			"worker_id": w.WorkerID,
			"job_id":    job.ID,
		}).WithError(err).Error("Stats worker error completing job")
		return
	}

	logger.WithFields(logrus.Fields{
		"worker_id": w.WorkerID,
		"job_id":    job.ID,
	}).Info("Stats worker completed job")
}

// ProcessJob processes a stats job
func (w *StatsWorker) ProcessJob(ctx context.Context, job *models.Job) error {
	logger.WithField("project_id", job.ProjectID).Info("Processing stats job for project")

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
		logger.WithField("repository_id", projectRepo.ID).Info("Clearing existing statistics for repository")
		if err := w.peopleStatsService.DeleteStatisticsByRepository(projectRepo.ID); err != nil {
			logger.WithFields(logrus.Fields{
				"repository_id": projectRepo.ID,
			}).WithError(err).Warn("Failed to clear existing statistics for repository")
			// Continue anyway - we'll recalculate from scratch
		}

		logger.WithField("repository_id", projectRepo.ID).Info("Calculating statistics for repository")
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
			logger.WithField("repository_id", projectRepo.ID).Info("Clearing existing statistics for repository")
			if err := w.peopleStatsService.DeleteStatisticsByRepository(projectRepo.ID); err != nil {
				logger.WithFields(logrus.Fields{
					"repository_id": projectRepo.ID,
				}).WithError(err).Warn("Failed to clear existing statistics for repository")
				// Continue anyway - we'll recalculate from scratch
			}

			logger.WithField("repository_id", projectRepo.ID).Info("Calculating statistics for repository")
			if err := w.peopleStatsService.CalculateStatisticsForRepository(job.ProjectID, projectRepo.ID, projectRepo.GithubRepoID); err != nil {
				logger.WithFields(logrus.Fields{
					"repository_id": projectRepo.ID,
				}).WithError(err).Error("Error calculating statistics for repository")
				continue
			}

			// Mark repository as analyzed after successful stats calculation
			now := time.Now()
			if err := w.projectRepositoryRepo.UpdateLastAnalyzed(projectRepo.ID, &now); err != nil {
				logger.WithFields(logrus.Fields{
					"repository_id": projectRepo.ID,
				}).WithError(err).Error("Error updating analysis status for repository")
			}
		}

		return nil
	}
}
