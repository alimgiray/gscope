package repositories

import (
	"database/sql"
	"time"

	"github.com/alimgiray/gscope/internal/models"
)

type ProjectRepositoryRepository struct {
	db *sql.DB
}

func NewProjectRepositoryRepository(db *sql.DB) *ProjectRepositoryRepository {
	return &ProjectRepositoryRepository{db: db}
}

// Create creates a new project repository relationship
func (r *ProjectRepositoryRepository) Create(projectRepo *models.ProjectRepository) error {
	query := `
		INSERT INTO project_repositories (
			id, project_id, github_repo_id, is_analyzed, is_tracked, last_analyzed, deleted_at
		) VALUES (?, ?, ?, ?, ?, ?, ?)
	`

	_, err := r.db.Exec(query,
		projectRepo.ID, projectRepo.ProjectID, projectRepo.GithubRepoID,
		projectRepo.IsAnalyzed, projectRepo.IsTracked, projectRepo.LastAnalyzed, projectRepo.DeletedAt,
	)

	return err
}

// GetByID retrieves a project repository by ID
func (r *ProjectRepositoryRepository) GetByID(id string) (*models.ProjectRepository, error) {
	query := `
		SELECT id, project_id, github_repo_id, is_analyzed, is_tracked, last_analyzed,
			   created_at, updated_at, deleted_at
		FROM project_repositories WHERE id = ? AND deleted_at IS NULL
	`

	projectRepo := &models.ProjectRepository{}
	err := r.db.QueryRow(query, id).Scan(
		&projectRepo.ID, &projectRepo.ProjectID, &projectRepo.GithubRepoID,
		&projectRepo.IsAnalyzed, &projectRepo.IsTracked, &projectRepo.LastAnalyzed,
		&projectRepo.CreatedAt, &projectRepo.UpdatedAt, &projectRepo.DeletedAt,
	)

	if err != nil {
		return nil, err
	}

	return projectRepo, nil
}

