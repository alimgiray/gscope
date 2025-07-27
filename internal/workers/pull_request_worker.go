package workers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/alimgiray/gscope/internal/models"
	"github.com/alimgiray/gscope/internal/repositories"
	"github.com/alimgiray/gscope/internal/services"
	"github.com/google/go-github/v57/github"
	"golang.org/x/oauth2"
)

type PullRequestWorker struct {
	*BaseWorker
	githubClient          *github.Client
	jobRepo               *repositories.JobRepository
	pullRequestService    *services.PullRequestService
	prReviewService       *services.PRReviewService
	githubPersonService   *services.GithubPersonService
	githubRepoService     *services.GitHubRepositoryService
	projectRepositoryRepo *repositories.ProjectRepositoryRepository
	projectRepo           *repositories.ProjectRepository
	userRepo              *repositories.UserRepository
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
) *PullRequestWorker {
	return &PullRequestWorker{
		BaseWorker:            NewBaseWorker(workerID, models.JobTypePullRequest),
		githubClient:          githubClient,
		jobRepo:               jobRepo,
		pullRequestService:    pullRequestService,
		prReviewService:       prReviewService,
		githubPersonService:   githubPersonService,
		githubRepoService:     githubRepoService,
		projectRepositoryRepo: projectRepositoryRepo,
		projectRepo:           projectRepo,
		userRepo:              userRepo,
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
			job, err := w.jobRepo.GetNextPendingJob(models.JobTypePullRequest)
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

		// Fetch pull requests from GitHub
		pullRequests, err := w.fetchPullRequests(ctx, userGithubClient, owner, repoName)
		if err != nil {
			return fmt.Errorf("failed to fetch pull requests for %s/%s: %s", owner, repoName, err)
		}

		// Process each pull request
		for _, pr := range pullRequests {
			if err := w.processPullRequest(ctx, userGithubClient, owner, repoName, pr, githubRepo.ID); err != nil {
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
				if err := w.processPullRequestReview(ctx, review, githubRepo.ID, pr.GetID(), userGithubClient); err != nil {
					log.Printf("Failed to process review %d: %s", review.GetID(), err)
					continue
				}
				totalReviews++
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
			pullRequests, err := w.fetchPullRequests(ctx, userGithubClient, owner, repoName)
			if err != nil {
				log.Printf("Failed to fetch pull requests for %s/%s: %s", owner, repoName, err)
				continue
			}

			// Process each pull request
			for _, pr := range pullRequests {
				if err := w.processPullRequest(ctx, userGithubClient, owner, repoName, pr, githubRepo.ID); err != nil {
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
					if err := w.processPullRequestReview(ctx, review, githubRepo.ID, pr.GetID(), userGithubClient); err != nil {
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

	log.Printf("Pull request job completed. Processed %d PRs, %d reviews, %d people", totalPRs, totalReviews, totalPeople)
	return nil
}

func (w *PullRequestWorker) fetchPullRequests(ctx context.Context, client *github.Client, owner, repo string) ([]*github.PullRequest, error) {
	var allPRs []*github.PullRequest
	opts := &github.PullRequestListOptions{
		State: "all", // Get both open and closed PRs
		ListOptions: github.ListOptions{
			PerPage: 100,
		},
	}

	for {
		prs, resp, err := client.PullRequests.List(ctx, owner, repo, opts)
		if err != nil {
			return nil, err
		}
		allPRs = append(allPRs, prs...)

		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	return allPRs, nil
}

func (w *PullRequestWorker) fetchPullRequestReviews(ctx context.Context, client *github.Client, owner, repo string, prNumber int) ([]*github.PullRequestReview, error) {
	var allReviews []*github.PullRequestReview
	opts := &github.ListOptions{
		PerPage: 100,
	}

	for {
		reviews, resp, err := client.PullRequests.ListReviews(ctx, owner, repo, prNumber, opts)
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

func (w *PullRequestWorker) processPullRequest(ctx context.Context, client *github.Client, owner, repo string, githubPR *github.PullRequest, repositoryID string) error {
	// Process the PR author
	if githubPR.User != nil {
		if err := w.processGithubPerson(githubPR.User, client); err != nil {
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

func (w *PullRequestWorker) processPullRequestReview(ctx context.Context, githubReview *github.PullRequestReview, repositoryID string, pullRequestID int64, client *github.Client) error {
	// Process the reviewer
	if githubReview.User != nil {
		if err := w.processGithubPerson(githubReview.User, client); err != nil {
			log.Printf("Failed to process review author: %s", err)
		}
	}

	// Convert GitHub review to our model
	review := &models.PRReview{
		RepositoryID:   repositoryID,
		PullRequestID:  strconv.FormatInt(pullRequestID, 10), // Convert to string since our ID is TEXT
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

func (w *PullRequestWorker) processGithubPerson(githubUser *github.User, client *github.Client) error {
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
	return w.githubPersonService.UpsertGithubPerson(person)
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
