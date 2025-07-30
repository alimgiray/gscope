package services

import (
	"encoding/json"
	"fmt"
	"sort"
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
	excludedFolderRepo    *repositories.ExcludedFolderRepository
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

	// Calculate statistics for each day from beginning to end
	for date := startDate; !date.After(endDate); date = date.AddDate(0, 0, 1) {
		if err := s.calculateDailyStatistics(projectID, projectRepositoryID, githubRepositoryID, date, scoreSettings, excludedExtensions, excludedFolders, emailMerges, personEmailMap); err != nil {
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
		// Apply email merges
		mergedEmail := s.getMergedEmail(personEmail, emailMerges)

		// Calculate commit statistics
		commits, additions, deletions = s.calculateCommitStats(githubRepositoryID, mergedEmail, date, excludedExtMap, excludedFolders)
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
		// Check if file path contains the excluded folder path
		folderPath := folder.FolderPath
		if len(filePath) >= len(folderPath) &&
			(filePath == folderPath ||
				(len(filePath) > len(folderPath) && filePath[:len(folderPath)] == folderPath && filePath[len(folderPath)] == '/') ||
				filePath[len(filePath)-len(folderPath):] == folderPath ||
				(len(filePath) > len(folderPath)+1 && filePath[len(filePath)-len(folderPath)-1:] == "/"+folder.FolderPath)) {
			return true
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
