package workers

import (
	"context"
	"log"
	"time"

	"github.com/alimgiray/gscope/internal/models"
	"github.com/alimgiray/gscope/internal/repositories"
)

// CommitWorker handles commit jobs
type CommitWorker struct {
	*BaseWorker
	jobRepo *repositories.JobRepository
}

// NewCommitWorker creates a new commit worker
func NewCommitWorker(workerID string, jobRepo *repositories.JobRepository) *CommitWorker {
	return &CommitWorker{
		BaseWorker: NewBaseWorker(workerID, models.JobTypeCommit),
		jobRepo:    jobRepo,
	}
}

// Start begins the commit worker process
func (w *CommitWorker) Start(ctx context.Context) error {
	w.Running = true
	log.Printf("Commit worker %s started", w.WorkerID)

	for {
		select {
		case <-ctx.Done():
			log.Printf("Commit worker %s stopping due to context cancellation", w.WorkerID)
			return ctx.Err()
		case <-w.StopChan:
			log.Printf("Commit worker %s stopping", w.WorkerID)
			return nil
		default:
			// Try to get a pending commit job
			job, err := w.jobRepo.GetNextPendingJob(models.JobTypeCommit)
			if err != nil {
				log.Printf("Commit worker %s error getting job: %v", w.WorkerID, err)
				time.Sleep(5 * time.Second)
				continue
			}

			if job == nil {
				// No jobs available, sleep and try again
				time.Sleep(10 * time.Second)
				continue
			}

			// Process the commit job
			w.processCommitJob(ctx, job)
		}
	}
}

// processCommitJob handles the actual commit job processing
func (w *CommitWorker) processCommitJob(ctx context.Context, job *models.Job) {
	log.Printf("Commit worker %s processing job %s", w.WorkerID, job.ID)

	// Mark job as started
	job.MarkStarted()
	if err := w.jobRepo.Update(job); err != nil {
		log.Printf("Commit worker %s error updating job %s: %v", w.WorkerID, job.ID, err)
		return
	}

	// TODO: Implement actual commit analysis logic here
	// For now, just simulate work
	time.Sleep(3 * time.Second)

	// Mark job as completed
	job.MarkCompleted()
	if err := w.jobRepo.Update(job); err != nil {
		log.Printf("Commit worker %s error completing job %s: %v", w.WorkerID, job.ID, err)
		return
	}

	log.Printf("Commit worker %s completed job %s", w.WorkerID, job.ID)
}
