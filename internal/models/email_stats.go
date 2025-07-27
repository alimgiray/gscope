package models

import "time"

// EmailStats represents email statistics from commits
type EmailStats struct {
	Email        string     `json:"email"`
	Name         *string    `json:"name"`
	FirstCommit  *time.Time `json:"first_commit"`
	LastCommit   *time.Time `json:"last_commit"`
	IsMerged     bool       `json:"is_merged"`     // Whether this email has been merged into another
	MergedEmails []string   `json:"merged_emails"` // List of emails that have been merged into this one
}
