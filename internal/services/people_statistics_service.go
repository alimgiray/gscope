package services

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/alimgiray/gscope/internal/models"
	"github.com/alimgiray/gscope/internal/repositories"
	"github.com/google/uuid"
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
	excludedFolderRepo    *repositories.ExcludedFolderRepository
	githubRepoRepo        *repositories.GitHubRepositoryRepository
	projectRepositoryRepo *repositories.ProjectRepositoryRepository
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
	excludedFolderRepo *repositories.ExcludedFolderRepository,
	githubRepoRepo *repositories.GitHubRepositoryRepository,
	projectRepositoryRepo *repositories.ProjectRepositoryRepository,
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
		excludedFolderRepo:    excludedFolderRepo,
		githubRepoRepo:        githubRepoRepo,
		projectRepositoryRepo: projectRepositoryRepo,
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

	// Get excluded folders for the project
	excludedFolders, err := s.excludedFolderRepo.GetByProjectID(projectID)
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

	// OPTIMIZATION: Pre-load all data for the repository to avoid N+1 queries
	allCommits, err := s.commitRepo.GetByRepositoryID(githubRepositoryID)
	if err != nil {
		return err
	}

	allPullRequests, err := s.pullRequestRepo.GetByRepositoryID(githubRepositoryID)
	if err != nil {
		return err
	}

	allPRReviews, err := s.prReviewRepo.GetByRepositoryID(githubRepositoryID)
	if err != nil {
		return err
	}

	// OPTIMIZATION: Find actual activity dates to avoid processing empty days
	activityDates := s.findActivityDates(allCommits, allPullRequests, allPRReviews, startDate, endDate)

	// Pre-load all commit files for the repository
	allCommitFiles := make(map[string][]*models.CommitFile)
	for _, commit := range allCommits {
		commitFiles, err := s.commitFileRepo.GetByCommitID(commit.ID)
		if err != nil {
			continue
		}
		allCommitFiles[commit.ID] = commitFiles
	}

	// Create a map of excluded extensions for quick lookup
	excludedExtMap := make(map[string]bool)
	for _, ext := range excludedExtensions {
		excludedExtMap[ext.Extension] = true
	}

	// Get all GitHub people for this project
	githubPeople, err := s.githubPersonRepo.GetByProjectID(projectID)
	if err != nil {
		return err
	}

	// OPTIMIZATION: Only process days that have actual activity
	for _, date := range activityDates {
		if err := s.calculateDailyStatisticsOptimized(
			projectID, projectRepositoryID, githubRepositoryID, date,
			scoreSettings, excludedExtMap, excludedFolders, emailMerges, personEmailMap,
			allCommits, allPullRequests, allPRReviews, allCommitFiles, githubPeople,
		); err != nil {
			return err
		}
	}

	return nil
}

// findActivityDates finds all dates that have actual activity (commits, PRs, or reviews)
func (s *PeopleStatisticsService) findActivityDates(
	allCommits []*models.Commit,
	allPullRequests []*models.PullRequest,
	allPRReviews []*models.PRReview,
	startDate, endDate time.Time,
) []time.Time {
	activityMap := make(map[string]bool)

	// Add commit dates
	for _, commit := range allCommits {
		commitDate := commit.CommitDate.Format("2006-01-02")
		if commit.CommitDate.After(startDate) && commit.CommitDate.Before(endDate.AddDate(0, 0, 1)) {
			activityMap[commitDate] = true
		}
	}

	// Add PR creation dates
	for _, pr := range allPullRequests {
		if pr.GithubCreatedAt != nil {
			prDate := pr.GithubCreatedAt.Format("2006-01-02")
			if pr.GithubCreatedAt.After(startDate) && pr.GithubCreatedAt.Before(endDate.AddDate(0, 0, 1)) {
				activityMap[prDate] = true
			}
		}
	}

	// Add review creation dates
	for _, review := range allPRReviews {
		if review.GithubCreatedAt != nil {
			reviewDate := review.GithubCreatedAt.Format("2006-01-02")
			if review.GithubCreatedAt.After(startDate) && review.GithubCreatedAt.Before(endDate.AddDate(0, 0, 1)) {
				activityMap[reviewDate] = true
			}
		}
	}

	// Convert map to sorted slice
	var activityDates []time.Time
	for dateStr := range activityMap {
		if date, err := time.Parse("2006-01-02", dateStr); err == nil {
			activityDates = append(activityDates, date)
		}
	}

	// Sort dates
	sort.Slice(activityDates, func(i, j int) bool {
		return activityDates[i].Before(activityDates[j])
	})

	return activityDates
}

// calculateDailyStatisticsOptimized calculates statistics for a specific date using pre-loaded data
func (s *PeopleStatisticsService) calculateDailyStatisticsOptimized(
	projectID, projectRepositoryID, githubRepositoryID string,
	date time.Time,
	scoreSettings *models.ScoreSettings,
	excludedExtMap map[string]bool,
	excludedFolders []*models.ExcludedFolder,
	emailMerges map[string]string,
	personEmailMap map[string]string,
	allCommits []*models.Commit,
	allPullRequests []*models.PullRequest,
	allPRReviews []*models.PRReview,
	allCommitFiles map[string][]*models.CommitFile,
	githubPeople []*models.GithubPerson,
) error {

	// Calculate statistics for each person
	for _, person := range githubPeople {
		stats := s.calculatePersonDailyStatsOptimized(
			projectID, projectRepositoryID, githubRepositoryID, person.ID, date,
			scoreSettings, excludedExtMap, excludedFolders, emailMerges, personEmailMap,
			allCommits, allPullRequests, allPRReviews, allCommitFiles,
		)

		if stats != nil && (stats.Score > 0 || stats.Commits > 0 || stats.PullRequests > 0 || stats.Comments > 0) {
			// Upsert if there's any activity (score > 0 or any commits/PRs/comments)
			if err := s.peopleStatsRepo.Upsert(stats); err != nil {
				return err
			}
		}
	}

	return nil
}

