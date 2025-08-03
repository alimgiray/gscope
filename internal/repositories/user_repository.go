package repositories

import (
	"database/sql"

	"github.com/alimgiray/gscope/internal/models"
	"github.com/google/uuid"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{
		db: db,
	}
}

// Create creates a new user
func (r *UserRepository) Create(user *models.User) error {
	query := `
		INSERT INTO users (id, name, username, profile_picture, access_token, github_access_token, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`

	_, err := r.db.Exec(query,
		user.ID.String(),
		user.Name,
		user.Username,
		user.ProfilePicture,
		user.AccessToken,
		user.GitHubAccessToken,
		user.CreatedAt,
	)
	return err
}

// GetByID retrieves a user by ID
func (r *UserRepository) GetByID(id string) (*models.User, error) {
	query := `SELECT id, name, username, profile_picture, access_token, github_access_token, created_at FROM users WHERE id = ?`

	var user models.User
	var userID string
	err := r.db.QueryRow(query, id).Scan(
		&userID,
		&user.Name,
		&user.Username,
		&user.ProfilePicture,
		&user.AccessToken,
		&user.GitHubAccessToken,
		&user.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	user.ID, err = uuid.Parse(userID)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

// GetByUsername retrieves a user by username
func (r *UserRepository) GetByUsername(username string) (*models.User, error) {
	query := `SELECT id, name, username, profile_picture, access_token, github_access_token, created_at FROM users WHERE username = ?`

	var user models.User
	var userID string
	err := r.db.QueryRow(query, username).Scan(
		&userID,
		&user.Name,
		&user.Username,
		&user.ProfilePicture,
		&user.AccessToken,
		&user.GitHubAccessToken,
		&user.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	user.ID, err = uuid.Parse(userID)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

// Update updates a user
func (r *UserRepository) Update(user *models.User) error {
	query := `
		UPDATE users 
		SET name = ?, username = ?, profile_picture = ?, access_token = ?, github_access_token = ?
		WHERE id = ?
	`

	_, err := r.db.Exec(query,
		user.Name,
		user.Username,
		user.ProfilePicture,
		user.AccessToken,
		user.GitHubAccessToken,
		user.ID.String(),
	)
	return err
}
