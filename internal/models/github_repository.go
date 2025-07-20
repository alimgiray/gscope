package models

import (
	"time"

	"github.com/google/uuid"
)

// GitHubRepository represents a GitHub repository
type GitHubRepository struct {
	ID              string     `json:"id"`
	GithubID        int64      `json:"github_id"`
	Name            string     `json:"name"`
	FullName        string     `json:"full_name"`
	Description     *string    `json:"description"`
	URL             string     `json:"url"`
	CloneURL        string     `json:"clone_url"`
	Language        *string    `json:"language"`
	Stars           int        `json:"stars"`
	Forks           int        `json:"forks"`
	Private         bool       `json:"private"`
	DefaultBranch   *string    `json:"default_branch"`
	LocalPath       *string    `json:"local_path"`
	IsCloned        bool       `json:"is_cloned"`
	LastCloned      *time.Time `json:"last_cloned"`
	GithubCreatedAt *time.Time `json:"github_created_at"`
	GithubUpdatedAt *time.Time `json:"github_updated_at"`
	GithubPushedAt  *time.Time `json:"github_pushed_at"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

// NewGitHubRepository creates a new GitHubRepository with a generated UUID
func NewGitHubRepository(githubID int64, name, fullName, url, cloneURL string) *GitHubRepository {
	return &GitHubRepository{
		ID:       uuid.New().String(),
		GithubID: githubID,
		Name:     name,
		FullName: fullName,
		URL:      url,
		CloneURL: cloneURL,
		Stars:    0,
		Forks:    0,
		Private:  false,
		IsCloned: false,
	}
}
