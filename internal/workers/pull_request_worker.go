package workers

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/alimgiray/gscope/internal/models"
	"github.com/alimgiray/gscope/internal/repositories"
	"github.com/alimgiray/gscope/internal/services"
	"github.com/google/go-github/v57/github"
	"golang.org/x/oauth2"
)

type PullRequestWorker struct {
	*BaseWorker
	githubClient               *github.Client
	jobRepo                    *repositories.JobRepository
	pullRequestService         *services.PullRequestService
	prReviewService            *services.PRReviewService
	githubPersonService        *services.GithubPersonService
	githubRepoService          *services.GitHubRepositoryService
	projectRepositoryRepo      *repositories.ProjectRepositoryRepository
	projectRepo                *repositories.ProjectRepository
	userRepo                   *repositories.UserRepository
	projectGithubPersonService *services.ProjectGithubPersonService
	pullRequestRepo            *repositories.PullRequestRepository
}

func NewPullRequestWorker(
	workerID string,
	githubClient *github.Client,
	jobRepo *repositories.JobRepository,
	pullRequestService *services.PullRequestService,
	prReviewService *services.PRReviewService,
	githubPersonService *services.GithubPersonService,
	githubRepoService *services.GitHubRepositoryService,
	projectRepositoryRepo *repositories.ProjectRepositoryRepository,
	projectRepo *repositories.ProjectRepository,
	userRepo *repositories.UserRepository,
	projectGithubPersonService *services.ProjectGithubPersonService,
	pullRequestRepo *repositories.PullRequestRepository,
) *PullRequestWorker {
	return &PullRequestWorker{
		BaseWorker:                 NewBaseWorker(workerID, models.JobTypePullRequest),
		githubClient:               githubClient,
		jobRepo:                    jobRepo,
		pullRequestService:         pullRequestService,
		prReviewService:            prReviewService,
		githubPersonService:        githubPersonService,
		githubRepoService:          githubRepoService,
		projectRepositoryRepo:      projectRepositoryRepo,
		projectRepo:                projectRepo,
		userRepo:                   userRepo,
		projectGithubPersonService: projectGithubPersonService,
		pullRequestRepo:            pullRequestRepo,
	}
}

// Start begins the pull request worker process
func (w *PullRequestWorker) Start(ctx context.Context) error {
	w.Running = true
	log.Printf("Pull request worker %s started", w.WorkerID)

	for {
		select {
		case <-ctx.Done():
			log.Printf("Pull request worker %s stopping due to context cancellation", w.WorkerID)
			return ctx.Err()
		case <-w.StopChan:
			log.Printf("Pull request worker %s stopping", w.WorkerID)
			return nil
		default:
			// Try to get a pending pull request job
			job, err := w.jobRepo.GetNextPendingJob(models.JobTypePullRequest, w.WorkerID)
			if err != nil {
				log.Printf("Pull request worker %s error getting job: %v", w.WorkerID, err)
				time.Sleep(5 * time.Second)
				continue
			}

			if job == nil {
				// No jobs available, sleep and try again
				time.Sleep(10 * time.Second)
				continue
			}

			// Process the pull request job
			w.processPullRequestJob(ctx, job)
		}
	}
}

// processPullRequestJob handles the actual pull request job processing
func (w *PullRequestWorker) processPullRequestJob(ctx context.Context, job *models.Job) {
	log.Printf("Pull request worker %s processing job %s", w.WorkerID, job.ID)

	// Mark job as started
	job.MarkStarted()
	if err := w.jobRepo.Update(job); err != nil {
		log.Printf("Pull request worker %s error updating job %s: %v", w.WorkerID, job.ID, err)
		return
	}

	// Process the pull request job
	if err := w.ProcessJob(ctx, job); err != nil {
		log.Printf("Pull request worker %s error processing job %s: %v", w.WorkerID, job.ID, err)
		job.SetError(err.Error())
		job.MarkFailed()
		if err := w.jobRepo.Update(job); err != nil {
			log.Printf("Pull request worker %s error marking job %s as failed: %v", w.WorkerID, job.ID, err)
		}
		return
	}

	log.Printf("Pull request worker %s completed job %s", w.WorkerID, job.ID)
}

