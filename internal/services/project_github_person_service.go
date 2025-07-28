package services

import (
	"github.com/alimgiray/gscope/internal/models"
	"github.com/alimgiray/gscope/internal/repositories"
)

type ProjectGithubPersonService struct {
	projectGithubPersonRepo *repositories.ProjectGithubPersonRepository
}

func NewProjectGithubPersonService(projectGithubPersonRepo *repositories.ProjectGithubPersonRepository) *ProjectGithubPersonService {
	return &ProjectGithubPersonService{
		projectGithubPersonRepo: projectGithubPersonRepo,
	}
}

// CreateProjectGithubPerson creates a new project-github person relationship
func (s *ProjectGithubPersonService) CreateProjectGithubPerson(projectID, githubPersonID, sourceType string) error {
	projectGithubPerson := models.NewProjectGithubPerson(projectID, githubPersonID, sourceType)
	return s.projectGithubPersonRepo.Upsert(projectGithubPerson)
}

// GetProjectGithubPeopleByProjectID retrieves all project-github person relationships for a project
func (s *ProjectGithubPersonService) GetProjectGithubPeopleByProjectID(projectID string) ([]*models.ProjectGithubPerson, error) {
	return s.projectGithubPersonRepo.GetByProjectID(projectID)
}

// GetProjectGithubPeopleByGithubPersonID retrieves all project-github person relationships for a GitHub person
func (s *ProjectGithubPersonService) GetProjectGithubPeopleByGithubPersonID(githubPersonID string) ([]*models.ProjectGithubPerson, error) {
	return s.projectGithubPersonRepo.GetByGithubPersonID(githubPersonID)
}

// DeleteProjectGithubPerson deletes a project-github person relationship
func (s *ProjectGithubPersonService) DeleteProjectGithubPerson(id string) error {
	return s.projectGithubPersonRepo.Delete(id)
}

// DeleteByProjectAndGithubPerson deletes a project-github person relationship by project and GitHub person IDs
func (s *ProjectGithubPersonService) DeleteByProjectAndGithubPerson(projectID, githubPersonID string) error {
	return s.projectGithubPersonRepo.DeleteByProjectAndGithubPerson(projectID, githubPersonID)
}