// GetByProjectID retrieves all repositories for a project
func (r *ProjectRepositoryRepository) GetByProjectID(projectID string) ([]*models.ProjectRepository, error) {
	query := `
		SELECT id, project_id, github_repo_id, is_analyzed, is_tracked, last_analyzed,
			   created_at, updated_at, deleted_at
		FROM project_repositories WHERE project_id = ? AND deleted_at IS NULL
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(query, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var projectRepos []*models.ProjectRepository
	for rows.Next() {
		projectRepo := &models.ProjectRepository{}
		err := rows.Scan(
			&projectRepo.ID, &projectRepo.ProjectID, &projectRepo.GithubRepoID,
			&projectRepo.IsAnalyzed, &projectRepo.IsTracked, &projectRepo.LastAnalyzed,
			&projectRepo.CreatedAt, &projectRepo.UpdatedAt, &projectRepo.DeletedAt,
		)
		if err != nil {
			return nil, err
		}
		projectRepos = append(projectRepos, projectRepo)
	}

	return projectRepos, nil
}

// UpdateLastAnalyzed updates the last_analyzed field for a project repository
func (r *ProjectRepositoryRepository) UpdateLastAnalyzed(id string, lastAnalyzed *time.Time) error {
	query := `
		UPDATE project_repositories 
		SET last_analyzed = ?, is_analyzed = TRUE, updated_at = CURRENT_TIMESTAMP
		WHERE id = ? AND deleted_at IS NULL
	`

	_, err := r.db.Exec(query, lastAnalyzed, id)
	return err
}

// GetByGithubRepoID retrieves all projects that use a specific GitHub repository
func (r *ProjectRepositoryRepository) GetByGithubRepoID(githubRepoID string) ([]*models.ProjectRepository, error) {
	query := `
		SELECT id, project_id, github_repo_id, is_analyzed, is_tracked, last_analyzed,
			   created_at, updated_at, deleted_at
		FROM project_repositories WHERE github_repo_id = ? AND deleted_at IS NULL
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(query, githubRepoID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var projectRepos []*models.ProjectRepository
	for rows.Next() {
		projectRepo := &models.ProjectRepository{}
		err := rows.Scan(
			&projectRepo.ID, &projectRepo.ProjectID, &projectRepo.GithubRepoID,
			&projectRepo.IsAnalyzed, &projectRepo.IsTracked, &projectRepo.LastAnalyzed,
			&projectRepo.CreatedAt, &projectRepo.UpdatedAt, &projectRepo.DeletedAt,
		)
		if err != nil {
			return nil, err
		}
		projectRepos = append(projectRepos, projectRepo)
	}

	return projectRepos, nil
}

// GetByProjectAndGithubRepo retrieves a specific project-repository relationship
func (r *ProjectRepositoryRepository) GetByProjectAndGithubRepo(projectID, githubRepoID string) (*models.ProjectRepository, error) {
	query := `
		SELECT id, project_id, github_repo_id, is_analyzed, is_tracked, last_analyzed,
			   created_at, updated_at, deleted_at
		FROM project_repositories 
		WHERE project_id = ? AND github_repo_id = ? AND deleted_at IS NULL
	`

	projectRepo := &models.ProjectRepository{}
	err := r.db.QueryRow(query, projectID, githubRepoID).Scan(
		&projectRepo.ID, &projectRepo.ProjectID, &projectRepo.GithubRepoID,
		&projectRepo.IsAnalyzed, &projectRepo.IsTracked, &projectRepo.LastAnalyzed,
		&projectRepo.CreatedAt, &projectRepo.UpdatedAt, &projectRepo.DeletedAt,
	)

	if err != nil {
		return nil, err
	}

	return projectRepo, nil
}

// Update updates a project repository
func (r *ProjectRepositoryRepository) Update(projectRepo *models.ProjectRepository) error {
	query := `
		UPDATE project_repositories SET
			project_id = ?, github_repo_id = ?, is_analyzed = ?, is_tracked = ?, last_analyzed = ?
		WHERE id = ?
	`

	_, err := r.db.Exec(query,
		projectRepo.ProjectID, projectRepo.GithubRepoID, projectRepo.IsAnalyzed,
		projectRepo.IsTracked, projectRepo.LastAnalyzed, projectRepo.ID,
	)

	return err
}

// Delete performs a soft delete of a project repository
func (r *ProjectRepositoryRepository) Delete(id string) error {
	query := `UPDATE project_repositories SET deleted_at = ? WHERE id = ?`
	_, err := r.db.Exec(query, time.Now(), id)
	return err
}

// HardDelete permanently deletes a project repository
func (r *ProjectRepositoryRepository) HardDelete(id string) error {
	query := `DELETE FROM project_repositories WHERE id = ?`
	_, err := r.db.Exec(query, id)
	return err
}

// UpdateAnalysisStatus updates the analysis status of a project repository
func (r *ProjectRepositoryRepository) UpdateAnalysisStatus(id string, isAnalyzed bool) error {
	query := `
		UPDATE project_repositories SET
			is_analyzed = ?, last_analyzed = ?
		WHERE id = ?
	`

	var lastAnalyzed *time.Time
	if isAnalyzed {
		now := time.Now()
		lastAnalyzed = &now
	}

	_, err := r.db.Exec(query, isAnalyzed, lastAnalyzed, id)
	return err
}

// UpdateTrackingStatus updates the tracking status of a project repository
func (r *ProjectRepositoryRepository) UpdateTrackingStatus(id string, isTracked bool) error {
	query := `
		UPDATE project_repositories SET
			is_tracked = ?
		WHERE id = ?
	`

	_, err := r.db.Exec(query, isTracked, id)
	return err
}

// ListAll retrieves all project repositories (not deleted)
func (r *ProjectRepositoryRepository) ListAll() ([]*models.ProjectRepository, error) {
	query := `
		SELECT id, project_id, github_repo_id, is_analyzed, is_tracked, last_analyzed,
			   created_at, updated_at, deleted_at
		FROM project_repositories WHERE deleted_at IS NULL
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var projectRepos []*models.ProjectRepository
	for rows.Next() {
		projectRepo := &models.ProjectRepository{}
		err := rows.Scan(
			&projectRepo.ID, &projectRepo.ProjectID, &projectRepo.GithubRepoID,
			&projectRepo.IsAnalyzed, &projectRepo.IsTracked, &projectRepo.LastAnalyzed,
			&projectRepo.CreatedAt, &projectRepo.UpdatedAt, &projectRepo.DeletedAt,
		)
		if err != nil {
			return nil, err
		}
		projectRepos = append(projectRepos, projectRepo)
	}

	return projectRepos, nil
}

// GetAnalyzedRepositories retrieves all analyzed repositories for a project
func (r *ProjectRepositoryRepository) GetAnalyzedRepositories(projectID string) ([]*models.ProjectRepository, error) {
	query := `
		SELECT id, project_id, github_repo_id, is_analyzed, is_tracked, last_analyzed,
			   created_at, updated_at, deleted_at
		FROM project_repositories 
		WHERE project_id = ? AND is_analyzed = 1 AND deleted_at IS NULL
		ORDER BY last_analyzed DESC
	`

	rows, err := r.db.Query(query, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var projectRepos []*models.ProjectRepository
	for rows.Next() {
		projectRepo := &models.ProjectRepository{}
		err := rows.Scan(
			&projectRepo.ID, &projectRepo.ProjectID, &projectRepo.GithubRepoID,
			&projectRepo.IsAnalyzed, &projectRepo.IsTracked, &projectRepo.LastAnalyzed,
			&projectRepo.CreatedAt, &projectRepo.UpdatedAt, &projectRepo.DeletedAt,
		)
		if err != nil {
			return nil, err
		}
		projectRepos = append(projectRepos, projectRepo)
	}

	return projectRepos, nil
}
