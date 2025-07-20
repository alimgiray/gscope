package workers

import (
	"context"

	"github.com/alimgiray/gscope/internal/models"
)

// Worker interface defines the contract for all workers
type Worker interface {
	// Start begins the worker process
	Start(ctx context.Context) error

	// Stop gracefully stops the worker
	Stop() error

	// GetJobType returns the type of job this worker handles
	GetJobType() models.JobType

	// GetWorkerID returns the unique identifier for this worker
	GetWorkerID() string
}

// BaseWorker provides common functionality for all workers
type BaseWorker struct {
	WorkerID string
	JobType  models.JobType
	Running  bool
	StopChan chan struct{}
}

// NewBaseWorker creates a new base worker
func NewBaseWorker(workerID string, jobType models.JobType) *BaseWorker {
	return &BaseWorker{
		WorkerID: workerID,
		JobType:  jobType,
		Running:  false,
		StopChan: make(chan struct{}),
	}
}

// GetJobType returns the job type this worker handles
func (w *BaseWorker) GetJobType() models.JobType {
	return w.JobType
}

// GetWorkerID returns the worker's unique identifier
func (w *BaseWorker) GetWorkerID() string {
	return w.WorkerID
}

// Stop gracefully stops the worker
func (w *BaseWorker) Stop() error {
	if w.Running {
		w.Running = false
		close(w.StopChan)
	}
	return nil
}

// IsRunning checks if the worker is currently running
func (w *BaseWorker) IsRunning() bool {
	return w.Running
}
