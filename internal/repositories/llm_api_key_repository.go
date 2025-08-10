package repositories

import (
	"database/sql"
	"fmt"

	"github.com/alimgiray/gscope/internal/models"
	"github.com/google/uuid"
)

type LLMAPIKeyRepository struct {
	db *sql.DB
}

func NewLLMAPIKeyRepository(db *sql.DB) *LLMAPIKeyRepository {
	return &LLMAPIKeyRepository{db: db}
}

// Create creates a new LLM API key
func (r *LLMAPIKeyRepository) Create(key *models.LLMAPIKey) error {
	// First, permanently delete any existing inactive records for this project/provider
	deleteQuery := `
		DELETE FROM llm_api_keys 
		WHERE project_id = ? AND provider = ? AND is_active = 0
	`
	_, err := r.db.Exec(deleteQuery, key.ProjectID, key.Provider)
	if err != nil {
		return err
	}

	// Now create the new record
	query := `
		INSERT INTO llm_api_keys (id, project_id, provider, api_key, is_active, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`
	_, err = r.db.Exec(query, key.ID, key.ProjectID, key.Provider, key.APIKey, key.IsActive, key.CreatedAt, key.UpdatedAt)
	return err
}

// GetByProjectID gets the LLM API key for a specific project
func (r *LLMAPIKeyRepository) GetByProjectID(projectID uuid.UUID) (*models.LLMAPIKey, error) {
	query := `
		SELECT id, project_id, provider, api_key, is_active, created_at, updated_at
		FROM llm_api_keys
		WHERE project_id = ? AND is_active = 1
		ORDER BY created_at DESC
		LIMIT 1
	`

	key := &models.LLMAPIKey{}
	err := r.db.QueryRow(query, projectID).Scan(
		&key.ID,
		&key.ProjectID,
		&key.Provider,
		&key.APIKey,
		&key.IsActive,
		&key.CreatedAt,
		&key.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	return key, err
}

// GetByID gets an LLM API key by ID
func (r *LLMAPIKeyRepository) GetByID(id uuid.UUID) (*models.LLMAPIKey, error) {
	query := `
		SELECT id, project_id, provider, api_key, is_active, created_at, updated_at
		FROM llm_api_keys
		WHERE id = ?
	`

	key := &models.LLMAPIKey{}
	err := r.db.QueryRow(query, id).Scan(
		&key.ID,
		&key.ProjectID,
		&key.Provider,
		&key.APIKey,
		&key.IsActive,
		&key.CreatedAt,
		&key.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	return key, err
}

// Update updates an existing LLM API key
func (r *LLMAPIKeyRepository) Update(key *models.LLMAPIKey) error {
	query := `
		UPDATE llm_api_keys
		SET provider = ?, api_key = ?, is_active = ?, updated_at = ?
		WHERE id = ?
	`
	result, err := r.db.Exec(query, key.Provider, key.APIKey, key.IsActive, key.UpdatedAt, key.ID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("no rows affected")
	}

	return nil
}

// Delete soft deletes an LLM API key by setting is_active to false
func (r *LLMAPIKeyRepository) Delete(id uuid.UUID) error {
	query := `
		UPDATE llm_api_keys
		SET is_active = 0, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
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
		return fmt.Errorf("no rows affected")
	}

	return nil
}

// DeleteByProjectID soft deletes all LLM API keys for a project
func (r *LLMAPIKeyRepository) DeleteByProjectID(projectID uuid.UUID) error {
	query := `
		UPDATE llm_api_keys
		SET is_active = 0, updated_at = CURRENT_TIMESTAMP
		WHERE project_id = ?
	`
	_, err := r.db.Exec(query, projectID)
	return err
}

// HasActiveKey checks if a project has an active LLM API key
func (r *LLMAPIKeyRepository) HasActiveKey(projectID uuid.UUID) (bool, error) {
	query := `
		SELECT COUNT(*)
		FROM llm_api_keys
		WHERE project_id = ? AND is_active = 1
	`

	var count int
	err := r.db.QueryRow(query, projectID).Scan(&count)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}
