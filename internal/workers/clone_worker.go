package workers

import (
	"context"
	"log"
	"time"

	"github.com/alimgiray/gscope/internal/models"
	"github.com/alimgiray/gscope/internal/repositories"
)

// CloneWorker handles clone jobs
type CloneWorker struct {
	*BaseWorker
	jobRepo *repositories.JobRepository
}

// NewCloneWorker creates a new clone worker
func NewCloneWorker(workerID string, jobRepo *repositories.JobRepository) *CloneWorker {
	return &CloneWorker{
		BaseWorker: NewBaseWorker(workerID, models.JobTypeClone),
		jobRepo:    jobRepo,
	}
}

// Start begins the clone worker process
func (w *CloneWorker) Start(ctx context.Context) error {
	w.Running = true
	log.Printf("Clone worker %s started", w.WorkerID)

	for {
		select {
		case <-ctx.Done():
			log.Printf("Clone worker %s stopping due to context cancellation", w.WorkerID)
			return ctx.Err()
		case <-w.StopChan:
			log.Printf("Clone worker %s stopping", w.WorkerID)
			return nil
		default:
			// Try to get a pending clone job
			job, err := w.jobRepo.GetNextPendingJob(models.JobTypeClone)
			if err != nil {
				log.Printf("Clone worker %s error getting job: %v", w.WorkerID, err)
				time.Sleep(5 * time.Second)
				continue
			}

			if job == nil {
				// No jobs available, sleep and try again
				time.Sleep(10 * time.Second)
				continue
			}

			// Process the clone job
			w.processCloneJob(ctx, job)
		}
	}
}

// processCloneJob handles the actual clone job processing
func (w *CloneWorker) processCloneJob(ctx context.Context, job *models.Job) {
	log.Printf("Clone worker %s processing job %s", w.WorkerID, job.ID)

	// Mark job as started
	job.MarkStarted()
	if err := w.jobRepo.Update(job); err != nil {
		log.Printf("Clone worker %s error updating job %s: %v", w.WorkerID, job.ID, err)
		return
	}

	// TODO: Implement actual clone logic here
	// For now, just simulate work
	time.Sleep(2 * time.Second)

	// Simulate potential failure (for demonstration)
	// In real implementation, this would be based on actual clone success/failure
	if job.ProjectRepositoryID != nil {
		// Mark job as completed
		job.MarkCompleted()
		log.Printf("Clone worker %s completed job %s", w.WorkerID, job.ID)
	} else {
		// Mark job as failed
		job.SetError("Project repository ID is required for clone jobs")
		job.MarkFailed()
		log.Printf("Clone worker %s failed job %s: %s", w.WorkerID, job.ID, *job.ErrorMessage)
	}

	if err := w.jobRepo.Update(job); err != nil {
		log.Printf("Clone worker %s error updating job %s: %v", w.WorkerID, job.ID, err)
		return
	}
}
