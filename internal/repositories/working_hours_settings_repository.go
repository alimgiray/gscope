package repositories

import (
	"database/sql"
	"fmt"

	"github.com/alimgiray/gscope/internal/models"
)

type WorkingHoursSettingsRepository struct {
	db *sql.DB
}

func NewWorkingHoursSettingsRepository(db *sql.DB) *WorkingHoursSettingsRepository {
	return &WorkingHoursSettingsRepository{db: db}
}

// Create creates a new working hours settings
func (r *WorkingHoursSettingsRepository) Create(settings *models.WorkingHoursSettings) error {
	query := `
		INSERT INTO working_hours_settings (
			id, project_id, start_hour, end_hour, 
			monday, tuesday, wednesday, thursday, friday, saturday, sunday
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`

	_, err := r.db.Exec(query,
		settings.ID, settings.ProjectID, settings.StartHour, settings.EndHour,
		settings.Monday, settings.Tuesday, settings.Wednesday, settings.Thursday,
		settings.Friday, settings.Saturday, settings.Sunday,
	)

	if err != nil {
		return fmt.Errorf("error creating working hours settings: %v", err)
	}

	return nil
}

// GetByProjectID retrieves working hours settings for a project
func (r *WorkingHoursSettingsRepository) GetByProjectID(projectID string) (*models.WorkingHoursSettings, error) {
	query := `
		SELECT id, project_id, start_hour, end_hour, 
		       monday, tuesday, wednesday, thursday, friday, saturday, sunday,
		       created_at, updated_at
		FROM working_hours_settings 
		WHERE project_id = $1
	`

	var settings models.WorkingHoursSettings
	err := r.db.QueryRow(query, projectID).Scan(
		&settings.ID, &settings.ProjectID, &settings.StartHour, &settings.EndHour,
		&settings.Monday, &settings.Tuesday, &settings.Wednesday, &settings.Thursday,
		&settings.Friday, &settings.Saturday, &settings.Sunday,
		&settings.CreatedAt, &settings.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // No settings found
		}
		return nil, fmt.Errorf("error getting working hours settings: %v", err)
	}

	return &settings, nil
}

// Update updates working hours settings
func (r *WorkingHoursSettingsRepository) Update(settings *models.WorkingHoursSettings) error {
	query := `
		UPDATE working_hours_settings 
		SET start_hour = $1, end_hour = $2, 
		    monday = $3, tuesday = $4, wednesday = $5, thursday = $6, 
		    friday = $7, saturday = $8, sunday = $9
		WHERE project_id = $10
	`

	result, err := r.db.Exec(query,
		settings.StartHour, settings.EndHour,
		settings.Monday, settings.Tuesday, settings.Wednesday, settings.Thursday,
		settings.Friday, settings.Saturday, settings.Sunday,
		settings.ProjectID,
	)

	if err != nil {
		return fmt.Errorf("error updating working hours settings: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error getting rows affected: %v", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("no working hours settings found for project %s", settings.ProjectID)
	}

	return nil
}

// CreateOrUpdate creates or updates working hours settings
func (r *WorkingHoursSettingsRepository) CreateOrUpdate(settings *models.WorkingHoursSettings) error {
	existing, err := r.GetByProjectID(settings.ProjectID)
	if err != nil {
		return err
	}
	
	if existing == nil {
		// Create new settings
		if settings.ID == "" {
			settings.ID = "default" // Simple ID for SQLite
		}
		return r.Create(settings)
	} else {
		// Update existing settings
		settings.ID = existing.ID
		return r.Update(settings)
	}
}

// DeleteByProjectID deletes working hours settings for a project
func (r *WorkingHoursSettingsRepository) DeleteByProjectID(projectID string) error {
	query := `DELETE FROM working_hours_settings WHERE project_id = $1`

	result, err := r.db.Exec(query, projectID)
	if err != nil {
		return fmt.Errorf("error deleting working hours settings: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error getting rows affected: %v", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("no working hours settings found for project %s", projectID)
	}

	return nil
}
