package services

import (
	"github.com/alimgiray/gscope/internal/models"
	"github.com/alimgiray/gscope/internal/repositories"
)

type GitHubPersonEmailService struct {
	githubPersonEmailRepo *repositories.GitHubPersonEmailRepository
}

func NewGitHubPersonEmailService(githubPersonEmailRepo *repositories.GitHubPersonEmailRepository) *GitHubPersonEmailService {
	return &GitHubPersonEmailService{
		githubPersonEmailRepo: githubPersonEmailRepo,
	}
}

// CreateGitHubPersonEmail creates a new GitHub person email association
func (s *GitHubPersonEmailService) CreateGitHubPersonEmail(projectID, githubPersonID, personID string) (*models.GitHubPersonEmail, error) {
	// Check if person is already associated with another GitHub person in this project
	existing, err := s.githubPersonEmailRepo.GetByPersonID(projectID, personID)
	if err == nil && existing != nil {
		return nil, &models.ValidationError{Message: "Person is already associated with another GitHub person"}
	}

	// Check if this GitHub person already has an email association in this project
	existingAssociations, err := s.githubPersonEmailRepo.GetByGitHubPersonID(projectID, githubPersonID)
	if err == nil && len(existingAssociations) > 0 {
		return nil, &models.ValidationError{Message: "GitHub person already has an email association"}
	}

	gpe := models.NewGitHubPersonEmail(projectID, githubPersonID, personID)
	err = s.githubPersonEmailRepo.Create(gpe)
	if err != nil {
		return nil, err
	}
	return gpe, nil
}

// GetGitHubPersonEmailByID retrieves a GitHub person email association by ID
func (s *GitHubPersonEmailService) GetGitHubPersonEmailByID(id string) (*models.GitHubPersonEmail, error) {
	return s.githubPersonEmailRepo.GetByID(id)
}

// GetGitHubPersonEmailsByProjectID retrieves all GitHub person email associations for a project
func (s *GitHubPersonEmailService) GetGitHubPersonEmailsByProjectID(projectID string) ([]*models.GitHubPersonEmail, error) {
	return s.githubPersonEmailRepo.GetByProjectID(projectID)
}

// GetGitHubPersonEmailByPersonID retrieves a GitHub person email association by person ID and project
func (s *GitHubPersonEmailService) GetGitHubPersonEmailByPersonID(projectID, personID string) (*models.GitHubPersonEmail, error) {
	return s.githubPersonEmailRepo.GetByPersonID(projectID, personID)
}

// GetGitHubPersonEmailsByGitHubPersonID retrieves all email associations for a GitHub person in a project
func (s *GitHubPersonEmailService) GetGitHubPersonEmailsByGitHubPersonID(projectID, githubPersonID string) ([]*models.GitHubPersonEmail, error) {
	return s.githubPersonEmailRepo.GetByGitHubPersonID(projectID, githubPersonID)
}

// UpdateGitHubPersonEmail updates a GitHub person email association
func (s *GitHubPersonEmailService) UpdateGitHubPersonEmail(gpe *models.GitHubPersonEmail) error {
	return s.githubPersonEmailRepo.Update(gpe)
}

// DeleteGitHubPersonEmail deletes a GitHub person email association by ID
func (s *GitHubPersonEmailService) DeleteGitHubPersonEmail(id string) error {
	return s.githubPersonEmailRepo.Delete(id)
}

// DeleteGitHubPersonEmailsByProjectID deletes all GitHub person email associations for a project
func (s *GitHubPersonEmailService) DeleteGitHubPersonEmailsByProjectID(projectID string) error {
	return s.githubPersonEmailRepo.DeleteByProjectID(projectID)
}

// DeleteGitHubPersonEmailByPersonID deletes a GitHub person email association by person ID and project
func (s *GitHubPersonEmailService) DeleteGitHubPersonEmailByPersonID(projectID, personID string) error {
	return s.githubPersonEmailRepo.DeleteByPersonID(projectID, personID)
}
