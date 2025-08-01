package models

import (
	"time"

	"github.com/google/uuid"
)

// ProjectGithubPerson represents the relationship between a project and a GitHub person
type ProjectGithubPerson struct {
	ID             string    `json:"id"`
	ProjectID      string    `json:"project_id"`
	GithubPersonID string    `json:"github_person_id"`
	SourceType     string    `json:"source_type"` // "pull_request", "contributor", "commit_author"
	IsDeleted      bool      `json:"is_deleted"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// NewProjectGithubPerson creates a new ProjectGithubPerson with a generated UUID
func NewProjectGithubPerson(projectID, githubPersonID, sourceType string) *ProjectGithubPerson {
	now := time.Now()
	return &ProjectGithubPerson{
		ID:             uuid.New().String(),
		ProjectID:      projectID,
		GithubPersonID: githubPersonID,
		SourceType:     sourceType,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
}

// Validate validates the ProjectGithubPerson fields
func (p *ProjectGithubPerson) Validate() error {
	if p.ProjectID == "" {
		return &ValidationError{Field: "project_id", Message: "Project ID is required"}
	}
	if p.GithubPersonID == "" {
		return &ValidationError{Field: "github_person_id", Message: "GitHub Person ID is required"}
	}
	if p.SourceType == "" {
		return &ValidationError{Field: "source_type", Message: "Source type is required"}
	}

	// Validate source type
	validSourceTypes := []string{"pull_request", "contributor", "commit_author"}
	isValid := false
	for _, validType := range validSourceTypes {
		if p.SourceType == validType {
			isValid = true
			break
		}
	}
	if !isValid {
		return &ValidationError{Field: "source_type", Message: "Source type must be one of: pull_request, contributor, commit_author"}
	}

	return nil
}
