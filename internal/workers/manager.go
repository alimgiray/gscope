package workers

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"sync"

	"github.com/alimgiray/gscope/internal/repositories"
	"github.com/alimgiray/gscope/internal/services"
	"github.com/google/go-github/v57/github"
)

// WorkerManager manages multiple workers of different types
type WorkerManager struct {
	workers               []Worker
	jobRepo               *repositories.JobRepository
	cloneService          *services.CloneService
	projectRepositoryRepo *repositories.ProjectRepositoryRepository
	commitRepo            *repositories.CommitRepository
	commitFileRepo        *repositories.CommitFileRepository
	personRepo            *repositories.PersonRepository
	githubRepoRepo        *repositories.GitHubRepositoryRepository
	githubRepoService     *services.GitHubRepositoryService
	pullRequestService    *services.PullRequestService
	prReviewService       *services.PRReviewService
	githubPersonService   *services.GithubPersonService
	githubClient          *github.Client
	projectRepo           *repositories.ProjectRepository
	userRepo              *repositories.UserRepository
	wg                    sync.WaitGroup
	ctx                   context.Context
	cancel                context.CancelFunc
}

// NewWorkerManager creates a new worker manager
func NewWorkerManager(
	jobRepo *repositories.JobRepository,
	cloneService *services.CloneService,
	projectRepositoryRepo *repositories.ProjectRepositoryRepository,
	commitRepo *repositories.CommitRepository,
	commitFileRepo *repositories.CommitFileRepository,
	personRepo *repositories.PersonRepository,
	githubRepoRepo *repositories.GitHubRepositoryRepository,
	githubRepoService *services.GitHubRepositoryService,
	pullRequestService *services.PullRequestService,
	prReviewService *services.PRReviewService,
	githubPersonService *services.GithubPersonService,
	githubClient *github.Client,
	projectRepo *repositories.ProjectRepository,
	userRepo *repositories.UserRepository,
) *WorkerManager {
	ctx, cancel := context.WithCancel(context.Background())
	return &WorkerManager{
		workers:               make([]Worker, 0),
		jobRepo:               jobRepo,
		cloneService:          cloneService,
		projectRepositoryRepo: projectRepositoryRepo,
		commitRepo:            commitRepo,
		commitFileRepo:        commitFileRepo,
		personRepo:            personRepo,
		githubRepoRepo:        githubRepoRepo,
		githubRepoService:     githubRepoService,
		pullRequestService:    pullRequestService,
		prReviewService:       prReviewService,
		githubPersonService:   githubPersonService,
		githubClient:          githubClient,
		projectRepo:           projectRepo,
		userRepo:              userRepo,
		ctx:                   ctx,
		cancel:                cancel,
	}
}

// StartAll starts all workers based on environment configuration
func (wm *WorkerManager) StartAll() error {
	// Get worker counts from environment variables
	cloneWorkers := wm.getWorkerCount("CLONE_WORKERS", 2)
	commitWorkers := wm.getWorkerCount("COMMIT_WORKERS", 2)
	pullRequestWorkers := wm.getWorkerCount("PULL_REQUEST_WORKERS", 2)
	statsWorkers := wm.getWorkerCount("STATS_WORKERS", 1)

	log.Printf("Starting workers - Clone: %d, Commit: %d, PullRequest: %d, Stats: %d",
		cloneWorkers, commitWorkers, pullRequestWorkers, statsWorkers)

	// Create and start clone workers
	for i := 0; i < cloneWorkers; i++ {
		worker := NewCloneWorker(fmt.Sprintf("clone-%d", i+1), wm.jobRepo, wm.cloneService)
		wm.workers = append(wm.workers, worker)
		wm.startWorker(worker)
	}

	// Create and start commit workers
	for i := 0; i < commitWorkers; i++ {
		worker := NewCommitWorker(fmt.Sprintf("commit-%d", i+1), wm.jobRepo, wm.commitRepo, wm.commitFileRepo, wm.personRepo, wm.projectRepositoryRepo, wm.githubRepoRepo)
		wm.workers = append(wm.workers, worker)
		wm.startWorker(worker)
	}

	// Create and start pull request workers
	for i := 0; i < pullRequestWorkers; i++ {
		worker := NewPullRequestWorker(
			fmt.Sprintf("pull-request-%d", i+1),
			wm.githubClient,
			wm.jobRepo,
			wm.pullRequestService,
			wm.prReviewService,
			wm.githubPersonService,
			wm.githubRepoService,
			wm.projectRepositoryRepo,
			wm.projectRepo,
			wm.userRepo,
		)
		wm.workers = append(wm.workers, worker)
		wm.startWorker(worker)
	}

	// Create and start stats workers
	for i := 0; i < statsWorkers; i++ {
		worker := NewStatsWorker(fmt.Sprintf("stats-%d", i+1), wm.jobRepo, wm.projectRepositoryRepo)
		wm.workers = append(wm.workers, worker)
		wm.startWorker(worker)
	}

	log.Printf("Started %d total workers", len(wm.workers))
	return nil
}

// StopAll gracefully stops all workers
func (wm *WorkerManager) StopAll() error {
	log.Println("Stopping all workers...")

	// Cancel the context to signal all workers to stop
	wm.cancel()

	// Stop each worker
	for _, worker := range wm.workers {
		if err := worker.Stop(); err != nil {
			log.Printf("Error stopping worker %s: %v", worker.GetWorkerID(), err)
		}
	}

	// Wait for all workers to finish
	wm.wg.Wait()

	log.Println("All workers stopped")
	return nil
}

// GetWorkerCount reads worker count from environment variable with fallback
func (wm *WorkerManager) getWorkerCount(envVar string, defaultValue int) int {
	if value := os.Getenv(envVar); value != "" {
		if count, err := strconv.Atoi(value); err == nil && count > 0 {
			return count
		}
		log.Printf("Invalid value for %s, using default: %d", envVar, defaultValue)
	}
	return defaultValue
}

// startWorker starts a single worker in a goroutine
func (wm *WorkerManager) startWorker(worker Worker) {
	wm.wg.Add(1)
	go func() {
		defer wm.wg.Done()
		if err := worker.Start(wm.ctx); err != nil {
			log.Printf("Worker %s stopped with error: %v", worker.GetWorkerID(), err)
		}
	}()
}

// GetWorkerStatus returns the status of all workers
func (wm *WorkerManager) GetWorkerStatus() map[string]bool {
	status := make(map[string]bool)
	for _, worker := range wm.workers {
		// Check if the worker has a BaseWorker embedded
		if cloneWorker, ok := worker.(*CloneWorker); ok {
			status[worker.GetWorkerID()] = cloneWorker.IsRunning()
		} else if commitWorker, ok := worker.(*CommitWorker); ok {
			status[worker.GetWorkerID()] = commitWorker.IsRunning()
		} else if pullRequestWorker, ok := worker.(*PullRequestWorker); ok {
			status[worker.GetWorkerID()] = pullRequestWorker.IsRunning()
		} else if statsWorker, ok := worker.(*StatsWorker); ok {
			status[worker.GetWorkerID()] = statsWorker.IsRunning()
		} else {
			status[worker.GetWorkerID()] = false
		}
	}
	return status
}
