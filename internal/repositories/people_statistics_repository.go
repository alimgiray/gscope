package repositories

import (
	"database/sql"
	"fmt"
	"sync"
	"time"

	"github.com/alimgiray/gscope/internal/models"
)

type PeopleStatisticsRepository struct {
	db *sql.DB
	mu sync.RWMutex
}

func NewPeopleStatisticsRepository(db *sql.DB) *PeopleStatisticsRepository {
	return &PeopleStatisticsRepository{db: db}
}

// Create creates a new people statistics record
func (r *PeopleStatisticsRepository) Create(stats *models.PeopleStatistics) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	query := `
		INSERT INTO people_statistics (
			id, project_id, repository_id, github_person_id, stat_date,
			commits, additions, deletions, comments, pull_requests, score,
			created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := r.db.Exec(query,
		stats.ID, stats.ProjectID, stats.RepositoryID, stats.GithubPersonID, stats.StatDate,
		stats.Commits, stats.Additions, stats.Deletions, stats.Comments, stats.PullRequests, stats.Score,
		stats.CreatedAt, stats.UpdatedAt,
	)

	return err
}

// GetByID retrieves people statistics by ID
func (r *PeopleStatisticsRepository) GetByID(id string) (*models.PeopleStatistics, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	query := `
		SELECT id, project_id, repository_id, github_person_id, stat_date,
		       commits, additions, deletions, comments, pull_requests, score,
		       created_at, updated_at
		FROM people_statistics WHERE id = ?
	`

	var stats models.PeopleStatistics
	err := r.db.QueryRow(query, id).Scan(
		&stats.ID, &stats.ProjectID, &stats.RepositoryID, &stats.GithubPersonID, &stats.StatDate,
		&stats.Commits, &stats.Additions, &stats.Deletions, &stats.Comments, &stats.PullRequests, &stats.Score,
		&stats.CreatedAt, &stats.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &stats, nil
}

// GetByProjectID retrieves all people statistics for a project
func (r *PeopleStatisticsRepository) GetByProjectID(projectID string) ([]*models.PeopleStatistics, error) {
	query := `
		SELECT id, project_id, repository_id, github_person_id, stat_date,
		       commits, additions, deletions, comments, pull_requests, score,
		       created_at, updated_at
		FROM people_statistics WHERE project_id = ?
		ORDER BY stat_date DESC, score DESC
	`

	rows, err := r.db.Query(query, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stats []*models.PeopleStatistics
	for rows.Next() {
		var stat models.PeopleStatistics
		err := rows.Scan(
			&stat.ID, &stat.ProjectID, &stat.RepositoryID, &stat.GithubPersonID, &stat.StatDate,
			&stat.Commits, &stat.Additions, &stat.Deletions, &stat.Comments, &stat.PullRequests, &stat.Score,
			&stat.CreatedAt, &stat.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		stats = append(stats, &stat)
	}

	return stats, nil
}

// GetByRepositoryID retrieves all people statistics for a repository
func (r *PeopleStatisticsRepository) GetByRepositoryID(repositoryID string) ([]*models.PeopleStatistics, error) {
	query := `
		SELECT id, project_id, repository_id, github_person_id, stat_date,
		       commits, additions, deletions, comments, pull_requests, score,
		       created_at, updated_at
		FROM people_statistics WHERE repository_id = ?
		ORDER BY stat_date DESC, score DESC
	`

	rows, err := r.db.Query(query, repositoryID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stats []*models.PeopleStatistics
	for rows.Next() {
		var stat models.PeopleStatistics
		err := rows.Scan(
			&stat.ID, &stat.ProjectID, &stat.RepositoryID, &stat.GithubPersonID, &stat.StatDate,
			&stat.Commits, &stat.Additions, &stat.Deletions, &stat.Comments, &stat.PullRequests, &stat.Score,
			&stat.CreatedAt, &stat.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		stats = append(stats, &stat)
	}

	return stats, nil
}

// GetByGithubPersonID retrieves all people statistics for a GitHub person
func (r *PeopleStatisticsRepository) GetByGithubPersonID(githubPersonID string) ([]*models.PeopleStatistics, error) {
	query := `
		SELECT id, project_id, repository_id, github_person_id, stat_date,
		       commits, additions, deletions, comments, pull_requests, score,
		       created_at, updated_at
		FROM people_statistics WHERE github_person_id = ?
		ORDER BY stat_date DESC, score DESC
	`

	rows, err := r.db.Query(query, githubPersonID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stats []*models.PeopleStatistics
	for rows.Next() {
		var stat models.PeopleStatistics
		err := rows.Scan(
			&stat.ID, &stat.ProjectID, &stat.RepositoryID, &stat.GithubPersonID, &stat.StatDate,
			&stat.Commits, &stat.Additions, &stat.Deletions, &stat.Comments, &stat.PullRequests, &stat.Score,
			&stat.CreatedAt, &stat.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		stats = append(stats, &stat)
	}

	return stats, nil
}

// GetByProjectAndPerson retrieves statistics for a specific person in a project
func (r *PeopleStatisticsRepository) GetByProjectAndPerson(projectID, githubPersonID string) ([]*models.PeopleStatistics, error) {
	query := `
		SELECT id, project_id, repository_id, github_person_id, stat_date,
		       commits, additions, deletions, comments, pull_requests, score,
		       created_at, updated_at
		FROM people_statistics WHERE project_id = ? AND github_person_id = ?
		ORDER BY stat_date DESC, score DESC
	`

	rows, err := r.db.Query(query, projectID, githubPersonID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stats []*models.PeopleStatistics
	for rows.Next() {
		var stat models.PeopleStatistics
		err := rows.Scan(
			&stat.ID, &stat.ProjectID, &stat.RepositoryID, &stat.GithubPersonID, &stat.StatDate,
			&stat.Commits, &stat.Additions, &stat.Deletions, &stat.Comments, &stat.PullRequests, &stat.Score,
			&stat.CreatedAt, &stat.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		stats = append(stats, &stat)
	}

	return stats, nil
}

// GetByRepositoryAndPerson retrieves statistics for a specific person in a repository
func (r *PeopleStatisticsRepository) GetByRepositoryAndPerson(repositoryID, githubPersonID string) ([]*models.PeopleStatistics, error) {
	query := `
		SELECT id, project_id, repository_id, github_person_id, stat_date,
		       commits, additions, deletions, comments, pull_requests, score,
		       created_at, updated_at
		FROM people_statistics WHERE repository_id = ? AND github_person_id = ?
		ORDER BY stat_date DESC, score DESC
	`

	rows, err := r.db.Query(query, repositoryID, githubPersonID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stats []*models.PeopleStatistics
	for rows.Next() {
		var stat models.PeopleStatistics
		err := rows.Scan(
			&stat.ID, &stat.ProjectID, &stat.RepositoryID, &stat.GithubPersonID, &stat.StatDate,
			&stat.Commits, &stat.Additions, &stat.Deletions, &stat.Comments, &stat.PullRequests, &stat.Score,
			&stat.CreatedAt, &stat.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		stats = append(stats, &stat)
	}

	return stats, nil
}

// GetByDateRange retrieves statistics within a date range for a project
func (r *PeopleStatisticsRepository) GetByDateRange(projectID string, startDate, endDate time.Time) ([]*models.PeopleStatistics, error) {
	query := `
		SELECT id, project_id, repository_id, github_person_id, stat_date,
		       commits, additions, deletions, comments, pull_requests, score,
		       created_at, updated_at
		FROM people_statistics WHERE project_id = ? AND stat_date BETWEEN ? AND ?
		ORDER BY stat_date DESC, score DESC
	`

	rows, err := r.db.Query(query, projectID, startDate, endDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stats []*models.PeopleStatistics
	for rows.Next() {
		var stat models.PeopleStatistics
		err := rows.Scan(
			&stat.ID, &stat.ProjectID, &stat.RepositoryID, &stat.GithubPersonID, &stat.StatDate,
			&stat.Commits, &stat.Additions, &stat.Deletions, &stat.Comments, &stat.PullRequests, &stat.Score,
			&stat.CreatedAt, &stat.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		stats = append(stats, &stat)
	}

	return stats, nil
}

// GetByDate retrieves statistics for a specific date
func (r *PeopleStatisticsRepository) GetByDate(projectID string, statDate time.Time) ([]*models.PeopleStatistics, error) {
	query := `
		SELECT id, project_id, repository_id, github_person_id, stat_date,
		       commits, additions, deletions, comments, pull_requests, score,
		       created_at, updated_at
		FROM people_statistics WHERE project_id = ? AND stat_date = ?
		ORDER BY score DESC
	`

	rows, err := r.db.Query(query, projectID, statDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stats []*models.PeopleStatistics
	for rows.Next() {
		var stat models.PeopleStatistics
		err := rows.Scan(
			&stat.ID, &stat.ProjectID, &stat.RepositoryID, &stat.GithubPersonID, &stat.StatDate,
			&stat.Commits, &stat.Additions, &stat.Deletions, &stat.Comments, &stat.PullRequests, &stat.Score,
			&stat.CreatedAt, &stat.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		stats = append(stats, &stat)
	}

	return stats, nil
}

// Upsert creates or updates people statistics
func (r *PeopleStatisticsRepository) Upsert(stats *models.PeopleStatistics) error {
	query := `
		INSERT INTO people_statistics (
			id, project_id, repository_id, github_person_id, stat_date,
			commits, additions, deletions, comments, pull_requests, score,
			created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(project_id, repository_id, github_person_id, stat_date)
		DO UPDATE SET
			commits = EXCLUDED.commits,
			additions = EXCLUDED.additions,
			deletions = EXCLUDED.deletions,
			comments = EXCLUDED.comments,
			pull_requests = EXCLUDED.pull_requests,
			score = EXCLUDED.score,
			updated_at = CURRENT_TIMESTAMP
	`

	_, err := r.db.Exec(query,
		stats.ID, stats.ProjectID, stats.RepositoryID, stats.GithubPersonID, stats.StatDate,
		stats.Commits, stats.Additions, stats.Deletions, stats.Comments, stats.PullRequests, stats.Score,
		stats.CreatedAt, stats.UpdatedAt,
	)

	return err
}

// Update updates an existing people statistics record
func (r *PeopleStatisticsRepository) Update(stats *models.PeopleStatistics) error {
	query := `
		UPDATE people_statistics SET
			commits = ?, additions = ?, deletions = ?, comments = ?, pull_requests = ?, score = ?,
			updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`

	_, err := r.db.Exec(query,
		stats.Commits, stats.Additions, stats.Deletions, stats.Comments, stats.PullRequests, stats.Score,
		stats.ID,
	)

	return err
}

// Delete deletes a people statistics record
func (r *PeopleStatisticsRepository) Delete(id string) error {
	query := `DELETE FROM people_statistics WHERE id = ?`
	_, err := r.db.Exec(query, id)
	return err
}

// DeleteByProjectID deletes all statistics for a project
func (r *PeopleStatisticsRepository) DeleteByProjectID(projectID string) error {
	query := `DELETE FROM people_statistics WHERE project_id = ?`
	_, err := r.db.Exec(query, projectID)
	return err
}

// DeleteByRepositoryID deletes all statistics for a repository
func (r *PeopleStatisticsRepository) DeleteByRepositoryID(repositoryID string) error {
	query := `DELETE FROM people_statistics WHERE repository_id = ?`
	_, err := r.db.Exec(query, repositoryID)
	return err
}

// GetByProjectAndPersonAndYear retrieves statistics for a specific person in a project for a specific year
func (r *PeopleStatisticsRepository) GetByProjectAndPersonAndYear(projectID, githubPersonID string, year int) ([]*models.PeopleStatistics, error) {
	query := `
		SELECT id, project_id, repository_id, github_person_id, stat_date,
		       commits, additions, deletions, comments, pull_requests, score,
		       created_at, updated_at
		FROM people_statistics 
		WHERE project_id = ? AND github_person_id = ? AND strftime('%Y', stat_date) = ?
		ORDER BY stat_date DESC
	`

	rows, err := r.db.Query(query, projectID, githubPersonID, fmt.Sprintf("%d", year))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stats []*models.PeopleStatistics
	for rows.Next() {
		var stat models.PeopleStatistics
		err := rows.Scan(
			&stat.ID, &stat.ProjectID, &stat.RepositoryID, &stat.GithubPersonID, &stat.StatDate,
			&stat.Commits, &stat.Additions, &stat.Deletions, &stat.Comments, &stat.PullRequests, &stat.Score,
			&stat.CreatedAt, &stat.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		stats = append(stats, &stat)
	}

	return stats, nil
}

// GetDateRangeForProject retrieves the earliest and latest dates for a project
func (r *PeopleStatisticsRepository) GetDateRangeForProject(projectID string) (*time.Time, *time.Time, error) {
	query := `
		SELECT MIN(stat_date), MAX(stat_date)
		FROM people_statistics 
		WHERE project_id = ?
	`

	var earliestDateStr, latestDateStr sql.NullString
	err := r.db.QueryRow(query, projectID).Scan(&earliestDateStr, &latestDateStr)
	if err != nil {
		return nil, nil, err
	}

	var earliest, latest *time.Time
	if earliestDateStr.Valid && earliestDateStr.String != "" {
		if parsed, err := time.Parse("2006-01-02", earliestDateStr.String); err == nil {
			earliest = &parsed
		}
	}
	if latestDateStr.Valid && latestDateStr.String != "" {
		if parsed, err := time.Parse("2006-01-02", latestDateStr.String); err == nil {
			latest = &parsed
		}
	}

	return earliest, latest, nil
}

// GetByProjectAndPersonAndMonth retrieves statistics for a specific person in a project for a specific month
func (r *PeopleStatisticsRepository) GetByProjectAndPersonAndMonth(projectID, githubPersonID string, year int, month int) ([]*models.PeopleStatistics, error) {
	query := `
		SELECT id, project_id, repository_id, github_person_id, stat_date,
		       commits, additions, deletions, comments, pull_requests, score,
		       created_at, updated_at
		FROM people_statistics 
		WHERE project_id = ? AND github_person_id = ? AND strftime('%Y', stat_date) = ? AND strftime('%m', stat_date) = ?
		ORDER BY stat_date DESC
	`

	rows, err := r.db.Query(query, projectID, githubPersonID, fmt.Sprintf("%d", year), fmt.Sprintf("%02d", month))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stats []*models.PeopleStatistics
	for rows.Next() {
		var stat models.PeopleStatistics
		err := rows.Scan(
			&stat.ID, &stat.ProjectID, &stat.RepositoryID, &stat.GithubPersonID, &stat.StatDate,
			&stat.Commits, &stat.Additions, &stat.Deletions, &stat.Comments, &stat.PullRequests, &stat.Score,
			&stat.CreatedAt, &stat.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		stats = append(stats, &stat)
	}

	return stats, nil
}

// GetByProjectAndPersonAndWeek retrieves statistics for a specific person in a project for a specific week
func (r *PeopleStatisticsRepository) GetByProjectAndPersonAndWeek(projectID, githubPersonID string, year int, week int) ([]*models.PeopleStatistics, error) {
	query := `
		SELECT id, project_id, repository_id, github_person_id, stat_date,
		       commits, additions, deletions, comments, pull_requests, score,
		       created_at, updated_at
		FROM people_statistics 
		WHERE project_id = ? AND github_person_id = ? AND strftime('%Y', stat_date) = ? AND strftime('%W', stat_date) = ?
		ORDER BY stat_date DESC
	`

	rows, err := r.db.Query(query, projectID, githubPersonID, fmt.Sprintf("%d", year), fmt.Sprintf("%02d", week))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stats []*models.PeopleStatistics
	for rows.Next() {
		var stat models.PeopleStatistics
		err := rows.Scan(
			&stat.ID, &stat.ProjectID, &stat.RepositoryID, &stat.GithubPersonID, &stat.StatDate,
			&stat.Commits, &stat.Additions, &stat.Deletions, &stat.Comments, &stat.PullRequests, &stat.Score,
			&stat.CreatedAt, &stat.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		stats = append(stats, &stat)
	}

	return stats, nil
}

// GetByProjectAndPersonAndDate retrieves statistics for a specific person in a project for a specific date
func (r *PeopleStatisticsRepository) GetByProjectAndPersonAndDate(projectID, githubPersonID string, date time.Time) ([]*models.PeopleStatistics, error) {
	query := `
		SELECT id, project_id, repository_id, github_person_id, stat_date,
		       commits, additions, deletions, comments, pull_requests, score,
		       created_at, updated_at
		FROM people_statistics 
		WHERE project_id = ? AND github_person_id = ? AND stat_date = ?
		ORDER BY stat_date DESC
	`

	rows, err := r.db.Query(query, projectID, githubPersonID, date.Format("2006-01-02"))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stats []*models.PeopleStatistics
	for rows.Next() {
		var stat models.PeopleStatistics
		err := rows.Scan(
			&stat.ID, &stat.ProjectID, &stat.RepositoryID, &stat.GithubPersonID, &stat.StatDate,
			&stat.Commits, &stat.Additions, &stat.Deletions, &stat.Comments, &stat.PullRequests, &stat.Score,
			&stat.CreatedAt, &stat.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		stats = append(stats, &stat)
	}

	return stats, nil
}
