package repositories

import (
	"database/sql"

	"github.com/alimgiray/gscope/internal/models"
)

type ScoreSettingsRepository struct {
	db *sql.DB
}

func NewScoreSettingsRepository(db *sql.DB) *ScoreSettingsRepository {
	return &ScoreSettingsRepository{
		db: db,
	}
}

// Create creates new score settings for a project
func (r *ScoreSettingsRepository) Create(settings *models.ScoreSettings) error {
	query := `
		INSERT INTO score_settings (id, project_id, additions, deletions, commits, pull_requests, comments)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	_, err := r.db.Exec(query,
		settings.ID,
		settings.ProjectID,
		settings.Additions,
		settings.Deletions,
		settings.Commits,
		settings.PullRequests,
		settings.Comments,
	)

	return err
}

// GetByProjectID retrieves score settings for a project
func (r *ScoreSettingsRepository) GetByProjectID(projectID string) (*models.ScoreSettings, error) {
	query := `
		SELECT id, project_id, additions, deletions, commits, pull_requests, comments, created_at, updated_at
		FROM score_settings 
		WHERE project_id = $1
	`

	settings := &models.ScoreSettings{}
	err := r.db.QueryRow(query, projectID).Scan(
		&settings.ID,
		&settings.ProjectID,
		&settings.Additions,
		&settings.Deletions,
		&settings.Commits,
		&settings.PullRequests,
		&settings.Comments,
		&settings.CreatedAt,
		&settings.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return settings, nil
}

// Update updates score settings
func (r *ScoreSettingsRepository) Update(settings *models.ScoreSettings) error {
	query := `
		UPDATE score_settings 
		SET additions = $1, deletions = $2, commits = $3, pull_requests = $4, comments = $5, updated_at = CURRENT_TIMESTAMP
		WHERE project_id = $6
	`

	result, err := r.db.Exec(query,
		settings.Additions,
		settings.Deletions,
		settings.Commits,
		settings.PullRequests,
		settings.Comments,
		settings.ProjectID,
	)

	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

// Delete deletes score settings for a project
func (r *ScoreSettingsRepository) Delete(projectID string) error {
	query := `DELETE FROM score_settings WHERE project_id = $1`

	result, err := r.db.Exec(query, projectID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}
