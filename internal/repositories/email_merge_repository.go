package repositories

import (
	"database/sql"
	"time"

	"github.com/alimgiray/gscope/internal/models"
)

type EmailMergeRepository struct {
	db *sql.DB
}

func NewEmailMergeRepository(db *sql.DB) *EmailMergeRepository {
	return &EmailMergeRepository{db: db}
}

// Create creates a new email merge
func (r *EmailMergeRepository) Create(merge *models.EmailMerge) error {
	query := `
		INSERT INTO email_merges (id, project_id, source_email, target_email, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`

	_, err := r.db.Exec(query,
		merge.ID, merge.ProjectID, merge.SourceEmail, merge.TargetEmail,
		merge.CreatedAt, merge.UpdatedAt,
	)

	return err
}

// GetByID retrieves an email merge by ID
func (r *EmailMergeRepository) GetByID(id string) (*models.EmailMerge, error) {
	query := `
		SELECT id, project_id, source_email, target_email, created_at, updated_at
		FROM email_merges WHERE id = ?
	`

	merge := &models.EmailMerge{}
	err := r.db.QueryRow(query, id).Scan(
		&merge.ID, &merge.ProjectID, &merge.SourceEmail, &merge.TargetEmail,
		&merge.CreatedAt, &merge.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return merge, nil
}

// GetByProjectID retrieves all email merges for a project
func (r *EmailMergeRepository) GetByProjectID(projectID string) ([]*models.EmailMerge, error) {
	query := `
		SELECT id, project_id, source_email, target_email, created_at, updated_at
		FROM email_merges WHERE project_id = ?
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(query, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var merges []*models.EmailMerge
	for rows.Next() {
		merge := &models.EmailMerge{}
		err := rows.Scan(
			&merge.ID, &merge.ProjectID, &merge.SourceEmail, &merge.TargetEmail,
			&merge.CreatedAt, &merge.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		merges = append(merges, merge)
	}

	return merges, nil
}

// GetMergedEmailsForProject returns a map of source_email -> target_email for a project
func (r *EmailMergeRepository) GetMergedEmailsForProject(projectID string) (map[string]string, error) {
	query := `
		SELECT source_email, target_email
		FROM email_merges WHERE project_id = ?
	`

	rows, err := r.db.Query(query, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	mergedEmails := make(map[string]string)
	for rows.Next() {
		var sourceEmail, targetEmail string
		err := rows.Scan(&sourceEmail, &targetEmail)
		if err != nil {
			return nil, err
		}
		mergedEmails[sourceEmail] = targetEmail
	}

	return mergedEmails, nil
}

// IsEmailMerged checks if an email has been merged in a project
func (r *EmailMergeRepository) IsEmailMerged(projectID, email string) (bool, string, error) {
	query := `
		SELECT target_email FROM email_merges 
		WHERE project_id = ? AND source_email = ?
	`

	var targetEmail string
	err := r.db.QueryRow(query, projectID, email).Scan(&targetEmail)
	if err == sql.ErrNoRows {
		return false, "", nil
	}
	if err != nil {
		return false, "", err
	}

	return true, targetEmail, nil
}

// Update updates an existing email merge
func (r *EmailMergeRepository) Update(merge *models.EmailMerge) error {
	merge.UpdatedAt = time.Now()

	query := `
		UPDATE email_merges SET 
			project_id = ?, source_email = ?, target_email = ?, updated_at = ?
		WHERE id = ?
	`

	_, err := r.db.Exec(query,
		merge.ProjectID, merge.SourceEmail, merge.TargetEmail, merge.UpdatedAt, merge.ID,
	)

	return err
}

// Delete deletes an email merge by ID
func (r *EmailMergeRepository) Delete(id string) error {
	query := `DELETE FROM email_merges WHERE id = ?`
	_, err := r.db.Exec(query, id)
	return err
}

// DeleteByProjectID deletes all email merges for a project
func (r *EmailMergeRepository) DeleteByProjectID(projectID string) error {
	query := `DELETE FROM email_merges WHERE project_id = ?`
	_, err := r.db.Exec(query, projectID)
	return err
}

// DeleteByTargetEmail deletes all email merges where target_email matches
func (r *EmailMergeRepository) DeleteByTargetEmail(projectID, targetEmail string) error {
	query := `DELETE FROM email_merges WHERE project_id = ? AND target_email = ?`
	_, err := r.db.Exec(query, projectID, targetEmail)
	return err
}

// DeleteBySourceAndTarget deletes a specific email merge
func (r *EmailMergeRepository) DeleteBySourceAndTarget(projectID, sourceEmail, targetEmail string) error {
	query := `DELETE FROM email_merges WHERE project_id = ? AND source_email = ? AND target_email = ?`
	_, err := r.db.Exec(query, projectID, sourceEmail, targetEmail)
	return err
}
