package services

import (
	"fmt"
	"time"

	"github.com/alimgiray/gscope/internal/models"
	"github.com/alimgiray/gscope/internal/repositories"
	"github.com/google/uuid"
)

type LLMAPIKeyService struct {
	llmAPIKeyRepo              *repositories.LLMAPIKeyRepository
	projectCollaboratorService *ProjectCollaboratorService
}

func NewLLMAPIKeyService(llmAPIKeyRepo *repositories.LLMAPIKeyRepository, projectCollaboratorService *ProjectCollaboratorService) *LLMAPIKeyService {
	return &LLMAPIKeyService{
		llmAPIKeyRepo:              llmAPIKeyRepo,
		projectCollaboratorService: projectCollaboratorService,
	}
}

// CreateOrUpdateAPIKey creates or updates an LLM API key for a project (owner-only)
func (s *LLMAPIKeyService) CreateOrUpdateAPIKey(projectID, userID uuid.UUID, request *models.LLMAPIKeyRequest) (*models.LLMAPIKey, error) {
	// Validate that user is project owner
	accessType, err := s.projectCollaboratorService.GetProjectAccessType(projectID.String(), userID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to check project access: %w", err)
	}

	if accessType != "owner" {
		return nil, fmt.Errorf("only project owners can manage LLM API keys")
	}

	// Validate request
	if err := request.Validate(); err != nil {
		return nil, err
	}

	// Check if key already exists
	existingKey, err := s.llmAPIKeyRepo.GetByProjectID(projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing key: %w", err)
	}

	now := time.Now()

	if existingKey != nil {
		// Update existing key
		existingKey.Provider = request.Provider
		existingKey.APIKey = request.APIKey
		existingKey.UpdatedAt = now

		err = s.llmAPIKeyRepo.Update(existingKey)
		if err != nil {
			return nil, fmt.Errorf("failed to update API key: %w", err)
		}

		return existingKey, nil
	} else {
		// Create new key
		key := &models.LLMAPIKey{
			ID:        uuid.New(),
			ProjectID: projectID,
			Provider:  request.Provider,
			APIKey:    request.APIKey,
			IsActive:  true,
			CreatedAt: now,
			UpdatedAt: now,
		}

		err = s.llmAPIKeyRepo.Create(key)
		if err != nil {
			return nil, fmt.Errorf("failed to create API key: %w", err)
		}

		return key, nil
	}
}

// GetAPIKey gets the LLM API key for a project (with access control)
func (s *LLMAPIKeyService) GetAPIKey(projectID, userID uuid.UUID) (*models.LLMAPIKey, error) {
	// Validate that user has access to project
	accessType, err := s.projectCollaboratorService.GetProjectAccessType(projectID.String(), userID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to check project access: %w", err)
	}

	if accessType == "none" {
		return nil, fmt.Errorf("access denied")
	}

	key, err := s.llmAPIKeyRepo.GetByProjectID(projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get API key: %w", err)
	}

	return key, nil
}

// DeleteAPIKey deletes an LLM API key (owner-only)
func (s *LLMAPIKeyService) DeleteAPIKey(projectID, userID uuid.UUID) error {
	// Validate that user is project owner
	accessType, err := s.projectCollaboratorService.GetProjectAccessType(projectID.String(), userID.String())
	if err != nil {
		return fmt.Errorf("failed to check project access: %w", err)
	}

	if accessType != "owner" {
		return fmt.Errorf("only project owners can delete LLM API keys")
	}

	// Get existing key to validate it exists
	existingKey, err := s.llmAPIKeyRepo.GetByProjectID(projectID)
	if err != nil {
		return fmt.Errorf("failed to get existing key: %w", err)
	}

	if existingKey == nil {
		return fmt.Errorf("no API key found for this project")
	}

	err = s.llmAPIKeyRepo.Delete(existingKey.ID)
	if err != nil {
		return fmt.Errorf("failed to delete API key: %w", err)
	}

	return nil
}

// HasActiveKey checks if a project has an active LLM API key
func (s *LLMAPIKeyService) HasActiveKey(projectID uuid.UUID) (bool, error) {
	return s.llmAPIKeyRepo.HasActiveKey(projectID)
}

// GetAPIKeyForInternal gets the raw API key for internal use (no user validation)
// This method is for internal services that need to use the API key
func (s *LLMAPIKeyService) GetAPIKeyForInternal(projectID uuid.UUID) (*models.LLMAPIKey, error) {
	return s.llmAPIKeyRepo.GetByProjectID(projectID)
}
