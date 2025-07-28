package services

import (
	"time"

	"github.com/alimgiray/gscope/internal/models"
	"github.com/alimgiray/gscope/internal/repositories"
)

type PeopleStatisticsService struct {
	peopleStatsRepo       *repositories.PeopleStatisticsRepository
	commitRepo            *repositories.CommitRepository
	commitFileRepo        *repositories.CommitFileRepository
	pullRequestRepo       *repositories.PullRequestRepository
	prReviewRepo          *repositories.PRReviewRepository
	githubPersonRepo      *repositories.GithubPersonRepository
	emailMergeRepo        *repositories.EmailMergeRepository
	githubPersonEmailRepo *repositories.GitHubPersonEmailRepository
	personRepo            *repositories.PersonRepository
	scoreSettingsRepo     *repositories.ScoreSettingsRepository
	excludedExtRepo       *repositories.ExcludedExtensionRepository
}

func NewPeopleStatisticsService(
	peopleStatsRepo *repositories.PeopleStatisticsRepository,
	commitRepo *repositories.CommitRepository,
	commitFileRepo *repositories.CommitFileRepository,
	pullRequestRepo *repositories.PullRequestRepository,
	prReviewRepo *repositories.PRReviewRepository,
	githubPersonRepo *repositories.GithubPersonRepository,
	emailMergeRepo *repositories.EmailMergeRepository,
	githubPersonEmailRepo *repositories.GitHubPersonEmailRepository,
	personRepo *repositories.PersonRepository,
	scoreSettingsRepo *repositories.ScoreSettingsRepository,
	excludedExtRepo *repositories.ExcludedExtensionRepository,
) *PeopleStatisticsService {
	return &PeopleStatisticsService{
		peopleStatsRepo:       peopleStatsRepo,
		commitRepo:            commitRepo,
		commitFileRepo:        commitFileRepo,
		pullRequestRepo:       pullRequestRepo,
		prReviewRepo:          prReviewRepo,
		githubPersonRepo:      githubPersonRepo,
		emailMergeRepo:        emailMergeRepo,
		githubPersonEmailRepo: githubPersonEmailRepo,
		personRepo:            personRepo,
		scoreSettingsRepo:     scoreSettingsRepo,
		excludedExtRepo:       excludedExtRepo,
	}
}

// CalculateStatisticsForRepository calculates daily statistics for a specific repository
func (s *PeopleStatisticsService) CalculateStatisticsForRepository(projectID, projectRepositoryID, githubRepositoryID string) error {
	// Get score settings for the project
	scoreSettings, err := s.scoreSettingsRepo.GetByProjectID(projectID)
	if err != nil {
		return err
	}

	// Get excluded extensions for the project
	excludedExtensions, err := s.excludedExtRepo.GetByProjectID(projectID)
	if err != nil {
		return err
	}

	// Get email merges for the project
	emailMerges, err := s.emailMergeRepo.GetMergedEmailsForProject(projectID)
	if err != nil {
		return err
	}

	// Get GitHub person to email associations for the project
	emailAssociations, err := s.githubPersonEmailRepo.GetByProjectID(projectID)
	if err != nil {
		return err
	}

	// Create a map of GitHub person ID to email
	personEmailMap := make(map[string]string)
	for _, assoc := range emailAssociations {
		person, err := s.personRepo.GetByID(assoc.PersonID)
		if err != nil {
			continue
		}
		personEmailMap[assoc.GitHubPersonID] = person.PrimaryEmail
	}

	// Get the date range from actual commits to ensure we calculate for the correct period
	minCommitDate, _, err := s.commitRepo.GetDateRangeByRepositoryID(githubRepositoryID)
	var startDate, endDate time.Time
	if err != nil {
		// Fallback to 1 year ago if we can't get commit dates
		startDate = time.Now().AddDate(-1, 0, 0)
		endDate = time.Now()
	} else {
		// Start from the first commit date and go to today
		startDate = minCommitDate
		endDate = time.Now()
	}

	// Delete existing statistics for this repository to ensure fresh calculation
	if err := s.peopleStatsRepo.DeleteByRepositoryID(projectRepositoryID); err != nil {
		return err
	}

	// Calculate statistics for each day from beginning to end
	for date := startDate; !date.After(endDate); date = date.AddDate(0, 0, 1) {
		if err := s.calculateDailyStatistics(projectID, projectRepositoryID, githubRepositoryID, date, scoreSettings, excludedExtensions, emailMerges, personEmailMap); err != nil {
			return err
		}
	}

	return nil
}

