package repositories

import (
	"database/sql"

	"github.com/alimgiray/gscope/internal/models"
)

type CommitRepository struct {
	db *sql.DB
}

func NewCommitRepository(db *sql.DB) *CommitRepository {
	return &CommitRepository{db: db}
}

// Create creates a new commit
func (r *CommitRepository) Create(commit *models.Commit) error {
	query := `
		INSERT INTO commits (
			id, github_repository_id, commit_sha, message, author_name, author_email,
			commit_date, is_merge_commit, merge_commit_sha, additions, deletions, changes
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := r.db.Exec(query,
		commit.ID, commit.GithubRepositoryID, commit.CommitSHA, commit.Message,
		commit.AuthorName, commit.AuthorEmail, commit.CommitDate, commit.IsMergeCommit,
		commit.MergeCommitSHA, commit.Additions, commit.Deletions, commit.Changes,
	)

	return err
}

// GetByID retrieves a commit by ID
func (r *CommitRepository) GetByID(id string) (*models.Commit, error) {
	query := `
		SELECT id, github_repository_id, commit_sha, message, author_name, author_email,
			   commit_date, is_merge_commit, merge_commit_sha, additions, deletions, changes, created_at
		FROM commits WHERE id = ?
	`

	commit := &models.Commit{}
	err := r.db.QueryRow(query, id).Scan(
		&commit.ID, &commit.GithubRepositoryID, &commit.CommitSHA, &commit.Message,
		&commit.AuthorName, &commit.AuthorEmail, &commit.CommitDate, &commit.IsMergeCommit,
		&commit.MergeCommitSHA, &commit.Additions, &commit.Deletions, &commit.Changes, &commit.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	return commit, nil
}

// GetByCommitSHA retrieves a commit by its SHA
func (r *CommitRepository) GetByCommitSHA(commitSHA string) (*models.Commit, error) {
	query := `
		SELECT id, github_repository_id, commit_sha, message, author_name, author_email,
			   commit_date, is_merge_commit, merge_commit_sha, additions, deletions, changes, created_at
		FROM commits WHERE commit_sha = ?
	`

	commit := &models.Commit{}
	err := r.db.QueryRow(query, commitSHA).Scan(
		&commit.ID, &commit.GithubRepositoryID, &commit.CommitSHA, &commit.Message,
		&commit.AuthorName, &commit.AuthorEmail, &commit.CommitDate, &commit.IsMergeCommit,
		&commit.MergeCommitSHA, &commit.Additions, &commit.Deletions, &commit.Changes, &commit.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	return commit, nil
}

// GetByRepositoryID retrieves all commits for a repository
func (r *CommitRepository) GetByRepositoryID(repositoryID string) ([]*models.Commit, error) {
	query := `
		SELECT id, github_repository_id, commit_sha, message, author_name, author_email,
			   commit_date, is_merge_commit, merge_commit_sha, additions, deletions, changes, created_at
		FROM commits WHERE github_repository_id = ?
		ORDER BY commit_date DESC
	`

	rows, err := r.db.Query(query, repositoryID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var commits []*models.Commit
	for rows.Next() {
		commit := &models.Commit{}
		err := rows.Scan(
			&commit.ID, &commit.GithubRepositoryID, &commit.CommitSHA, &commit.Message,
			&commit.AuthorName, &commit.AuthorEmail, &commit.CommitDate, &commit.IsMergeCommit,
			&commit.MergeCommitSHA, &commit.Additions, &commit.Deletions, &commit.Changes, &commit.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		commits = append(commits, commit)
	}

	return commits, nil
}

// Update updates an existing commit
func (r *CommitRepository) Update(commit *models.Commit) error {
	query := `
		UPDATE commits SET
			message = ?, author_name = ?, author_email = ?, commit_date = ?,
			is_merge_commit = ?, merge_commit_sha = ?, additions = ?, deletions = ?, changes = ?
		WHERE id = ?
	`

	_, err := r.db.Exec(query,
		commit.Message, commit.AuthorName, commit.AuthorEmail, commit.CommitDate,
		commit.IsMergeCommit, commit.MergeCommitSHA, commit.Additions, commit.Deletions, commit.Changes,
		commit.ID,
	)

	return err
}

// Delete deletes a commit by ID
func (r *CommitRepository) Delete(id string) error {
	query := `DELETE FROM commits WHERE id = ?`
	_, err := r.db.Exec(query, id)
	return err
}

// DeleteByRepositoryID deletes all commits for a repository
func (r *CommitRepository) DeleteByRepositoryID(repositoryID string) error {
	query := `DELETE FROM commits WHERE github_repository_id = ?`
	_, err := r.db.Exec(query, repositoryID)
	return err
}

// ExistsByCommitSHA checks if a commit exists by its SHA
func (r *CommitRepository) ExistsByCommitSHA(commitSHA string) (bool, error) {
	query := `SELECT COUNT(*) FROM commits WHERE commit_sha = ?`
	var count int
	err := r.db.QueryRow(query, commitSHA).Scan(&count)
	return count > 0, err
}
