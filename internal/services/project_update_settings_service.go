package services

import (
	"time"

	"github.com/alimgiray/gscope/internal/models"
	"github.com/alimgiray/gscope/internal/repositories"
)

type ProjectUpdateSettingsService struct {
	projectUpdateSettingsRepo *repositories.ProjectUpdateSettingsRepository
}

func NewProjectUpdateSettingsService(projectUpdateSettingsRepo *repositories.ProjectUpdateSettingsRepository) *ProjectUpdateSettingsService {
	return &ProjectUpdateSettingsService{
		projectUpdateSettingsRepo: projectUpdateSettingsRepo,
	}
}

// GetProjectUpdateSettings retrieves project update settings by project ID
func (s *ProjectUpdateSettingsService) GetProjectUpdateSettings(projectID string) (*models.ProjectUpdateSettings, error) {
	if projectID == "" {
		return nil, &models.ValidationError{Message: "Project ID is required"}
	}

	return s.projectUpdateSettingsRepo.GetByProjectID(projectID)
}

// CreateProjectUpdateSettings creates new project update settings
func (s *ProjectUpdateSettingsService) CreateProjectUpdateSettings(projectID string, isEnabled bool, hour int) (*models.ProjectUpdateSettings, error) {
	// Validate input
	if projectID == "" {
		return nil, &models.ValidationError{Message: "Project ID is required"}
	}
	if hour < 0 || hour > 23 {
		return nil, &models.ValidationError{Message: "Hour must be between 0 and 23"}
	}

	// Create new settings
	settings := models.NewProjectUpdateSettings(projectID, isEnabled, hour)
	if err := settings.Validate(); err != nil {
		return nil, err
	}

	// Save to database
	if err := s.projectUpdateSettingsRepo.Create(settings); err != nil {
		return nil, err
	}

	return settings, nil
}

// UpdateProjectUpdateSettings updates project update settings
func (s *ProjectUpdateSettingsService) UpdateProjectUpdateSettings(projectID string, isEnabled bool, hour int) (*models.ProjectUpdateSettings, error) {
	// Validate input
	if projectID == "" {
		return nil, &models.ValidationError{Message: "Project ID is required"}
	}
	if hour < 0 || hour > 23 {
		return nil, &models.ValidationError{Message: "Hour must be between 0 and 23"}
	}

	// Get existing settings
	existing, err := s.projectUpdateSettingsRepo.GetByProjectID(projectID)
	if err != nil {
		// If no existing settings, create new ones
		return s.CreateProjectUpdateSettings(projectID, isEnabled, hour)
	}

	// Update settings
	existing.IsEnabled = isEnabled
	existing.Hour = hour
	existing.UpdatedAt = time.Now()

	if err := existing.Validate(); err != nil {
		return nil, err
	}

	// Save to database
	if err := s.projectUpdateSettingsRepo.Update(existing); err != nil {
		return nil, err
	}

	return existing, nil
}

// UpsertProjectUpdateSettings creates or updates project update settings
func (s *ProjectUpdateSettingsService) UpsertProjectUpdateSettings(projectID string, isEnabled bool, hour int) (*models.ProjectUpdateSettings, error) {
	// Validate input
	if projectID == "" {
		return nil, &models.ValidationError{Message: "Project ID is required"}
	}
	if hour < 0 || hour > 23 {
		return nil, &models.ValidationError{Message: "Hour must be between 0 and 23"}
	}

	// Create or update settings
	settings := models.NewProjectUpdateSettings(projectID, isEnabled, hour)
	if err := settings.Validate(); err != nil {
		return nil, err
	}

	// Upsert to database
	if err := s.projectUpdateSettingsRepo.Upsert(settings); err != nil {
		return nil, err
	}

	return settings, nil
}

// DeleteProjectUpdateSettings deletes project update settings
func (s *ProjectUpdateSettingsService) DeleteProjectUpdateSettings(projectID string) error {
	if projectID == "" {
		return &models.ValidationError{Message: "Project ID is required"}
	}

	return s.projectUpdateSettingsRepo.DeleteByProjectID(projectID)
}

// GetAllEnabledProjectUpdateSettings retrieves all enabled project update settings
func (s *ProjectUpdateSettingsService) GetAllEnabledProjectUpdateSettings() ([]*models.ProjectUpdateSettings, error) {
	return s.projectUpdateSettingsRepo.GetAllEnabled()
}
