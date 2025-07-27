package models

import (
	"time"

	"github.com/google/uuid"
)

// GitHubPersonEmail represents an association between a GitHub person and a person (email)
type GitHubPersonEmail struct {
	ID             string    `json:"id"`
	ProjectID      string    `json:"project_id"`
	GitHubPersonID string    `json:"github_person_id"`
	PersonID       string    `json:"person_id"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// NewGitHubPersonEmail creates a new GitHub person email association with a generated UUID
func NewGitHubPersonEmail(projectID, githubPersonID, personID string) *GitHubPersonEmail {
	return &GitHubPersonEmail{
		ID:             uuid.New().String(),
		ProjectID:      projectID,
		GitHubPersonID: githubPersonID,
		PersonID:       personID,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
}
