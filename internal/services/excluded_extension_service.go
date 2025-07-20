package services

import (
	"errors"
	"strings"

	"github.com/alimgiray/gscope/internal/models"
	"github.com/alimgiray/gscope/internal/repositories"
	"github.com/google/uuid"
)

type ExcludedExtensionService struct {
	excludedExtensionRepo *repositories.ExcludedExtensionRepository
}

func NewExcludedExtensionService(excludedExtensionRepo *repositories.ExcludedExtensionRepository) *ExcludedExtensionService {
	return &ExcludedExtensionService{
		excludedExtensionRepo: excludedExtensionRepo,
	}
}

// CreateExcludedExtension creates a new excluded extension for a project
func (s *ExcludedExtensionService) CreateExcludedExtension(projectID, extension string) error {
	if projectID == "" {
		return errors.New("project ID is required")
	}

	if extension == "" {
		return errors.New("extension is required")
	}

	// Validate UUID format
	if _, err := uuid.Parse(projectID); err != nil {
		return errors.New("invalid project ID format")
	}

	// Normalize extension (remove dot if present, convert to lowercase)
	extension = strings.TrimPrefix(strings.ToLower(extension), ".")
	if extension == "" {
		return errors.New("invalid extension format")
	}

	excludedExt := models.NewExcludedExtension(projectID, extension)
	return s.excludedExtensionRepo.Create(excludedExt)
}

// GetExcludedExtensionsByProjectID retrieves all excluded extensions for a project
func (s *ExcludedExtensionService) GetExcludedExtensionsByProjectID(projectID string) ([]*models.ExcludedExtension, error) {
	if projectID == "" {
		return nil, errors.New("project ID is required")
	}

	// Validate UUID format
	if _, err := uuid.Parse(projectID); err != nil {
		return nil, errors.New("invalid project ID format")
	}

	return s.excludedExtensionRepo.GetByProjectID(projectID)
}

// DeleteExcludedExtension deletes an excluded extension by ID
func (s *ExcludedExtensionService) DeleteExcludedExtension(id string) error {
	if id == "" {
		return errors.New("extension ID is required")
	}

	// Validate UUID format
	if _, err := uuid.Parse(id); err != nil {
		return errors.New("invalid extension ID format")
	}

	return s.excludedExtensionRepo.Delete(id)
}

// DeleteExcludedExtensionsByProjectID deletes all excluded extensions for a project
func (s *ExcludedExtensionService) DeleteExcludedExtensionsByProjectID(projectID string) error {
	if projectID == "" {
		return errors.New("project ID is required")
	}

	// Validate UUID format
	if _, err := uuid.Parse(projectID); err != nil {
		return errors.New("invalid project ID format")
	}

	return s.excludedExtensionRepo.DeleteByProjectID(projectID)
}
