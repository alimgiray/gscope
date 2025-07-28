package repositories

import (
	"database/sql"
	"time"

	"github.com/alimgiray/gscope/internal/models"
	"github.com/google/uuid"
)

type GithubPersonRepository struct {
	db *sql.DB
}

func NewGithubPersonRepository(db *sql.DB) *GithubPersonRepository {
	return &GithubPersonRepository{db: db}
}

func (r *GithubPersonRepository) Create(person *models.GithubPerson) error {
	person.ID = uuid.New().String()

	query := `
		INSERT INTO github_people (
			id, github_user_id, username, display_name, avatar_url, profile_url, type
		) VALUES (?, ?, ?, ?, ?, ?, ?)
	`

	_, err := r.db.Exec(query,
		person.ID, person.GithubUserID, person.Username, person.DisplayName, person.AvatarURL,
		person.ProfileURL, person.Type,
	)

	return err
}

func (r *GithubPersonRepository) GetByID(id string) (*models.GithubPerson, error) {
	query := `SELECT * FROM github_people WHERE id = ?`

	var person models.GithubPerson
	err := r.db.QueryRow(query, id).Scan(
		&person.ID, &person.GithubUserID, &person.Username, &person.DisplayName, &person.AvatarURL,
		&person.ProfileURL, &person.Type, &person.CreatedAt, &person.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &person, nil
}

func (r *GithubPersonRepository) GetByGithubUserID(githubUserID int) (*models.GithubPerson, error) {
	query := `SELECT * FROM github_people WHERE github_user_id = ?`

	var person models.GithubPerson
	err := r.db.QueryRow(query, githubUserID).Scan(
		&person.ID, &person.GithubUserID, &person.Username, &person.DisplayName, &person.AvatarURL,
		&person.ProfileURL, &person.Type, &person.CreatedAt, &person.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &person, nil
}

func (r *GithubPersonRepository) GetByUsername(username string) (*models.GithubPerson, error) {
	query := `SELECT * FROM github_people WHERE username = ?`

	var person models.GithubPerson
	err := r.db.QueryRow(query, username).Scan(
		&person.ID, &person.GithubUserID, &person.Username, &person.DisplayName, &person.AvatarURL,
		&person.ProfileURL, &person.Type, &person.CreatedAt, &person.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &person, nil
}

func (r *GithubPersonRepository) Update(person *models.GithubPerson) error {
	person.UpdatedAt = time.Now()

	query := `
		UPDATE github_people SET 
			github_user_id = ?, username = ?, display_name = ?, avatar_url = ?,
			profile_url = ?, type = ?
		WHERE id = ?
	`

	_, err := r.db.Exec(query,
		person.GithubUserID, person.Username, person.DisplayName, person.AvatarURL,
		person.ProfileURL, person.Type, person.ID,
	)

	return err
}

func (r *GithubPersonRepository) Delete(id string) error {
	query := `DELETE FROM github_people WHERE id = ?`
	_, err := r.db.Exec(query, id)
	return err
}

func (r *GithubPersonRepository) Upsert(person *models.GithubPerson) error {
	existing, err := r.GetByGithubUserID(person.GithubUserID)
	if err != nil && err != sql.ErrNoRows {
		return err
	}

	if existing != nil {
		person.ID = existing.ID
		person.CreatedAt = existing.CreatedAt
		return r.Update(person)
	}

	return r.Create(person)
}

// GetByProjectID retrieves all GitHub people for a project
func (r *GithubPersonRepository) GetByProjectID(projectID string) ([]*models.GithubPerson, error) {
	query := `
		SELECT DISTINCT gp.id, gp.github_user_id, gp.username, gp.display_name, 
		       gp.avatar_url, gp.profile_url, gp.type, gp.created_at, gp.updated_at
		FROM github_people gp
		INNER JOIN project_github_people pgp ON gp.id = pgp.github_person_id
		WHERE pgp.project_id = ?
		ORDER BY gp.username ASC
	`

	rows, err := r.db.Query(query, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var people []*models.GithubPerson
	for rows.Next() {
		person := &models.GithubPerson{}
		err := rows.Scan(
			&person.ID,
			&person.GithubUserID,
			&person.Username,
			&person.DisplayName,
			&person.AvatarURL,
			&person.ProfileURL,
			&person.Type,
			&person.CreatedAt,
			&person.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		people = append(people, person)
	}

	return people, nil
}
