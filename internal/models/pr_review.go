package models

import (
	"time"
)

// PRReview represents a GitHub pull request review
type PRReview struct {
	ID                string     `json:"id" db:"id"`
	RepositoryID      string     `json:"repository_id" db:"repository_id"`
	PullRequestID     string     `json:"pull_request_id" db:"pull_request_id"`
	GithubReviewID    int        `json:"github_review_id" db:"github_review_id"`
	ReviewerID        int        `json:"reviewer_id" db:"reviewer_id"`
	ReviewerLogin     string     `json:"reviewer_login" db:"reviewer_login"`
	Body              *string    `json:"body" db:"body"`
	State             string     `json:"state" db:"state"`
	AuthorAssociation *string    `json:"author_association" db:"author_association"`
	SubmittedAt       *time.Time `json:"submitted_at" db:"submitted_at"`
	CommitID          string     `json:"commit_id" db:"commit_id"`
	HTMLURL           *string    `json:"html_url" db:"html_url"`
	GithubCreatedAt   *time.Time `json:"github_created_at" db:"github_created_at"`
	GithubUpdatedAt   *time.Time `json:"github_updated_at" db:"github_updated_at"`
	CreatedAt         time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at" db:"updated_at"`
}
