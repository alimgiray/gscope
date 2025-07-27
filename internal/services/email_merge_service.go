package services

import (
	"github.com/alimgiray/gscope/internal/models"
	"github.com/alimgiray/gscope/internal/repositories"
)

type EmailMergeService struct {
	emailMergeRepo *repositories.EmailMergeRepository
}

func NewEmailMergeService(emailMergeRepo *repositories.EmailMergeRepository) *EmailMergeService {
	return &EmailMergeService{
		emailMergeRepo: emailMergeRepo,
	}
}

// CreateEmailMerge creates a new email merge
func (s *EmailMergeService) CreateEmailMerge(projectID, sourceEmail, targetEmail string) error {
	merge := models.NewEmailMerge(projectID, sourceEmail, targetEmail)
	return s.emailMergeRepo.Create(merge)
}

// GetEmailMergeByID retrieves an email merge by ID
func (s *EmailMergeService) GetEmailMergeByID(id string) (*models.EmailMerge, error) {
	return s.emailMergeRepo.GetByID(id)
}

// GetEmailMergesByProjectID retrieves all email merges for a project
func (s *EmailMergeService) GetEmailMergesByProjectID(projectID string) ([]*models.EmailMerge, error) {
	return s.emailMergeRepo.GetByProjectID(projectID)
}

// GetMergedEmailsForProject returns a map of source_email -> target_email for a project
func (s *EmailMergeService) GetMergedEmailsForProject(projectID string) (map[string]string, error) {
	return s.emailMergeRepo.GetMergedEmailsForProject(projectID)
}

// IsEmailMerged checks if an email has been merged in a project
func (s *EmailMergeService) IsEmailMerged(projectID, email string) (bool, string, error) {
	return s.emailMergeRepo.IsEmailMerged(projectID, email)
}

// UpdateEmailMerge updates an existing email merge
func (s *EmailMergeService) UpdateEmailMerge(merge *models.EmailMerge) error {
	return s.emailMergeRepo.Update(merge)
}

// DeleteEmailMerge deletes an email merge by ID
func (s *EmailMergeService) DeleteEmailMerge(id string) error {
	return s.emailMergeRepo.Delete(id)
}

// DeleteEmailMergesByProjectID deletes all email merges for a project
func (s *EmailMergeService) DeleteEmailMergesByProjectID(projectID string) error {
	return s.emailMergeRepo.DeleteByProjectID(projectID)
}

// DeleteEmailMergesByTargetEmail deletes all email merges where target_email matches
func (s *EmailMergeService) DeleteEmailMergesByTargetEmail(projectID, targetEmail string) error {
	return s.emailMergeRepo.DeleteByTargetEmail(projectID, targetEmail)
}

// DeleteEmailMergeBySourceAndTarget deletes a specific email merge
func (s *EmailMergeService) DeleteEmailMergeBySourceAndTarget(projectID, sourceEmail, targetEmail string) error {
	return s.emailMergeRepo.DeleteBySourceAndTarget(projectID, sourceEmail, targetEmail)
}
