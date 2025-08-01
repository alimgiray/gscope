package services

import (
	"fmt"
	"time"

	"github.com/alimgiray/gscope/internal/models"
	"github.com/alimgiray/gscope/internal/repositories"
)

type WorkingHoursSettingsService struct {
	workingHoursSettingsRepo *repositories.WorkingHoursSettingsRepository
}

func NewWorkingHoursSettingsService(workingHoursSettingsRepo *repositories.WorkingHoursSettingsRepository) *WorkingHoursSettingsService {
	return &WorkingHoursSettingsService{
		workingHoursSettingsRepo: workingHoursSettingsRepo,
	}
}

// GetByProjectID retrieves working hours settings for a project
func (s *WorkingHoursSettingsService) GetByProjectID(projectID string) (*models.WorkingHoursSettings, error) {
	settings, err := s.workingHoursSettingsRepo.GetByProjectID(projectID)
	if err != nil {
		return nil, fmt.Errorf("error getting working hours settings: %v", err)
	}

		// If no settings exist, return default settings
	if settings == nil {
		settings = &models.WorkingHoursSettings{
			ProjectID: projectID,
			StartHour: 9,
			EndHour:   18,
			Monday:    true,
			Tuesday:   true,
			Wednesday: true,
			Thursday:  true,
			Friday:    true,
			Saturday:  false,
			Sunday:    false,
		}
	}

	return settings, nil
}

// CreateOrUpdate creates or updates working hours settings for a project
func (s *WorkingHoursSettingsService) CreateOrUpdate(settings *models.WorkingHoursSettings) error {
	// Validate settings
	if err := s.validateSettings(settings); err != nil {
		return err
	}

	return s.workingHoursSettingsRepo.CreateOrUpdate(settings)
}

// DeleteByProjectID deletes working hours settings for a project
func (s *WorkingHoursSettingsService) DeleteByProjectID(projectID string) error {
	return s.workingHoursSettingsRepo.DeleteByProjectID(projectID)
}

// validateSettings validates working hours settings
func (s *WorkingHoursSettingsService) validateSettings(settings *models.WorkingHoursSettings) error {
	if settings.StartHour < 0 || settings.StartHour > 23 {
		return fmt.Errorf("start hour must be between 0 and 23")
	}

	if settings.EndHour < 0 || settings.EndHour > 23 {
		return fmt.Errorf("end hour must be between 0 and 23")
	}

	if settings.StartHour >= settings.EndHour {
		return fmt.Errorf("start hour must be before end hour")
	}

	// At least one day must be selected
	if !settings.Monday && !settings.Tuesday && !settings.Wednesday &&
		!settings.Thursday && !settings.Friday && !settings.Saturday && !settings.Sunday {
		return fmt.Errorf("at least one working day must be selected")
	}

	return nil
}

// IsOvertime checks if a given time is outside working hours based on project settings
func (s *WorkingHoursSettingsService) IsOvertime(projectID string, t time.Time) (bool, error) {
	settings, err := s.GetByProjectID(projectID)
	if err != nil {
		return false, err
	}

	// For commit_date: time already has timezone info, use as-is
	// For github_created_at: time is UTC, convert to local
	var localTime time.Time
	if t.Location() == time.UTC {
		// UTC time (from PR reviews) - convert to local
		localTime = t.Local()
	} else {
		// Timezone-aware time (from commits) - use as-is
		localTime = t
	}

	// Check if it's a working day
	weekday := localTime.Weekday()
	if !settings.IsWorkingDay(weekday) {
		return true, nil // Overtime if not a working day
	}

	// Check if it's outside working hours
	hour := localTime.Hour()
	isOvertime := hour < settings.StartHour || hour >= settings.EndHour
	return isOvertime, nil
}
