package services

import (
	"fmt"
	"time"

	"github.com/alimgiray/gscope/internal/models"
	"github.com/alimgiray/gscope/internal/repositories"
)

type ProjectCollaboratorService struct {
	collaboratorRepo *repositories.ProjectCollaboratorRepository
	userRepo         *repositories.UserRepository
	projectRepo      *repositories.ProjectRepository
}

func NewProjectCollaboratorService(
	collaboratorRepo *repositories.ProjectCollaboratorRepository,
	userRepo *repositories.UserRepository,
	projectRepo *repositories.ProjectRepository,
) *ProjectCollaboratorService {
	return &ProjectCollaboratorService{
		collaboratorRepo: collaboratorRepo,
		userRepo:         userRepo,
		projectRepo:      projectRepo,
	}
}

// AddCollaborator adds a user as a collaborator to a project
func (s *ProjectCollaboratorService) AddCollaborator(projectID, username string) error {
	// Validate project exists
	_, err := s.projectRepo.GetByID(projectID)
	if err != nil {
		return fmt.Errorf("project not found: %w", err)
	}

	// Find user by username
	user, err := s.userRepo.GetByUsername(username)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	// Check if collaboration already exists
	exists, err := s.collaboratorRepo.ExistsByProjectAndUserID(projectID, user.ID.String())
	if err != nil {
		return fmt.Errorf("error checking existing collaboration: %w", err)
	}

	if exists {
		return fmt.Errorf("user is already a collaborator on this project")
	}

	// Create new collaboration
	collaborator := models.NewProjectCollaborator(projectID, user.ID.String())
	now := time.Now()
	collaborator.CreatedAt = now
	collaborator.UpdatedAt = now

	if err := collaborator.Validate(); err != nil {
		return fmt.Errorf("invalid collaborator data: %w", err)
	}

	return s.collaboratorRepo.Create(collaborator)
}

// RemoveCollaborator removes a user as a collaborator from a project
func (s *ProjectCollaboratorService) RemoveCollaborator(projectID, username string) error {
	// Find user by username
	user, err := s.userRepo.GetByUsername(username)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	// Check if collaboration exists
	exists, err := s.collaboratorRepo.ExistsByProjectAndUserID(projectID, user.ID.String())
	if err != nil {
		return fmt.Errorf("error checking collaboration: %w", err)
	}

	if !exists {
		return fmt.Errorf("user is not a collaborator on this project")
	}

	return s.collaboratorRepo.DeleteByProjectAndUserID(projectID, user.ID.String())
}

// GetProjectCollaborators retrieves all collaborators for a project
func (s *ProjectCollaboratorService) GetProjectCollaborators(projectID string) ([]*models.User, error) {
	collaborators, err := s.collaboratorRepo.GetByProjectID(projectID)
	if err != nil {
		return nil, fmt.Errorf("error fetching collaborators: %w", err)
	}

	var users []*models.User
	for _, collaborator := range collaborators {
		user, err := s.userRepo.GetByID(collaborator.UserID)
		if err != nil {
			// Skip users that can't be found
			continue
		}
		users = append(users, user)
	}

	return users, nil
}

// GetUserCollaborations retrieves all projects where a user is a collaborator
func (s *ProjectCollaboratorService) GetUserCollaborations(userID string) ([]*models.Project, error) {
	collaborations, err := s.collaboratorRepo.GetByUserID(userID)
	if err != nil {
		return nil, fmt.Errorf("error fetching collaborations: %w", err)
	}

	var projects []*models.Project
	for _, collaboration := range collaborations {
		project, err := s.projectRepo.GetByID(collaboration.ProjectID)
		if err != nil {
			// Skip projects that can't be found
			continue
		}
		projects = append(projects, project)
	}

	return projects, nil
}

// IsUserCollaborator checks if a user is a collaborator on a project
func (s *ProjectCollaboratorService) IsUserCollaborator(projectID, userID string) (bool, error) {
	return s.collaboratorRepo.ExistsByProjectAndUserID(projectID, userID)
}

// GetUserAccessibleProjects retrieves all projects a user can access (owned + collaborated)
func (s *ProjectCollaboratorService) GetUserAccessibleProjects(userID string) ([]*models.Project, error) {
	// Get owned projects
	ownedProjects, err := s.projectRepo.GetByOwnerID(userID)
	if err != nil {
		return nil, fmt.Errorf("error fetching owned projects: %w", err)
	}

	// Get collaborated projects
	collaboratedProjects, err := s.GetUserCollaborations(userID)
	if err != nil {
		return nil, fmt.Errorf("error fetching collaborated projects: %w", err)
	}

	// Combine and deduplicate
	projectMap := make(map[string]*models.Project)

	// Add owned projects
	for _, project := range ownedProjects {
		projectMap[project.ID.String()] = project
	}

	// Add collaborated projects
	for _, project := range collaboratedProjects {
		projectMap[project.ID.String()] = project
	}

	// Convert map back to slice
	var allProjects []*models.Project
	for _, project := range projectMap {
		allProjects = append(allProjects, project)
	}

	return allProjects, nil
}

// GetProjectAccessType determines if a user is an owner or collaborator of a project
func (s *ProjectCollaboratorService) GetProjectAccessType(projectID, userID string) (string, error) {
	// Check if user is the owner
	project, err := s.projectRepo.GetByID(projectID)
	if err != nil {
		return "", fmt.Errorf("project not found: %w", err)
	}

	if project.OwnerID.String() == userID {
		return "owner", nil
	}

	// Check if user is a collaborator
	isCollaborator, err := s.IsUserCollaborator(projectID, userID)
	if err != nil {
		return "", fmt.Errorf("error checking collaboration: %w", err)
	}

	if isCollaborator {
		return "collaborator", nil
	}

	return "none", nil
}
