package services

import (
	"context"
	"fmt"
	"strings"

	"github.com/alimgiray/gscope/internal/models"
	"github.com/alimgiray/gscope/internal/repositories"
	"github.com/google/go-github/v57/github"
	"golang.org/x/oauth2"
)

type GitHubRepositoryService struct {
	githubRepoRepo  *repositories.GitHubRepositoryRepository
	projectRepoRepo *repositories.ProjectRepositoryRepository
}

func NewGitHubRepositoryService(
	githubRepoRepo *repositories.GitHubRepositoryRepository,
	projectRepoRepo *repositories.ProjectRepositoryRepository,
) *GitHubRepositoryService {
	return &GitHubRepositoryService{
		githubRepoRepo:  githubRepoRepo,
		projectRepoRepo: projectRepoRepo,
	}
}

// createGitHubClient creates a GitHub client with the provided token
func (s *GitHubRepositoryService) createGitHubClient(token string) *github.Client {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)
	return github.NewClient(tc)
}

// FetchUserRepositories fetches all repositories the user has access to
func (s *GitHubRepositoryService) FetchUserRepositories(projectID, token string) error {
	if token == "" {
		return fmt.Errorf("GitHub token is required")
	}

	githubClient := s.createGitHubClient(token)
	ctx := context.Background()

	// Get authenticated user (we don't need the user data, just checking auth)
	_, _, err := githubClient.Users.Get(ctx, "")
	if err != nil {
		return fmt.Errorf("failed to get authenticated user: %w", err)
	}

	// Get all repositories for the user
	opt := &github.RepositoryListOptions{
		Type:        "all", // all, owner, public, private, member
		Sort:        "updated",
		Direction:   "desc",
		ListOptions: github.ListOptions{PerPage: 100},
	}

	var allRepos []*github.Repository
	for {
		repos, resp, err := githubClient.Repositories.List(ctx, "", opt)
		if err != nil {
			return fmt.Errorf("failed to list repositories: %w", err)
		}
		allRepos = append(allRepos, repos...)
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}

	// Process each repository
	for _, repo := range allRepos {
		if err := s.processRepository(repo, projectID); err != nil {
			// Log error but continue processing other repos
			fmt.Printf("Error processing repository %s: %v\n", repo.GetFullName(), err)
		}
	}

	return nil
}

// processRepository processes a single GitHub repository
func (s *GitHubRepositoryService) processRepository(repo *github.Repository, projectID string) error {
	// Check if repository already exists in our database
	existingRepo, err := s.githubRepoRepo.GetByGithubID(repo.GetID())

	var githubRepo *models.GitHubRepository

	if err != nil {
		// Repository doesn't exist, create new one
		githubRepo = s.createGitHubRepositoryFromAPI(repo)
		if err := s.githubRepoRepo.Create(githubRepo); err != nil {
			return fmt.Errorf("failed to create GitHub repository: %w", err)
		}
	} else {
		// Repository exists, update it
		githubRepo = s.updateGitHubRepositoryFromAPI(existingRepo, repo)
		if err := s.githubRepoRepo.Update(githubRepo); err != nil {
			return fmt.Errorf("failed to update GitHub repository: %w", err)
		}
	}

	// Check if project-repository relationship already exists
	_, err = s.projectRepoRepo.GetByProjectAndGithubRepo(projectID, githubRepo.ID)

	if err != nil {
		// Relationship doesn't exist, create new one
		projectRepo := models.NewProjectRepository(projectID, githubRepo.ID)
		projectRepo.IsTracked = false // Default to false as requested
		if err := s.projectRepoRepo.Create(projectRepo); err != nil {
			return fmt.Errorf("failed to create project repository relationship: %w", err)
		}
	}

	return nil
}

// createGitHubRepositoryFromAPI creates a new GitHubRepository from GitHub API data
func (s *GitHubRepositoryService) createGitHubRepositoryFromAPI(repo *github.Repository) *models.GitHubRepository {
	githubRepo := models.NewGitHubRepository(
		repo.GetID(),
		repo.GetName(),
		repo.GetFullName(),
		repo.GetHTMLURL(),
		repo.GetCloneURL(),
	)

	// Set optional fields
	if repo.Description != nil {
		githubRepo.Description = repo.Description
	}
	if repo.Language != nil {
		githubRepo.Language = repo.Language
	}
	githubRepo.Stars = repo.GetStargazersCount()
	githubRepo.Forks = repo.GetForksCount()
	githubRepo.Private = repo.GetPrivate()
	if repo.DefaultBranch != nil {
		githubRepo.DefaultBranch = repo.DefaultBranch
	}
	if repo.CreatedAt != nil {
		githubRepo.GithubCreatedAt = &repo.CreatedAt.Time
	}
	if repo.UpdatedAt != nil {
		githubRepo.GithubUpdatedAt = &repo.UpdatedAt.Time
	}
	if repo.PushedAt != nil {
		githubRepo.GithubPushedAt = &repo.PushedAt.Time
	}

	return githubRepo
}