// calculateDailyStatistics calculates statistics for a specific date
func (s *PeopleStatisticsService) calculateDailyStatistics(
	projectID, projectRepositoryID, githubRepositoryID string,
	date time.Time,
	scoreSettings *models.ScoreSettings,
	excludedExtensions []*models.ExcludedExtension,
	emailMerges map[string]string,
	personEmailMap map[string]string,
) error {

	// Get all GitHub people for this project
	githubPeople, err := s.githubPersonRepo.GetByProjectID(projectID)
	if err != nil {
		return err
	}

	// Create a map of excluded extensions for quick lookup
	excludedExtMap := make(map[string]bool)
	for _, ext := range excludedExtensions {
		excludedExtMap[ext.Extension] = true
	}

	// Calculate statistics for each person
	for _, person := range githubPeople {
		stats := s.calculatePersonDailyStats(
			projectID, projectRepositoryID, githubRepositoryID, person.ID, date,
			scoreSettings, excludedExtMap, emailMerges, personEmailMap,
		)

		if stats != nil && stats.Score > 0 {
			// Only insert if score is greater than 0
			if err := s.peopleStatsRepo.Create(stats); err != nil {
				return err
			}
		}
	}

	return nil
}

// calculatePersonDailyStats calculates daily statistics for a specific person
func (s *PeopleStatisticsService) calculatePersonDailyStats(
	projectID, projectRepositoryID, githubRepositoryID, githubPersonID string,
	date time.Time,
	scoreSettings *models.ScoreSettings,
	excludedExtMap map[string]bool,
	emailMerges map[string]string,
	personEmailMap map[string]string,
) *models.PeopleStatistics {

	// Get the person's associated email
	personEmail := personEmailMap[githubPersonID]
	if personEmail == "" {
		// No email association, skip this person
		return nil
	}

	// Apply email merges
	mergedEmail := s.getMergedEmail(personEmail, emailMerges)

	// Calculate commit statistics
	commits, additions, deletions := s.calculateCommitStats(githubRepositoryID, mergedEmail, date, excludedExtMap)

	// Calculate PR statistics
	pullRequests := s.calculatePRStats(githubRepositoryID, githubPersonID, date)

	// Calculate comment statistics
	comments := s.calculateCommentStats(githubRepositoryID, githubPersonID, date)

	// Create statistics record
	stats := models.NewPeopleStatistics(projectID, projectRepositoryID, githubPersonID, date)
	stats.Commits = commits
	stats.Additions = additions
	stats.Deletions = deletions
	stats.PullRequests = pullRequests
	stats.Comments = comments

	// Calculate score
	stats.CalculateScore(scoreSettings)

	return stats
}

// getMergedEmail returns the target email if the given email is merged, otherwise returns the original email
func (s *PeopleStatisticsService) getMergedEmail(email string, emailMerges map[string]string) string {
	if targetEmail, exists := emailMerges[email]; exists {
		return targetEmail
	}
	return email
}

// calculateCommitStats calculates commit-related statistics for a person on a specific date
func (s *PeopleStatisticsService) calculateCommitStats(repositoryID, email string, date time.Time, excludedExtMap map[string]bool) (int, int, int) {
	// Get all commits for this repository
	commits, err := s.commitRepo.GetByRepositoryID(repositoryID)
	if err != nil {
		return 0, 0, 0
	}

	totalCommits := 0
	totalAdditions := 0
	totalDeletions := 0

	// Filter commits by author email and date
	for _, commit := range commits {
		if commit.AuthorEmail != nil && *commit.AuthorEmail == email {
			// Check if commit is on the specified date (compare year, month, day)
			commitYear, commitMonth, commitDay := commit.CommitDate.Date()
			dateYear, dateMonth, dateDay := date.Date()

			if commitYear == dateYear && commitMonth == dateMonth && commitDay == dateDay {
				// Get commit files for this commit
				commitFiles, err := s.commitFileRepo.GetByCommitID(commit.ID)
				if err != nil {
					// If we can't get commit files, skip this commit
					continue
				}

				// Check if any files in this commit are not excluded
				hasNonExcludedFiles := false
				commitAdditions := 0
				commitDeletions := 0

				for _, commitFile := range commitFiles {
					// Check if this file has an excluded extension
					if !s.isExcludedExtension(commitFile.Filename, excludedExtMap) {
						hasNonExcludedFiles = true
						commitAdditions += commitFile.Additions
						commitDeletions += commitFile.Deletions
					}
				}

				// Only count this commit if it has non-excluded files
				if hasNonExcludedFiles {
					totalCommits++
					totalAdditions += commitAdditions
					totalDeletions += commitDeletions
				}
			}
		}
	}

	return totalCommits, totalAdditions, totalDeletions
}

