package repositories

import (
	"database/sql"
	"sync"

	"github.com/alimgiray/gscope/internal/models"
)

type ProjectUpdateSettingsRepository struct {
	db *sql.DB
	mu sync.RWMutex
}

func NewProjectUpdateSettingsRepository(db *sql.DB) *ProjectUpdateSettingsRepository {
	return &ProjectUpdateSettingsRepository{db: db}
}

// Create creates a new project update settings
func (r *ProjectUpdateSettingsRepository) Create(settings *models.ProjectUpdateSettings) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	query := `
		INSERT INTO project_update_settings (id, project_id, is_enabled, hour, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`

	_, err := r.db.Exec(query,
		settings.ID, settings.ProjectID, settings.IsEnabled, settings.Hour, settings.CreatedAt, settings.UpdatedAt,
	)

	return err
}

// GetByID retrieves project update settings by ID
func (r *ProjectUpdateSettingsRepository) GetByID(id string) (*models.ProjectUpdateSettings, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	query := `
		SELECT id, project_id, is_enabled, hour, created_at, updated_at, deleted_at
		FROM project_update_settings WHERE id = ? AND deleted_at IS NULL
	`

	var settings models.ProjectUpdateSettings
	err := r.db.QueryRow(query, id).Scan(
		&settings.ID, &settings.ProjectID, &settings.IsEnabled, &settings.Hour, &settings.CreatedAt, &settings.UpdatedAt, &settings.DeletedAt,
	)

	if err != nil {
		return nil, err
	}

	return &settings, nil
}

// GetByProjectID retrieves project update settings by project ID
func (r *ProjectUpdateSettingsRepository) GetByProjectID(projectID string) (*models.ProjectUpdateSettings, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	query := `
		SELECT id, project_id, is_enabled, hour, created_at, updated_at, deleted_at
		FROM project_update_settings WHERE project_id = ? AND deleted_at IS NULL
	`

	var settings models.ProjectUpdateSettings
	err := r.db.QueryRow(query, projectID).Scan(
		&settings.ID, &settings.ProjectID, &settings.IsEnabled, &settings.Hour, &settings.CreatedAt, &settings.UpdatedAt, &settings.DeletedAt,
	)

	if err != nil {
		return nil, err
	}

	return &settings, nil
}

// Update updates project update settings
func (r *ProjectUpdateSettingsRepository) Update(settings *models.ProjectUpdateSettings) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	query := `
		UPDATE project_update_settings
		SET project_id = ?, is_enabled = ?, hour = ?, updated_at = ?
		WHERE id = ? AND deleted_at IS NULL
	`

	_, err := r.db.Exec(query,
		settings.ProjectID, settings.IsEnabled, settings.Hour, settings.UpdatedAt, settings.ID,
	)

	return err
}

// Upsert creates or updates project update settings
func (r *ProjectUpdateSettingsRepository) Upsert(settings *models.ProjectUpdateSettings) error {
	existing, err := r.GetByProjectID(settings.ProjectID)
	if err != nil && err != sql.ErrNoRows {
		return err
	}

	if existing != nil {
		settings.ID = existing.ID
		settings.CreatedAt = existing.CreatedAt
		return r.Update(settings)
	}

	return r.Create(settings)
}

// Delete soft deletes project update settings
func (r *ProjectUpdateSettingsRepository) Delete(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	query := `UPDATE project_update_settings SET deleted_at = CURRENT_TIMESTAMP WHERE id = ?`
	_, err := r.db.Exec(query, id)
	return err
}

// DeleteByProjectID soft deletes project update settings by project ID
func (r *ProjectUpdateSettingsRepository) DeleteByProjectID(projectID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	query := `UPDATE project_update_settings SET deleted_at = CURRENT_TIMESTAMP WHERE project_id = ?`
	_, err := r.db.Exec(query, projectID)
	return err
}

// GetAllEnabled retrieves all enabled project update settings
func (r *ProjectUpdateSettingsRepository) GetAllEnabled() ([]*models.ProjectUpdateSettings, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	query := `
		SELECT id, project_id, is_enabled, hour, created_at, updated_at, deleted_at
		FROM project_update_settings
		WHERE is_enabled = 1 AND deleted_at IS NULL
		ORDER BY hour ASC
	`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var settings []*models.ProjectUpdateSettings
	for rows.Next() {
		var setting models.ProjectUpdateSettings
		err := rows.Scan(
			&setting.ID, &setting.ProjectID, &setting.IsEnabled, &setting.Hour, &setting.CreatedAt, &setting.UpdatedAt, &setting.DeletedAt,
		)
		if err != nil {
			return nil, err
		}
		settings = append(settings, &setting)
	}

	return settings, nil
}
