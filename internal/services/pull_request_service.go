package services

import (
	"github.com/alimgiray/gscope/internal/models"
	"github.com/alimgiray/gscope/internal/repositories"
)

type PullRequestService struct {
	pullRequestRepo *repositories.PullRequestRepository
}

func NewPullRequestService(pullRequestRepo *repositories.PullRequestRepository) *PullRequestService {
	return &PullRequestService{
		pullRequestRepo: pullRequestRepo,
	}
}

func (s *PullRequestService) CreatePullRequest(pr *models.PullRequest) error {
	return s.pullRequestRepo.Create(pr)
}

func (s *PullRequestService) GetPullRequestByID(id string) (*models.PullRequest, error) {
	return s.pullRequestRepo.GetByID(id)
}

func (s *PullRequestService) GetPullRequestsByRepositoryID(repositoryID string) ([]*models.PullRequest, error) {
	return s.pullRequestRepo.GetByRepositoryID(repositoryID)
}

func (s *PullRequestService) GetPullRequestByGithubPRID(githubPRID int) (*models.PullRequest, error) {
	return s.pullRequestRepo.GetByGithubPRID(githubPRID)
}

func (s *PullRequestService) UpdatePullRequest(pr *models.PullRequest) error {
	return s.pullRequestRepo.Update(pr)
}

func (s *PullRequestService) DeletePullRequest(id string) error {
	return s.pullRequestRepo.Delete(id)
}

func (s *PullRequestService) UpsertPullRequest(pr *models.PullRequest) error {
	return s.pullRequestRepo.Upsert(pr)
}
