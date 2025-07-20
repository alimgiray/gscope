package models

import (
	"time"

	"github.com/google/uuid"
)

type ExcludedExtension struct {
	ID        string    `json:"id"`
	ProjectID string    `json:"project_id"`
	Extension string    `json:"extension"`
	CreatedAt time.Time `json:"created_at"`
}

func NewExcludedExtension(projectID, extension string) *ExcludedExtension {
	return &ExcludedExtension{
		ID:        uuid.New().String(),
		ProjectID: projectID,
		Extension: extension,
	}
}
