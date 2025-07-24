package models

import (
	"time"
)

// PullRequest represents a GitHub pull request
type PullRequest struct {
	ID                 string     `json:"id" db:"id"`
	RepositoryID       string     `json:"repository_id" db:"repository_id"`
	GithubPRNumber     int        `json:"github_pr_number" db:"github_pr_number"`
	GithubPRID         int        `json:"github_pr_id" db:"github_pr_id"`
	Title              string     `json:"title" db:"title"`
	Body               *string    `json:"body" db:"body"`
	State              string     `json:"state" db:"state"`
	MergedAt           *time.Time `json:"merged_at" db:"merged_at"`
	MergeCommitSHA     *string    `json:"merge_commit_sha" db:"merge_commit_sha"`
	ClosedAt           *time.Time `json:"closed_at" db:"closed_at"`
	User               *string    `json:"user" db:"user"`                               // JSON object with user information
	RequestedReviewers *string    `json:"requested_reviewers" db:"requested_reviewers"` // JSON array of reviewer objects
	RequestedTeams     *string    `json:"requested_teams" db:"requested_teams"`         // JSON array of team objects
	Draft              bool       `json:"draft" db:"draft"`
	GithubCreatedAt    *time.Time `json:"github_created_at" db:"github_created_at"`
	GithubUpdatedAt    *time.Time `json:"github_updated_at" db:"github_updated_at"`
	CreatedAt          time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at" db:"updated_at"`
}
