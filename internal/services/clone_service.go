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

	// Create project-specific directory
	projectClonePath := filepath.Join(s.cloneBasePath, project.ID.String())
	if err := os.MkdirAll(projectClonePath, 0755); err != nil {
		return fmt.Errorf("failed to create project clone directory: %w", err)
	}

	// Create repository-specific directory
	repoClonePath := filepath.Join(projectClonePath, githubRepo.Name)

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

	// Clone the repository with authentication
	cmd := exec.Command("git", "clone", authURL, repoPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to clone repository: %w", err)
	}

	// Update GitHub repository record
	now := time.Now()
	githubRepo.IsCloned = true
	githubRepo.LastCloned = &now
	githubRepo.LocalPath = &repoPath

	if err := s.githubRepoRepo.Update(githubRepo); err != nil {
		return fmt.Errorf("failed to update GitHub repository record: %w", err)
	}

	return nil
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

	// Pull the repository
	cmd = exec.Command("git", "pull")
	cmd.Dir = repoPath
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to pull repository: %w", err)
	}

	// Update GitHub repository record
	now := time.Now()
	githubRepo.LastCloned = &now

	if err := s.githubRepoRepo.Update(githubRepo); err != nil {
		return fmt.Errorf("failed to update GitHub repository record: %w", err)
	}

	return nil
}

// GetClonePath returns the local path where a repository is cloned
func (s *CloneService) GetClonePath(repoName string) string {
	return filepath.Join(s.cloneBasePath, repoName)
}
