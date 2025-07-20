package models

import (
	"time"

	"github.com/google/uuid"
)

// ProjectRepository represents the relationship between a project and a GitHub repository
type ProjectRepository struct {
	ID           string     `json:"id"`
	ProjectID    string     `json:"project_id"`
	GithubRepoID string     `json:"github_repo_id"`
	IsAnalyzed   bool       `json:"is_analyzed"`
	IsTracked    bool       `json:"is_tracked"`
	LastAnalyzed *time.Time `json:"last_analyzed"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	DeletedAt    *time.Time `json:"deleted_at"`
}

// NewProjectRepository creates a new ProjectRepository with a generated UUID
func NewProjectRepository(projectID, githubRepoID string) *ProjectRepository {
	return &ProjectRepository{
		ID:           uuid.New().String(),
		ProjectID:    projectID,
		GithubRepoID: githubRepoID,
		IsAnalyzed:   false,
		IsTracked:    false,
	}
}
