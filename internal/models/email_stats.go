package models

import "time"

// EmailStats represents email statistics from commits
type EmailStats struct {
	Email       string     `json:"email"`
	Name        *string    `json:"name"`
	CommitCount int        `json:"commit_count"`
	FirstCommit *time.Time `json:"first_commit"`
	LastCommit  *time.Time `json:"last_commit"`
}
