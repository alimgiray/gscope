package models

import (
	"time"

	"github.com/google/uuid"
)

type Project struct {
	ID          uuid.UUID  `json:"id"`
	Name        string     `json:"name"`
	OwnerID     uuid.UUID  `json:"owner_id"`
	Description string     `json:"description"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty"`
}

func (p *Project) Validate() error {
	if p.Name == "" {
		return ErrProjectNameRequired
	}
	return nil
}

// Common errors
var (
	ErrProjectNameRequired = &ValidationError{Field: "name", Message: "Project name is required"}
)

// ValidationError is defined in user.go, reusing it here
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}
