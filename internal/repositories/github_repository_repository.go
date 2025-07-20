package repositories

import (
	"database/sql"
	"time"

	"github.com/alimgiray/gscope/internal/models"
)

type GitHubRepositoryRepository struct {
	db *sql.DB
}

func NewGitHubRepositoryRepository(db *sql.DB) *GitHubRepositoryRepository {
	return &GitHubRepositoryRepository{db: db}
}

// Create creates a new GitHub repository
func (r *GitHubRepositoryRepository) Create(repo *models.GitHubRepository) error {
	query := `
		INSERT INTO github_repositories (
			id, github_id, name, full_name, description, url, clone_url, language,
			stars, forks, private, default_branch, local_path, is_cloned, last_cloned,
			github_created_at, github_updated_at, github_pushed_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := r.db.Exec(query,
		repo.ID, repo.GithubID, repo.Name, repo.FullName, repo.Description,
		repo.URL, repo.CloneURL, repo.Language, repo.Stars, repo.Forks,
		repo.Private, repo.DefaultBranch, repo.LocalPath, repo.IsCloned,
		repo.LastCloned, repo.GithubCreatedAt, repo.GithubUpdatedAt,
		repo.GithubPushedAt,
	)

	return err
}

// GetByID retrieves a GitHub repository by ID
func (r *GitHubRepositoryRepository) GetByID(id string) (*models.GitHubRepository, error) {
	query := `
		SELECT id, github_id, name, full_name, description, url, clone_url, language,
			   stars, forks, private, default_branch, local_path, is_cloned, last_cloned,
			   github_created_at, github_updated_at, github_pushed_at, created_at, updated_at
		FROM github_repositories WHERE id = ?
	`

	repo := &models.GitHubRepository{}
	err := r.db.QueryRow(query, id).Scan(
		&repo.ID, &repo.GithubID, &repo.Name, &repo.FullName, &repo.Description,
		&repo.URL, &repo.CloneURL, &repo.Language, &repo.Stars, &repo.Forks,
		&repo.Private, &repo.DefaultBranch, &repo.LocalPath, &repo.IsCloned,
		&repo.LastCloned, &repo.GithubCreatedAt, &repo.GithubUpdatedAt,
		&repo.GithubPushedAt, &repo.CreatedAt, &repo.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return repo, nil
}

// GetByGithubID retrieves a GitHub repository by GitHub ID
func (r *GitHubRepositoryRepository) GetByGithubID(githubID int64) (*models.GitHubRepository, error) {
	query := `
		SELECT id, github_id, name, full_name, description, url, clone_url, language,
			   stars, forks, private, default_branch, local_path, is_cloned, last_cloned,
			   github_created_at, github_updated_at, github_pushed_at, created_at, updated_at
		FROM github_repositories WHERE github_id = ?
	`

	repo := &models.GitHubRepository{}
	err := r.db.QueryRow(query, githubID).Scan(
		&repo.ID, &repo.GithubID, &repo.Name, &repo.FullName, &repo.Description,
		&repo.URL, &repo.CloneURL, &repo.Language, &repo.Stars, &repo.Forks,
		&repo.Private, &repo.DefaultBranch, &repo.LocalPath, &repo.IsCloned,
		&repo.LastCloned, &repo.GithubCreatedAt, &repo.GithubUpdatedAt,
		&repo.GithubPushedAt, &repo.CreatedAt, &repo.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return repo, nil
}

// GetByFullName retrieves a GitHub repository by full name
func (r *GitHubRepositoryRepository) GetByFullName(fullName string) (*models.GitHubRepository, error) {
	query := `
		SELECT id, github_id, name, full_name, description, url, clone_url, language,
			   stars, forks, private, default_branch, local_path, is_cloned, last_cloned,
			   github_created_at, github_updated_at, github_pushed_at, created_at, updated_at
		FROM github_repositories WHERE full_name = ?
	`

	repo := &models.GitHubRepository{}
	err := r.db.QueryRow(query, fullName).Scan(
		&repo.ID, &repo.GithubID, &repo.Name, &repo.FullName, &repo.Description,
		&repo.URL, &repo.CloneURL, &repo.Language, &repo.Stars, &repo.Forks,
		&repo.Private, &repo.DefaultBranch, &repo.LocalPath, &repo.IsCloned,
		&repo.LastCloned, &repo.GithubCreatedAt, &repo.GithubUpdatedAt,
		&repo.GithubPushedAt, &repo.CreatedAt, &repo.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return repo, nil
}

// Update updates a GitHub repository
func (r *GitHubRepositoryRepository) Update(repo *models.GitHubRepository) error {
	query := `
		UPDATE github_repositories SET
			github_id = ?, name = ?, full_name = ?, description = ?, url = ?,
			clone_url = ?, language = ?, stars = ?, forks = ?, private = ?,
			default_branch = ?, local_path = ?, is_cloned = ?, last_cloned = ?,
			github_created_at = ?, github_updated_at = ?, github_pushed_at = ?
		WHERE id = ?
	`

	_, err := r.db.Exec(query,
		repo.GithubID, repo.Name, repo.FullName, repo.Description, repo.URL,
		repo.CloneURL, repo.Language, repo.Stars, repo.Forks, repo.Private,
		repo.DefaultBranch, repo.LocalPath, repo.IsCloned, repo.LastCloned,
		repo.GithubCreatedAt, repo.GithubUpdatedAt, repo.GithubPushedAt,
		repo.ID,
	)

	return err
}

// Delete deletes a GitHub repository
func (r *GitHubRepositoryRepository) Delete(id string) error {
	query := `DELETE FROM github_repositories WHERE id = ?`
	_, err := r.db.Exec(query, id)
	return err
}

// ListAll retrieves all GitHub repositories
func (r *GitHubRepositoryRepository) ListAll() ([]*models.GitHubRepository, error) {
	query := `
		SELECT id, github_id, name, full_name, description, url, clone_url, language,
			   stars, forks, private, default_branch, local_path, is_cloned, last_cloned,
			   github_created_at, github_updated_at, github_pushed_at, created_at, updated_at
		FROM github_repositories ORDER BY created_at DESC
	`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var repos []*models.GitHubRepository
	for rows.Next() {
		repo := &models.GitHubRepository{}
		err := rows.Scan(
			&repo.ID, &repo.GithubID, &repo.Name, &repo.FullName, &repo.Description,
			&repo.URL, &repo.CloneURL, &repo.Language, &repo.Stars, &repo.Forks,
			&repo.Private, &repo.DefaultBranch, &repo.LocalPath, &repo.IsCloned,
			&repo.LastCloned, &repo.GithubCreatedAt, &repo.GithubUpdatedAt,
			&repo.GithubPushedAt, &repo.CreatedAt, &repo.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		repos = append(repos, repo)
	}

	return repos, nil
}

// ListByLanguage retrieves repositories by language
func (r *GitHubRepositoryRepository) ListByLanguage(language string) ([]*models.GitHubRepository, error) {
	query := `
		SELECT id, github_id, name, full_name, description, url, clone_url, language,
			   stars, forks, private, default_branch, local_path, is_cloned, last_cloned,
			   github_created_at, github_updated_at, github_pushed_at, created_at, updated_at
		FROM github_repositories WHERE language = ? ORDER BY stars DESC
	`

	rows, err := r.db.Query(query, language)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var repos []*models.GitHubRepository
	for rows.Next() {
		repo := &models.GitHubRepository{}
		err := rows.Scan(
			&repo.ID, &repo.GithubID, &repo.Name, &repo.FullName, &repo.Description,
			&repo.URL, &repo.CloneURL, &repo.Language, &repo.Stars, &repo.Forks,
			&repo.Private, &repo.DefaultBranch, &repo.LocalPath, &repo.IsCloned,
			&repo.LastCloned, &repo.GithubCreatedAt, &repo.GithubUpdatedAt,
			&repo.GithubPushedAt, &repo.CreatedAt, &repo.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		repos = append(repos, repo)
	}

	return repos, nil
}

// UpdateCloneStatus updates the clone status of a repository
func (r *GitHubRepositoryRepository) UpdateCloneStatus(id string, isCloned bool, localPath *string) error {
	query := `
		UPDATE github_repositories SET
			is_cloned = ?, local_path = ?, last_cloned = ?
		WHERE id = ?
	`

	var lastCloned *time.Time
	if isCloned {
		now := time.Now()
		lastCloned = &now
	}

	_, err := r.db.Exec(query, isCloned, localPath, lastCloned, id)
	return err
}
