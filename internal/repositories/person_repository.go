package repositories

import (
	"database/sql"
	"strings"

	"github.com/alimgiray/gscope/internal/models"
)

type PersonRepository struct {
	db *sql.DB
}

func NewPersonRepository(db *sql.DB) *PersonRepository {
	return &PersonRepository{db: db}
}

// Create creates a new person
func (r *PersonRepository) Create(person *models.Person) error {
	query := `
		INSERT INTO people (
			id, name, primary_email
		) VALUES (?, ?, ?)
	`

	_, err := r.db.Exec(query, person.ID, person.Name, person.PrimaryEmail)
	return err
}

// GetByID retrieves a person by ID
func (r *PersonRepository) GetByID(id string) (*models.Person, error) {
	query := `
		SELECT id, name, primary_email, created_at, updated_at
		FROM people WHERE id = ?
	`

	person := &models.Person{}
	err := r.db.QueryRow(query, id).Scan(
		&person.ID, &person.Name, &person.PrimaryEmail, &person.CreatedAt, &person.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return person, nil
}

// GetByEmail retrieves a person by email
func (r *PersonRepository) GetByEmail(email string) (*models.Person, error) {
	query := `
		SELECT id, name, primary_email, created_at, updated_at
		FROM people WHERE primary_email = ?
	`

	person := &models.Person{}
	err := r.db.QueryRow(query, email).Scan(
		&person.ID, &person.Name, &person.PrimaryEmail, &person.CreatedAt, &person.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return person, nil
}

// Update updates an existing person
func (r *PersonRepository) Update(person *models.Person) error {
	query := `
		UPDATE people SET
			name = ?, primary_email = ?
		WHERE id = ?
	`

	_, err := r.db.Exec(query, person.Name, person.PrimaryEmail, person.ID)
	return err
}

// Delete deletes a person by ID
func (r *PersonRepository) Delete(id string) error {
	query := `DELETE FROM people WHERE id = ?`
	_, err := r.db.Exec(query, id)
	return err
}

// ExistsByEmail checks if a person exists by email
func (r *PersonRepository) ExistsByEmail(email string) (bool, error) {
	query := `SELECT COUNT(*) FROM people WHERE primary_email = ?`
	var count int
	err := r.db.QueryRow(query, email).Scan(&count)
	return count > 0, err
}

// GetOrCreateByEmail gets a person by email or creates a new one
func (r *PersonRepository) GetOrCreateByEmail(name, email string) (*models.Person, error) {
	// Try to get existing person first
	person, err := r.GetByEmail(email)
	if err == nil {
		// Person exists, return it (don't create a new one)
		return person, nil
	}

	// If not found, create new person
	if err == sql.ErrNoRows {
		person = models.NewPerson(name, email)
		if err := r.Create(person); err != nil {
			// If creation fails due to unique constraint violation (race condition),
			// try to get the person again
			if strings.Contains(err.Error(), "UNIQUE constraint failed") {
				person, err = r.GetByEmail(email)
				if err == nil {
					return person, nil
				}
			}
			return nil, err
		}
		return person, nil
	}

	// Some other error occurred
	return nil, err
}

// GetOrCreateByEmailWithNameUpdate gets a person by email, creates if not exists, or updates name if different
func (r *PersonRepository) GetOrCreateByEmailWithNameUpdate(name, email string) (*models.Person, error) {
	person, err := r.GetOrCreateByEmail(name, email)
	if err != nil {
		return nil, err
	}

	// Update person's name if it's different (in case they used different names)
	if person.Name != name {
		person.Name = name
		if err := r.Update(person); err != nil {
			return nil, err
		}
	}

	return person, nil
}

// GetPeopleByRepositoryID gets all people who have committed to a specific repository
func (r *PersonRepository) GetPeopleByRepositoryID(repositoryID string) ([]*models.Person, error) {
	query := `
		SELECT DISTINCT p.id, p.name, p.primary_email, p.created_at, p.updated_at
		FROM people p
		INNER JOIN commits c ON p.primary_email = c.author_email
		WHERE c.github_repository_id = ?
		ORDER BY p.name
	`

	rows, err := r.db.Query(query, repositoryID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var people []*models.Person
	for rows.Next() {
		person := &models.Person{}
		err := rows.Scan(
			&person.ID, &person.Name, &person.PrimaryEmail, &person.CreatedAt, &person.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		people = append(people, person)
	}

	return people, nil
}
