package models

import (
	"time"

	"github.com/google/uuid"
)

type ScoreSettings struct {
	ID           string    `json:"id"`
	ProjectID    string    `json:"project_id"`
	Additions    int       `json:"additions"`
	Deletions    int       `json:"deletions"`
	Commits      int       `json:"commits"`
	PullRequests int       `json:"pull_requests"`
	Comments     int       `json:"comments"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func NewScoreSettings(projectID string) *ScoreSettings {
	return &ScoreSettings{
		ID:           uuid.New().String(),
		ProjectID:    projectID,
		Additions:    1,
		Deletions:    3,
		Commits:      10,
		PullRequests: 20,
		Comments:     100,
	}
}