func (w *PullRequestWorker) ProcessJob(ctx context.Context, job *models.Job) error {
	log.Printf("Processing pull_request job for project: %s", job.ProjectID)

	// Update job status to in-progress
	job.Status = "in-progress"
	job.StartedAt = &time.Time{}
	*job.StartedAt = time.Now()
	if err := w.jobRepo.Update(job); err != nil {
		return fmt.Errorf("failed to update job status: %w", err)
	}

	// Get GitHub token for the project owner
	user, err := w.getUserByProjectID(job.ProjectID)
	if err != nil {
		return fmt.Errorf("failed to get user for project: %s", err)
	}

	// Create GitHub client with user's token
	userGithubClient := w.githubClient
	if user.GitHubAccessToken != "" {
		userGithubClient = w.createAuthenticatedClient(user.GitHubAccessToken)
	}

	totalPRs := 0
	totalReviews := 0
	totalPeople := 0

	// Check if this is a repository-specific job
	if job.ProjectRepositoryID != nil {
		// Process only the specific repository
		projectRepo, err := w.projectRepositoryRepo.GetByID(*job.ProjectRepositoryID)
		if err != nil {
			return fmt.Errorf("failed to get project repository %s: %s", *job.ProjectRepositoryID, err)
		}

		if !projectRepo.IsTracked {
			return fmt.Errorf("repository %s is not tracked", *job.ProjectRepositoryID)
		}

		// Get GitHub repository info
		githubRepo, err := w.githubRepoService.GetGitHubRepository(projectRepo.GithubRepoID)
		if err != nil {
			return fmt.Errorf("failed to get GitHub repository %s: %s", projectRepo.GithubRepoID, err)
		}

		// Parse owner and repo name from full name
		owner, repoName, err := parseRepoFullName(githubRepo.FullName)
		if err != nil {
			return fmt.Errorf("failed to parse repository name %s: %s", githubRepo.FullName, err)
		}

		log.Printf("Processing pull requests for %s/%s", owner, repoName)

		// Step 1: Fetch new pull requests from GitHub (PRs created after our last PR)
		newPullRequests, err := w.fetchPullRequests(ctx, userGithubClient, owner, repoName, githubRepo.ID)
		if err != nil {
			return fmt.Errorf("failed to fetch new pull requests for %s/%s: %s", owner, repoName, err)
		}

		// Step 2: Fetch and update existing open PRs from GitHub
		existingOpenPRs, err := w.fetchExistingOpenPullRequests(ctx, userGithubClient, owner, repoName, githubRepo.ID)
		if err != nil {
			log.Printf("Failed to fetch existing open PRs for %s/%s: %s", owner, repoName, err)
		}

		// Combine new and existing PRs
		allPullRequests := append(newPullRequests, existingOpenPRs...)

		// Process each pull request
		for _, pr := range allPullRequests {
			if err := w.processPullRequest(ctx, userGithubClient, owner, repoName, pr, githubRepo.ID, job.ProjectID); err != nil {
				log.Printf("Failed to process pull request #%d: %s", pr.GetNumber(), err)
				continue
			}
			totalPRs++

			// Fetch and process reviews for this PR
			reviews, err := w.fetchPullRequestReviews(ctx, userGithubClient, owner, repoName, pr.GetNumber())
			if err != nil {
				log.Printf("Failed to fetch reviews for PR #%d: %s", pr.GetNumber(), err)
				continue
			}

			for _, review := range reviews {
				if err := w.processPullRequestReview(ctx, review, githubRepo.ID, pr.GetID(), userGithubClient, job.ProjectID); err != nil {
					log.Printf("Failed to process review %d: %s", review.GetID(), err)
					continue
				}
				totalReviews++
			}
		}

		// Fetch and process repository contributors (even if no PRs exist)
		contributors, err := w.fetchRepositoryContributors(ctx, userGithubClient, owner, repoName)
		if err != nil {
			log.Printf("Failed to fetch contributors for %s/%s: %s", owner, repoName, err)
		} else {
			for _, contributor := range contributors {
				if err := w.processGithubPerson(contributor, userGithubClient, job.ProjectID, "contributor"); err != nil {
					log.Printf("Failed to process contributor %s: %s", contributor.GetLogin(), err)
					continue
				}
				totalPeople++
			}
		}
	} else {
		// Legacy: Process all tracked repositories (for backward compatibility)
		projectRepos, err := w.githubRepoService.GetProjectRepositories(job.ProjectID)
		if err != nil {
			return fmt.Errorf("failed to get project repositories: %s", err)
		}

		// Process each tracked repository
		for _, projectRepo := range projectRepos {
			if !projectRepo.IsTracked {
				continue
			}

			// Get GitHub repository info
			githubRepo, err := w.githubRepoService.GetGitHubRepository(projectRepo.GithubRepoID)
			if err != nil {
				log.Printf("Failed to get GitHub repository %s: %s", projectRepo.GithubRepoID, err)
				continue
			}

			// Parse owner and repo name from full name
			owner, repoName, err := parseRepoFullName(githubRepo.FullName)
			if err != nil {
				log.Printf("Failed to parse repository name %s: %s", githubRepo.FullName, err)
				continue
			}

			log.Printf("Processing pull requests for %s/%s", owner, repoName)

			// Fetch pull requests from GitHub
			pullRequests, err := w.fetchPullRequests(ctx, userGithubClient, owner, repoName, githubRepo.ID)
			if err != nil {
				log.Printf("Failed to fetch pull requests for %s/%s: %s", owner, repoName, err)
				continue
			}

			// Process each pull request
			for _, pr := range pullRequests {
				if err := w.processPullRequest(ctx, userGithubClient, owner, repoName, pr, githubRepo.ID, job.ProjectID); err != nil {
					log.Printf("Failed to process pull request #%d: %s", pr.GetNumber(), err)
					continue
				}
				totalPRs++

				// Fetch and process reviews for this PR
				reviews, err := w.fetchPullRequestReviews(ctx, userGithubClient, owner, repoName, pr.GetNumber())
				if err != nil {
					log.Printf("Failed to fetch reviews for PR #%d: %s", pr.GetNumber(), err)
					continue
				}

				for _, review := range reviews {
					if err := w.processPullRequestReview(ctx, review, githubRepo.ID, pr.GetID(), userGithubClient, job.ProjectID); err != nil {
						log.Printf("Failed to process review %d: %s", review.GetID(), err)
						continue
					}
					totalReviews++
				}
			}
		}
	}

	// Update job status to completed
	job.Status = "completed"
	job.CompletedAt = &time.Time{}
	*job.CompletedAt = time.Now()
	if err := w.jobRepo.Update(job); err != nil {
		return fmt.Errorf("failed to update job status: %w", err)
	}

	// Update last_fetched timestamp for the project repository if this was a repository-specific job
	if job.ProjectRepositoryID != nil {
		now := time.Now()
		if err := w.projectRepositoryRepo.UpdateLastFetched(*job.ProjectRepositoryID, &now); err != nil {
			log.Printf("Warning: failed to update last_fetched for project repository %s: %v", *job.ProjectRepositoryID, err)
		}
	}

	log.Printf("Pull request job completed. Processed %d PRs, %d reviews, %d people", totalPRs, totalReviews, totalPeople)
	return nil
}

