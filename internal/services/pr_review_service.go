package services

import (
	"github.com/alimgiray/gscope/internal/models"
	"github.com/alimgiray/gscope/internal/repositories"
)

type PRReviewService struct {
	prReviewRepo *repositories.PRReviewRepository
}

func NewPRReviewService(prReviewRepo *repositories.PRReviewRepository) *PRReviewService {
	return &PRReviewService{
		prReviewRepo: prReviewRepo,
	}
}

func (s *PRReviewService) CreatePRReview(review *models.PRReview) error {
	return s.prReviewRepo.Create(review)
}

func (s *PRReviewService) GetPRReviewByID(id string) (*models.PRReview, error) {
	return s.prReviewRepo.GetByID(id)
}

func (s *PRReviewService) GetPRReviewsByPullRequestID(pullRequestID string) ([]*models.PRReview, error) {
	return s.prReviewRepo.GetByPullRequestID(pullRequestID)
}

func (s *PRReviewService) GetPRReviewByGithubReviewID(githubReviewID int) (*models.PRReview, error) {
	return s.prReviewRepo.GetByGithubReviewID(githubReviewID)
}

func (s *PRReviewService) UpdatePRReview(review *models.PRReview) error {
	return s.prReviewRepo.Update(review)
}

func (s *PRReviewService) DeletePRReview(id string) error {
	return s.prReviewRepo.Delete(id)
}

func (s *PRReviewService) UpsertPRReview(review *models.PRReview) error {
	return s.prReviewRepo.Upsert(review)
}
