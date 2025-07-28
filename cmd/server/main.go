package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/alimgiray/gscope/internal/handlers"
	"github.com/alimgiray/gscope/internal/middleware"
	"github.com/alimgiray/gscope/internal/repositories"
	"github.com/alimgiray/gscope/internal/services"
	"github.com/alimgiray/gscope/internal/workers"
	"github.com/alimgiray/gscope/pkg/config"
	"github.com/alimgiray/gscope/pkg/database"
	"github.com/gin-gonic/gin"
	"github.com/google/go-github/v57/github"
)

func main() {
	// Set Gin mode
	gin.SetMode(gin.ReleaseMode)

	// Load configuration
	if err := config.Load(); err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize database
	if err := database.Init(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
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

	// People statistics service
	peopleStatsRepo := repositories.NewPeopleStatisticsRepository(database.DB)
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
	)

	// Project update settings service
	projectUpdateSettingsRepo := repositories.NewProjectUpdateSettingsRepository(database.DB)
	projectUpdateSettingsService := services.NewProjectUpdateSettingsService(projectUpdateSettingsRepo)

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
	setupRoutes(router, userService, projectService, scoreSettingsService, excludedExtensionService, excludedFolderService, githubRepoService, jobService, jobRepo, commitRepo, githubPersonRepo, personRepo, emailMergeService, githubPersonEmailService, textSimilarityService, peopleStatsService, projectUpdateSettingsService)
	loadTemplates(router)

	// Start workers
	if err := workerManager.StartAll(); err != nil {
		log.Fatalf("Failed to start workers: %v", err)
	}
	defer workerManager.StopAll()

	// Start scheduler
	schedulerService.StartScheduler()
	log.Println("Automatic update scheduler started")

	// Setup server
	server := &http.Server{
		Addr:    ":8080",
		Handler: router,
	}

	// Graceful shutdown
	go func() {
		log.Printf("Server starting on :8080")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")
	log.Println("Server stopped")
}

func setupRoutes(router *gin.Engine, userService *services.UserService, projectService *services.ProjectService, scoreSettingsService *services.ScoreSettingsService, excludedExtensionService *services.ExcludedExtensionService, excludedFolderService *services.ExcludedFolderService, githubRepoService *services.GitHubRepositoryService, jobService *services.JobService, jobRepo *repositories.JobRepository, commitRepo *repositories.CommitRepository, githubPersonRepo *repositories.GithubPersonRepository, personRepo *repositories.PersonRepository, emailMergeService *services.EmailMergeService, githubPersonEmailService *services.GitHubPersonEmailService, textSimilarityService *services.TextSimilarityService, peopleStatsService *services.PeopleStatisticsService, projectUpdateSettingsService *services.ProjectUpdateSettingsService) {
	// Initialize handlers
	homeHandler := handlers.NewHomeHandler(userService)
	authHandler := handlers.NewAuthHandler(userService)
	dashboardHandler := handlers.NewDashboardHandler(userService, projectService)
	projectHandler := handlers.NewProjectHandler(projectService, userService, scoreSettingsService, excludedExtensionService, excludedFolderService, githubRepoService, jobService, jobRepo, commitRepo, githubPersonRepo, personRepo, emailMergeService, githubPersonEmailService, textSimilarityService, peopleStatsService, projectUpdateSettingsService)
	healthHandler := handlers.NewHealthHandler()

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
		projects.POST("/:id/people/associate", projectHandler.CreateGitHubPersonEmailAssociation)
		projects.POST("/:id/people/detach", projectHandler.DeleteGitHubPersonEmailAssociation)
		projects.GET("/:id/reports", projectHandler.ViewProjectReports)
		projects.GET("/:id/reports/daily", projectHandler.ViewProjectReportsDaily)
		projects.GET("/:id/reports/weekly", projectHandler.ViewProjectReportsWeekly)
		projects.GET("/:id/reports/monthly", projectHandler.ViewProjectReportsMonthly)
		projects.GET("/:id/reports/yearly", projectHandler.ViewProjectReportsYearly)
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
		projects.POST("/:id/settings/delete", projectHandler.DeleteProject)
	}

	// Health check endpoint
	router.GET("/health", healthHandler.HealthCheck)
}

func loadTemplates(router *gin.Engine) {
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatal("Couldn't get working directory:", err)
	}
	log.Println("Working dir:", cwd)

	router.LoadHTMLFiles(
		filepath.Join(cwd, "web/templates/layouts/header.html"),
		filepath.Join(cwd, "web/templates/layouts/footer.html"),
		filepath.Join(cwd, "web/templates/index.html"),
		filepath.Join(cwd, "web/templates/login.html"),
		filepath.Join(cwd, "web/templates/dashboard.html"),
		filepath.Join(cwd, "web/templates/projects/create.html"),
		filepath.Join(cwd, "web/templates/projects/view.html"),
		filepath.Join(cwd, "web/templates/projects/settings.html"),
		filepath.Join(cwd, "web/templates/projects/emails.html"),
		filepath.Join(cwd, "web/templates/projects/people.html"),
		filepath.Join(cwd, "web/templates/projects/reports.html"),
		filepath.Join(cwd, "web/templates/projects/reports_daily.html"),
		filepath.Join(cwd, "web/templates/projects/reports_weekly.html"),
		filepath.Join(cwd, "web/templates/projects/reports_monthly.html"),
		filepath.Join(cwd, "web/templates/projects/reports_yearly.html"),
	)
}