// calculatePersonDailyStatsOptimized calculates daily statistics for a specific person using pre-loaded data
func (s *PeopleStatisticsService) calculatePersonDailyStatsOptimized(
	projectID, projectRepositoryID, githubRepositoryID, githubPersonID string,
	date time.Time,
	scoreSettings *models.ScoreSettings,
	excludedExtMap map[string]bool,
	excludedFolders []*models.ExcludedFolder,
	emailMerges map[string]string,
	personEmailMap map[string]string,
	allCommits []*models.Commit,
	allPullRequests []*models.PullRequest,
	allPRReviews []*models.PRReview,
	allCommitFiles map[string][]*models.CommitFile,
) *models.PeopleStatistics {

	// Get the person's email
	personEmail, exists := personEmailMap[githubPersonID]
	if !exists {
		return nil // No email association, skip
	}

	// Calculate commit statistics using pre-loaded data
	commits, additions, deletions := s.calculateCommitStatsOptimized(
		allCommits, allCommitFiles, personEmail, date, excludedExtMap, excludedFolders, emailMerges,
	)

	// Calculate PR statistics using pre-loaded data
	pullRequests := s.calculatePRStatsOptimized(allPullRequests, githubPersonID, date)

	// Calculate comment statistics using pre-loaded data
	comments := s.calculateCommentStatsOptimized(allPRReviews, githubPersonID, date)

	// Calculate score based on score settings
	score := s.calculateScore(commits, additions, deletions, pullRequests, comments, scoreSettings)

	// Create statistics record
	stats := &models.PeopleStatistics{
		ID:             uuid.New().String(),
		ProjectID:      projectID,
		RepositoryID:   projectRepositoryID,
		GithubPersonID: githubPersonID,
		StatDate:       date,
		Commits:        commits,
		Additions:      additions,
		Deletions:      deletions,
		Comments:       comments,
		PullRequests:   pullRequests,
		Score:          score,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	return stats
}

// calculateDailyStatistics calculates statistics for a specific date
func (s *PeopleStatisticsService) calculateDailyStatistics(
	projectID, projectRepositoryID, githubRepositoryID string,
	date time.Time,
	scoreSettings *models.ScoreSettings,
	excludedExtensions []*models.ExcludedExtension,
	excludedFolders []*models.ExcludedFolder,
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
			scoreSettings, excludedExtMap, excludedFolders, emailMerges, personEmailMap,
		)

		if stats != nil && (stats.Score > 0 || stats.Commits > 0 || stats.PullRequests > 0 || stats.Comments > 0) {
			// Upsert if there's any activity (score > 0 or any commits/PRs/comments)
			if err := s.peopleStatsRepo.Upsert(stats); err != nil {
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
	excludedFolders []*models.ExcludedFolder,
	emailMerges map[string]string,
	personEmailMap map[string]string,
) *models.PeopleStatistics {

	// Get the person's associated email
	personEmail := personEmailMap[githubPersonID]

	// Initialize commit statistics
	commits := 0
	additions := 0
	deletions := 0

	// Only calculate commit statistics if the person has an email association
	if personEmail != "" {
		// Get all emails that should be attributed to this person (including merged emails)
		emailsToCheck := s.getEmailsForPerson(personEmail, emailMerges)

		// Calculate commit statistics for all emails
		commits, additions, deletions = s.calculateCommitStatsForEmails(githubRepositoryID, emailsToCheck, date, excludedExtMap, excludedFolders)
	}

	// Calculate PR statistics (always calculate, regardless of email association)
	pullRequests := s.calculatePRStats(githubRepositoryID, githubPersonID, date)

	// Calculate comment statistics (always calculate, regardless of email association)
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

// getEmailsForPerson returns all emails that should be attributed to a person (including merged emails)
func (s *PeopleStatisticsService) getEmailsForPerson(personEmail string, emailMerges map[string]string) []string {
	emails := []string{personEmail} // Always include the person's main email

	// Find all emails that merge to this person's email
	for sourceEmail, targetEmail := range emailMerges {
		if targetEmail == personEmail {
			emails = append(emails, sourceEmail)
		}
	}

	return emails
}

// calculateCommitStatsForEmails calculates commit statistics for multiple emails
func (s *PeopleStatisticsService) calculateCommitStatsForEmails(repositoryID string, emails []string, date time.Time, excludedExtMap map[string]bool, excludedFolders []*models.ExcludedFolder) (int, int, int) {
	totalCommits := 0
	totalAdditions := 0
	totalDeletions := 0

	// Calculate stats for each email and sum them up
	for _, email := range emails {
		commits, additions, deletions := s.calculateCommitStats(repositoryID, email, date, excludedExtMap, excludedFolders)
		totalCommits += commits
		totalAdditions += additions
		totalDeletions += deletions
	}

	return totalCommits, totalAdditions, totalDeletions
}

// calculateCommitStats calculates commit-related statistics for a person on a specific date
func (s *PeopleStatisticsService) calculateCommitStats(repositoryID, email string, date time.Time, excludedExtMap map[string]bool, excludedFolders []*models.ExcludedFolder) (int, int, int) {
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
						// Check if this file is in an excluded folder
						if !s.isExcludedFolder(commitFile.Filename, excludedFolders) {
							hasNonExcludedFiles = true
							commitAdditions += commitFile.Additions
							commitDeletions += commitFile.Deletions
						}
					}
				}

				// Only count this commit if it has non-excluded files
				if hasNonExcludedFiles {
					// Optimization 1: Skip commits with more than 20,000 total changes
					totalChanges := commitAdditions + commitDeletions
					if totalChanges > 20000 {
						continue
					}

					// Optimization 2: Skip commits that are deletion-only with more than 5,000 deletions
					if commitAdditions == 0 && commitDeletions > 5000 {
						continue
					}

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

	// Get the GitHub person to get their username
	githubPerson, err := s.githubPersonRepo.GetByID(githubPersonID)
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
				// Parse the user JSON to get the login (username)
				if pr.User != nil {
					var userData map[string]interface{}
					if err := json.Unmarshal([]byte(*pr.User), &userData); err == nil {
						if login, ok := userData["login"].(string); ok {
							// Check if the PR was created by this GitHub person
							if login == githubPerson.Username {
								count++
							}
						}
					}
				}
			}
		}
	}

	return count
}

