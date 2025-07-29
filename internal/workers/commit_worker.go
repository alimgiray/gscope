package workers

import (
	"context"
	"fmt"
	"log"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/alimgiray/gscope/internal/models"
	"github.com/alimgiray/gscope/internal/repositories"
)

// CommitWorker handles commit jobs
type CommitWorker struct {
	*BaseWorker
	jobRepo               *repositories.JobRepository
	commitRepo            *repositories.CommitRepository
	commitFileRepo        *repositories.CommitFileRepository
	personRepo            *repositories.PersonRepository
	projectRepositoryRepo *repositories.ProjectRepositoryRepository
	githubRepoRepo        *repositories.GitHubRepositoryRepository
}

// NewCommitWorker creates a new commit worker
func NewCommitWorker(workerID string, jobRepo *repositories.JobRepository, commitRepo *repositories.CommitRepository, commitFileRepo *repositories.CommitFileRepository, personRepo *repositories.PersonRepository, projectRepositoryRepo *repositories.ProjectRepositoryRepository, githubRepoRepo *repositories.GitHubRepositoryRepository) *CommitWorker {
	return &CommitWorker{
		BaseWorker:            NewBaseWorker(workerID, models.JobTypeCommit),
		jobRepo:               jobRepo,
		commitRepo:            commitRepo,
		commitFileRepo:        commitFileRepo,
		personRepo:            personRepo,
		projectRepositoryRepo: projectRepositoryRepo,
		githubRepoRepo:        githubRepoRepo,
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
			job, err := w.jobRepo.GetNextPendingJob(models.JobTypeCommit, w.WorkerID)
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

	// Implement actual commit analysis logic
	if err := w.processCommitJobLogic(job); err != nil {
		log.Printf("Commit worker %s error processing job %s: %v", w.WorkerID, job.ID, err)
		job.SetError(err.Error())
		job.MarkFailed()
		if err := w.jobRepo.Update(job); err != nil {
			log.Printf("Commit worker %s error marking job %s as failed: %v", w.WorkerID, job.ID, err)
		}
		return
	}

	// Mark job as completed
	job.MarkCompleted()
	if err := w.jobRepo.Update(job); err != nil {
		log.Printf("Commit worker %s error completing job %s: %v", w.WorkerID, job.ID, err)
		return
	}

	log.Printf("Commit worker %s completed job %s", w.WorkerID, job.ID)
}

// processCommitJobLogic handles the actual commit analysis logic
func (w *CommitWorker) processCommitJobLogic(job *models.Job) error {
	if job.ProjectRepositoryID == nil {
		return fmt.Errorf("project repository ID is required for commit analysis")
	}

	// Get project repository details
	projectRepo, err := w.projectRepositoryRepo.GetByID(*job.ProjectRepositoryID)
	if err != nil {
		return fmt.Errorf("failed to get project repository: %w", err)
	}

	// Get GitHub repository details
	githubRepo, err := w.githubRepoRepo.GetByID(projectRepo.GithubRepoID)
	if err != nil {
		return fmt.Errorf("failed to get GitHub repository: %w", err)
	}

	// Check if repository is cloned
	if !githubRepo.IsCloned || githubRepo.LocalPath == nil {
		return fmt.Errorf("repository must be cloned before commit analysis")
	}

	// Analyze commits in the repository
	return w.analyzeRepositoryCommits(githubRepo)
}

func (w *CommitWorker) analyzeRepositoryCommits(githubRepo *models.GitHubRepository) error {
	repoPath := *githubRepo.LocalPath

	// Get the latest commit date from the database for this repository
	latestCommitDate, err := w.commitRepo.GetLatestCommitDateByRepositoryID(githubRepo.ID)
	if err != nil {
		log.Printf("Warning: failed to get latest commit date for repository %s: %v", githubRepo.ID, err)
		// If we can't get the latest date, process all commits
		latestCommitDate = time.Time{}
	}

	// Build git log command with date filter if we have a latest commit date
	var cmd *exec.Cmd
	if !latestCommitDate.IsZero() {
		// Only get commits after the latest commit date
		dateFilter := latestCommitDate.Format("2006-01-02 15:04:05")
		cmd = exec.Command("git", "log", "--pretty=format:%H|%aN|%aE|%ad|%s", "--date=iso", "--numstat", "--since="+dateFilter)
		log.Printf("Processing commits after %s for repository %s", dateFilter, githubRepo.ID)
	} else {
		// Get all commits if no previous commits exist
		cmd = exec.Command("git", "log", "--pretty=format:%H|%aN|%aE|%ad|%s", "--date=iso", "--numstat")
		log.Printf("Processing all commits for repository %s (no previous commits found)", githubRepo.ID)
	}

	cmd.Dir = repoPath
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to get git log: %w", err)
	}

	lines := strings.Split(string(output), "\n")

	var currentCommit *models.Commit
	var currentCommitSHA string

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Check if this is a commit header line (contains |)
		if strings.Contains(line, "|") {
			// Update stats for the previous commit before starting a new one
			if currentCommit != nil {
				if err := w.commitRepo.Update(currentCommit); err != nil {
					log.Printf("Warning: failed to update commit stats for %s: %v", currentCommitSHA, err)
				}
			}

			parts := strings.Split(line, "|")
			if len(parts) >= 5 {
				currentCommitSHA = parts[0]
				authorName := strings.TrimSpace(parts[1])
				authorEmail := strings.TrimSpace(parts[2])
				commitDateStr := parts[3]
				message := parts[4]

				// Parse commit date
				commitDate, err := time.Parse("2006-01-02 15:04:05 -0700", commitDateStr)
				if err != nil {
					log.Printf("Warning: failed to parse commit date for %s: %v", currentCommitSHA, err)
					commitDate = time.Now()
				}

				// Check if commit already exists
				exists, err := w.commitRepo.ExistsByCommitSHA(currentCommitSHA)
				if err != nil {
					log.Printf("Warning: failed to check commit existence for %s: %v", currentCommitSHA, err)
					continue
				}

				if exists {
					// Skip if commit already exists
					currentCommit = nil
					continue
				}

				// Create or get person, updating name if different
				person, err := w.personRepo.GetOrCreateByEmailWithNameUpdate(authorName, authorEmail)
				if err != nil {
					log.Printf("Warning: failed to get/create person for %s: %v", authorEmail, err)
					continue
				}

				// Create new commit
				currentCommit = models.NewCommit(githubRepo.ID, currentCommitSHA, message, person.Name, &person.PrimaryEmail, commitDate)

				// Check if it's a merge commit
				if strings.Contains(strings.ToLower(message), "merge") {
					currentCommit.SetMergeCommit("") // We'll get the actual merge commit SHA later if needed
				}

				// Save commit
				if err := w.commitRepo.Create(currentCommit); err != nil {
					log.Printf("Warning: failed to create commit %s: %v", currentCommitSHA, err)
					currentCommit = nil
					continue
				}
			}
		} else if currentCommit != nil {
			// This is a file change line (numstat format: additions deletions filename)
			parts := strings.Fields(line)
			if len(parts) >= 3 {
				additions, _ := strconv.Atoi(parts[0])
				deletions, _ := strconv.Atoi(parts[1])
				filename := strings.Join(parts[2:], " ") // Handle filenames with spaces

				// Skip if filename is empty or is a binary file indicator
				if filename == "" || filename == "-" {
					continue
				}

				// Determine file status
				var status models.FileStatus
				if additions > 0 && deletions == 0 {
					status = models.FileStatusAdded
				} else if additions == 0 && deletions > 0 {
					status = models.FileStatusRemoved
				} else {
					status = models.FileStatusModified
				}

				// Create commit file
				commitFile := models.NewCommitFile(currentCommit.ID, filename, status)
				commitFile.SetStats(additions, deletions, additions+deletions)

				// Save commit file
				if err := w.commitFileRepo.Create(commitFile); err != nil {
					log.Printf("Warning: failed to create commit file %s for commit %s: %v", filename, currentCommitSHA, err)
					continue
				}

				// Update commit stats
				currentCommit.Additions += additions
				currentCommit.Deletions += deletions
				currentCommit.Changes += additions + deletions
			}
		}
	}

	// Update stats for the last commit
	if currentCommit != nil {
		if err := w.commitRepo.Update(currentCommit); err != nil {
			log.Printf("Warning: failed to update commit stats for %s: %v", currentCommitSHA, err)
		}
	}

	return nil
}
