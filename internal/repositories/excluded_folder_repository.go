package repositories

import (
	"database/sql"
	"sync"

	"github.com/alimgiray/gscope/internal/models"
)

type ExcludedFolderRepository struct {
	db *sql.DB
	mu sync.RWMutex
}

func NewExcludedFolderRepository(db *sql.DB) *ExcludedFolderRepository {
	return &ExcludedFolderRepository{db: db}
}

// Create creates a new excluded folder
func (r *ExcludedFolderRepository) Create(folder *models.ExcludedFolder) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	query := `
		INSERT INTO excluded_folders (id, project_id, folder_path, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?)
	`

	_, err := r.db.Exec(query,
		folder.ID, folder.ProjectID, folder.FolderPath, folder.CreatedAt, folder.UpdatedAt,
	)

	return err
}

// GetByID retrieves an excluded folder by ID
func (r *ExcludedFolderRepository) GetByID(id string) (*models.ExcludedFolder, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	query := `
		SELECT id, project_id, folder_path, created_at, updated_at, deleted_at
		FROM excluded_folders WHERE id = ? AND deleted_at IS NULL
	`

	var folder models.ExcludedFolder
	err := r.db.QueryRow(query, id).Scan(
		&folder.ID, &folder.ProjectID, &folder.FolderPath, &folder.CreatedAt, &folder.UpdatedAt, &folder.DeletedAt,
	)

	if err != nil {
		return nil, err
	}

	return &folder, nil
}

// GetByProjectID retrieves all excluded folders for a project
func (r *ExcludedFolderRepository) GetByProjectID(projectID string) ([]*models.ExcludedFolder, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	query := `
		SELECT id, project_id, folder_path, created_at, updated_at, deleted_at
		FROM excluded_folders 
		WHERE project_id = ? AND deleted_at IS NULL
		ORDER BY folder_path ASC
	`

	rows, err := r.db.Query(query, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var folders []*models.ExcludedFolder
	for rows.Next() {
		var folder models.ExcludedFolder
		err := rows.Scan(
			&folder.ID, &folder.ProjectID, &folder.FolderPath, &folder.CreatedAt, &folder.UpdatedAt, &folder.DeletedAt,
		)
		if err != nil {
			return nil, err
		}
		folders = append(folders, &folder)
	}

	return folders, nil
}

// Update updates an excluded folder
func (r *ExcludedFolderRepository) Update(folder *models.ExcludedFolder) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	query := `
		UPDATE excluded_folders 
		SET project_id = ?, folder_path = ?, updated_at = ?
		WHERE id = ? AND deleted_at IS NULL
	`

	_, err := r.db.Exec(query,
		folder.ProjectID, folder.FolderPath, folder.UpdatedAt, folder.ID,
	)

	return err
}

// Delete soft deletes an excluded folder
func (r *ExcludedFolderRepository) Delete(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	query := `UPDATE excluded_folders SET deleted_at = CURRENT_TIMESTAMP WHERE id = ?`
	_, err := r.db.Exec(query, id)
	return err
}

// DeleteByProjectID soft deletes all excluded folders for a project
func (r *ExcludedFolderRepository) DeleteByProjectID(projectID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	query := `UPDATE excluded_folders SET deleted_at = CURRENT_TIMESTAMP WHERE project_id = ?`
	_, err := r.db.Exec(query, projectID)
	return err
}

// ExistsByProjectIDAndFolderPath checks if an excluded folder exists for a project
func (r *ExcludedFolderRepository) ExistsByProjectIDAndFolderPath(projectID, folderPath string) (bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	query := `
		SELECT COUNT(*) FROM excluded_folders 
		WHERE project_id = ? AND folder_path = ? AND deleted_at IS NULL
	`

	var count int
	err := r.db.QueryRow(query, projectID, folderPath).Scan(&count)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}
