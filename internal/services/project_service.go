package services

import (
	"errors"
	"strings"

	"github.com/alimgiray/gscope/internal/models"
	"github.com/alimgiray/gscope/internal/repositories"
	"github.com/google/uuid"
)

type ProjectService struct {
	projectRepo          *repositories.ProjectRepository
	scoreSettingsService *ScoreSettingsService
}

func NewProjectService(projectRepo *repositories.ProjectRepository, scoreSettingsService *ScoreSettingsService) *ProjectService {
	return &ProjectService{
		projectRepo:          projectRepo,
		scoreSettingsService: scoreSettingsService,
	}
}

// CreateProject creates a new project
func (s *ProjectService) CreateProject(project *models.Project) error {
	// Validate project
	if err := project.Validate(); err != nil {
		return err
	}

	// Trim whitespace from name
	project.Name = strings.TrimSpace(project.Name)
	if project.Name == "" {
		return models.ErrProjectNameRequired
	}

	// Validate owner ID
	if project.OwnerID == uuid.Nil {
		return errors.New("owner ID is required")
	}

	// Create the project
	if err := s.projectRepo.Create(project); err != nil {
		return err
	}

	// Create default score settings for the project
	return s.scoreSettingsService.CreateScoreSettings(project.ID.String())
}

// GetProjectByID retrieves a project by ID
func (s *ProjectService) GetProjectByID(id string) (*models.Project, error) {
	if id == "" {
		return nil, errors.New("project ID is required")
	}

	// Validate UUID format
	if _, err := uuid.Parse(id); err != nil {
		return nil, errors.New("invalid project ID format")
	}

	return s.projectRepo.GetByID(id)
}

// GetProjectsByOwnerID retrieves all projects for an owner
func (s *ProjectService) GetProjectsByOwnerID(ownerID string) ([]*models.Project, error) {
	if ownerID == "" {
		return nil, errors.New("owner ID is required")
	}

	// Validate UUID format
	if _, err := uuid.Parse(ownerID); err != nil {
		return nil, errors.New("invalid owner ID format")
	}

	return s.projectRepo.GetByOwnerID(ownerID)
}

// UpdateProject updates a project
func (s *ProjectService) UpdateProject(project *models.Project) error {
	// Validate project
	if err := project.Validate(); err != nil {
		return err
	}

	// Trim whitespace from name
	project.Name = strings.TrimSpace(project.Name)
	if project.Name == "" {
		return models.ErrProjectNameRequired
	}

	// Validate project ID
	if project.ID == uuid.Nil {
		return errors.New("project ID is required")
	}

	// Check if project exists
	_, err := s.projectRepo.GetByID(project.ID.String())
	if err != nil {
		return err
	}

	return s.projectRepo.Update(project)
}

// DeleteProject performs a soft delete of a project
func (s *ProjectService) DeleteProject(id string) error {
	if id == "" {
		return errors.New("project ID is required")
	}

	// Validate UUID format
	if _, err := uuid.Parse(id); err != nil {
		return errors.New("invalid project ID format")
	}

	// Check if project exists
	_, err := s.projectRepo.GetByID(id)
	if err != nil {
		return err
	}

	return s.projectRepo.Delete(id)
}
