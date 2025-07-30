package repositories

import (
	"database/sql"
	"sync"
	"time"

	"github.com/alimgiray/gscope/internal/models"
)

type CommitRepository struct {
	db *sql.DB
	mu sync.RWMutex
}

func NewCommitRepository(db *sql.DB) *CommitRepository {
	return &CommitRepository{db: db}
}

// Create creates a new commit
func (r *CommitRepository) Create(commit *models.Commit) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	query := `
		INSERT INTO commits (
			id, github_repository_id, commit_sha, message, author_name, author_email,
			commit_date, is_merge_commit, merge_commit_sha, additions, deletions, changes
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := r.db.Exec(query,
		commit.ID, commit.GithubRepositoryID, commit.CommitSHA, commit.Message,
		commit.AuthorName, commit.AuthorEmail, commit.CommitDate, commit.IsMergeCommit,
		commit.MergeCommitSHA, commit.Additions, commit.Deletions, commit.Changes,
	)

	return err
}

// GetByID retrieves a commit by ID
func (r *CommitRepository) GetByID(id string) (*models.Commit, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	query := `
		SELECT id, github_repository_id, commit_sha, message, author_name, author_email,
			   commit_date, is_merge_commit, merge_commit_sha, additions, deletions, changes, created_at
		FROM commits WHERE id = ?
	`

	commit := &models.Commit{}
	err := r.db.QueryRow(query, id).Scan(
		&commit.ID, &commit.GithubRepositoryID, &commit.CommitSHA, &commit.Message,
		&commit.AuthorName, &commit.AuthorEmail, &commit.CommitDate, &commit.IsMergeCommit,
		&commit.MergeCommitSHA, &commit.Additions, &commit.Deletions, &commit.Changes, &commit.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	return commit, nil
}

// GetByCommitSHA retrieves a commit by its SHA
func (r *CommitRepository) GetByCommitSHA(commitSHA string) (*models.Commit, error) {
	query := `
		SELECT id, github_repository_id, commit_sha, message, author_name, author_email,
			   commit_date, is_merge_commit, merge_commit_sha, additions, deletions, changes, created_at
		FROM commits WHERE commit_sha = ?
	`

	commit := &models.Commit{}
	err := r.db.QueryRow(query, commitSHA).Scan(
		&commit.ID, &commit.GithubRepositoryID, &commit.CommitSHA, &commit.Message,
		&commit.AuthorName, &commit.AuthorEmail, &commit.CommitDate, &commit.IsMergeCommit,
		&commit.MergeCommitSHA, &commit.Additions, &commit.Deletions, &commit.Changes, &commit.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	return commit, nil
}

// GetByRepositoryID retrieves all commits for a repository
func (r *CommitRepository) GetByRepositoryID(repositoryID string) ([]*models.Commit, error) {
	query := `
		SELECT id, github_repository_id, commit_sha, message, author_name, author_email,
			   commit_date, is_merge_commit, merge_commit_sha, additions, deletions, changes, created_at
		FROM commits WHERE github_repository_id = ?
		ORDER BY commit_date DESC
	`

	rows, err := r.db.Query(query, repositoryID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var commits []*models.Commit
	for rows.Next() {
		commit := &models.Commit{}
		err := rows.Scan(
			&commit.ID, &commit.GithubRepositoryID, &commit.CommitSHA, &commit.Message,
			&commit.AuthorName, &commit.AuthorEmail, &commit.CommitDate, &commit.IsMergeCommit,
			&commit.MergeCommitSHA, &commit.Additions, &commit.Deletions, &commit.Changes, &commit.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		commits = append(commits, commit)
	}

	return commits, nil
}

// GetEmailStatsByProjectID retrieves email statistics for a project
func (r *CommitRepository) GetEmailStatsByProjectID(projectID string, mergedEmails map[string]string) ([]*models.EmailStats, error) {

	query := `
		SELECT 
			c.author_email,
			c.author_name,
			MIN(c.commit_date) as first_commit,
			MAX(c.commit_date) as last_commit
		FROM commits c
		INNER JOIN github_repositories gr ON c.github_repository_id = gr.id
		INNER JOIN project_repositories pr ON gr.id = pr.github_repo_id
		WHERE pr.project_id = ?
		GROUP BY c.author_email, c.author_name
		ORDER BY c.author_email ASC
	`

	rows, err := r.db.Query(query, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var emailStats []*models.EmailStats
	for rows.Next() {
		stats := &models.EmailStats{}
		var authorName *string
		var firstCommitStr, lastCommitStr *string

		err := rows.Scan(
			&stats.Email,
			&authorName,
			&firstCommitStr,
			&lastCommitStr,
		)
		if err != nil {
			return nil, err
		}

		stats.Name = authorName

		// Parse time strings to time.Time
		if firstCommitStr != nil {
			if t, err := time.Parse("2006-01-02 15:04:05-07:00", *firstCommitStr); err == nil {
				stats.FirstCommit = &t
			} else if t, err := time.Parse("2006-01-02 15:04:05", *firstCommitStr); err == nil {
				stats.FirstCommit = &t
			}
		}

		if lastCommitStr != nil {
			if t, err := time.Parse("2006-01-02 15:04:05-07:00", *lastCommitStr); err == nil {
				stats.LastCommit = &t
			} else if t, err := time.Parse("2006-01-02 15:04:05", *lastCommitStr); err == nil {
				stats.LastCommit = &t
			}
		}

		emailStats = append(emailStats, stats)
	}

	// Apply email merges
	if mergedEmails != nil {
		emailStatsMap := make(map[string]*models.EmailStats)
		mergedIntoMap := make(map[string][]string) // target_email -> []source_emails

		// First pass: collect all stats and apply merges
		for _, stats := range emailStats {
			// Check if this email should be merged
			if targetEmail, exists := mergedEmails[stats.Email]; exists {
				// This email should be merged into targetEmail
				if targetStats, targetExists := emailStatsMap[targetEmail]; targetExists {
					// Target email already exists, merge the stats
					if stats.FirstCommit != nil && (targetStats.FirstCommit == nil || stats.FirstCommit.Before(*targetStats.FirstCommit)) {
						targetStats.FirstCommit = stats.FirstCommit
					}
					if stats.LastCommit != nil && (targetStats.LastCommit == nil || stats.LastCommit.After(*targetStats.LastCommit)) {
						targetStats.LastCommit = stats.LastCommit
					}
				} else {
					// Target email doesn't exist yet, create it
					mergedStats := &models.EmailStats{
						Email:       targetEmail,
						Name:        stats.Name,
						FirstCommit: stats.FirstCommit,
						LastCommit:  stats.LastCommit,
					}
					emailStatsMap[targetEmail] = mergedStats
				}

				// Track which emails are merged into this target
				mergedIntoMap[targetEmail] = append(mergedIntoMap[targetEmail], stats.Email)
			} else {
				// This email is not merged, check if it's a target email
				// Only add it if it's not already in the map (from a previous merge)
				if _, isTarget := emailStatsMap[stats.Email]; !isTarget {
					emailStatsMap[stats.Email] = stats
				}
			}
		}

		// Second pass: handle target emails that exist as separate entries
		for _, stats := range emailStats {
			// Skip if this email is already processed (merged into another)
			if _, isMerged := mergedEmails[stats.Email]; isMerged {
				continue
			}

			// Check if this email is a target email (has other emails merged into it)
			if _, hasMerged := mergedIntoMap[stats.Email]; hasMerged {
				// This is a target email, merge its stats with any existing target stats
				if targetStats, targetExists := emailStatsMap[stats.Email]; targetExists {
					// Merge the stats
					if stats.FirstCommit != nil && (targetStats.FirstCommit == nil || stats.FirstCommit.Before(*targetStats.FirstCommit)) {
						targetStats.FirstCommit = stats.FirstCommit
					}
					if stats.LastCommit != nil && (targetStats.LastCommit == nil || stats.LastCommit.After(*targetStats.LastCommit)) {
						targetStats.LastCommit = stats.LastCommit
					}
				} else {
					// Target email doesn't exist in map yet, add it
					emailStatsMap[stats.Email] = stats
				}
			}
		}

		// Second pass: mark merged emails and populate merged emails list
		for email, stats := range emailStatsMap {
			// Check if this email has been merged into another
			if _, isMerged := mergedEmails[email]; isMerged {
				stats.IsMerged = true
			}

			// Add list of emails merged into this one
			if mergedEmails, hasMerged := mergedIntoMap[email]; hasMerged {
				stats.MergedEmails = mergedEmails
			}
		}

		// Convert map back to slice (only include target emails, not merged ones)
		emailStats = make([]*models.EmailStats, 0, len(emailStatsMap))
		for _, stats := range emailStatsMap {
			emailStats = append(emailStats, stats)
		}

		// Re-sort alphabetically by email
		for i := 0; i < len(emailStats)-1; i++ {
			for j := i + 1; j < len(emailStats); j++ {
				if emailStats[i].Email > emailStats[j].Email {
					emailStats[i], emailStats[j] = emailStats[j], emailStats[i]
				}
			}
		}
	}

	return emailStats, nil
}

// Update updates an existing commit
func (r *CommitRepository) Update(commit *models.Commit) error {
	query := `
		UPDATE commits SET
			message = ?, author_name = ?, author_email = ?, commit_date = ?,
			is_merge_commit = ?, merge_commit_sha = ?, additions = ?, deletions = ?, changes = ?
		WHERE id = ?
	`

	_, err := r.db.Exec(query,
		commit.Message, commit.AuthorName, commit.AuthorEmail, commit.CommitDate,
		commit.IsMergeCommit, commit.MergeCommitSHA, commit.Additions, commit.Deletions, commit.Changes,
		commit.ID,
	)

	return err
}

// Delete deletes a commit by ID
func (r *CommitRepository) Delete(id string) error {
	query := `DELETE FROM commits WHERE id = ?`
	_, err := r.db.Exec(query, id)
	return err
}

// DeleteByRepositoryID deletes all commits for a repository
func (r *CommitRepository) DeleteByRepositoryID(repositoryID string) error {
	query := `DELETE FROM commits WHERE github_repository_id = ?`
	_, err := r.db.Exec(query, repositoryID)
	return err
}

// ExistsByCommitSHA checks if a commit exists by its SHA
func (r *CommitRepository) ExistsByCommitSHA(commitSHA string) (bool, error) {
	query := `SELECT COUNT(*) FROM commits WHERE commit_sha = ?`
	var count int
	err := r.db.QueryRow(query, commitSHA).Scan(&count)
	return count > 0, err
}

// GetDateRangeByRepositoryID gets the minimum and maximum commit dates for a repository
func (r *CommitRepository) GetDateRangeByRepositoryID(repositoryID string) (time.Time, time.Time, error) {
	query := `SELECT MIN(commit_date), MAX(commit_date) FROM commits WHERE github_repository_id = ?`

	var minDateStr, maxDateStr string
	err := r.db.QueryRow(query, repositoryID).Scan(&minDateStr, &maxDateStr)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}

	// Parse the date strings
	minDate, err := time.Parse("2006-01-02 15:04:05-07:00", minDateStr)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}

	maxDate, err := time.Parse("2006-01-02 15:04:05-07:00", maxDateStr)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}

	return minDate, maxDate, nil
}

// GetLatestCommitDateByRepositoryID gets the latest commit date for a repository
func (r *CommitRepository) GetLatestCommitDateByRepositoryID(repositoryID string) (time.Time, error) {
	query := `SELECT MAX(commit_date) FROM commits WHERE github_repository_id = ?`

	var latestDateStr sql.NullString
	err := r.db.QueryRow(query, repositoryID).Scan(&latestDateStr)
	if err != nil {
		return time.Time{}, err
	}

	// If no commits found, return zero time
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

// GetDateRangeByProjectID retrieves the earliest and latest commit dates for a project
func (r *CommitRepository) GetDateRangeByProjectID(projectID string) (time.Time, time.Time, error) {
	query := `
		SELECT MIN(c.commit_date), MAX(c.commit_date)
		FROM commits c
		INNER JOIN github_repositories gr ON c.github_repository_id = gr.id
		INNER JOIN project_repositories pr ON gr.id = pr.github_repo_id
		WHERE pr.project_id = ?
	`

	var earliestDateStr, latestDateStr sql.NullString
	err := r.db.QueryRow(query, projectID).Scan(&earliestDateStr, &latestDateStr)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}

	var earliest, latest time.Time
	if earliestDateStr.Valid && earliestDateStr.String != "" {
		if parsed, err := time.Parse("2006-01-02 15:04:05-07:00", earliestDateStr.String); err == nil {
			earliest = parsed
		}
	}
	if latestDateStr.Valid && latestDateStr.String != "" {
		if parsed, err := time.Parse("2006-01-02 15:04:05-07:00", latestDateStr.String); err == nil {
			latest = parsed
		}
	}

	return earliest, latest, nil
}
