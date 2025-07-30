package services

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/alimgiray/gscope/internal/models"
	"github.com/alimgiray/gscope/internal/repositories"
)

// CloneService handles repository cloning operations
type CloneService struct {
	projectRepo           *repositories.ProjectRepository
	userRepo              *repositories.UserRepository
	githubRepoRepo        *repositories.GitHubRepositoryRepository
	projectRepositoryRepo *repositories.ProjectRepositoryRepository
	cloneBasePath         string
}

// NewCloneService creates a new clone service
func NewCloneService(
	projectRepo *repositories.ProjectRepository,
	userRepo *repositories.UserRepository,
	githubRepoRepo *repositories.GitHubRepositoryRepository,
	projectRepositoryRepo *repositories.ProjectRepositoryRepository,
) *CloneService {
	return &CloneService{
		projectRepo:           projectRepo,
		userRepo:              userRepo,
		githubRepoRepo:        githubRepoRepo,
		projectRepositoryRepo: projectRepositoryRepo,
		cloneBasePath:         "./clones",
	}
}

// CloneRepository clones or pulls a repository
func (s *CloneService) CloneRepository(job *models.Job) error {
	// Get the project to access owner information
	project, err := s.projectRepo.GetByID(job.ProjectID)
	if err != nil {
		return fmt.Errorf("failed to get project: %w", err)
	}

	// Get the project owner to access GitHub token
	owner, err := s.userRepo.GetByID(project.OwnerID.String())
	if err != nil {
		return fmt.Errorf("failed to get project owner: %w", err)
	}

	if owner.GitHubAccessToken == "" {
		return fmt.Errorf("GitHub access token not found for project owner")
	}

	// Get the project repository to access GitHub repository info
	projectRepo, err := s.projectRepositoryRepo.GetByID(*job.ProjectRepositoryID)
	if err != nil {
		return fmt.Errorf("failed to get project repository: %w", err)
	}

	// Get the GitHub repository details
	githubRepo, err := s.githubRepoRepo.GetByID(projectRepo.GithubRepoID)
	if err != nil {
		return fmt.Errorf("failed to get GitHub repository: %w", err)
	}

	// Create clones directory if it doesn't exist
	if err := os.MkdirAll(s.cloneBasePath, 0755); err != nil {
		return fmt.Errorf("failed to create clones directory: %w", err)
	}

	// Create repository-specific directory based on full name to avoid conflicts
	// Use full_name to ensure unique paths for repositories with same name from different owners
	repoClonePath := filepath.Join(s.cloneBasePath, githubRepo.FullName)

	// Check if repository is already cloned
	if s.isRepositoryCloned(repoClonePath) {
		// Repository exists, do a git pull
		return s.pullRepository(repoClonePath, githubRepo, owner.GitHubAccessToken)
	} else {
		// Repository doesn't exist, do a full clone
		return s.cloneRepository(repoClonePath, githubRepo, owner.GitHubAccessToken)
	}
}

// isRepositoryCloned checks if a repository is already cloned
func (s *CloneService) isRepositoryCloned(repoPath string) bool {
	gitDir := filepath.Join(repoPath, ".git")
	info, err := os.Stat(gitDir)
	return err == nil && info.IsDir()
}

// cloneRepository performs a full clone of the repository
func (s *CloneService) cloneRepository(repoPath string, githubRepo *models.GitHubRepository, token string) error {
	// Remove directory if it exists but is not a git repo
	if err := os.RemoveAll(repoPath); err != nil {
		return fmt.Errorf("failed to clean repository directory: %w", err)
	}

	// Create authenticated clone URL with token
	authURL := strings.Replace(githubRepo.CloneURL, "https://", "https://"+token+"@", 1)

	// Try to clone with retry logic for potential lock issues
	maxRetries := 3
	for attempt := 1; attempt <= maxRetries; attempt++ {
		// Clone the repository with authentication
		cmd := exec.Command("git", "clone", authURL, repoPath)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			if attempt < maxRetries {
				fmt.Printf("Git clone failed (attempt %d/%d), cleaning up and retrying in 2 seconds: %v\n", attempt, maxRetries, err)
				// Clean up the failed clone attempt
				if err := os.RemoveAll(repoPath); err != nil {
					fmt.Printf("Warning: failed to clean up failed clone: %v\n", err)
				}
				time.Sleep(2 * time.Second)
				continue
			}
			return fmt.Errorf("failed to clone repository after %d attempts: %w", maxRetries, err)
		}

		// Success - update GitHub repository record
		now := time.Now()
		githubRepo.IsCloned = true
		githubRepo.LastCloned = &now
		githubRepo.LocalPath = &repoPath

		if err := s.githubRepoRepo.Update(githubRepo); err != nil {
			return fmt.Errorf("failed to update GitHub repository record: %w", err)
		}

		return nil
	}

	return fmt.Errorf("failed to clone repository after %d attempts", maxRetries)
}

