package services

import (
	"time"

	"github.com/alimgiray/gscope/internal/models"
	"github.com/alimgiray/gscope/internal/repositories"
)

type ExcludedFolderService struct {
	excludedFolderRepo *repositories.ExcludedFolderRepository
}

func NewExcludedFolderService(excludedFolderRepo *repositories.ExcludedFolderRepository) *ExcludedFolderService {
	return &ExcludedFolderService{
		excludedFolderRepo: excludedFolderRepo,
	}
}

// CreateExcludedFolder creates a new excluded folder
func (s *ExcludedFolderService) CreateExcludedFolder(projectID, folderPath string) (*models.ExcludedFolder, error) {
	// Validate input
	if projectID == "" {
		return nil, &models.ValidationError{Message: "Project ID is required"}
	}
	if folderPath == "" {
		return nil, &models.ValidationError{Message: "Folder path is required"}
	}

	// Check if folder path already exists for this project
	exists, err := s.excludedFolderRepo.ExistsByProjectIDAndFolderPath(projectID, folderPath)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, &models.ValidationError{Message: "Folder path already exists for this project"}
	}

	// Create new excluded folder
	folder := models.NewExcludedFolder(projectID, folderPath)
	if err := folder.Validate(); err != nil {
		return nil, err
	}

	// Save to database
	if err := s.excludedFolderRepo.Create(folder); err != nil {
		return nil, err
	}

	return folder, nil
}

// GetExcludedFoldersByProjectID retrieves all excluded folders for a project
func (s *ExcludedFolderService) GetExcludedFoldersByProjectID(projectID string) ([]*models.ExcludedFolder, error) {
	if projectID == "" {
		return nil, &models.ValidationError{Message: "Project ID is required"}
	}

	return s.excludedFolderRepo.GetByProjectID(projectID)
}

// UpdateExcludedFolder updates an excluded folder
func (s *ExcludedFolderService) UpdateExcludedFolder(id, folderPath string) (*models.ExcludedFolder, error) {
	if id == "" {
		return nil, &models.ValidationError{Message: "ID is required"}
	}
	if folderPath == "" {
		return nil, &models.ValidationError{Message: "Folder path is required"}
	}

	// Get existing folder
	folder, err := s.excludedFolderRepo.GetByID(id)
	if err != nil {
		return nil, err
	}

	// Check if new folder path already exists for this project
	exists, err := s.excludedFolderRepo.ExistsByProjectIDAndFolderPath(folder.ProjectID, folderPath)
	if err != nil {
		return nil, err
	}
	if exists && folder.FolderPath != folderPath {
		return nil, &models.ValidationError{Message: "Folder path already exists for this project"}
	}

	// Update folder path
	folder.FolderPath = folderPath
	folder.UpdatedAt = time.Now()

	if err := folder.Validate(); err != nil {
		return nil, err
	}

	// Save to database
	if err := s.excludedFolderRepo.Update(folder); err != nil {
		return nil, err
	}

	return folder, nil
}

// DeleteExcludedFolder deletes an excluded folder
func (s *ExcludedFolderService) DeleteExcludedFolder(id string) error {
	if id == "" {
		return &models.ValidationError{Message: "ID is required"}
	}

	return s.excludedFolderRepo.Delete(id)
}

// DeleteExcludedFoldersByProjectID deletes all excluded folders for a project
func (s *ExcludedFolderService) DeleteExcludedFoldersByProjectID(projectID string) error {
	if projectID == "" {
		return &models.ValidationError{Message: "Project ID is required"}
	}

	return s.excludedFolderRepo.DeleteByProjectID(projectID)
}

// IsExcludedFolder checks if a file path should be excluded based on excluded folders
func (s *ExcludedFolderService) IsExcludedFolder(filePath string, excludedFolders []*models.ExcludedFolder) bool {
	for _, folder := range excludedFolders {
		// Check if file path contains the excluded folder path
		if contains(filePath, folder.FolderPath) {
			return true
		}
	}
	return false
}

// contains checks if a string contains another string
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr ||
		(len(s) > len(substr) && (s[:len(substr)] == substr && s[len(substr)] == '/') ||
			s[len(s)-len(substr):] == substr ||
			(len(s) > len(substr)+1 && s[len(s)-len(substr)-1:] == "/"+substr)))
}
