package workers

import (
	"context"
	"log"
	"time"

	"github.com/alimgiray/gscope/internal/models"
	"github.com/alimgiray/gscope/internal/repositories"
)

// PullRequestWorker handles pull request jobs
type PullRequestWorker struct {
	*BaseWorker
	jobRepo *repositories.JobRepository
}

// NewPullRequestWorker creates a new pull request worker
func NewPullRequestWorker(workerID string, jobRepo *repositories.JobRepository) *PullRequestWorker {
	return &PullRequestWorker{
		BaseWorker: NewBaseWorker(workerID, models.JobTypePullRequest),
		jobRepo:    jobRepo,
	}
}

// Start begins the pull request worker process
func (w *PullRequestWorker) Start(ctx context.Context) error {
	w.Running = true
	log.Printf("Pull request worker %s started", w.WorkerID)

	for {
		select {
		case <-ctx.Done():
			log.Printf("Pull request worker %s stopping due to context cancellation", w.WorkerID)
			return ctx.Err()
		case <-w.StopChan:
			log.Printf("Pull request worker %s stopping", w.WorkerID)
			return nil
		default:
			// Try to get a pending pull request job
			job, err := w.jobRepo.GetNextPendingJob(models.JobTypePullRequest)
			if err != nil {
				log.Printf("Pull request worker %s error getting job: %v", w.WorkerID, err)
				time.Sleep(5 * time.Second)
				continue
			}

			if job == nil {
				// No jobs available, sleep and try again
				time.Sleep(10 * time.Second)
				continue
			}

			// Process the pull request job
			w.processPullRequestJob(ctx, job)
		}
	}
}

// processPullRequestJob handles the actual pull request job processing
func (w *PullRequestWorker) processPullRequestJob(ctx context.Context, job *models.Job) {
	log.Printf("Pull request worker %s processing job %s", w.WorkerID, job.ID)

	// Mark job as started
	job.MarkStarted()
	if err := w.jobRepo.Update(job); err != nil {
		log.Printf("Pull request worker %s error updating job %s: %v", w.WorkerID, job.ID, err)
		return
	}

	// TODO: Implement actual pull request analysis logic here
	// For now, just simulate work
	time.Sleep(4 * time.Second)

	// Mark job as completed
	job.MarkCompleted()
	if err := w.jobRepo.Update(job); err != nil {
		log.Printf("Pull request worker %s error completing job %s: %v", w.WorkerID, job.ID, err)
		return
	}

	log.Printf("Pull request worker %s completed job %s", w.WorkerID, job.ID)
}
