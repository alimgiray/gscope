package repositories

import (
	"database/sql"
	"sync"

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

func (r *PRReviewRepository) Delete(id string) error {
	query := `DELETE FROM pr_reviews WHERE id = ?`
	_, err := r.db.Exec(query, id)
	return err
}

func (r *PRReviewRepository) Upsert(review *models.PRReview) error {
	existing, err := r.GetByGithubReviewID(review.GithubReviewID)
	if err != nil && err != sql.ErrNoRows {
		return err
	}

	if existing != nil {
		review.ID = existing.ID
		review.CreatedAt = existing.CreatedAt
		return r.Update(review)
	}

	return r.Create(review)
}
