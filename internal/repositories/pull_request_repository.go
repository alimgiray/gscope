package repositories

import (
	"database/sql"
	"sync"
	"time"

	"github.com/alimgiray/gscope/internal/models"
	"github.com/google/uuid"
)

type PullRequestRepository struct {
	db *sql.DB
	mu sync.RWMutex
}

func NewPullRequestRepository(db *sql.DB) *PullRequestRepository {
	return &PullRequestRepository{db: db}
}

func (r *PullRequestRepository) Create(pr *models.PullRequest) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	pr.ID = uuid.New().String()

	query := `
		INSERT INTO pull_requests (
			id, repository_id, github_pr_number, github_pr_id, title, body, 
			state, merged_at, merge_commit_sha, closed_at, user, 
			requested_reviewers, requested_teams, draft, github_created_at, 
			github_updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := r.db.Exec(query,
		pr.ID, pr.RepositoryID, pr.GithubPRNumber, pr.GithubPRID, pr.Title, pr.Body,
		pr.State, pr.MergedAt, pr.MergeCommitSHA, pr.ClosedAt, pr.User,
		pr.RequestedReviewers, pr.RequestedTeams, pr.Draft, pr.GithubCreatedAt,
		pr.GithubUpdatedAt,
	)

	return err
}

func (r *PullRequestRepository) GetByID(id string) (*models.PullRequest, error) {
	query := `SELECT * FROM pull_requests WHERE id = ?`

	var pr models.PullRequest
	err := r.db.QueryRow(query, id).Scan(
		&pr.ID, &pr.RepositoryID, &pr.GithubPRNumber, &pr.GithubPRID, &pr.Title, &pr.Body,
		&pr.State, &pr.MergedAt, &pr.MergeCommitSHA, &pr.ClosedAt, &pr.User,
		&pr.RequestedReviewers, &pr.RequestedTeams, &pr.Draft, &pr.GithubCreatedAt,
		&pr.GithubUpdatedAt, &pr.CreatedAt, &pr.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &pr, nil
}

func (r *PullRequestRepository) GetByRepositoryID(repositoryID string) ([]*models.PullRequest, error) {
	query := `SELECT * FROM pull_requests WHERE repository_id = ? ORDER BY github_pr_number DESC`

	rows, err := r.db.Query(query, repositoryID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var pullRequests []*models.PullRequest
	for rows.Next() {
		var pr models.PullRequest
		err := rows.Scan(
			&pr.ID, &pr.RepositoryID, &pr.GithubPRNumber, &pr.GithubPRID, &pr.Title, &pr.Body,
			&pr.State, &pr.MergedAt, &pr.MergeCommitSHA, &pr.ClosedAt, &pr.User,
			&pr.RequestedReviewers, &pr.RequestedTeams, &pr.Draft, &pr.GithubCreatedAt,
			&pr.GithubUpdatedAt, &pr.CreatedAt, &pr.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		pullRequests = append(pullRequests, &pr)
	}

	return pullRequests, nil
}

func (r *PullRequestRepository) GetByGithubPRID(githubPRID int) (*models.PullRequest, error) {
	query := `SELECT * FROM pull_requests WHERE github_pr_id = ?`

	var pr models.PullRequest
	err := r.db.QueryRow(query, githubPRID).Scan(
		&pr.ID, &pr.RepositoryID, &pr.GithubPRNumber, &pr.GithubPRID, &pr.Title, &pr.Body,
		&pr.State, &pr.MergedAt, &pr.MergeCommitSHA, &pr.ClosedAt, &pr.User,
		&pr.RequestedReviewers, &pr.RequestedTeams, &pr.Draft, &pr.GithubCreatedAt,
		&pr.GithubUpdatedAt, &pr.CreatedAt, &pr.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &pr, nil
}

func (r *PullRequestRepository) Update(pr *models.PullRequest) error {
	query := `
		UPDATE pull_requests SET 
			repository_id = ?, github_pr_number = ?, github_pr_id = ?, title = ?, body = ?,
			state = ?, merged_at = ?, merge_commit_sha = ?, closed_at = ?, user = ?,
			requested_reviewers = ?, requested_teams = ?, draft = ?, github_created_at = ?,
			github_updated_at = ?
		WHERE id = ?
	`

	_, err := r.db.Exec(query,
		pr.RepositoryID, pr.GithubPRNumber, pr.GithubPRID, pr.Title, pr.Body,
		pr.State, pr.MergedAt, pr.MergeCommitSHA, pr.ClosedAt, pr.User,
		pr.RequestedReviewers, pr.RequestedTeams, pr.Draft, pr.GithubCreatedAt,
		pr.GithubUpdatedAt, pr.ID,
	)

	return err
}

func (r *PullRequestRepository) Delete(id string) error {
	query := `DELETE FROM pull_requests WHERE id = ?`
	_, err := r.db.Exec(query, id)
	return err
}

func (r *PullRequestRepository) Upsert(pr *models.PullRequest) error {
	existing, err := r.GetByGithubPRID(pr.GithubPRID)
	if err != nil && err != sql.ErrNoRows {
		return err
	}

	if existing != nil {
		pr.ID = existing.ID
		pr.CreatedAt = existing.CreatedAt
		return r.Update(pr)
	}

	return r.Create(pr)
}

// GetEarliestOpenPRDateByRepositoryID gets the earliest open PR date for a repository
func (r *PullRequestRepository) GetEarliestOpenPRDateByRepositoryID(repositoryID string) (time.Time, error) {
	query := `SELECT MIN(github_created_at) FROM pull_requests WHERE repository_id = ? AND state = 'open'`

	var earliestDateStr sql.NullString
	err := r.db.QueryRow(query, repositoryID).Scan(&earliestDateStr)
	if err != nil {
		return time.Time{}, err
	}

	// If no open PRs found, return zero time
	if !earliestDateStr.Valid {
		return time.Time{}, nil
	}

	// Parse the date string
	earliestDate, err := time.Parse("2006-01-02 15:04:05-07:00", earliestDateStr.String)
	if err != nil {
		return time.Time{}, err
	}

	return earliestDate, nil
}

// GetLatestPRDateByRepositoryID gets the latest PR date for a repository (any PR, not just open ones)
func (r *PullRequestRepository) GetLatestPRDateByRepositoryID(repositoryID string) (time.Time, error) {
	query := `SELECT MAX(github_created_at) FROM pull_requests WHERE repository_id = ?`

	var latestDateStr sql.NullString
	err := r.db.QueryRow(query, repositoryID).Scan(&latestDateStr)
	if err != nil {
		return time.Time{}, err
	}

	// If no PRs found, return zero time
	if !latestDateStr.Valid {
		return time.Time{}, nil
	}

	// Parse the date string
	latestDate, err := time.Parse("2006-01-02 15:04:05-07:00", latestDateStr.String)
	if err != nil {
		return time.Time{}, err
	}

	return latestDate, nil
}

// GetOpenPRNumbersByRepositoryID gets the PR numbers of all open PRs for a repository
func (r *PullRequestRepository) GetOpenPRNumbersByRepositoryID(repositoryID string) ([]int, error) {
	query := `SELECT github_pr_number FROM pull_requests WHERE repository_id = ? AND state = 'open' ORDER BY github_pr_number`

	rows, err := r.db.Query(query, repositoryID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var prNumbers []int
	for rows.Next() {
		var prNumber int
		err := rows.Scan(&prNumber)
		if err != nil {
			return nil, err
		}
		prNumbers = append(prNumbers, prNumber)
	}

	return prNumbers, nil
}
