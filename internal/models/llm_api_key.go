package models

import (
	"time"

	"github.com/google/uuid"
)

// LLMAPIKey represents an LLM API key configuration for a project
type LLMAPIKey struct {
	ID        uuid.UUID `json:"id" db:"id"`
	ProjectID uuid.UUID `json:"project_id" db:"project_id"`
	Provider  string    `json:"provider" db:"provider"` // Currently only "anthropic"
	APIKey    string    `json:"api_key" db:"api_key"`   // Stored as plaintext per requirement
	IsActive  bool      `json:"is_active" db:"is_active"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`

	// Associated project (for joins)
	Project *Project `json:"project,omitempty"`
}

// LLMAPIKeyRequest represents the request payload for creating/updating LLM API keys
type LLMAPIKeyRequest struct {
	Provider string `json:"provider" form:"provider" binding:"required"`
	APIKey   string `json:"api_key" form:"api_key" binding:"required"`
}

// Validate validates the LLM API key request
func (r *LLMAPIKeyRequest) Validate() error {
	if r.Provider != "anthropic" {
		return &ValidationError{Field: "provider", Message: "Only 'anthropic' provider is supported"}
	}

	if len(r.APIKey) < 10 {
		return &ValidationError{Field: "api_key", Message: "API key is too short"}
	}

	// Basic Anthropic API key format validation (starts with sk-ant-)
	if r.Provider == "anthropic" && len(r.APIKey) > 0 {
		if len(r.APIKey) < 20 || r.APIKey[:7] != "sk-ant-" {
			return &ValidationError{Field: "api_key", Message: "Invalid Anthropic API key format"}
		}
	}

	return nil
}

// MaskAPIKey returns a masked version of the API key for display purposes
func (k *LLMAPIKey) MaskAPIKey() string {
	if len(k.APIKey) < 8 {
		return "****"
	}
	// Show first 7 chars (sk-ant-) and last 4 chars
	return k.APIKey[:7] + "****" + k.APIKey[len(k.APIKey)-4:]
}

// GetDisplayProvider returns a user-friendly provider name
func (k *LLMAPIKey) GetDisplayProvider() string {
	switch k.Provider {
	case "anthropic":
		return "Anthropic Claude"
	default:
		return k.Provider
	}
}
