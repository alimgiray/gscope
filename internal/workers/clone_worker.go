package workers

import (
	"context"
	"log"
	"time"

	"github.com/alimgiray/gscope/internal/models"
	"github.com/alimgiray/gscope/internal/repositories"
	"github.com/alimgiray/gscope/internal/services"
)

// CloneWorker handles clone jobs
type CloneWorker struct {
	*BaseWorker
	jobRepo      *repositories.JobRepository
	cloneService *services.CloneService
}

// NewCloneWorker creates a new clone worker
func NewCloneWorker(workerID string, jobRepo *repositories.JobRepository, cloneService *services.CloneService) *CloneWorker {
	return &CloneWorker{
		BaseWorker:   NewBaseWorker(workerID, models.JobTypeClone),
		jobRepo:      jobRepo,
		cloneService: cloneService,
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
			job, err := w.jobRepo.GetNextPendingJob(models.JobTypeClone, w.WorkerID)
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

	// Check if project repository ID is provided
	if job.ProjectRepositoryID == nil {
		job.SetError("Project repository ID is required for clone jobs")
		job.MarkFailed()
		log.Printf("Clone worker %s failed job %s: %s", w.WorkerID, job.ID, *job.ErrorMessage)
		if err := w.jobRepo.Update(job); err != nil {
			log.Printf("Clone worker %s error updating failed job %s: %v", w.WorkerID, job.ID, err)
		}
		return
	}

	// Perform the actual clone operation
	if err := w.cloneService.CloneRepository(job); err != nil {
		job.SetError(err.Error())
		job.MarkFailed()
		log.Printf("Clone worker %s failed job %s: %s", w.WorkerID, job.ID, *job.ErrorMessage)
	} else {
		job.MarkCompleted()
		log.Printf("Clone worker %s completed job %s", w.WorkerID, job.ID)
	}

	if err := w.jobRepo.Update(job); err != nil {
		log.Printf("Clone worker %s error updating job %s: %v", w.WorkerID, job.ID, err)
		return
	}
}
