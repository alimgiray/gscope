package models

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// PeopleStatistics represents daily statistics for a GitHub person in a repository
type PeopleStatistics struct {
	ID             string    `json:"id" db:"id"`
	ProjectID      string    `json:"project_id" db:"project_id"`
	RepositoryID   string    `json:"repository_id" db:"repository_id"`
	GithubPersonID string    `json:"github_person_id" db:"github_person_id"`
	StatDate       time.Time `json:"stat_date" db:"stat_date"`
	Commits        int       `json:"commits" db:"commits"`
	Additions      int       `json:"additions" db:"additions"`
	Deletions      int       `json:"deletions" db:"deletions"`
	Comments       int       `json:"comments" db:"comments"`
	PullRequests   int       `json:"pull_requests" db:"pull_requests"`
	Score          int       `json:"score" db:"score"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time `json:"updated_at" db:"updated_at"`
}

// NewPeopleStatistics creates a new PeopleStatistics with a generated UUID
func NewPeopleStatistics(projectID, repositoryID, githubPersonID string, statDate time.Time) *PeopleStatistics {
	return &PeopleStatistics{
		ID:             uuid.New().String(),
		ProjectID:      projectID,
		RepositoryID:   repositoryID,
		GithubPersonID: githubPersonID,
		StatDate:       statDate,
		Commits:        0,
		Additions:      0,
		Deletions:      0,
		Comments:       0,
		PullRequests:   0,
		Score:          0,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
}

// Validate validates the PeopleStatistics fields
func (ps *PeopleStatistics) Validate() error {
	if ps.ProjectID == "" {
		return errors.New("project ID is required")
	}
	if ps.RepositoryID == "" {
		return errors.New("repository ID is required")
	}
	if ps.GithubPersonID == "" {
		return errors.New("GitHub person ID is required")
	}
	if ps.StatDate.IsZero() {
		return errors.New("stat date is required")
	}
	if ps.Commits < 0 {
		return errors.New("commits cannot be negative")
	}
	if ps.Additions < 0 {
		return errors.New("additions cannot be negative")
	}
	if ps.Deletions < 0 {
		return errors.New("deletions cannot be negative")
	}
	if ps.Comments < 0 {
		return errors.New("comments cannot be negative")
	}
	if ps.PullRequests < 0 {
		return errors.New("pull requests cannot be negative")
	}
	return nil
}

// CalculateScore calculates the score based on the statistics and score settings
func (ps *PeopleStatistics) CalculateScore(scoreSettings *ScoreSettings) {
	score := 0
	score += ps.Commits * scoreSettings.Commits
	score += ps.Additions * scoreSettings.Additions
	score += ps.Deletions * scoreSettings.Deletions
	score += ps.Comments * scoreSettings.Comments
	score += ps.PullRequests * scoreSettings.PullRequests
	ps.Score = score
}
