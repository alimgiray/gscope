package models

import (
	"time"

	"github.com/google/uuid"
)

// ProjectUpdateSettings represents automatic update settings for a project
type ProjectUpdateSettings struct {
	ID        string     `json:"id"`
	ProjectID string     `json:"project_id"`
	IsEnabled bool       `json:"is_enabled"`
	Hour      int        `json:"hour"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
}

// NewProjectUpdateSettings creates a new ProjectUpdateSettings with a generated ID
func NewProjectUpdateSettings(projectID string, isEnabled bool, hour int) *ProjectUpdateSettings {
	return &ProjectUpdateSettings{
		ID:        uuid.New().String(),
		ProjectID: projectID,
		IsEnabled: isEnabled,
		Hour:      hour,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

// Validate validates the ProjectUpdateSettings fields
func (pus *ProjectUpdateSettings) Validate() error {
	if pus.ProjectID == "" {
		return &ValidationError{Field: "project_id", Message: "Project ID is required"}
	}
	if pus.Hour < 0 || pus.Hour > 23 {
		return &ValidationError{Field: "hour", Message: "Hour must be between 0 and 23"}
	}
	return nil
}
