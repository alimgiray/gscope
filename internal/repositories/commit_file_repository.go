package repositories

import (
	"database/sql"

	"github.com/alimgiray/gscope/internal/models"
)

type CommitFileRepository struct {
	db *sql.DB
}

func NewCommitFileRepository(db *sql.DB) *CommitFileRepository {
	return &CommitFileRepository{db: db}
}

// Create creates a new commit file
func (r *CommitFileRepository) Create(commitFile *models.CommitFile) error {
	query := `
		INSERT INTO commit_files (
			id, commit_id, filename, status, additions, deletions, changes
		) VALUES (?, ?, ?, ?, ?, ?, ?)
	`

	_, err := r.db.Exec(query,
		commitFile.ID, commitFile.CommitID, commitFile.Filename, commitFile.Status,
		commitFile.Additions, commitFile.Deletions, commitFile.Changes,
	)

	return err
}

// GetByID retrieves a commit file by ID
func (r *CommitFileRepository) GetByID(id string) (*models.CommitFile, error) {
	query := `
		SELECT id, commit_id, filename, status, additions, deletions, changes, created_at
		FROM commit_files WHERE id = ?
	`

	commitFile := &models.CommitFile{}
	err := r.db.QueryRow(query, id).Scan(
		&commitFile.ID, &commitFile.CommitID, &commitFile.Filename, &commitFile.Status,
		&commitFile.Additions, &commitFile.Deletions, &commitFile.Changes, &commitFile.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	return commitFile, nil
}

// GetByCommitID retrieves all files for a commit
func (r *CommitFileRepository) GetByCommitID(commitID string) ([]*models.CommitFile, error) {
	query := `
		SELECT id, commit_id, filename, status, additions, deletions, changes, created_at
		FROM commit_files WHERE commit_id = ?
		ORDER BY filename
	`

	rows, err := r.db.Query(query, commitID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var commitFiles []*models.CommitFile
	for rows.Next() {
		commitFile := &models.CommitFile{}
		err := rows.Scan(
			&commitFile.ID, &commitFile.CommitID, &commitFile.Filename, &commitFile.Status,
			&commitFile.Additions, &commitFile.Deletions, &commitFile.Changes, &commitFile.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		commitFiles = append(commitFiles, commitFile)
	}

	return commitFiles, nil
}

// Update updates an existing commit file
func (r *CommitFileRepository) Update(commitFile *models.CommitFile) error {
	query := `
		UPDATE commit_files SET
			filename = ?, status = ?, additions = ?, deletions = ?, changes = ?
		WHERE id = ?
	`

	_, err := r.db.Exec(query,
		commitFile.Filename, commitFile.Status, commitFile.Additions, commitFile.Deletions, commitFile.Changes,
		commitFile.ID,
	)

	return err
}

// Delete deletes a commit file by ID
func (r *CommitFileRepository) Delete(id string) error {
	query := `DELETE FROM commit_files WHERE id = ?`
	_, err := r.db.Exec(query, id)
	return err
}

// DeleteByCommitID deletes all files for a commit
func (r *CommitFileRepository) DeleteByCommitID(commitID string) error {
	query := `DELETE FROM commit_files WHERE commit_id = ?`
	_, err := r.db.Exec(query, commitID)
	return err
}

// ExistsByCommitIDAndFilename checks if a file exists for a commit
func (r *CommitFileRepository) ExistsByCommitIDAndFilename(commitID, filename string) (bool, error) {
	query := `SELECT COUNT(*) FROM commit_files WHERE commit_id = ? AND filename = ?`
	var count int
	err := r.db.QueryRow(query, commitID, filename).Scan(&count)
	return count > 0, err
}
