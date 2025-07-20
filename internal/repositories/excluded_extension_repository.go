package repositories

import (
	"database/sql"

	"github.com/alimgiray/gscope/internal/models"
)

type ExcludedExtensionRepository struct {
	db *sql.DB
}

func NewExcludedExtensionRepository(db *sql.DB) *ExcludedExtensionRepository {
	return &ExcludedExtensionRepository{
		db: db,
	}
}

// Create creates a new excluded extension
func (r *ExcludedExtensionRepository) Create(extension *models.ExcludedExtension) error {
	query := `
		INSERT INTO excluded_extensions (id, project_id, extension)
		VALUES ($1, $2, $3)
	`

	_, err := r.db.Exec(query,
		extension.ID,
		extension.ProjectID,
		extension.Extension,
	)

	return err
}

// GetByProjectID retrieves all excluded extensions for a project
func (r *ExcludedExtensionRepository) GetByProjectID(projectID string) ([]*models.ExcludedExtension, error) {
	query := `
		SELECT id, project_id, extension, created_at
		FROM excluded_extensions 
		WHERE project_id = $1
		ORDER BY extension
	`

	rows, err := r.db.Query(query, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var extensions []*models.ExcludedExtension
	for rows.Next() {
		extension := &models.ExcludedExtension{}
		err := rows.Scan(
			&extension.ID,
			&extension.ProjectID,
			&extension.Extension,
			&extension.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		extensions = append(extensions, extension)
	}

	return extensions, nil
}

// Delete deletes an excluded extension by ID
func (r *ExcludedExtensionRepository) Delete(id string) error {
	query := `DELETE FROM excluded_extensions WHERE id = $1`

	result, err := r.db.Exec(query, id)
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

// DeleteByProjectID deletes all excluded extensions for a project
func (r *ExcludedExtensionRepository) DeleteByProjectID(projectID string) error {
	query := `DELETE FROM excluded_extensions WHERE project_id = $1`

	_, err := r.db.Exec(query, projectID)
	return err
}