// pullRepository performs a git pull on an existing repository
func (s *CloneService) pullRepository(repoPath string, githubRepo *models.GitHubRepository, token string) error {
	// Configure git to use the token for authentication
	cmd := exec.Command("git", "config", "credential.helper", "store")
	cmd.Dir = repoPath
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to configure git credential helper: %w", err)
	}

	// Set up the remote URL with token
	authURL := strings.Replace(githubRepo.CloneURL, "https://", "https://"+token+"@", 1)
	cmd = exec.Command("git", "remote", "set-url", "origin", authURL)
	cmd.Dir = repoPath
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to set remote URL: %w", err)
	}

	// Try to pull with retry logic for Git lock errors
	maxRetries := 3
	for attempt := 1; attempt <= maxRetries; attempt++ {
		// Clean up any potential lock files before attempting pull
		if err := s.cleanupGitLocks(repoPath); err != nil {
			// Log but don't fail - this is just cleanup
			fmt.Printf("Warning: failed to cleanup Git locks (attempt %d): %v\n", attempt, err)
		}

		// Pull the repository
		cmd = exec.Command("git", "pull")
		cmd.Dir = repoPath
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			if attempt < maxRetries {
				fmt.Printf("Git pull failed (attempt %d/%d), trying fetch and reset: %v\n", attempt, maxRetries, err)

				// Try fetch and reset as an alternative to pull
				if fetchErr := s.fetchAndReset(repoPath); fetchErr != nil {
					fmt.Printf("Fetch and reset also failed, retrying pull in 2 seconds: %v\n", fetchErr)
					time.Sleep(2 * time.Second)
					continue
				} else {
					// Fetch and reset succeeded, break out of retry loop
					break
				}
			}
			return fmt.Errorf("failed to pull repository after %d attempts: %w", maxRetries, err)
		}

		// Success - update GitHub repository record
		now := time.Now()
		githubRepo.IsCloned = true
		githubRepo.LastCloned = &now
		githubRepo.LocalPath = &repoPath

		if err := s.githubRepoRepo.Update(githubRepo); err != nil {
			return fmt.Errorf("failed to update GitHub repository record: %w", err)
		}

		return nil
	}

	return fmt.Errorf("failed to pull repository after %d attempts", maxRetries)
}

// fetchAndReset performs a git fetch and reset to handle remote ref lock issues
func (s *CloneService) fetchAndReset(repoPath string) error {
	// Fetch all remotes
	cmd := exec.Command("git", "fetch", "--all")
	cmd.Dir = repoPath
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to fetch all remotes: %w", err)
	}

	// Get current branch
	cmd = exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	cmd.Dir = repoPath
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to get current branch: %w", err)
	}
	currentBranch := strings.TrimSpace(string(output))

	// Reset to origin/branch
	cmd = exec.Command("git", "reset", "--hard", "origin/"+currentBranch)
	cmd.Dir = repoPath
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to reset to origin/%s: %w", currentBranch, err)
	}

	return nil
}

// cleanupGitLocks removes Git lock files that might be causing issues
func (s *CloneService) cleanupGitLocks(repoPath string) error {
	gitDir := filepath.Join(repoPath, ".git")

	// List of common Git lock files to remove
	lockFiles := []string{
		filepath.Join(gitDir, "index.lock"),
		filepath.Join(gitDir, "refs", "heads", "*.lock"),
		filepath.Join(gitDir, "refs", "remotes", "*.lock"),
		filepath.Join(gitDir, "MERGE_HEAD.lock"),
		filepath.Join(gitDir, "CHERRY_PICK_HEAD.lock"),
		filepath.Join(gitDir, "REBASE_HEAD.lock"),
	}

	for _, lockFile := range lockFiles {
		// Use glob pattern matching for wildcard patterns
		if strings.Contains(lockFile, "*") {
			matches, err := filepath.Glob(lockFile)
			if err != nil {
				continue
			}
			for _, match := range matches {
				if err := os.Remove(match); err != nil && !os.IsNotExist(err) {
					// Log but don't fail - this is just cleanup
					fmt.Printf("Warning: failed to remove lock file %s: %v\n", match, err)
				}
			}
		} else {
			if err := os.Remove(lockFile); err != nil && !os.IsNotExist(err) {
				// Log but don't fail - this is just cleanup
				fmt.Printf("Warning: failed to remove lock file %s: %v\n", lockFile, err)
			}
		}
	}

	return nil
}

// GetClonePath returns the local path where a repository is cloned
func (s *CloneService) GetClonePath(fullName string) string {
	return filepath.Join(s.cloneBasePath, fullName)
}
