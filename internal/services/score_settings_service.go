package services

import (
	"errors"

	"github.com/alimgiray/gscope/internal/models"
	"github.com/alimgiray/gscope/internal/repositories"
	"github.com/google/uuid"
)

type ScoreSettingsService struct {
	scoreSettingsRepo *repositories.ScoreSettingsRepository
}

func NewScoreSettingsService(scoreSettingsRepo *repositories.ScoreSettingsRepository) *ScoreSettingsService {
	return &ScoreSettingsService{
		scoreSettingsRepo: scoreSettingsRepo,
	}
}

// CreateScoreSettings creates default score settings for a project
func (s *ScoreSettingsService) CreateScoreSettings(projectID string) error {
	if projectID == "" {
		return errors.New("project ID is required")
	}

	// Validate UUID format
	if _, err := uuid.Parse(projectID); err != nil {
		return errors.New("invalid project ID format")
	}

	settings := models.NewScoreSettings(projectID)
	return s.scoreSettingsRepo.Create(settings)
}

// GetScoreSettingsByProjectID retrieves score settings for a project
func (s *ScoreSettingsService) GetScoreSettingsByProjectID(projectID string) (*models.ScoreSettings, error) {
	if projectID == "" {
		return nil, errors.New("project ID is required")
	}

	// Validate UUID format
	if _, err := uuid.Parse(projectID); err != nil {
		return nil, errors.New("invalid project ID format")
	}

	return s.scoreSettingsRepo.GetByProjectID(projectID)
}

// UpdateScoreSettings updates score settings for a project
func (s *ScoreSettingsService) UpdateScoreSettings(settings *models.ScoreSettings) error {
	if settings.ProjectID == "" {
		return errors.New("project ID is required")
	}

	// Validate UUID format
	if _, err := uuid.Parse(settings.ProjectID); err != nil {
		return errors.New("invalid project ID format")
	}

	// Validate score values (should be positive)
	if settings.Additions < 0 || settings.Deletions < 0 || settings.Commits < 0 ||
		settings.PullRequests < 0 || settings.Comments < 0 {
		return errors.New("score values must be non-negative")
	}

	return s.scoreSettingsRepo.Update(settings)
}

// DeleteScoreSettings deletes score settings for a project
func (s *ScoreSettingsService) DeleteScoreSettings(projectID string) error {
	if projectID == "" {
		return errors.New("project ID is required")
	}

	// Validate UUID format
	if _, err := uuid.Parse(projectID); err != nil {
		return errors.New("invalid project ID format")
	}

	return s.scoreSettingsRepo.Delete(projectID)
}
