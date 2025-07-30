package repositories

import (
	"database/sql"
	"sync"
	"time"

	"github.com/alimgiray/gscope/internal/models"
	"github.com/google/uuid"
)

type PRReviewRepository struct {
	db *sql.DB
	mu sync.RWMutex
}

func NewPRReviewRepository(db *sql.DB) *PRReviewRepository {
	return &PRReviewRepository{db: db}
}

func (r *PRReviewRepository) Create(review *models.PRReview) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	review.ID = uuid.New().String()

	query := `
		INSERT INTO pr_reviews (
			id, repository_id, pull_request_id, github_review_id, reviewer_id,
			reviewer_login, body, state, author_association, submitted_at, commit_id,
			html_url, github_created_at, github_updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := r.db.Exec(query,
		review.ID, review.RepositoryID, review.PullRequestID, review.GithubReviewID, review.ReviewerID,
		review.ReviewerLogin, review.Body, review.State, review.AuthorAssociation, review.SubmittedAt, review.CommitID,
		review.HTMLURL, review.GithubCreatedAt, review.GithubUpdatedAt,
	)

	return err
}

func (r *PRReviewRepository) GetByID(id string) (*models.PRReview, error) {
	query := `SELECT * FROM pr_reviews WHERE id = ?`

	var review models.PRReview
	err := r.db.QueryRow(query, id).Scan(
		&review.ID, &review.RepositoryID, &review.PullRequestID, &review.GithubReviewID, &review.ReviewerID,
		&review.ReviewerLogin, &review.Body, &review.State, &review.AuthorAssociation, &review.SubmittedAt, &review.CommitID,
		&review.HTMLURL, &review.GithubCreatedAt, &review.GithubUpdatedAt,
		&review.CreatedAt, &review.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &review, nil
}

func (r *PRReviewRepository) GetByPullRequestID(pullRequestID string) ([]*models.PRReview, error) {
	query := `SELECT * FROM pr_reviews WHERE pull_request_id = ? ORDER BY submitted_at DESC`

	rows, err := r.db.Query(query, pullRequestID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reviews []*models.PRReview
	for rows.Next() {
		var review models.PRReview
		err := rows.Scan(
			&review.ID, &review.RepositoryID, &review.PullRequestID, &review.GithubReviewID, &review.ReviewerID,
			&review.ReviewerLogin, &review.Body, &review.State, &review.AuthorAssociation, &review.SubmittedAt, &review.CommitID,
			&review.HTMLURL, &review.GithubCreatedAt, &review.GithubUpdatedAt,
			&review.CreatedAt, &review.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		reviews = append(reviews, &review)
	}

	return reviews, nil
}

func (r *PRReviewRepository) GetByGithubReviewID(githubReviewID int) (*models.PRReview, error) {
	query := `SELECT * FROM pr_reviews WHERE github_review_id = ?`

	var review models.PRReview
	err := r.db.QueryRow(query, githubReviewID).Scan(
		&review.ID, &review.RepositoryID, &review.PullRequestID, &review.GithubReviewID, &review.ReviewerID,
		&review.ReviewerLogin, &review.Body, &review.State, &review.AuthorAssociation, &review.SubmittedAt, &review.CommitID,
		&review.HTMLURL, &review.GithubCreatedAt, &review.GithubUpdatedAt,
		&review.CreatedAt, &review.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &review, nil
}

func (r *PRReviewRepository) GetByRepositoryID(repositoryID string) ([]*models.PRReview, error) {
	query := `SELECT * FROM pr_reviews WHERE repository_id = ? ORDER BY github_created_at DESC`

	rows, err := r.db.Query(query, repositoryID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reviews []*models.PRReview
	for rows.Next() {
		var review models.PRReview
		err := rows.Scan(
			&review.ID, &review.RepositoryID, &review.PullRequestID, &review.GithubReviewID, &review.ReviewerID,
			&review.ReviewerLogin, &review.Body, &review.State, &review.AuthorAssociation, &review.SubmittedAt, &review.CommitID,
			&review.HTMLURL, &review.GithubCreatedAt, &review.GithubUpdatedAt,
			&review.CreatedAt, &review.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		reviews = append(reviews, &review)
	}

	return reviews, nil
}

func (r *PRReviewRepository) Update(review *models.PRReview) error {
	query := `
		UPDATE pr_reviews SET 
			repository_id = ?, pull_request_id = ?, github_review_id = ?, reviewer_id = ?,
			reviewer_login = ?, body = ?, state = ?, author_association = ?, submitted_at = ?, commit_id = ?,
			html_url = ?, github_created_at = ?, github_updated_at = ?
		WHERE id = ?
	`

	_, err := r.db.Exec(query,
		review.RepositoryID, review.PullRequestID, review.GithubReviewID, review.ReviewerID,
		review.ReviewerLogin, review.Body, review.State, review.AuthorAssociation, review.SubmittedAt, review.CommitID,
		review.HTMLURL, review.GithubCreatedAt, review.GithubUpdatedAt,
		review.ID,
	)

	return err
}

// Delete function removed to prevent accidental deletion of PR reviews

func (r *PRReviewRepository) Upsert(review *models.PRReview) error {
	// Use a transaction to handle race conditions
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Try to get existing review within transaction
	query := `SELECT id, created_at FROM pr_reviews WHERE github_review_id = ?`
	var existingID string
	var existingCreatedAt time.Time
	err = tx.QueryRow(query, review.GithubReviewID).Scan(&existingID, &existingCreatedAt)

	if err == nil {
		// Review exists, update it
		review.ID = existingID
		review.CreatedAt = existingCreatedAt
		review.UpdatedAt = time.Now()

		updateQuery := `
			UPDATE pr_reviews SET 
				repository_id = ?, pull_request_id = ?, github_review_id = ?, reviewer_id = ?,
				reviewer_login = ?, body = ?, state = ?, author_association = ?, submitted_at = ?, commit_id = ?,
				html_url = ?, github_created_at = ?, github_updated_at = ?, updated_at = ?
			WHERE id = ?
		`

		_, err = tx.Exec(updateQuery,
			review.RepositoryID, review.PullRequestID, review.GithubReviewID, review.ReviewerID,
			review.ReviewerLogin, review.Body, review.State, review.AuthorAssociation, review.SubmittedAt, review.CommitID,
			review.HTMLURL, review.GithubCreatedAt, review.GithubUpdatedAt, review.UpdatedAt,
			review.ID,
		)
	} else if err == sql.ErrNoRows {
		// Review doesn't exist, create it
		review.ID = uuid.New().String()
		now := time.Now()
		review.CreatedAt = now
		review.UpdatedAt = now

		insertQuery := `
			INSERT INTO pr_reviews (
				id, repository_id, pull_request_id, github_review_id, reviewer_id,
				reviewer_login, body, state, author_association, submitted_at, commit_id,
				html_url, github_created_at, github_updated_at, created_at, updated_at
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		`

		_, err = tx.Exec(insertQuery,
			review.ID, review.RepositoryID, review.PullRequestID, review.GithubReviewID, review.ReviewerID,
			review.ReviewerLogin, review.Body, review.State, review.AuthorAssociation, review.SubmittedAt, review.CommitID,
			review.HTMLURL, review.GithubCreatedAt, review.GithubUpdatedAt, review.CreatedAt, review.UpdatedAt,
		)
	} else {
		// Some other error occurred
		return err
	}

	if err != nil {
		return err
	}

	return tx.Commit()
}
