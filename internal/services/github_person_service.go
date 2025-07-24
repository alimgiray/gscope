package services

import (
	"github.com/alimgiray/gscope/internal/models"
	"github.com/alimgiray/gscope/internal/repositories"
)

type GithubPersonService struct {
	githubPersonRepo *repositories.GithubPersonRepository
}

func NewGithubPersonService(githubPersonRepo *repositories.GithubPersonRepository) *GithubPersonService {
	return &GithubPersonService{
		githubPersonRepo: githubPersonRepo,
	}
}

func (s *GithubPersonService) CreateGithubPerson(person *models.GithubPerson) error {
	return s.githubPersonRepo.Create(person)
}

func (s *GithubPersonService) GetGithubPersonByID(id string) (*models.GithubPerson, error) {
	return s.githubPersonRepo.GetByID(id)
}

func (s *GithubPersonService) GetGithubPersonByGithubUserID(githubUserID int) (*models.GithubPerson, error) {
	return s.githubPersonRepo.GetByGithubUserID(githubUserID)
}

func (s *GithubPersonService) GetGithubPersonByUsername(username string) (*models.GithubPerson, error) {
	return s.githubPersonRepo.GetByUsername(username)
}

func (s *GithubPersonService) UpdateGithubPerson(person *models.GithubPerson) error {
	return s.githubPersonRepo.Update(person)
}

func (s *GithubPersonService) DeleteGithubPerson(id string) error {
	return s.githubPersonRepo.Delete(id)
}

func (s *GithubPersonService) UpsertGithubPerson(person *models.GithubPerson) error {
	return s.githubPersonRepo.Upsert(person)
}
