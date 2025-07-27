package models

import (
	"time"

	"github.com/google/uuid"
)

// EmailMerge represents an email merge operation
type EmailMerge struct {
	ID          string    `json:"id"`
	ProjectID   string    `json:"project_id"`
	SourceEmail string    `json:"source_email"` // The email being merged (will be hidden)
	TargetEmail string    `json:"target_email"` // The email to merge into (will be shown)
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// NewEmailMerge creates a new email merge with a generated UUID
func NewEmailMerge(projectID, sourceEmail, targetEmail string) *EmailMerge {
	return &EmailMerge{
		ID:          uuid.New().String(),
		ProjectID:   projectID,
		SourceEmail: sourceEmail,
		TargetEmail: targetEmail,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}
