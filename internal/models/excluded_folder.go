package models

import (
	"time"

	"github.com/google/uuid"
)

// ExcludedFolder represents a folder that should be excluded from statistics
type ExcludedFolder struct {
	ID         string     `json:"id"`
	ProjectID  string     `json:"project_id"`
	FolderPath string     `json:"folder_path"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
	DeletedAt  *time.Time `json:"deleted_at,omitempty"`
}

// NewExcludedFolder creates a new ExcludedFolder with a generated ID
func NewExcludedFolder(projectID, folderPath string) *ExcludedFolder {
	return &ExcludedFolder{
		ID:         uuid.New().String(),
		ProjectID:  projectID,
		FolderPath: folderPath,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
}

// Validate validates the ExcludedFolder fields
func (ef *ExcludedFolder) Validate() error {
	if ef.ProjectID == "" {
		return &ValidationError{Field: "project_id", Message: "Project ID is required"}
	}
	if ef.FolderPath == "" {
		return &ValidationError{Field: "folder_path", Message: "Folder path is required"}
	}
	return nil
}