// updateGitHubRepositoryFromAPI updates an existing GitHubRepository from GitHub API data
func (s *GitHubRepositoryService) updateGitHubRepositoryFromAPI(existingRepo *models.GitHubRepository, repo *github.Repository) *models.GitHubRepository {
	// Update fields that might have changed
	existingRepo.Name = repo.GetName()
	existingRepo.FullName = repo.GetFullName()
	existingRepo.URL = repo.GetHTMLURL()
	existingRepo.CloneURL = repo.GetCloneURL()

	if repo.Description != nil {
		existingRepo.Description = repo.Description
	}
	if repo.Language != nil {
		existingRepo.Language = repo.Language
	}
	existingRepo.Stars = repo.GetStargazersCount()
	existingRepo.Forks = repo.GetForksCount()
	existingRepo.Private = repo.GetPrivate()
	if repo.DefaultBranch != nil {
		existingRepo.DefaultBranch = repo.DefaultBranch
	}
	if repo.UpdatedAt != nil {
		existingRepo.GithubUpdatedAt = &repo.UpdatedAt.Time
	}
	if repo.PushedAt != nil {
		existingRepo.GithubPushedAt = &repo.PushedAt.Time
	}

	return existingRepo
}

// GetProjectRepositories gets all repositories for a project, sorted by tracked status and name
func (s *GitHubRepositoryService) GetProjectRepositories(projectID string) ([]*models.ProjectRepository, error) {
	projectRepos, err := s.projectRepoRepo.GetByProjectID(projectID)
	if err != nil {
		return nil, err
	}

	// Sort: tracked repos first, then alphabetically by name
	// We'll need to fetch the GitHub repository names for sorting
	var sortedRepos []*models.ProjectRepository
	var untrackedRepos []*models.ProjectRepository

	for _, projectRepo := range projectRepos {
		// Get the GitHub repository details
		_, err := s.githubRepoRepo.GetByID(projectRepo.GithubRepoID)
		if err != nil {
			// Skip repos we can't find
			continue
		}

		if projectRepo.IsTracked {
			sortedRepos = append(sortedRepos, projectRepo)
		} else {
			untrackedRepos = append(untrackedRepos, projectRepo)
		}
	}

	// Sort tracked repos alphabetically by name
	sortedRepos = s.sortRepositoriesByName(sortedRepos)
	untrackedRepos = s.sortRepositoriesByName(untrackedRepos)

	// Combine: tracked first, then untracked
	return append(sortedRepos, untrackedRepos...), nil
}

// sortRepositoriesByName sorts repositories alphabetically by their GitHub repository name
func (s *GitHubRepositoryService) sortRepositoriesByName(repos []*models.ProjectRepository) []*models.ProjectRepository {
	// Simple bubble sort for now - could be optimized
	for i := 0; i < len(repos)-1; i++ {
		for j := 0; j < len(repos)-i-1; j++ {
			repo1, _ := s.githubRepoRepo.GetByID(repos[j].GithubRepoID)
			repo2, _ := s.githubRepoRepo.GetByID(repos[j+1].GithubRepoID)

			if repo1 != nil && repo2 != nil && strings.Compare(repo1.FullName, repo2.FullName) > 0 {
				repos[j], repos[j+1] = repos[j+1], repos[j]
			}
		}
	}
	return repos
}

// GetGitHubRepository retrieves a GitHub repository by ID
func (s *GitHubRepositoryService) GetGitHubRepository(id string) (*models.GitHubRepository, error) {
	return s.githubRepoRepo.GetByID(id)
}

// GetProjectRepository retrieves a project repository by ID
func (s *GitHubRepositoryService) GetProjectRepository(id string) (*models.ProjectRepository, error) {
	return s.projectRepoRepo.GetByID(id)
}

// UpdateProjectRepository updates a project repository
func (s *GitHubRepositoryService) UpdateProjectRepository(projectRepo *models.ProjectRepository) error {
	return s.projectRepoRepo.Update(projectRepo)
}
