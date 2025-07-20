package repositories

import (
	"database/sql"

	"github.com/alimgiray/gscope/internal/models"
	"github.com/google/uuid"
)

type ProjectRepository struct {
	db *sql.DB
}

func NewProjectRepository(db *sql.DB) *ProjectRepository {
	return &ProjectRepository{
		db: db,
	}
}

// Create creates a new project
func (r *ProjectRepository) Create(project *models.Project) error {
	query := `
		INSERT INTO projects (id, name, owner_id, description)
		VALUES ($1, $2, $3, $4)
	`

	project.ID = uuid.New()

	_, err := r.db.Exec(query,
		project.ID,
		project.Name,
		project.OwnerID,
		project.Description,
	)

	return err
}

// GetByID retrieves a project by ID (excluding soft deleted)
func (r *ProjectRepository) GetByID(id string) (*models.Project, error) {
	query := `
		SELECT id, name, owner_id, description, created_at, updated_at, deleted_at
		FROM projects 
		WHERE id = $1 AND deleted_at IS NULL
	`

	project := &models.Project{}
	err := r.db.QueryRow(query, id).Scan(
		&project.ID,
		&project.Name,
		&project.OwnerID,
		&project.Description,
		&project.CreatedAt,
		&project.UpdatedAt,
		&project.DeletedAt,
	)

	if err != nil {
		return nil, err
	}

	return project, nil
}

// GetByOwnerID retrieves all projects for an owner (excluding soft deleted)
func (r *ProjectRepository) GetByOwnerID(ownerID string) ([]*models.Project, error) {
	query := `
		SELECT id, name, owner_id, description, created_at, updated_at, deleted_at
		FROM projects 
		WHERE owner_id = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(query, ownerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var projects []*models.Project
	for rows.Next() {
		project := &models.Project{}
		err := rows.Scan(
			&project.ID,
			&project.Name,
			&project.OwnerID,
			&project.Description,
			&project.CreatedAt,
			&project.UpdatedAt,
			&project.DeletedAt,
		)
		if err != nil {
			return nil, err
		}
		projects = append(projects, project)
	}

	return projects, nil
}

// Update updates a project
func (r *ProjectRepository) Update(project *models.Project) error {
	query := `
		UPDATE projects 
		SET name = $1, description = $2, updated_at = CURRENT_TIMESTAMP
		WHERE id = $3 AND deleted_at IS NULL
	`

	result, err := r.db.Exec(query,
		project.Name,
		project.Description,
		project.ID,
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

// Delete performs a soft delete of a project
func (r *ProjectRepository) Delete(id string) error {
	query := `
		UPDATE projects 
		SET deleted_at = CURRENT_TIMESTAMP
		WHERE id = $1 AND deleted_at IS NULL
	`

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
