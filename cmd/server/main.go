package main

import (
	"context"
	"html/template"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/alimgiray/gscope/internal/handlers"
	"github.com/alimgiray/gscope/internal/middleware"
	"github.com/alimgiray/gscope/internal/repositories"
	"github.com/alimgiray/gscope/internal/services"
	"github.com/alimgiray/gscope/internal/workers"
	"github.com/alimgiray/gscope/pkg/config"
	"github.com/alimgiray/gscope/pkg/database"
	"github.com/alimgiray/gscope/pkg/logger"
	"github.com/gin-gonic/gin"
	"github.com/google/go-github/v57/github"
)

func main() {
	// Initialize logger
	logger.Init()

	// Load configuration
	if err := config.Load(); err != nil {
		logger.WithError(err).Fatal("Failed to load config")
	}

	// Set Gin mode from config
	gin.SetMode(config.AppConfig.Server.Mode)

	// Initialize database
	if err := database.Init(); err != nil {
		logger.WithError(err).Fatal("Failed to initialize database")
	}
	defer database.Close()

	// Initialize dependencies
	userRepo := repositories.NewUserRepository(database.DB)
	userService := services.NewUserService(userRepo)
	projectRepo := repositories.NewProjectRepository(database.DB)
	scoreSettingsRepo := repositories.NewScoreSettingsRepository(database.DB)
	scoreSettingsService := services.NewScoreSettingsService(scoreSettingsRepo)
	excludedExtensionRepo := repositories.NewExcludedExtensionRepository(database.DB)
	excludedExtensionService := services.NewExcludedExtensionService(excludedExtensionRepo)
	excludedFolderRepo := repositories.NewExcludedFolderRepository(database.DB)
	excludedFolderService := services.NewExcludedFolderService(excludedFolderRepo)
	githubRepoRepo := repositories.NewGitHubRepositoryRepository(database.DB)
	projectRepoRepo := repositories.NewProjectRepositoryRepository(database.DB)
	githubRepoService := services.NewGitHubRepositoryService(githubRepoRepo, projectRepoRepo)
	projectService := services.NewProjectService(projectRepo, scoreSettingsService)

	// Job and worker services
	jobRepo := repositories.NewJobRepository(database.DB)
	commitRepo := repositories.NewCommitRepository(database.DB)
	commitFileRepo := repositories.NewCommitFileRepository(database.DB)
	personRepo := repositories.NewPersonRepository(database.DB)
	jobService := services.NewJobService(jobRepo)
	cloneService := services.NewCloneService(projectRepo, userRepo, githubRepoRepo, projectRepoRepo)

	// Pull request related services
	pullRequestRepo := repositories.NewPullRequestRepository(database.DB)
	pullRequestService := services.NewPullRequestService(pullRequestRepo)
	prReviewRepo := repositories.NewPRReviewRepository(database.DB)
	prReviewService := services.NewPRReviewService(prReviewRepo)
	githubPersonRepo := repositories.NewGithubPersonRepository(database.DB)
	githubPersonService := services.NewGithubPersonService(githubPersonRepo)
	emailMergeRepo := repositories.NewEmailMergeRepository(database.DB)
	emailMergeService := services.NewEmailMergeService(emailMergeRepo)
	githubPersonEmailRepo := repositories.NewGitHubPersonEmailRepository(database.DB)
	githubPersonEmailService := services.NewGitHubPersonEmailService(githubPersonEmailRepo)
	textSimilarityService := services.NewTextSimilarityService()
	projectGithubPersonRepo := repositories.NewProjectGithubPersonRepository(database.DB)
	projectGithubPersonService := services.NewProjectGithubPersonService(projectGithubPersonRepo)

	// Working hours settings service
	workingHoursSettingsRepo := repositories.NewWorkingHoursSettingsRepository(database.DB)
	workingHoursSettingsService := services.NewWorkingHoursSettingsService(workingHoursSettingsRepo)

	// People statistics service
	peopleStatsRepo := repositories.NewPeopleStatisticsRepository(database.DB)
	projectRepositoryRepo := repositories.NewProjectRepositoryRepository(database.DB)
	peopleStatsService := services.NewPeopleStatisticsService(
		peopleStatsRepo,
		commitRepo,
		commitFileRepo,
		pullRequestRepo,
		prReviewRepo,
		githubPersonRepo,
		emailMergeRepo,
		githubPersonEmailRepo,
		personRepo,
		scoreSettingsRepo,
		excludedExtensionRepo,
		excludedFolderRepo,
		githubRepoRepo,
		projectRepositoryRepo,
		workingHoursSettingsService,
		projectGithubPersonService,
	)

	// Project update settings service
	projectUpdateSettingsRepo := repositories.NewProjectUpdateSettingsRepository(database.DB)
	projectUpdateSettingsService := services.NewProjectUpdateSettingsService(projectUpdateSettingsRepo)

	// Project collaborator service
	projectCollaboratorRepo := repositories.NewProjectCollaboratorRepository(database.DB)
	projectCollaboratorService := services.NewProjectCollaboratorService(projectCollaboratorRepo, userRepo, projectRepo)

	// LLM API key service
	llmAPIKeyRepo := repositories.NewLLMAPIKeyRepository(database.DB)
	llmAPIKeyService := services.NewLLMAPIKeyService(llmAPIKeyRepo, projectCollaboratorService)

	// Scheduler service
	schedulerService := services.NewSchedulerService(projectUpdateSettingsRepo, jobRepo, githubRepoService)

	// Initialize GitHub client
	githubClient := github.NewClient(nil)

	// Initialize worker manager
	workerManager := workers.NewWorkerManager(
		jobRepo, cloneService, projectRepoRepo, commitRepo, commitFileRepo, personRepo, githubRepoRepo,
		githubRepoService, pullRequestService, prReviewService, githubPersonService, peopleStatsService, githubClient,
		projectRepo, userRepo, projectGithubPersonService, pullRequestRepo,
	)

	// Initialize router
	router := gin.Default()

	// Apply middleware
	router.Use(middleware.SessionMiddleware())

	// Setup static files
	router.Static("/static", "./web/static")

	// Setup routes
	setupRoutes(router, userService, projectService, scoreSettingsService, excludedExtensionService, excludedFolderService, githubRepoService, jobService, jobRepo, commitRepo, commitFileRepo, pullRequestRepo, prReviewRepo, githubPersonRepo, personRepo, emailMergeService, githubPersonEmailService, textSimilarityService, peopleStatsService, projectUpdateSettingsService, projectCollaboratorService, workingHoursSettingsService, projectGithubPersonService, llmAPIKeyService)
	loadTemplates(router)

	// Start workers
	if err := workerManager.StartAll(); err != nil {
		logger.WithError(err).Fatal("Failed to start workers")
	}
	defer workerManager.StopAll()

	// Start scheduler
	schedulerService.StartScheduler()
	logger.Info("Automatic update scheduler started")

	// Setup server
	server := &http.Server{
		Addr:    ":" + config.AppConfig.Server.Port,
		Handler: router,
	}

	// Graceful shutdown
	go func() {
		logger.WithField("port", config.AppConfig.Server.Port).Info("Server starting")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.WithError(err).Fatal("Server failed to start")
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	// Create a context with 2-second timeout for graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Handle multiple Ctrl+C presses - force shutdown after 2 seconds
	go func() {
		<-quit
		logger.Info("Force shutdown requested...")
		cancel() // Cancel the context to force immediate shutdown
	}()

	// Attempt graceful shutdown
	if err := server.Shutdown(ctx); err != nil {
		logger.WithError(err).Error("Server forced to shutdown")
	}

	// Stop all workers
	workerManager.StopAll()
	logger.Info("Workers stopped")

	logger.Info("Server stopped")
}

func setupRoutes(router *gin.Engine, userService *services.UserService, projectService *services.ProjectService, scoreSettingsService *services.ScoreSettingsService, excludedExtensionService *services.ExcludedExtensionService, excludedFolderService *services.ExcludedFolderService, githubRepoService *services.GitHubRepositoryService, jobService *services.JobService, jobRepo *repositories.JobRepository, commitRepo *repositories.CommitRepository, commitFileRepo *repositories.CommitFileRepository, pullRequestRepo *repositories.PullRequestRepository, prReviewRepo *repositories.PRReviewRepository, githubPersonRepo *repositories.GithubPersonRepository, personRepo *repositories.PersonRepository, emailMergeService *services.EmailMergeService, githubPersonEmailService *services.GitHubPersonEmailService, textSimilarityService *services.TextSimilarityService, peopleStatsService *services.PeopleStatisticsService, projectUpdateSettingsService *services.ProjectUpdateSettingsService, projectCollaboratorService *services.ProjectCollaboratorService, workingHoursSettingsService *services.WorkingHoursSettingsService, projectGithubPersonService *services.ProjectGithubPersonService, llmAPIKeyService *services.LLMAPIKeyService) {
	// Initialize handlers
	homeHandler := handlers.NewHomeHandler(userService)
	authHandler := handlers.NewAuthHandler(userService)
	dashboardHandler := handlers.NewDashboardHandler(userService, projectService, projectCollaboratorService)
	projectHandler := handlers.NewProjectHandler(projectService, userService, scoreSettingsService, excludedExtensionService, excludedFolderService, githubRepoService, jobService, jobRepo, commitRepo, commitFileRepo, pullRequestRepo, prReviewRepo, githubPersonRepo, personRepo, emailMergeService, githubPersonEmailService, textSimilarityService, peopleStatsService, projectUpdateSettingsService, projectCollaboratorService, workingHoursSettingsService, projectGithubPersonService, llmAPIKeyService)
	workingHoursSettingsHandler := handlers.NewWorkingHoursSettingsHandler(workingHoursSettingsService)
	llmAPIKeyHandler := handlers.NewLLMAPIKeyHandler(llmAPIKeyService, projectService, projectCollaboratorService)
	healthHandler := handlers.NewHealthHandler()
	notFoundHandler := handlers.NewNotFoundHandler()

	// Home page
	router.GET("/", homeHandler.Index)

	// Auth routes
	router.GET("/login", authHandler.Login)
	router.GET("/logout", authHandler.Logout)
	router.GET("/auth/github", authHandler.GitHubLogin)
	router.GET("/auth/github/callback", authHandler.GitHubCallback)

	// Protected routes
	dashboard := router.Group("/dashboard")
	dashboard.Use(middleware.AuthRequired())
	{
		dashboard.GET("/", dashboardHandler.Dashboard)
	}

	projects := router.Group("/projects")
	projects.Use(middleware.AuthRequired())
	{
		projects.GET("/create", projectHandler.CreateProjectForm)
		projects.POST("/create", projectHandler.CreateProject)
		projects.GET("/:id", projectHandler.ViewProject)
		projects.GET("/:id/emails", projectHandler.ViewProjectEmails)
		projects.POST("/:id/emails/merge", projectHandler.CreateEmailMerge)
		projects.POST("/:id/emails/detach", projectHandler.DetachEmailMerge)
		projects.GET("/:id/people", projectHandler.ViewProjectPeople)
		projects.GET("/:id/people/:person_id", projectHandler.ViewPersonStats)
		projects.GET("/:id/repositories/:repository_id", projectHandler.ViewRepository)
		projects.POST("/:id/people/associate", projectHandler.CreateGitHubPersonEmailAssociation)
		projects.POST("/:id/people/detach", projectHandler.DeleteGitHubPersonEmailAssociation)
		projects.POST("/:id/people/remove", projectHandler.SoftDeletePerson)
		projects.POST("/:id/people/restore", projectHandler.RestorePerson)
		projects.GET("/:id/reports", projectHandler.ViewProjectReports)
		projects.GET("/:id/reports/daily", projectHandler.ViewProjectReportsDaily)
		projects.GET("/:id/reports/weekly", projectHandler.ViewProjectReportsWeekly)
		projects.GET("/:id/reports/monthly", projectHandler.ViewProjectReportsMonthly)
		projects.GET("/:id/reports/monthly/export", projectHandler.ExportMonthlyReportsToExcel)
		projects.GET("/:id/reports/yearly", projectHandler.ViewProjectReportsYearly)
		projects.GET("/:id/reports/yearly/export", projectHandler.ExportYearlyReportsToExcel)
		projects.POST("/:id/fetch-repositories", projectHandler.FetchRepositories)
		projects.POST("/:id/analyze", projectHandler.CreateAnalyzeJobs)
		projects.POST("/:id/repositories/:repository_id/clone", projectHandler.CreateCloneJob)
		projects.POST("/:id/repositories/:repository_id/fetch-github", projectHandler.CreateFetchGithubJob)
		projects.POST("/:id/repositories/:repository_id/analyze", projectHandler.CreateAnalyzeJobs)
		projects.POST("/:id/clone-all", projectHandler.CloneAllRepositories)
		projects.POST("/:id/track-all", projectHandler.TrackAllRepositories)
		projects.POST("/:id/fetch-all", projectHandler.FetchAllRepositories)
		projects.POST("/:id/analyze-all", projectHandler.AnalyzeAllRepositories)
		projects.POST("/:id/update-all", projectHandler.UpdateAllRepositories)
		projects.POST("/:id/repositories/:repository_id/toggle-track", projectHandler.ToggleRepositoryTracking)
		projects.GET("/:id/settings", projectHandler.ProjectSettings)
		projects.POST("/:id/settings/name", projectHandler.UpdateProjectName)
		projects.POST("/:id/settings/scores", projectHandler.UpdateScoreSettings)
		projects.POST("/:id/settings/extensions", projectHandler.AddExcludedExtension)
		projects.POST("/:id/settings/extensions/:extension_id/delete", projectHandler.DeleteExcludedExtension)
		projects.POST("/:id/settings/folders", projectHandler.AddExcludedFolder)
		projects.POST("/:id/settings/folders/:folder_id/delete", projectHandler.DeleteExcludedFolder)
		projects.POST("/:id/settings/update-settings", projectHandler.UpdateProjectUpdateSettings)
		projects.GET("/:id/working-hours-settings", workingHoursSettingsHandler.WorkingHoursSettingsForm)
		projects.POST("/:id/working-hours-settings", workingHoursSettingsHandler.UpdateWorkingHoursSettings)

		// LLM API Key routes
		projects.GET("/:id/llm-settings", llmAPIKeyHandler.ViewLLMSettings)
		projects.POST("/:id/llm/api-key", llmAPIKeyHandler.CreateOrUpdateAPIKey)
		projects.DELETE("/:id/llm/api-key", llmAPIKeyHandler.DeleteAPIKey)
		projects.POST("/:id/settings/delete", projectHandler.DeleteProject)
		projects.GET("/:id/collaborators", projectHandler.ViewProjectCollaborators)
		projects.POST("/:id/collaborators/add", projectHandler.AddProjectCollaborator)
		projects.POST("/:id/collaborators/remove", projectHandler.RemoveProjectCollaborator)
		projects.POST("/jobs/:job_id/retry", projectHandler.RetryFailedJob)
	}

	// Health check endpoint
	router.GET("/health", healthHandler.HealthCheck)

	// Handle 404 errors for non-existent routes
	router.NoRoute(notFoundHandler.NotFound)
}

func loadTemplates(router *gin.Engine) {
	cwd, err := os.Getwd()
	if err != nil {
		logger.WithError(err).Fatal("Couldn't get working directory")
	}
	logger.WithField("working_dir", cwd).Info("Template loading")

	// Add custom template functions
	router.SetFuncMap(template.FuncMap{
		"add1": func(i int) int {
			return i + 1
		},
		"seq": func(start, end int) []int {
			var result []int
			for i := start; i <= end; i++ {
				result = append(result, i)
			}
			return result
		},
	})

	router.LoadHTMLFiles(
		filepath.Join(cwd, "web/templates/layouts/header.html"),
		filepath.Join(cwd, "web/templates/layouts/footer.html"),
		filepath.Join(cwd, "web/templates/index.html"),
		filepath.Join(cwd, "web/templates/login.html"),
		filepath.Join(cwd, "web/templates/dashboard.html"),
		filepath.Join(cwd, "web/templates/projects/create.html"),
		filepath.Join(cwd, "web/templates/projects/view.html"),
		filepath.Join(cwd, "web/templates/projects/view_ajax.html"),
		filepath.Join(cwd, "web/templates/projects/settings.html"),
		filepath.Join(cwd, "web/templates/projects/emails.html"),
		filepath.Join(cwd, "web/templates/projects/people.html"),
		filepath.Join(cwd, "web/templates/projects/reports.html"),
		filepath.Join(cwd, "web/templates/projects/reports_daily.html"),
		filepath.Join(cwd, "web/templates/projects/reports_weekly.html"),
		filepath.Join(cwd, "web/templates/projects/reports_monthly.html"),
		filepath.Join(cwd, "web/templates/projects/reports_yearly.html"),
		filepath.Join(cwd, "web/templates/projects/collaborators.html"),
		filepath.Join(cwd, "web/templates/projects/person_stats.html"),
		filepath.Join(cwd, "web/templates/projects/repository.html"),
		filepath.Join(cwd, "web/templates/projects/llm_settings.html"),
		filepath.Join(cwd, "web/templates/error.html"),
		filepath.Join(cwd, "web/templates/404.html"),
	)
}
