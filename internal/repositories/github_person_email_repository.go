package repositories

import (
	"database/sql"

	"github.com/alimgiray/gscope/internal/models"
)

type GitHubPersonEmailRepository struct {
	db *sql.DB
}

func NewGitHubPersonEmailRepository(db *sql.DB) *GitHubPersonEmailRepository {
	return &GitHubPersonEmailRepository{db: db}
}

// Create creates a new GitHub person email association
func (r *GitHubPersonEmailRepository) Create(gpe *models.GitHubPersonEmail) error {
	query := `INSERT INTO github_people_emails (id, project_id, github_person_id, person_id, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?)`
	_, err := r.db.Exec(query, gpe.ID, gpe.ProjectID, gpe.GitHubPersonID, gpe.PersonID, gpe.CreatedAt, gpe.UpdatedAt)
	return err
}

// GetByID retrieves a GitHub person email association by ID
func (r *GitHubPersonEmailRepository) GetByID(id string) (*models.GitHubPersonEmail, error) {
	query := `SELECT id, project_id, github_person_id, person_id, created_at, updated_at FROM github_people_emails WHERE id = ?`
	gpe := &models.GitHubPersonEmail{}
	err := r.db.QueryRow(query, id).Scan(&gpe.ID, &gpe.ProjectID, &gpe.GitHubPersonID, &gpe.PersonID, &gpe.CreatedAt, &gpe.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return gpe, nil
}

// GetByProjectID retrieves all GitHub person email associations for a project
func (r *GitHubPersonEmailRepository) GetByProjectID(projectID string) ([]*models.GitHubPersonEmail, error) {
	query := `SELECT id, project_id, github_person_id, person_id, created_at, updated_at FROM github_people_emails WHERE project_id = ? ORDER BY created_at DESC`
	rows, err := r.db.Query(query, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var associations []*models.GitHubPersonEmail
	for rows.Next() {
		gpe := &models.GitHubPersonEmail{}
		err := rows.Scan(&gpe.ID, &gpe.ProjectID, &gpe.GitHubPersonID, &gpe.PersonID, &gpe.CreatedAt, &gpe.UpdatedAt)
		if err != nil {
			return nil, err
		}
		associations = append(associations, gpe)
	}
	return associations, nil
}

// GetByPersonID retrieves a GitHub person email association by person ID and project
func (r *GitHubPersonEmailRepository) GetByPersonID(projectID, personID string) (*models.GitHubPersonEmail, error) {
	query := `SELECT id, project_id, github_person_id, person_id, created_at, updated_at FROM github_people_emails WHERE project_id = ? AND person_id = ?`
	gpe := &models.GitHubPersonEmail{}
	err := r.db.QueryRow(query, projectID, personID).Scan(&gpe.ID, &gpe.ProjectID, &gpe.GitHubPersonID, &gpe.PersonID, &gpe.CreatedAt, &gpe.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return gpe, nil
}

// GetByGitHubPersonID retrieves all email associations for a GitHub person in a project
func (r *GitHubPersonEmailRepository) GetByGitHubPersonID(projectID, githubPersonID string) ([]*models.GitHubPersonEmail, error) {
	query := `SELECT id, project_id, github_person_id, person_id, created_at, updated_at FROM github_people_emails WHERE project_id = ? AND github_person_id = ? ORDER BY created_at DESC`
	rows, err := r.db.Query(query, projectID, githubPersonID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var associations []*models.GitHubPersonEmail
	for rows.Next() {
		gpe := &models.GitHubPersonEmail{}
		err := rows.Scan(&gpe.ID, &gpe.ProjectID, &gpe.GitHubPersonID, &gpe.PersonID, &gpe.CreatedAt, &gpe.UpdatedAt)
		if err != nil {
			return nil, err
		}
		associations = append(associations, gpe)
	}
	return associations, nil
}

// Update updates a GitHub person email association
func (r *GitHubPersonEmailRepository) Update(gpe *models.GitHubPersonEmail) error {
	query := `UPDATE github_people_emails SET project_id = ?, github_person_id = ?, person_id = ?, updated_at = ? WHERE id = ?`
	_, err := r.db.Exec(query, gpe.ProjectID, gpe.GitHubPersonID, gpe.PersonID, gpe.UpdatedAt, gpe.ID)
	return err
}

// Delete deletes a GitHub person email association by ID
func (r *GitHubPersonEmailRepository) Delete(id string) error {
	query := `DELETE FROM github_people_emails WHERE id = ?`
	_, err := r.db.Exec(query, id)
	return err
}

// DeleteByProjectID deletes all GitHub person email associations for a project
func (r *GitHubPersonEmailRepository) DeleteByProjectID(projectID string) error {
	query := `DELETE FROM github_people_emails WHERE project_id = ?`
	_, err := r.db.Exec(query, projectID)
	return err
}

// DeleteByPersonID deletes a GitHub person email association by person ID and project
func (r *GitHubPersonEmailRepository) DeleteByPersonID(projectID, personID string) error {
	query := `DELETE FROM github_people_emails WHERE project_id = ? AND person_id = ?`
	_, err := r.db.Exec(query, projectID, personID)
	return err
}
