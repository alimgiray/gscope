package models

import (
	"time"

	"github.com/google/uuid"
)

// ProjectCollaborator represents a user's collaboration on a project
type ProjectCollaborator struct {
	ID        string    `json:"id"`
	ProjectID string    `json:"project_id"`
	UserID    string    `json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// NewProjectCollaborator creates a new ProjectCollaborator with a generated UUID
func NewProjectCollaborator(projectID, userID string) *ProjectCollaborator {
	return &ProjectCollaborator{
		ID:        uuid.New().String(),
		ProjectID: projectID,
		UserID:    userID,
	}
}

// Validate validates the ProjectCollaborator
func (pc *ProjectCollaborator) Validate() error {
	if pc.ProjectID == "" {
		return &ValidationError{Field: "project_id", Message: "Project ID is required"}
	}
	if pc.UserID == "" {
		return &ValidationError{Field: "user_id", Message: "User ID is required"}
	}
	return nil
}
