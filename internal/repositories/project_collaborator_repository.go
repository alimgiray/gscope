package repositories

import (
	"database/sql"
	"sync"

	"github.com/alimgiray/gscope/internal/models"
)

type ProjectCollaboratorRepository struct {
	db *sql.DB
	mu sync.RWMutex
}

func NewProjectCollaboratorRepository(db *sql.DB) *ProjectCollaboratorRepository {
	return &ProjectCollaboratorRepository{db: db}
}

// Create creates a new project collaborator
func (r *ProjectCollaboratorRepository) Create(collaborator *models.ProjectCollaborator) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	query := `
		INSERT INTO project_collaborators (
			id, project_id, user_id, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?)
	`

	_, err := r.db.Exec(query,
		collaborator.ID, collaborator.ProjectID, collaborator.UserID,
		collaborator.CreatedAt, collaborator.UpdatedAt,
	)

	return err
}

// GetByID retrieves a project collaborator by ID
func (r *ProjectCollaboratorRepository) GetByID(id string) (*models.ProjectCollaborator, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	query := `
		SELECT id, project_id, user_id, created_at, updated_at
		FROM project_collaborators WHERE id = ?
	`

	collaborator := &models.ProjectCollaborator{}
	err := r.db.QueryRow(query, id).Scan(
		&collaborator.ID, &collaborator.ProjectID, &collaborator.UserID,
		&collaborator.CreatedAt, &collaborator.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return collaborator, nil
}

// GetByProjectID retrieves all collaborators for a project
func (r *ProjectCollaboratorRepository) GetByProjectID(projectID string) ([]*models.ProjectCollaborator, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	query := `
		SELECT id, project_id, user_id, created_at, updated_at
		FROM project_collaborators WHERE project_id = ?
		ORDER BY created_at ASC
	`

	rows, err := r.db.Query(query, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var collaborators []*models.ProjectCollaborator
	for rows.Next() {
		collaborator := &models.ProjectCollaborator{}
		err := rows.Scan(
			&collaborator.ID, &collaborator.ProjectID, &collaborator.UserID,
			&collaborator.CreatedAt, &collaborator.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		collaborators = append(collaborators, collaborator)
	}

	return collaborators, nil
}

// GetByUserID retrieves all projects where a user is a collaborator
func (r *ProjectCollaboratorRepository) GetByUserID(userID string) ([]*models.ProjectCollaborator, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	query := `
		SELECT id, project_id, user_id, created_at, updated_at
		FROM project_collaborators WHERE user_id = ?
		ORDER BY created_at ASC
	`

	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var collaborators []*models.ProjectCollaborator
	for rows.Next() {
		collaborator := &models.ProjectCollaborator{}
		err := rows.Scan(
			&collaborator.ID, &collaborator.ProjectID, &collaborator.UserID,
			&collaborator.CreatedAt, &collaborator.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		collaborators = append(collaborators, collaborator)
	}

	return collaborators, nil
}

// GetByProjectAndUserID retrieves a specific collaboration
func (r *ProjectCollaboratorRepository) GetByProjectAndUserID(projectID, userID string) (*models.ProjectCollaborator, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	query := `
		SELECT id, project_id, user_id, created_at, updated_at
		FROM project_collaborators WHERE project_id = ? AND user_id = ?
	`

	collaborator := &models.ProjectCollaborator{}
	err := r.db.QueryRow(query, projectID, userID).Scan(
		&collaborator.ID, &collaborator.ProjectID, &collaborator.UserID,
		&collaborator.CreatedAt, &collaborator.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return collaborator, nil
}

// Delete deletes a project collaborator by ID
func (r *ProjectCollaboratorRepository) Delete(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	query := `DELETE FROM project_collaborators WHERE id = ?`
	_, err := r.db.Exec(query, id)
	return err
}

// DeleteByProjectAndUserID deletes a specific collaboration
func (r *ProjectCollaboratorRepository) DeleteByProjectAndUserID(projectID, userID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	query := `DELETE FROM project_collaborators WHERE project_id = ? AND user_id = ?`
	_, err := r.db.Exec(query, projectID, userID)
	return err
}

// DeleteByProjectID deletes all collaborators for a project
func (r *ProjectCollaboratorRepository) DeleteByProjectID(projectID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	query := `DELETE FROM project_collaborators WHERE project_id = ?`
	_, err := r.db.Exec(query, projectID)
	return err
}

// DeleteByUserID deletes all collaborations for a user
func (r *ProjectCollaboratorRepository) DeleteByUserID(userID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	query := `DELETE FROM project_collaborators WHERE user_id = ?`
	_, err := r.db.Exec(query, userID)
	return err
}

// ExistsByProjectAndUserID checks if a collaboration exists
func (r *ProjectCollaboratorRepository) ExistsByProjectAndUserID(projectID, userID string) (bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	query := `SELECT COUNT(*) FROM project_collaborators WHERE project_id = ? AND user_id = ?`
	var count int
	err := r.db.QueryRow(query, projectID, userID).Scan(&count)
	return count > 0, err
}
