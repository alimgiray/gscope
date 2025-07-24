package models

import (
	"time"
)

// GithubPerson represents a GitHub user, bot, or organization
type GithubPerson struct {
	ID           string    `json:"id" db:"id"`
	GithubUserID int       `json:"github_user_id" db:"github_user_id"`
	Username     string    `json:"username" db:"username"`
	DisplayName  *string   `json:"display_name" db:"display_name"`
	AvatarURL    *string   `json:"avatar_url" db:"avatar_url"`
	ProfileURL   *string   `json:"profile_url" db:"profile_url"`
	Type         *string   `json:"type" db:"type"` // "User", "Bot", "Organization"
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}