func (w *PullRequestWorker) fetchPullRequests(ctx context.Context, client *github.Client, owner, repo string, repositoryID string) ([]*github.PullRequest, error) {
	// Get the latest PR date from the database for this repository (any PR, not just open ones)
	latestPRDate, err := w.pullRequestRepo.GetLatestPRDateByRepositoryID(repositoryID)
	if err != nil {
		log.Printf("Warning: failed to get latest PR date for repository %s: %v", repositoryID, err)
		// If we can't get the latest date, process all PRs
		latestPRDate = time.Time{}
	}

	var allPRs []*github.PullRequest
	opts := &github.PullRequestListOptions{
		State: "all", // Get both open and closed PRs
		ListOptions: github.ListOptions{
			PerPage: 100,
		},
	}

	if !latestPRDate.IsZero() {
		log.Printf("Processing PRs after %s for repository %s", latestPRDate.Format("2006-01-02 15:04:05"), repositoryID)
	} else {
		log.Printf("Processing all PRs for repository %s (no previous PRs found)", repositoryID)
	}

	for {
		prs, resp, err := w.makeGitHubRequestWithRetry(ctx, func() ([]*github.PullRequest, *github.Response, error) {
			return client.PullRequests.List(ctx, owner, repo, opts)
		})
		if err != nil {
			return nil, err
		}

		// Filter PRs by date if we have a latest date
		if !latestPRDate.IsZero() {
			for _, pr := range prs {
				// Only include PRs created after our latest PR date
				if pr.GetCreatedAt().After(latestPRDate) {
					allPRs = append(allPRs, pr)
				}
			}
		} else {
			allPRs = append(allPRs, prs...)
		}

		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	return allPRs, nil
}

// fetchExistingOpenPullRequests fetches all open PRs from GitHub to update existing ones
func (w *PullRequestWorker) fetchExistingOpenPullRequests(ctx context.Context, client *github.Client, owner, repo string, repositoryID string) ([]*github.PullRequest, error) {
	// Get existing open PR numbers from our database
	existingOpenPRs, err := w.pullRequestRepo.GetOpenPRNumbersByRepositoryID(repositoryID)
	if err != nil {
		log.Printf("Warning: failed to get existing open PRs for repository %s: %v", repositoryID, err)
		return nil, nil
	}

	if len(existingOpenPRs) == 0 {
		log.Printf("No existing open PRs found for repository %s", repositoryID)
		return nil, nil
	}

	var allPRs []*github.PullRequest
	opts := &github.PullRequestListOptions{
		State: "open", // Only get open PRs
		ListOptions: github.ListOptions{
			PerPage: 100,
		},
	}

	for {
		prs, resp, err := w.makeGitHubRequestWithRetry(ctx, func() ([]*github.PullRequest, *github.Response, error) {
			return client.PullRequests.List(ctx, owner, repo, opts)
		})
		if err != nil {
			return nil, err
		}

		// Only include PRs that exist in our database
		for _, pr := range prs {
			for _, existingPRNumber := range existingOpenPRs {
				if pr.GetNumber() == existingPRNumber {
					allPRs = append(allPRs, pr)
					break
				}
			}
		}

		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	log.Printf("Found %d existing open PRs to update for repository %s", len(allPRs), repositoryID)
	return allPRs, nil
}

func (w *PullRequestWorker) fetchPullRequestReviews(ctx context.Context, client *github.Client, owner, repo string, prNumber int) ([]*github.PullRequestReview, error) {
	var allReviews []*github.PullRequestReview
	opts := &github.ListOptions{
		PerPage: 100,
	}

	for {
		reviews, resp, err := w.makeGitHubRequestWithRetryReviews(ctx, func() ([]*github.PullRequestReview, *github.Response, error) {
			return client.PullRequests.ListReviews(ctx, owner, repo, prNumber, opts)
		})
		if err != nil {
			return nil, err
		}
		allReviews = append(allReviews, reviews...)

		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	return allReviews, nil
}

// fetchRepositoryContributors fetches contributors for a repository
func (w *PullRequestWorker) fetchRepositoryContributors(ctx context.Context, client *github.Client, owner, repo string) ([]*github.User, error) {
	var allContributors []*github.User
	opts := &github.ListContributorsOptions{
		ListOptions: github.ListOptions{
			PerPage: 100,
		},
	}

	for {
		contributors, resp, err := client.Repositories.ListContributors(ctx, owner, repo, opts)
		if err != nil {
			return nil, err
		}

		// Convert contributors to users by fetching full user details
		for _, contributor := range contributors {
			user, _, err := client.Users.Get(ctx, contributor.GetLogin())
			if err != nil {
				log.Printf("Failed to fetch user details for %s: %s", contributor.GetLogin(), err)
				continue
			}
			allContributors = append(allContributors, user)
		}

		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	return allContributors, nil
}

func (w *PullRequestWorker) processPullRequest(ctx context.Context, client *github.Client, owner, repo string, githubPR *github.PullRequest, repositoryID string, projectID string) error {
	// Process the PR author
	if githubPR.User != nil {
		if err := w.processGithubPerson(githubPR.User, client, projectID, "pull_request"); err != nil {
			log.Printf("Failed to process PR author: %s", err)
		}
	}

	// Convert GitHub PR to our model
	pr := &models.PullRequest{
		RepositoryID:   repositoryID,
		GithubPRNumber: githubPR.GetNumber(),
		GithubPRID:     int(githubPR.GetID()),
		Title:          githubPR.GetTitle(),
		State:          githubPR.GetState(),
		Draft:          githubPR.GetDraft(),
	}

	// Handle GitHub timestamps
	if githubPR.CreatedAt != nil {
		pr.GithubCreatedAt = &githubPR.CreatedAt.Time
	}
	if githubPR.UpdatedAt != nil {
		pr.GithubUpdatedAt = &githubPR.UpdatedAt.Time
	}

	// Handle optional fields
	if githubPR.Body != nil {
		body := githubPR.GetBody()
		pr.Body = &body
	}
	if githubPR.MergedAt != nil {
		pr.MergedAt = &githubPR.MergedAt.Time
	}
	if githubPR.MergeCommitSHA != nil {
		sha := githubPR.GetMergeCommitSHA()
		pr.MergeCommitSHA = &sha
	}
	if githubPR.ClosedAt != nil {
		pr.ClosedAt = &githubPR.ClosedAt.Time
	}

	// Convert user data to JSON
	if githubPR.User != nil {
		userJSON, err := json.Marshal(githubPR.User)
		if err == nil {
			userStr := string(userJSON)
			pr.User = &userStr
		}
	}

	// Convert requested reviewers to JSON
	if githubPR.RequestedReviewers != nil {
		reviewersJSON, err := json.Marshal(githubPR.RequestedReviewers)
		if err == nil {
			reviewersStr := string(reviewersJSON)
			pr.RequestedReviewers = &reviewersStr
		}
	}

	// Convert requested teams to JSON
	if githubPR.RequestedTeams != nil {
		teamsJSON, err := json.Marshal(githubPR.RequestedTeams)
		if err == nil {
			teamsStr := string(teamsJSON)
			pr.RequestedTeams = &teamsStr
		}
	}

	// Upsert the pull request
	return w.pullRequestService.UpsertPullRequest(pr)
}

func (w *PullRequestWorker) processPullRequestReview(ctx context.Context, githubReview *github.PullRequestReview, repositoryID string, pullRequestID int64, client *github.Client, projectID string) error {
	// Process the reviewer
	if githubReview.User != nil {
		if err := w.processGithubPerson(githubReview.User, client, projectID, "pull_request"); err != nil {
			log.Printf("Failed to process review author: %s", err)
		}
	}

	// Get the pull request from our database using the GitHub PR ID
	// Try a few times in case the PR was just created and hasn't been committed yet
	var pullRequest *models.PullRequest
	var err error
	for i := 0; i < 3; i++ {
		pullRequest, err = w.pullRequestRepo.GetByGithubPRID(int(pullRequestID))
		if err == nil {
			break
		}
		if err == sql.ErrNoRows {
			if i < 2 { // Don't log on the last attempt
				log.Printf("Warning: Pull request with GitHub PR ID %d not found in database, retrying...", pullRequestID)
				time.Sleep(100 * time.Millisecond) // Small delay before retry
				continue
			}
			log.Printf("Warning: Pull request with GitHub PR ID %d not found in database after retries, skipping review", pullRequestID)
			return nil // Skip this review if the PR doesn't exist yet
		}
		return fmt.Errorf("failed to get pull request with GitHub PR ID %d: %w", pullRequestID, err)
	}

	// Convert GitHub review to our model
	review := &models.PRReview{
		RepositoryID:   repositoryID,
		PullRequestID:  pullRequest.ID, // Use the database ID, not the GitHub PR ID
		GithubReviewID: int(githubReview.GetID()),
		ReviewerID:     int(githubReview.User.GetID()),
		ReviewerLogin:  githubReview.User.GetLogin(),
		State:          githubReview.GetState(),
		CommitID:       githubReview.GetCommitID(),
	}

	// Handle optional fields
	if githubReview.Body != nil {
		body := githubReview.GetBody()
		review.Body = &body
	}
	if githubReview.AuthorAssociation != nil {
		assoc := githubReview.GetAuthorAssociation()
		review.AuthorAssociation = &assoc
	}
	if githubReview.SubmittedAt != nil {
		review.SubmittedAt = &githubReview.SubmittedAt.Time
		// Set GitHub timestamps from submitted_at
		review.GithubCreatedAt = &githubReview.SubmittedAt.Time
		review.GithubUpdatedAt = &githubReview.SubmittedAt.Time
	}
	if githubReview.HTMLURL != nil {
		htmlURL := githubReview.GetHTMLURL()
		review.HTMLURL = &htmlURL
	}

	// Upsert the review
	return w.prReviewService.UpsertPRReview(review)
}

func (w *PullRequestWorker) processGithubPerson(githubUser *github.User, client *github.Client, projectID, sourceType string) error {
	person := &models.GithubPerson{
		GithubUserID: int(githubUser.GetID()),
		Username:     githubUser.GetLogin(),
	}

	// Handle optional fields
	if githubUser.Name != nil {
		displayName := githubUser.GetName()
		person.DisplayName = &displayName
	} else {
		// Try to fetch full user details to get the name using the authenticated client
		if fullUser, err := w.fetchFullUserDetails(githubUser.GetLogin(), client); err == nil && fullUser.Name != nil {
			displayName := fullUser.GetName()
			person.DisplayName = &displayName
		}
	}
	if githubUser.AvatarURL != nil {
		avatarURL := githubUser.GetAvatarURL()
		person.AvatarURL = &avatarURL
	}
	if githubUser.HTMLURL != nil {
		profileURL := githubUser.GetHTMLURL()
		person.ProfileURL = &profileURL
	}
	if githubUser.Type != nil {
		userType := githubUser.GetType()
		person.Type = &userType
	}

	// Upsert the person
	if err := w.githubPersonService.UpsertGithubPerson(person); err != nil {
		return err
	}

	// Create project-github person relationship
	return w.projectGithubPersonService.CreateProjectGithubPerson(projectID, person.ID, sourceType)
}

// fetchFullUserDetails fetches complete user details from GitHub API
func (w *PullRequestWorker) fetchFullUserDetails(username string, client *github.Client) (*github.User, error) {
	// Try to get the user details using the authenticated client
	user, _, err := client.Users.Get(context.Background(), username)
	if err != nil {
		log.Printf("Failed to fetch full user details for %s: %s", username, err)
		return nil, err
	}

	return user, nil
}

func parseRepoFullName(fullName string) (owner, repo string, err error) {
	// Full name format: "owner/repo"
	// This is a simple implementation - you might want to make it more robust
	if len(fullName) == 0 {
		return "", "", fmt.Errorf("empty repository name")
	}

	// Find the last slash
	lastSlash := -1
	for i := len(fullName) - 1; i >= 0; i-- {
		if fullName[i] == '/' {
			lastSlash = i
			break
		}
	}

	if lastSlash == -1 || lastSlash == 0 || lastSlash == len(fullName)-1 {
		return "", "", fmt.Errorf("invalid repository name format: %s", fullName)
	}

	owner = fullName[:lastSlash]
	repo = fullName[lastSlash+1:]
	return owner, repo, nil
}

// getUserByProjectID gets the user who owns the project
func (w *PullRequestWorker) getUserByProjectID(projectID string) (*models.User, error) {
	// Get the project to find the owner
	project, err := w.getProjectByID(projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get project: %w", err)
	}

	// Get the user by ID
	user, err := w.getUserByID(project.OwnerID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return user, nil
}

// getProjectByID gets a project by ID
func (w *PullRequestWorker) getProjectByID(projectID string) (*models.Project, error) {
	return w.projectRepo.GetByID(projectID)
}

// getUserByID gets a user by ID
func (w *PullRequestWorker) getUserByID(userID string) (*models.User, error) {
	return w.userRepo.GetByID(userID)
}

// createAuthenticatedClient creates a GitHub client with the provided token
func (w *PullRequestWorker) createAuthenticatedClient(token string) *github.Client {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)
	return github.NewClient(tc)
}

// makeGitHubRequestWithRetry performs a GitHub API request with rate limit handling and exponential backoff
func (w *PullRequestWorker) makeGitHubRequestWithRetry(ctx context.Context, requestFunc func() ([]*github.PullRequest, *github.Response, error)) ([]*github.PullRequest, *github.Response, error) {
	maxRetries := 5

	for attempt := 1; attempt <= maxRetries; attempt++ {
		result, resp, err := requestFunc()

		if err == nil {
			return result, resp, nil
		}

		// Check if it's a rate limit error
		if resp != nil && resp.StatusCode == 403 {
			// Check if we have rate limit info in headers
			if resetTime := resp.Header.Get("X-RateLimit-Reset"); resetTime != "" {
				if resetUnix, parseErr := time.Parse("2006-01-02 15:04:05", resetTime); parseErr == nil {
					waitTime := time.Until(resetUnix)
					if waitTime > 0 {
						log.Printf("Rate limit exceeded, waiting until %s (%.0f seconds)", resetUnix.Format("2006-01-02 15:04:05"), waitTime.Seconds())
						time.Sleep(waitTime)
						continue
					}
				}
			}

			// If we can't parse the reset time, use gradual backoff: 1min, 3min, 6min, 10min, 15min
			var waitTime time.Duration
			switch attempt {
			case 1:
				waitTime = 1 * time.Minute
			case 2:
				waitTime = 3 * time.Minute
			case 3:
				waitTime = 6 * time.Minute
			case 4:
				waitTime = 10 * time.Minute
			case 5:
				waitTime = 15 * time.Minute
			default:
				waitTime = 15 * time.Minute
			}
			log.Printf("Rate limit exceeded (attempt %d/%d), waiting %v before retry", attempt, maxRetries, waitTime)
			time.Sleep(waitTime)
			continue
		}

		// For other errors, use exponential backoff
		if attempt < maxRetries {
			var waitTime time.Duration
			switch attempt {
			case 1:
				waitTime = 1 * time.Minute
			case 2:
				waitTime = 3 * time.Minute
			case 3:
				waitTime = 6 * time.Minute
			case 4:
				waitTime = 10 * time.Minute
			case 5:
				waitTime = 15 * time.Minute
			default:
				waitTime = 15 * time.Minute
			}
			log.Printf("GitHub API request failed (attempt %d/%d): %v, waiting %v before retry", attempt, maxRetries, err, waitTime)
			time.Sleep(waitTime)
			continue
		}

		// Last attempt failed
		return result, resp, err
	}

	// This should never be reached, but just in case
	return nil, nil, fmt.Errorf("all retry attempts failed")
}

// makeGitHubRequestWithRetryReviews performs a GitHub API request for reviews with rate limit handling and exponential backoff
func (w *PullRequestWorker) makeGitHubRequestWithRetryReviews(ctx context.Context, requestFunc func() ([]*github.PullRequestReview, *github.Response, error)) ([]*github.PullRequestReview, *github.Response, error) {
	maxRetries := 5

	for attempt := 1; attempt <= maxRetries; attempt++ {
		result, resp, err := requestFunc()

		if err == nil {
			return result, resp, nil
		}

		// Check if it's a rate limit error
		if resp != nil && resp.StatusCode == 403 {
			// Check if we have rate limit info in headers
			if resetTime := resp.Header.Get("X-RateLimit-Reset"); resetTime != "" {
				if resetUnix, parseErr := time.Parse("2006-01-02 15:04:05", resetTime); parseErr == nil {
					waitTime := time.Until(resetUnix)
					if waitTime > 0 {
						log.Printf("Rate limit exceeded, waiting until %s (%.0f seconds)", resetUnix.Format("2006-01-02 15:04:05"), waitTime.Seconds())
						time.Sleep(waitTime)
						continue
					}
				}
			}

			// If we can't parse the reset time, use gradual backoff: 1min, 3min, 6min, 10min, 15min
			var waitTime time.Duration
			switch attempt {
			case 1:
				waitTime = 1 * time.Minute
			case 2:
				waitTime = 3 * time.Minute
			case 3:
				waitTime = 6 * time.Minute
			case 4:
				waitTime = 10 * time.Minute
			case 5:
				waitTime = 15 * time.Minute
			default:
				waitTime = 15 * time.Minute
			}
			log.Printf("Rate limit exceeded (attempt %d/%d), waiting %v before retry", attempt, maxRetries, waitTime)
			time.Sleep(waitTime)
			continue
		}

		// For other errors, use exponential backoff
		if attempt < maxRetries {
			var waitTime time.Duration
			switch attempt {
			case 1:
				waitTime = 1 * time.Minute
			case 2:
				waitTime = 3 * time.Minute
			case 3:
				waitTime = 6 * time.Minute
			case 4:
				waitTime = 10 * time.Minute
			case 5:
				waitTime = 15 * time.Minute
			default:
				waitTime = 15 * time.Minute
			}
			log.Printf("GitHub API request failed (attempt %d/%d): %v, waiting %v before retry", attempt, maxRetries, err, waitTime)
			time.Sleep(waitTime)
			continue
		}

		// Last attempt failed
		return result, resp, err
	}

	// This should never be reached, but just in case
	return nil, nil, fmt.Errorf("all retry attempts failed")
}
