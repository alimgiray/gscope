package workers

import (
	"context"
	"log"
	"time"

	"github.com/alimgiray/gscope/internal/models"
	"github.com/alimgiray/gscope/internal/repositories"
)

// StatsWorker handles stats jobs
type StatsWorker struct {
	*BaseWorker
	jobRepo *repositories.JobRepository
}

// NewStatsWorker creates a new stats worker
func NewStatsWorker(workerID string, jobRepo *repositories.JobRepository) *StatsWorker {
	return &StatsWorker{
		BaseWorker: NewBaseWorker(workerID, models.JobTypeStats),
		jobRepo:    jobRepo,
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
			job, err := w.jobRepo.GetNextPendingJob(models.JobTypeStats)
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

	// TODO: Implement actual stats generation logic here
	// For now, just simulate work
	time.Sleep(5 * time.Second)

	// Mark job as completed
	job.MarkCompleted()
	if err := w.jobRepo.Update(job); err != nil {
		log.Printf("Stats worker %s error completing job %s: %v", w.WorkerID, job.ID, err)
		return
	}

	log.Printf("Stats worker %s completed job %s", w.WorkerID, job.ID)
}
