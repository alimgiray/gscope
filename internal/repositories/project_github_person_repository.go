package repositories

import (
	"database/sql"
	"time"

	"github.com/alimgiray/gscope/internal/models"
)

type ProjectGithubPersonRepository struct {
	db *sql.DB
}

func NewProjectGithubPersonRepository(db *sql.DB) *ProjectGithubPersonRepository {
	return &ProjectGithubPersonRepository{db: db}
}

func (r *ProjectGithubPersonRepository) Create(projectGithubPerson *models.ProjectGithubPerson) error {
	if err := projectGithubPerson.Validate(); err != nil {
		return err
	}

	query := `
		INSERT INTO project_github_people (id, project_id, github_person_id, source_type, is_deleted, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`

	_, err := r.db.Exec(query,
		projectGithubPerson.ID,
		projectGithubPerson.ProjectID,
		projectGithubPerson.GithubPersonID,
		projectGithubPerson.SourceType,
		projectGithubPerson.IsDeleted,
		projectGithubPerson.CreatedAt,
		projectGithubPerson.UpdatedAt,
	)

	return err
}

func (r *ProjectGithubPersonRepository) GetByID(id string) (*models.ProjectGithubPerson, error) {
	query := `
		SELECT id, project_id, github_person_id, source_type, is_deleted, created_at, updated_at
		FROM project_github_people
		WHERE id = ?
	`

	var projectGithubPerson models.ProjectGithubPerson
	err := r.db.QueryRow(query, id).Scan(
		&projectGithubPerson.ID,
		&projectGithubPerson.ProjectID,
		&projectGithubPerson.GithubPersonID,
		&projectGithubPerson.SourceType,
		&projectGithubPerson.IsDeleted,
		&projectGithubPerson.CreatedAt,
		&projectGithubPerson.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &projectGithubPerson, nil
}

func (r *ProjectGithubPersonRepository) GetByProjectID(projectID string) ([]*models.ProjectGithubPerson, error) {
	query := `
		SELECT id, project_id, github_person_id, source_type, is_deleted, created_at, updated_at
		FROM project_github_people
		WHERE project_id = ? AND is_deleted = 0
		ORDER BY created_at ASC
	`

	rows, err := r.db.Query(query, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var projectGithubPeople []*models.ProjectGithubPerson
	for rows.Next() {
		var projectGithubPerson models.ProjectGithubPerson
		err := rows.Scan(
			&projectGithubPerson.ID,
			&projectGithubPerson.ProjectID,
			&projectGithubPerson.GithubPersonID,
			&projectGithubPerson.SourceType,
			&projectGithubPerson.IsDeleted,
			&projectGithubPerson.CreatedAt,
			&projectGithubPerson.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		projectGithubPeople = append(projectGithubPeople, &projectGithubPerson)
	}

	return projectGithubPeople, nil
}

// GetByProjectIDIncludingDeleted returns all project github people including deleted ones
func (r *ProjectGithubPersonRepository) GetByProjectIDIncludingDeleted(projectID string) ([]*models.ProjectGithubPerson, error) {
	query := `
		SELECT id, project_id, github_person_id, source_type, is_deleted, created_at, updated_at
		FROM project_github_people
		WHERE project_id = ?
		ORDER BY is_deleted ASC, created_at ASC
	`

	rows, err := r.db.Query(query, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var projectGithubPeople []*models.ProjectGithubPerson
	for rows.Next() {
		var projectGithubPerson models.ProjectGithubPerson
		err := rows.Scan(
			&projectGithubPerson.ID,
			&projectGithubPerson.ProjectID,
			&projectGithubPerson.GithubPersonID,
			&projectGithubPerson.SourceType,
			&projectGithubPerson.IsDeleted,
			&projectGithubPerson.CreatedAt,
			&projectGithubPerson.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		projectGithubPeople = append(projectGithubPeople, &projectGithubPerson)
	}

	return projectGithubPeople, nil
}

func (r *ProjectGithubPersonRepository) GetByGithubPersonID(githubPersonID string) ([]*models.ProjectGithubPerson, error) {
	query := `
		SELECT id, project_id, github_person_id, source_type, is_deleted, created_at, updated_at
		FROM project_github_people
		WHERE github_person_id = ? AND is_deleted = 0
		ORDER BY created_at ASC
	`

	rows, err := r.db.Query(query, githubPersonID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var projectGithubPeople []*models.ProjectGithubPerson
	for rows.Next() {
		var projectGithubPerson models.ProjectGithubPerson
		err := rows.Scan(
			&projectGithubPerson.ID,
			&projectGithubPerson.ProjectID,
			&projectGithubPerson.GithubPersonID,
			&projectGithubPerson.SourceType,
			&projectGithubPerson.IsDeleted,
			&projectGithubPerson.CreatedAt,
			&projectGithubPerson.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		projectGithubPeople = append(projectGithubPeople, &projectGithubPerson)
	}

	return projectGithubPeople, nil
}

func (r *ProjectGithubPersonRepository) Upsert(projectGithubPerson *models.ProjectGithubPerson) error {
	if err := projectGithubPerson.Validate(); err != nil {
		return err
	}

	query := `
		INSERT INTO project_github_people (id, project_id, github_person_id, source_type, is_deleted, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(project_id, github_person_id, is_deleted) DO UPDATE SET
			source_type = excluded.source_type,
			updated_at = excluded.updated_at
	`

	projectGithubPerson.UpdatedAt = time.Now()
	_, err := r.db.Exec(query,
		projectGithubPerson.ID,
		projectGithubPerson.ProjectID,
		projectGithubPerson.GithubPersonID,
		projectGithubPerson.SourceType,
		projectGithubPerson.IsDeleted,
		projectGithubPerson.CreatedAt,
		projectGithubPerson.UpdatedAt,
	)

	return err
}

func (r *ProjectGithubPersonRepository) Delete(id string) error {
	query := `DELETE FROM project_github_people WHERE id = ?`
	_, err := r.db.Exec(query, id)
	return err
}

func (r *ProjectGithubPersonRepository) DeleteByProjectAndGithubPerson(projectID, githubPersonID string) error {
	query := `DELETE FROM project_github_people WHERE project_id = ? AND github_person_id = ?`
	_, err := r.db.Exec(query, projectID, githubPersonID)
	return err
}

// SoftDelete marks a project github person as deleted
func (r *ProjectGithubPersonRepository) SoftDelete(id string) error {
	query := `UPDATE project_github_people SET is_deleted = 1, updated_at = ? WHERE id = ?`
	_, err := r.db.Exec(query, time.Now(), id)
	return err
}

// SoftDeleteByProjectAndGithubPerson marks a project github person as deleted by project and github person IDs
func (r *ProjectGithubPersonRepository) SoftDeleteByProjectAndGithubPerson(projectID, githubPersonID string) error {
	query := `UPDATE project_github_people SET is_deleted = 1, updated_at = ? WHERE project_id = ? AND github_person_id = ?`
	_, err := r.db.Exec(query, time.Now(), projectID, githubPersonID)
	return err
}

// Restore restores a soft-deleted project github person
func (r *ProjectGithubPersonRepository) Restore(id string) error {
	query := `UPDATE project_github_people SET is_deleted = 0, updated_at = ? WHERE id = ?`
	_, err := r.db.Exec(query, time.Now(), id)
	return err
}

// RestoreByProjectAndGithubPerson restores a soft-deleted project github person by project and github person IDs
func (r *ProjectGithubPersonRepository) RestoreByProjectAndGithubPerson(projectID, githubPersonID string) error {
	query := `UPDATE project_github_people SET is_deleted = 0, updated_at = ? WHERE project_id = ? AND github_person_id = ?`
	_, err := r.db.Exec(query, time.Now(), projectID, githubPersonID)
	return err
}