// calculateCommentStats calculates comment statistics for a person on a specific date
func (s *PeopleStatisticsService) calculateCommentStats(repositoryID, githubPersonID string, date time.Time) int {
	// Get the GitHub person to get their username
	githubPerson, err := s.githubPersonRepo.GetByID(githubPersonID)
	if err != nil {
		return 0
	}

	// Get all PR reviews for this repository
	reviews, err := s.prReviewRepo.GetByRepositoryID(repositoryID)
	if err != nil {
		return 0
	}

	count := 0
	for _, review := range reviews {
		// Check if review was created by this person on the specified date
		if review.GithubCreatedAt != nil {
			reviewYear, reviewMonth, reviewDay := review.GithubCreatedAt.Date()
			dateYear, dateMonth, dateDay := date.Date()

			if reviewYear == dateYear && reviewMonth == dateMonth && reviewDay == dateDay {
				// Check if the review was created by this GitHub person
				if review.ReviewerLogin == githubPerson.Username {
					count++
				}
			}
		}
	}

	return count
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

// isExcludedFolder checks if a file is in an excluded folder
func (s *PeopleStatisticsService) isExcludedFolder(filePath string, excludedFolders []*models.ExcludedFolder) bool {
	for _, folder := range excludedFolders {
		folderPath := folder.FolderPath
		// Check if the file path starts with the excluded folder path
		if len(filePath) >= len(folderPath) && filePath[:len(folderPath)] == folderPath {
			// If it's exactly the folder path or starts with folder path + "/"
			if len(filePath) == len(folderPath) || filePath[len(folderPath)] == '/' {
				return true
			}
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

// GetAllTimeStatisticsByProject retrieves aggregated statistics for all GitHub people in a project
func (s *PeopleStatisticsService) GetAllTimeStatisticsByProject(projectID string) ([]*models.GitHubPersonStats, error) {
	// Get all GitHub people for this project
	people, err := s.githubPersonRepo.GetByProjectID(projectID)
	if err != nil {
		return nil, err
	}

	var results []*models.GitHubPersonStats
	for _, person := range people {
		// Get all statistics for this person in this project
		stats, err := s.peopleStatsRepo.GetByProjectAndPerson(projectID, person.ID)
		if err != nil {
			continue
		}

		// Aggregate the statistics
		totalCommits := 0
		totalAdditions := 0
		totalDeletions := 0
		totalComments := 0
		totalPullRequests := 0
		totalScore := 0

		for _, stat := range stats {
			totalCommits += stat.Commits
			totalAdditions += stat.Additions
			totalDeletions += stat.Deletions
			totalComments += stat.Comments
			totalPullRequests += stat.PullRequests
			totalScore += stat.Score
		}

		// Create the aggregated stats
		personStats := &models.GitHubPersonStats{
			GitHubPerson:      person,
			TotalCommits:      totalCommits,
			TotalAdditions:    totalAdditions,
			TotalDeletions:    totalDeletions,
			TotalComments:     totalComments,
			TotalPullRequests: totalPullRequests,
			TotalScore:        totalScore,
		}

		results = append(results, personStats)
	}

	// Sort by total score (descending)
	sort.Slice(results, func(i, j int) bool {
		return results[i].TotalScore > results[j].TotalScore
	})

	return results, nil
}

// GetYearlyStatisticsByProject retrieves yearly statistics for a project
func (s *PeopleStatisticsService) GetYearlyStatisticsByProject(projectID string, year int) ([]*models.GitHubPersonStats, error) {
	// Get all GitHub people for this project
	people, err := s.githubPersonRepo.GetByProjectID(projectID)
	if err != nil {
		return nil, err
	}

	var results []*models.GitHubPersonStats
	for _, person := range people {
		// Get statistics for this person in this project for the specified year
		stats, err := s.peopleStatsRepo.GetByProjectAndPersonAndYear(projectID, person.ID, year)
		if err != nil {
			continue
		}

		// Aggregate the statistics
		totalCommits := 0
		totalAdditions := 0
		totalDeletions := 0
		totalComments := 0
		totalPullRequests := 0
		totalScore := 0

		for _, stat := range stats {
			totalCommits += stat.Commits
			totalAdditions += stat.Additions
			totalDeletions += stat.Deletions
			totalComments += stat.Comments
			totalPullRequests += stat.PullRequests
			totalScore += stat.Score
		}

		// Create the aggregated stats
		personStats := &models.GitHubPersonStats{
			GitHubPerson:      person,
			TotalCommits:      totalCommits,
			TotalAdditions:    totalAdditions,
			TotalDeletions:    totalDeletions,
			TotalComments:     totalComments,
			TotalPullRequests: totalPullRequests,
			TotalScore:        totalScore,
		}

		results = append(results, personStats)
	}

	// Sort by total score (descending)
	sort.Slice(results, func(i, j int) bool {
		return results[i].TotalScore > results[j].TotalScore
	})

	return results, nil
}

// GetAvailableYearsForProject retrieves all available years for a project
func (s *PeopleStatisticsService) GetAvailableYearsForProject(projectID string) ([]int, error) {
	years, err := s.peopleStatsRepo.GetAvailableYearsForProject(projectID)
	if err != nil {
		return nil, err
	}

	if len(years) == 0 {
		// If no statistics exist, return current year only
		currentYear := time.Now().Year()
		return []int{currentYear}, nil
	}

	return years, nil
}

// GetMonthlyStatisticsByProject retrieves monthly statistics for a project
func (s *PeopleStatisticsService) GetMonthlyStatisticsByProject(projectID string, year int, month int) ([]*models.GitHubPersonStats, error) {
	// Get all GitHub people for this project
	people, err := s.githubPersonRepo.GetByProjectID(projectID)
	if err != nil {
		return nil, err
	}

	var results []*models.GitHubPersonStats
	for _, person := range people {
		// Get statistics for this person in this project for the specified month
		stats, err := s.peopleStatsRepo.GetByProjectAndPersonAndMonth(projectID, person.ID, year, month)
		if err != nil {
			continue
		}

		// Aggregate the statistics
		totalCommits := 0
		totalAdditions := 0
		totalDeletions := 0
		totalComments := 0
		totalPullRequests := 0
		totalScore := 0

		for _, stat := range stats {
			totalCommits += stat.Commits
			totalAdditions += stat.Additions
			totalDeletions += stat.Deletions
			totalComments += stat.Comments
			totalPullRequests += stat.PullRequests
			totalScore += stat.Score
		}

		// Create the aggregated stats
		personStats := &models.GitHubPersonStats{
			GitHubPerson:      person,
			TotalCommits:      totalCommits,
			TotalAdditions:    totalAdditions,
			TotalDeletions:    totalDeletions,
			TotalComments:     totalComments,
			TotalPullRequests: totalPullRequests,
			TotalScore:        totalScore,
		}

		results = append(results, personStats)
	}

	// Sort by total score (descending)
	sort.Slice(results, func(i, j int) bool {
		return results[i].TotalScore > results[j].TotalScore
	})

	return results, nil
}

// GetAvailableMonthsForProject retrieves all available months for a project
func (s *PeopleStatisticsService) GetAvailableMonthsForProject(projectID string) ([]string, error) {
	months, err := s.peopleStatsRepo.GetAvailableMonthsForProject(projectID)
	if err != nil {
		return nil, err
	}

	if len(months) == 0 {
		// If no statistics exist, return current month only
		now := time.Now()
		currentMonth := fmt.Sprintf("%d-%02d", now.Year(), now.Month())
		return []string{currentMonth}, nil
	}

	return months, nil
}

// GetWeeklyStatisticsByProject retrieves weekly statistics for a project
func (s *PeopleStatisticsService) GetWeeklyStatisticsByProject(projectID string, year int, week int) ([]*models.GitHubPersonStats, error) {
	// Get all GitHub people for this project
	people, err := s.githubPersonRepo.GetByProjectID(projectID)
	if err != nil {
		return nil, err
	}

	var results []*models.GitHubPersonStats
	for _, person := range people {
		// Get statistics for this person in this project for the specified week
		stats, err := s.peopleStatsRepo.GetByProjectAndPersonAndWeek(projectID, person.ID, year, week)
		if err != nil {
			continue
		}

		// Aggregate the statistics
		totalCommits := 0
		totalAdditions := 0
		totalDeletions := 0
		totalComments := 0
		totalPullRequests := 0
		totalScore := 0

		for _, stat := range stats {
			totalCommits += stat.Commits
			totalAdditions += stat.Additions
			totalDeletions += stat.Deletions
			totalComments += stat.Comments
			totalPullRequests += stat.PullRequests
			totalScore += stat.Score
		}

		// Create the aggregated stats
		personStats := &models.GitHubPersonStats{
			GitHubPerson:      person,
			TotalCommits:      totalCommits,
			TotalAdditions:    totalAdditions,
			TotalDeletions:    totalDeletions,
			TotalComments:     totalComments,
			TotalPullRequests: totalPullRequests,
			TotalScore:        totalScore,
		}

		results = append(results, personStats)
	}

	// Sort by total score (descending)
	sort.Slice(results, func(i, j int) bool {
		return results[i].TotalScore > results[j].TotalScore
	})

	return results, nil
}

// GetAvailableWeeksForProject retrieves all available weeks for a project
func (s *PeopleStatisticsService) GetAvailableWeeksForProject(projectID string) ([]string, error) {
	weeks, err := s.peopleStatsRepo.GetAvailableWeeksForProject(projectID)
	if err != nil {
		return nil, err
	}

	if len(weeks) == 0 {
		// If no statistics exist, return current week only
		now := time.Now()
		year, week := now.ISOWeek()
		currentWeek := fmt.Sprintf("%d-W%02d", year, week)
		return []string{currentWeek}, nil
	}

	return weeks, nil
}

// GetDailyStatisticsByProject retrieves daily statistics for a project
func (s *PeopleStatisticsService) GetDailyStatisticsByProject(projectID string, date time.Time) ([]*models.GitHubPersonStats, error) {
	// Get all GitHub people for this project
	people, err := s.githubPersonRepo.GetByProjectID(projectID)
	if err != nil {
		return nil, err
	}

	var results []*models.GitHubPersonStats
	for _, person := range people {
		// Get statistics for this person in this project for the specified date
		stats, err := s.peopleStatsRepo.GetByProjectAndPersonAndDate(projectID, person.ID, date)
		if err != nil {
			continue
		}

		// Aggregate the statistics (should be only one record per day)
		totalCommits := 0
		totalAdditions := 0
		totalDeletions := 0
		totalComments := 0
		totalPullRequests := 0
		totalScore := 0

		for _, stat := range stats {
			totalCommits += stat.Commits
			totalAdditions += stat.Additions
			totalDeletions += stat.Deletions
			totalComments += stat.Comments
			totalPullRequests += stat.PullRequests
			totalScore += stat.Score
		}

		// Create the aggregated stats
		personStats := &models.GitHubPersonStats{
			GitHubPerson:      person,
			TotalCommits:      totalCommits,
			TotalAdditions:    totalAdditions,
			TotalDeletions:    totalDeletions,
			TotalComments:     totalComments,
			TotalPullRequests: totalPullRequests,
			TotalScore:        totalScore,
		}

		results = append(results, personStats)
	}

	// Sort by total score (descending)
	sort.Slice(results, func(i, j int) bool {
		return results[i].TotalScore > results[j].TotalScore
	})

	return results, nil
}

// GetAvailableDaysForProject retrieves all available days for a project (last 30 days)
func (s *PeopleStatisticsService) GetAvailableDaysForProject(projectID string) ([]string, error) {
	// Get the earliest and latest dates from people_statistics table for this project
	earliestDate, latestDate, err := s.peopleStatsRepo.GetDateRangeForProject(projectID)
	if err != nil {
		return nil, err
	}

	if earliestDate == nil || latestDate == nil {
		// If no statistics exist, return current day only
		now := time.Now()
		currentDay := now.Format("2006-01-02")
		return []string{currentDay}, nil
	}

	// Generate days from the last 30 days
	var days []string
	endDate := *latestDate
	startDate := endDate.AddDate(0, 0, -30) // 30 days ago

	// If no data, default to current day
	if startDate.IsZero() || endDate.IsZero() {
		now := time.Now()
		currentDay := now.Format("2006-01-02")
		return []string{currentDay}, nil
	}

	// Generate days in descending order (newest first)
	current := endDate
	for current.After(startDate) || current.Equal(startDate) {
		dayStr := current.Format("2006-01-02")
		days = append(days, dayStr)

		// Move to previous day
		current = current.AddDate(0, 0, -1)
	}

	return days, nil
}

// calculateCommitStatsOptimized calculates commit statistics using pre-loaded data
func (s *PeopleStatisticsService) calculateCommitStatsOptimized(
	allCommits []*models.Commit,
	allCommitFiles map[string][]*models.CommitFile,
	personEmail string,
	date time.Time,
	excludedExtMap map[string]bool,
	excludedFolders []*models.ExcludedFolder,
	emailMerges map[string]string,
) (int, int, int) {
	// Get all emails that should be attributed to this person
	emailsToCheck := s.getEmailsForPerson(personEmail, emailMerges)

	totalCommits := 0
	totalAdditions := 0
	totalDeletions := 0

	// Filter commits by author email and date
	for _, commit := range allCommits {
		if commit.AuthorEmail != nil {
			// Check if this commit is by one of the person's emails
			isPersonCommit := false
			for _, email := range emailsToCheck {
				if *commit.AuthorEmail == email {
					isPersonCommit = true
					break
				}
			}

			if isPersonCommit {
				// Check if commit is on the specified date
				commitYear, commitMonth, commitDay := commit.CommitDate.Date()
				dateYear, dateMonth, dateDay := date.Date()

				if commitYear == dateYear && commitMonth == dateMonth && commitDay == dateDay {
					// Get commit files for this commit (from pre-loaded data)
					commitFiles, exists := allCommitFiles[commit.ID]
					if !exists {
						continue
					}

					// Check if any files in this commit are not excluded
					hasNonExcludedFiles := false
					commitAdditions := 0
					commitDeletions := 0

					for _, commitFile := range commitFiles {
						// Check if this file has an excluded extension
						if !s.isExcludedExtension(commitFile.Filename, excludedExtMap) {
							// Check if this file is in an excluded folder
							if !s.isExcludedFolder(commitFile.Filename, excludedFolders) {
								hasNonExcludedFiles = true
								commitAdditions += commitFile.Additions
								commitDeletions += commitFile.Deletions
							}
						}
					}

					// Only count this commit if it has non-excluded files
					if hasNonExcludedFiles {
						// Optimization 1: Skip commits with more than 20,000 total changes
						totalChanges := commitAdditions + commitDeletions
						if totalChanges > 20000 {
							continue
						}

						// Optimization 2: Skip commits that are deletion-only with more than 5,000 deletions
						if commitAdditions == 0 && commitDeletions > 5000 {
							continue
						}

						totalCommits++
						totalAdditions += commitAdditions
						totalDeletions += commitDeletions
					}
				}
			}
		}
	}

	return totalCommits, totalAdditions, totalDeletions
}

// calculatePRStatsOptimized calculates pull request statistics using pre-loaded data
func (s *PeopleStatisticsService) calculatePRStatsOptimized(
	allPullRequests []*models.PullRequest,
	githubPersonID string,
	date time.Time,
) int {
	// Get the GitHub person to get their username
	githubPerson, err := s.githubPersonRepo.GetByID(githubPersonID)
	if err != nil {
		return 0
	}

	count := 0
	for _, pr := range allPullRequests {
		// Check if PR was created by this person on the specified date
		if pr.GithubCreatedAt != nil {
			prYear, prMonth, prDay := pr.GithubCreatedAt.Date()
			dateYear, dateMonth, dateDay := date.Date()

			if prYear == dateYear && prMonth == dateMonth && prDay == dateDay {
				// Parse the user JSON to get the login (username)
				if pr.User != nil {
					var userData map[string]interface{}
					if err := json.Unmarshal([]byte(*pr.User), &userData); err == nil {
						if login, ok := userData["login"].(string); ok {
							// Check if the PR was created by this GitHub person
							if login == githubPerson.Username {
								count++
							}
						}
					}
				}
			}
		}
	}

	return count
}

// calculateCommentStatsOptimized calculates comment statistics using pre-loaded data
func (s *PeopleStatisticsService) calculateCommentStatsOptimized(
	allPRReviews []*models.PRReview,
	githubPersonID string,
	date time.Time,
) int {
	// Get the GitHub person to get their username
	githubPerson, err := s.githubPersonRepo.GetByID(githubPersonID)
	if err != nil {
		return 0
	}

	count := 0
	for _, review := range allPRReviews {
		// Check if review was created by this person on the specified date
		if review.GithubCreatedAt != nil {
			reviewYear, reviewMonth, reviewDay := review.GithubCreatedAt.Date()
			dateYear, dateMonth, dateDay := date.Date()

			if reviewYear == dateYear && reviewMonth == dateMonth && reviewDay == dateDay {
				// Check if the review was created by this GitHub person
				if review.ReviewerLogin == githubPerson.Username {
					count++
				}
			}
		}
	}

	return count
}

// calculateScore calculates the score based on activity and score settings
func (s *PeopleStatisticsService) calculateScore(
	commits, additions, deletions, pullRequests, comments int,
	scoreSettings *models.ScoreSettings,
) int {
	score := 0

	// Add points for commits
	score += commits * scoreSettings.Commits

	// Add points for additions
	score += additions * scoreSettings.Additions

	// Add points for deletions
	score += deletions * scoreSettings.Deletions

	// Add points for pull requests
	score += pullRequests * scoreSettings.PullRequests

	// Add points for comments
	score += comments * scoreSettings.Comments

	return score
}

// GetPersonWeeklyAverages calculates weekly averages for a specific person in a project
func (s *PeopleStatisticsService) GetPersonWeeklyAverages(projectID, githubPersonID string) (map[string]float64, error) {
	// Get all statistics for this person in this project
	stats, err := s.peopleStatsRepo.GetByProjectAndPerson(projectID, githubPersonID)
	if err != nil {
		return nil, fmt.Errorf("error fetching person statistics: %w", err)
	}

	if len(stats) == 0 {
		// Return zero averages if no data
		return map[string]float64{
			"commits":       0,
			"additions":     0,
			"deletions":     0,
			"comments":      0,
			"pull_requests": 0,
		}, nil
	}

	// Calculate totals
	var totalCommits, totalAdditions, totalDeletions, totalComments, totalPullRequests int
	for _, stat := range stats {
		totalCommits += stat.Commits
		totalAdditions += stat.Additions
		totalDeletions += stat.Deletions
		totalComments += stat.Comments
		totalPullRequests += stat.PullRequests
	}

	// Calculate weeks (minimum 1 week)
	weeks := 1
	if len(stats) > 0 {
		// Calculate the date range
		earliestDate := stats[0].StatDate
		latestDate := stats[0].StatDate

		for _, stat := range stats {
			if stat.StatDate.Before(earliestDate) {
				earliestDate = stat.StatDate
			}
			if stat.StatDate.After(latestDate) {
				latestDate = stat.StatDate
			}
		}

		// Calculate weeks between earliest and latest date
		daysDiff := int(latestDate.Sub(earliestDate).Hours() / 24)
		weeks = (daysDiff / 7) + 1 // Add 1 to include the current week
		if weeks < 1 {
			weeks = 1
		}
	}

	// Calculate weekly averages
	averages := map[string]float64{
		"commits":       float64(totalCommits) / float64(weeks),
		"additions":     float64(totalAdditions) / float64(weeks),
		"deletions":     float64(totalDeletions) / float64(weeks),
		"comments":      float64(totalComments) / float64(weeks),
		"pull_requests": float64(totalPullRequests) / float64(weeks),
	}

	return averages, nil
}

// GetPersonScoreHistory returns the person's score history for graphing
func (s *PeopleStatisticsService) GetPersonScoreHistory(projectID, githubPersonID string) ([]map[string]interface{}, error) {
	// Get all statistics for this person in this project
	stats, err := s.peopleStatsRepo.GetByProjectAndPerson(projectID, githubPersonID)
	if err != nil {
		return nil, fmt.Errorf("error fetching person statistics: %w", err)
	}

	if len(stats) == 0 {
		return []map[string]interface{}{}, nil
	}

	// Sort by date
	sort.Slice(stats, func(i, j int) bool {
		return stats[i].StatDate.Before(stats[j].StatDate)
	})

	// Group by month and calculate monthly score
	var result []map[string]interface{}

	// Find the first commit date
	firstDate := stats[0].StatDate
	firstMonth := time.Date(firstDate.Year(), firstDate.Month(), 1, 0, 0, 0, 0, firstDate.Location())

	// Initialize monthly data starting from the first month
	currentMonth := firstMonth
	lastMonth := time.Now()
	lastMonth = time.Date(lastMonth.Year(), lastMonth.Month(), 1, 0, 0, 0, 0, lastMonth.Location())

	// Process each month from first commit to current month
	for currentMonth.Before(lastMonth) || currentMonth.Equal(lastMonth) {
		monthlyScore := 0

		// Add scores for this month
		for _, stat := range stats {
			statMonth := time.Date(stat.StatDate.Year(), stat.StatDate.Month(), 1, 0, 0, 0, 0, stat.StatDate.Location())
			if statMonth.Equal(currentMonth) {
				monthlyScore += stat.Score
			}
		}

		// Add to result
		result = append(result, map[string]interface{}{
			"date":  currentMonth.Format("2006-01"),
			"score": monthlyScore,
		})

		// Move to next month
		currentMonth = currentMonth.AddDate(0, 1, 0)
	}

	return result, nil
}

// GetPersonTopReposAndLanguages returns top repositories and languages for a person
func (s *PeopleStatisticsService) GetPersonTopReposAndLanguages(projectID, githubPersonID string) (map[string]interface{}, error) {
	// Get all statistics for this person in this project
	stats, err := s.peopleStatsRepo.GetByProjectAndPerson(projectID, githubPersonID)
	if err != nil {
		return nil, fmt.Errorf("error fetching person statistics: %w", err)
	}

	if len(stats) == 0 {
		return map[string]interface{}{
			"TopRepos":     []map[string]interface{}{},
			"TopLanguages": []map[string]interface{}{},
		}, nil
	}

	// Group by repository and get repository details
	repoStats := make(map[string]int)
	repoDetails := make(map[string]*models.GitHubRepository)

	for _, stat := range stats {
		repoStats[stat.RepositoryID] += stat.Score
	}

	// Get repository details
	for repoID := range repoStats {
		// First get the project repository to get the GitHub repository ID
		projectRepo, err := s.projectRepositoryRepo.GetByID(repoID)
		if err == nil && projectRepo != nil {
			// Now get the GitHub repository details
			repo, err := s.githubRepoRepo.GetByID(projectRepo.GithubRepoID)
			if err == nil && repo != nil {
				repoDetails[repoID] = repo
			}
		}
	}

	// Get top 3 repositories
	var topRepos []map[string]interface{}
	for repoID, score := range repoStats {
		repoName := repoID
		if repo, exists := repoDetails[repoID]; exists {
			repoName = repo.Name
		}

		topRepos = append(topRepos, map[string]interface{}{
			"ID":    repoID,
			"Name":  repoName,
			"Score": score,
		})
	}

	// Sort by score and take top 3
	sort.Slice(topRepos, func(i, j int) bool {
		return topRepos[i]["Score"].(int) > topRepos[j]["Score"].(int)
	})
	if len(topRepos) > 3 {
		topRepos = topRepos[:3]
	}

	// Get actual language data from commit files
	languageStats := make(map[string]int)

	// Get all commits for this person in this project
	commits, err := s.commitRepo.GetByProjectAndPerson(projectID, githubPersonID)

	if err == nil {
		for _, commit := range commits {
			// Get commit files for this commit
			commitFiles, err := s.commitFileRepo.GetByCommitID(commit.ID)
			if err == nil {
				for _, file := range commitFiles {
					// Extract file extension and map to language
					language := getLanguageFromFile(file.Filename)
					if language != "" {
						languageStats[language] += file.Additions + file.Deletions
					}
				}
			}
		}
	}

	// Convert to slice and sort by score
	var topLanguages []map[string]interface{}
	for lang, score := range languageStats {
		topLanguages = append(topLanguages, map[string]interface{}{
			"Name":  lang,
			"Score": score,
		})
	}

	// Sort by score and take top 3
	sort.Slice(topLanguages, func(i, j int) bool {
		return topLanguages[i]["Score"].(int) > topLanguages[j]["Score"].(int)
	})
	if len(topLanguages) > 3 {
		topLanguages = topLanguages[:3]
	}

	// If no languages found, provide some basic info
	if len(topLanguages) == 0 {
		topLanguages = []map[string]interface{}{
			{"Name": "No language data", "Score": 0},
		}
	}

	return map[string]interface{}{
		"TopRepos":     topRepos,
		"TopLanguages": topLanguages,
	}, nil
}

// GetPersonTopCommitsAndPRs returns top commits and PRs for a person
func (s *PeopleStatisticsService) GetPersonTopCommitsAndPRs(projectID, githubPersonID string) (map[string]interface{}, error) {
	// Get all commits for this person in this project
	commits, err := s.commitRepo.GetByProjectAndPerson(projectID, githubPersonID)
	if err != nil {
		return nil, fmt.Errorf("error fetching person commits: %w", err)
	}

	// Get top 3 commits by LoC (additions + deletions)
	var topCommits []map[string]interface{}
	for _, commit := range commits {
		loc := commit.Additions + commit.Deletions
		topCommits = append(topCommits, map[string]interface{}{
			"ID":        commit.ID,
			"SHA":       commit.CommitSHA[:8],
			"Message":   commit.Message,
			"LoC":       loc,
			"Additions": commit.Additions,
			"Deletions": commit.Deletions,
			"Date":      commit.CommitDate.Format("2006-01-02"),
		})
	}

	// Sort by LoC and take top 3
	sort.Slice(topCommits, func(i, j int) bool {
		return topCommits[i]["LoC"].(int) > topCommits[j]["LoC"].(int)
	})
	if len(topCommits) > 3 {
		topCommits = topCommits[:3]
	}

	// Get all PRs for this person in this project
	prs, err := s.pullRequestRepo.GetByProjectAndPerson(projectID, githubPersonID)
	if err != nil {
		// If no PRs found, return empty list
		return map[string]interface{}{
			"TopCommits": topCommits,
			"TopPRs":     []map[string]interface{}{},
		}, nil
	}

	// Get the GitHub person to get their username
	githubPerson, err := s.githubPersonRepo.GetByID(githubPersonID)
	if err != nil {
		return map[string]interface{}{
			"TopCommits": topCommits,
			"TopPRs":     []map[string]interface{}{},
		}, nil
	}

	// Filter PRs by username
	var filteredPRs []*models.PullRequest
	for _, pr := range prs {
		if pr.User != nil {
			var userData map[string]interface{}
			if err := json.Unmarshal([]byte(*pr.User), &userData); err == nil {
				if login, ok := userData["login"].(string); ok {
					if login == githubPerson.Username {
						filteredPRs = append(filteredPRs, pr)
					}
				}
			}
		}
	}

	// Get top 3 PRs by comment count
	var topPRs []map[string]interface{}
	for _, pr := range filteredPRs {
		// Get comment count for this PR
		commentCount, err := s.prReviewRepo.GetCommentCountByPRID(pr.ID)
		if err != nil {
			commentCount = 0
		}

		topPRs = append(topPRs, map[string]interface{}{
			"ID":        pr.ID,
			"Number":    pr.GithubPRNumber,
			"Title":     pr.Title,
			"Comments":  commentCount,
			"State":     pr.State,
			"CreatedAt": pr.CreatedAt.Format("2006-01-02"),
		})
	}

	// Sort by comment count and take top 3
	sort.Slice(topPRs, func(i, j int) bool {
		return topPRs[i]["Comments"].(int) > topPRs[j]["Comments"].(int)
	})
	if len(topPRs) > 3 {
		topPRs = topPRs[:3]
	}

	return map[string]interface{}{
		"TopCommits": topCommits,
		"TopPRs":     topPRs,
	}, nil
}

// getLanguageFromFile maps file extensions to programming languages
func getLanguageFromFile(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))

	languageMap := map[string]string{
		".go":           "Go",
		".js":           "JavaScript",
		".ts":           "TypeScript",
		".py":           "Python",
		".java":         "Java",
		".cpp":          "C++",
		".c":            "C",
		".cs":           "C#",
		".php":          "PHP",
		".rb":           "Ruby",
		".rs":           "Rust",
		".swift":        "Swift",
		".kt":           "Kotlin",
		".scala":        "Scala",
		".html":         "HTML",
		".css":          "CSS",
		".scss":         "SCSS",
		".sass":         "Sass",
		".vue":          "Vue",
		".jsx":          "React",
		".tsx":          "React",
		".sh":           "Shell",
		".bash":         "Shell",
		".zsh":          "Shell",
		".sql":          "SQL",
		".json":         "JSON",
		".xml":          "XML",
		".yaml":         "YAML",
		".yml":          "YAML",
		".toml":         "TOML",
		".ini":          "INI",
		".md":           "Markdown",
		".txt":          "Text",
		".dockerfile":   "Docker",
		".dockerignore": "Docker",
	}

	if lang, exists := languageMap[ext]; exists {
		return lang
	}
	return ""
}

// GetPersonDetailedStats returns detailed statistics for a person
func (s *PeopleStatisticsService) GetPersonDetailedStats(projectID, githubPersonID string) (map[string]interface{}, error) {
	// Get all statistics for this person in this project
	stats, err := s.peopleStatsRepo.GetByProjectAndPerson(projectID, githubPersonID)
	if err != nil {
		return nil, fmt.Errorf("error fetching person statistics: %w", err)
	}

	if len(stats) == 0 {
		return map[string]interface{}{
			"PeakMonth":         "N/A",
			"ConsistencyScore":  0,
			"TotalCommits":      0,
			"TotalAdditions":    0,
			"TotalDeletions":    0,
			"TotalPullRequests": 0,
			"FirstActivity":     "N/A",
			"LastActivity":      "N/A",
			"ActiveDuration":    "N/A",
			"EfficiencyScore":   "N/A",
			"ProductivityScore": "N/A",
			"EngagementScore":   "N/A",
		}, nil
	}

	// Sort by date
	sort.Slice(stats, func(i, j int) bool {
		return stats[i].StatDate.Before(stats[j].StatDate)
	})

	// Calculate totals
	var totalCommits, totalAdditions, totalDeletions, totalPullRequests, totalComments int
	monthlyScores := make(map[string]int)
	activeMonths := make(map[string]bool)

	for _, stat := range stats {
		totalCommits += stat.Commits
		totalAdditions += stat.Additions
		totalDeletions += stat.Deletions
		totalPullRequests += stat.PullRequests
		totalComments += stat.Comments

		// Track monthly scores for peak performance
		monthKey := stat.StatDate.Format("2006-01")
		monthlyScores[monthKey] += stat.Score
		activeMonths[monthKey] = true
	}

	// Find peak month
	var peakMonth string
	var maxScore int
	for month, score := range monthlyScores {
		if score > maxScore {
			maxScore = score
			peakMonth = month
		}
	}

	// Calculate consistency (percentage of months with activity)
	firstDate := stats[0].StatDate
	lastDate := stats[len(stats)-1].StatDate
	totalMonths := int(lastDate.Sub(firstDate).Hours()/24/30) + 1
	consistencyScore := 0
	if totalMonths > 0 {
		consistencyScore = (len(activeMonths) * 100) / totalMonths
	}

	// Calculate efficiency metrics
	commitSizeScore := "N/A"
	if totalCommits > 0 {
		commitSizeScore = fmt.Sprintf("%.1f", float64(totalAdditions+totalDeletions)/float64(totalCommits))
	}

	refactorRatioScore := "N/A"
	if totalAdditions > 0 {
		originalRatio := float64(totalAdditions) / float64(totalAdditions+totalDeletions) * 100
		refactorRatioScore = fmt.Sprintf("%.1f", 100-originalRatio)
	}

	engagementScore := "N/A"
	if len(activeMonths) > 0 {
		engagementScore = fmt.Sprintf("%.1f", float64(totalPullRequests+totalComments)/float64(len(activeMonths)))
	}

	// Format dates
	firstActivity := firstDate.Format("2006-01-02")
	lastActivity := lastDate.Format("2006-01-02")

	// Calculate duration
	duration := lastDate.Sub(firstDate)
	activeDuration := fmt.Sprintf("%.0f days", duration.Hours()/24)

	return map[string]interface{}{
		"PeakMonth":          peakMonth,
		"ConsistencyScore":   consistencyScore,
		"TotalCommits":       totalCommits,
		"TotalAdditions":     totalAdditions,
		"TotalDeletions":     totalDeletions,
		"TotalPullRequests":  totalPullRequests,
		"TotalComments":      totalComments,
		"FirstActivity":      firstActivity,
		"LastActivity":       lastActivity,
		"ActiveDuration":     activeDuration,
		"CommitSizeScore":    commitSizeScore,
		"RefactorRatioScore": refactorRatioScore,
		"EngagementScore":    engagementScore,
	}, nil
}