// calculatePRStats calculates pull request statistics for a person on a specific date
func (s *PeopleStatisticsService) calculatePRStats(repositoryID, githubPersonID string, date time.Time) int {
	// Get all pull requests for this repository
	pullRequests, err := s.pullRequestRepo.GetByRepositoryID(repositoryID)
	if err != nil {
		return 0
	}

	count := 0
	for _, pr := range pullRequests {
		// Check if PR was created by this person on the specified date
		if pr.GithubCreatedAt != nil {
			prYear, prMonth, prDay := pr.GithubCreatedAt.Date()
			dateYear, dateMonth, dateDay := date.Date()

			if prYear == dateYear && prMonth == dateMonth && prDay == dateDay {
				// Check if the PR user matches the github person ID
				// This would require parsing the user JSON field
				// For now, we'll count all PRs on that date (simplified)
				count++
			}
		}
	}

	return count
}

// calculateCommentStats calculates comment statistics for a person on a specific date
func (s *PeopleStatisticsService) calculateCommentStats(repositoryID, githubPersonID string, date time.Time) int {
	// For now, we'll return 0 since we need to implement a proper method to get reviews by repository
	// This would require adding a GetByRepositoryID method to PRReviewRepository
	// or using a different approach to get reviews for a specific repository
	// TODO: Implement proper review statistics calculation
	return 0
}

// isExcludedExtension checks if a file has an excluded extension
func (s *PeopleStatisticsService) isExcludedExtension(filename string, excludedExtMap map[string]bool) bool {
	// Extract file extension
	for i := len(filename) - 1; i >= 0; i-- {
		if filename[i] == '.' {
			ext := filename[i+1:]
			return excludedExtMap[ext]
		}
	}
	return false
}

// GetStatisticsByProject retrieves all statistics for a project
func (s *PeopleStatisticsService) GetStatisticsByProject(projectID string) ([]*models.PeopleStatistics, error) {
	return s.peopleStatsRepo.GetByProjectID(projectID)
}

// GetStatisticsByRepository retrieves all statistics for a repository
func (s *PeopleStatisticsService) GetStatisticsByRepository(repositoryID string) ([]*models.PeopleStatistics, error) {
	return s.peopleStatsRepo.GetByRepositoryID(repositoryID)
}

// GetStatisticsByPerson retrieves all statistics for a GitHub person
func (s *PeopleStatisticsService) GetStatisticsByPerson(githubPersonID string) ([]*models.PeopleStatistics, error) {
	return s.peopleStatsRepo.GetByGithubPersonID(githubPersonID)
}

// GetStatisticsByDateRange retrieves statistics within a date range
func (s *PeopleStatisticsService) GetStatisticsByDateRange(projectID string, startDate, endDate time.Time) ([]*models.PeopleStatistics, error) {
	return s.peopleStatsRepo.GetByDateRange(projectID, startDate, endDate)
}

// DeleteStatisticsByProject deletes all statistics for a project
func (s *PeopleStatisticsService) DeleteStatisticsByProject(projectID string) error {
	return s.peopleStatsRepo.DeleteByProjectID(projectID)
}

// DeleteStatisticsByRepository deletes all statistics for a repository
func (s *PeopleStatisticsService) DeleteStatisticsByRepository(repositoryID string) error {
	return s.peopleStatsRepo.DeleteByRepositoryID(repositoryID)
}
