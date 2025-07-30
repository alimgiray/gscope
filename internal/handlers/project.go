package handlers

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/alimgiray/gscope/internal/middleware"
	"github.com/alimgiray/gscope/internal/models"
	"github.com/alimgiray/gscope/internal/repositories"
	"github.com/alimgiray/gscope/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ProjectHandler struct {
	projectService               *services.ProjectService
	userService                  *services.UserService
	scoreSettingsService         *services.ScoreSettingsService
	excludedExtensionService     *services.ExcludedExtensionService
	excludedFolderService        *services.ExcludedFolderService
	githubRepoService            *services.GitHubRepositoryService
	jobService                   *services.JobService
	jobRepo                      *repositories.JobRepository
	commitRepo                   *repositories.CommitRepository
	githubPersonRepo             *repositories.GithubPersonRepository
	personRepo                   *repositories.PersonRepository
	emailMergeService            *services.EmailMergeService
	githubPersonEmailService     *services.GitHubPersonEmailService
	textSimilarityService        *services.TextSimilarityService
	peopleStatsService           *services.PeopleStatisticsService
	projectUpdateSettingsService *services.ProjectUpdateSettingsService
}

func NewProjectHandler(projectService *services.ProjectService, userService *services.UserService,
	scoreSettingsService *services.ScoreSettingsService, excludedExtensionService *services.ExcludedExtensionService,
	excludedFolderService *services.ExcludedFolderService, githubRepoService *services.GitHubRepositoryService, jobService *services.JobService,
	jobRepo *repositories.JobRepository, commitRepo *repositories.CommitRepository, githubPersonRepo *repositories.GithubPersonRepository,
	personRepo *repositories.PersonRepository, emailMergeService *services.EmailMergeService,
	githubPersonEmailService *services.GitHubPersonEmailService, textSimilarityService *services.TextSimilarityService,
	peopleStatsService *services.PeopleStatisticsService, projectUpdateSettingsService *services.ProjectUpdateSettingsService) *ProjectHandler {
	return &ProjectHandler{
		projectService:               projectService,
		userService:                  userService,
		scoreSettingsService:         scoreSettingsService,
		excludedExtensionService:     excludedExtensionService,
		excludedFolderService:        excludedFolderService,
		githubRepoService:            githubRepoService,
		jobService:                   jobService,
		jobRepo:                      jobRepo,
		commitRepo:                   commitRepo,
		githubPersonRepo:             githubPersonRepo,
		personRepo:                   personRepo,
		emailMergeService:            emailMergeService,
		githubPersonEmailService:     githubPersonEmailService,
		textSimilarityService:        textSimilarityService,
		peopleStatsService:           peopleStatsService,
		projectUpdateSettingsService: projectUpdateSettingsService,
	}
}

// CreateProjectForm displays the create project form
func (h *ProjectHandler) CreateProjectForm(c *gin.Context) {
	session := middleware.GetSession(c)
	if session == nil {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	data := gin.H{
		"Title":   "Create Project",
		"User":    session,
		"Project": &models.Project{},
	}

	c.HTML(http.StatusOK, "create_project", data)
}

// CreateProject handles project creation
func (h *ProjectHandler) CreateProject(c *gin.Context) {
	session := middleware.GetSession(c)
	if session == nil {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	// Get and validate form data
	name := strings.TrimSpace(c.PostForm("name"))
	if name == "" {
		data := gin.H{
			"Title":   "Create Project",
			"User":    session,
			"Project": &models.Project{Name: name},
			"Error":   "Project name is required",
		}
		c.HTML(http.StatusBadRequest, "create_project", data)
		return
	}

	// Parse user ID
	userID, err := uuid.Parse(session.UserID)
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error", gin.H{
			"Title": "Error",
			"User":  session,
			"Error": "Invalid user session",
		})
		return
	}

	// Create project
	project := &models.Project{
		Name:    name,
		OwnerID: userID,
	}

	if err := h.projectService.CreateProject(project); err != nil {
		data := gin.H{
			"Title":   "Create Project",
			"User":    session,
			"Project": &models.Project{Name: name},
			"Error":   err.Error(),
		}
		c.HTML(http.StatusBadRequest, "create_project", data)
		return
	}

	// Redirect to dashboard on success
	c.Redirect(http.StatusFound, "/dashboard")
}

// GetProjectsByOwner retrieves all projects for the current owner
func (h *ProjectHandler) GetProjectsByOwner(c *gin.Context) {
	session := middleware.GetSession(c)
	if session == nil {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	projects, err := h.projectService.GetProjectsByOwnerID(session.UserID)
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error", gin.H{
			"Title": "Error",
			"User":  session,
			"Error": "Failed to load projects",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"projects": projects,
	})
}

// GetProjectByID retrieves a specific project
func (h *ProjectHandler) GetProjectByID(c *gin.Context) {
	session := middleware.GetSession(c)
	if session == nil {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	projectID := c.Param("id")
	if projectID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Project ID is required"})
		return
	}

	project, err := h.projectService.GetProjectByID(projectID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
		return
	}

	// Check if the project belongs to the current owner
	userID, err := uuid.Parse(session.UserID)
	if err != nil || project.OwnerID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"project": project,
	})
}

// ViewProject displays a single project page
func (h *ProjectHandler) ViewProject(c *gin.Context) {
	session := middleware.GetSession(c)
	if session == nil {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	projectID := c.Param("id")
	if projectID == "" {
		c.HTML(http.StatusBadRequest, "error", gin.H{
			"Title": "Error",
			"User":  session,
			"Error": "Project ID is required",
		})
		return
	}

	project, err := h.projectService.GetProjectByID(projectID)
	if err != nil {
		c.HTML(http.StatusNotFound, "error", gin.H{
			"Title": "Project Not Found",
			"User":  session,
			"Error": "The requested project could not be found.",
		})
		return
	}

	// Check if the project belongs to the current owner
	userID, err := uuid.Parse(session.UserID)
	if err != nil || project.OwnerID != userID {
		c.HTML(http.StatusForbidden, "error", gin.H{
			"Title": "Access Denied",
			"User":  session,
			"Error": "You don't have permission to view this project.",
		})
		return
	}

	// Get project repositories
	projectRepos, err := h.githubRepoService.GetProjectRepositories(projectID)
	if err != nil {
		projectRepos = []*models.ProjectRepository{}
	}

	// Get GitHub repository details and job status for each project repository
	var repositories []map[string]interface{}
	for _, projectRepo := range projectRepos {
		githubRepo, err := h.githubRepoService.GetGitHubRepository(projectRepo.GithubRepoID)
		if err != nil {
			continue
		}

		// Get all jobs for this repository
		allJobs, err := h.jobService.GetProjectRepositoryJobs(projectRepo.ID)
		if err != nil {
			allJobs = []*models.Job{}
		}

		// Calculate clone job status
		hasActiveCloneJobs := false
		latestCloneJobFailed := false
		latestCloneJobError := ""
		failedCloneJobID := ""

		// Check for active clone or commit jobs (commit depends on clone)
		for _, job := range allJobs {
			if job.JobType == models.JobTypeClone || job.JobType == models.JobTypeCommit {
				if job.Status == models.JobStatusPending || job.Status == models.JobStatusInProgress {
					hasActiveCloneJobs = true
					break
				}
				if job.JobType == models.JobTypeClone && job.Status == models.JobStatusFailed {
					latestCloneJobFailed = true
					failedCloneJobID = job.ID
					if job.ErrorMessage != nil {
						latestCloneJobError = *job.ErrorMessage
					}
				}
			}
		}

		// Get repository-specific jobs for analyze status
		repoJobs, err := h.jobService.GetProjectRepositoryJobs(projectRepo.ID)
		if err != nil {
			repoJobs = []*models.Job{}
		}

		// Calculate analyze job status (repository-specific jobs)
		hasActiveAnalyzeJobs := false
		latestAnalyzeJobFailed := false
		latestAnalyzeJobError := ""
		failedAnalyzeJobID := ""

		// Check for active pull_request or stats jobs (stats depends on pull_request)
		for _, job := range repoJobs {
			if job.JobType == models.JobTypePullRequest || job.JobType == models.JobTypeStats {
				if job.Status == models.JobStatusPending || job.Status == models.JobStatusInProgress {
					hasActiveAnalyzeJobs = true
					break
				}
				if job.JobType == models.JobTypePullRequest && job.Status == models.JobStatusFailed {
					latestAnalyzeJobFailed = true
					failedAnalyzeJobID = job.ID
					if job.ErrorMessage != nil {
						latestAnalyzeJobError = *job.ErrorMessage
					}
				}
			}
		}

		// Check if "fetch github" has been called (has completed pull_request jobs)
		hasCompletedPullRequestJobs := false
		for _, job := range repoJobs {
			if job.JobType == models.JobTypePullRequest && job.Status == models.JobStatusCompleted {
				hasCompletedPullRequestJobs = true
				break
			}
		}

		// Check if repository has been cloned
		isRepositoryCloned := githubRepo.IsCloned

		// Check if there are GitHub people with email associations for this project
		hasGitHubPeopleWithEmails := false
		_, err = h.githubPersonRepo.GetByProjectID(projectID)
		if err == nil {
			// Get email associations for this project
			emailAssociations, err := h.githubPersonEmailService.GetGitHubPersonEmailsByProjectID(projectID)
			if err == nil && len(emailAssociations) > 0 {
				hasGitHubPeopleWithEmails = true
			}
		}

		// Get all failed jobs for this repository for display
		var failedJobs []map[string]interface{}
		for _, job := range allJobs {
			if job.Status == models.JobStatusFailed {
				failedJobs = append(failedJobs, map[string]interface{}{
					"ID":           job.ID,
					"JobType":      job.JobType,
					"ErrorMessage": job.ErrorMessage,
					"CreatedAt":    job.CreatedAt,
				})
			}
		}

		repositories = append(repositories, map[string]interface{}{
			"ProjectRepo":                 projectRepo,
			"GitHubRepo":                  githubRepo,
			"HasActiveCloneJobs":          hasActiveCloneJobs,
			"LatestCloneJobFailed":        latestCloneJobFailed,
			"LatestCloneJobError":         latestCloneJobError,
			"FailedCloneJobID":            failedCloneJobID,
			"HasActiveAnalyzeJobs":        hasActiveAnalyzeJobs,
			"LatestAnalyzeJobFailed":      latestAnalyzeJobFailed,
			"LatestAnalyzeJobError":       latestAnalyzeJobError,
			"FailedAnalyzeJobID":          failedAnalyzeJobID,
			"HasCompletedPullRequestJobs": hasCompletedPullRequestJobs,
			"IsRepositoryCloned":          isRepositoryCloned,
			"HasGitHubPeopleWithEmails":   hasGitHubPeopleWithEmails,
			"FailedJobs":                  failedJobs,
		})
	}

	// Get message from query parameter
	message := c.Query("message")

	data := gin.H{
		"Title":        project.Name,
		"User":         session,
		"Project":      project,
		"Repositories": repositories,
		"Message":      message,
	}

	c.HTML(http.StatusOK, "project_view", data)
}

// ProjectSettings displays the project settings page
func (h *ProjectHandler) ProjectSettings(c *gin.Context) {
	session := middleware.GetSession(c)
	if session == nil {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	projectID := c.Param("id")
	if projectID == "" {
		c.HTML(http.StatusBadRequest, "error", gin.H{
			"Title": "Error",
			"User":  session,
			"Error": "Project ID is required",
		})
		return
	}

	project, err := h.projectService.GetProjectByID(projectID)
	if err != nil {
		c.HTML(http.StatusNotFound, "error", gin.H{
			"Title": "Project Not Found",
			"User":  session,
			"Error": "The requested project could not be found.",
		})
		return
	}

	// Check if the project belongs to the current owner
	userID, err := uuid.Parse(session.UserID)
	if err != nil || project.OwnerID != userID {
		c.HTML(http.StatusForbidden, "error", gin.H{
			"Title": "Access Denied",
			"User":  session,
			"Error": "You don't have permission to access this project's settings.",
		})
		return
	}

	// Get score settings
	scoreSettings, err := h.scoreSettingsService.GetScoreSettingsByProjectID(projectID)
	if err != nil {
		// If no score settings found, create default ones
		scoreSettings = models.NewScoreSettings(projectID)
	}

	// Get excluded extensions
	excludedExtensions, err := h.excludedExtensionService.GetExcludedExtensionsByProjectID(projectID)
	if err != nil {
		excludedExtensions = []*models.ExcludedExtension{}
	}

	// Get excluded folders
	excludedFolders, err := h.excludedFolderService.GetExcludedFoldersByProjectID(projectID)
	if err != nil {
		excludedFolders = []*models.ExcludedFolder{}
	}

	// Get project update settings
	updateSettings, err := h.projectUpdateSettingsService.GetProjectUpdateSettings(projectID)
	if err != nil && err != sql.ErrNoRows {
		// If there's an error other than "no rows", log it but continue
		log.Printf("Error getting project update settings: %v", err)
		updateSettings = nil
	}

	data := gin.H{
		"Title":              "Project Settings",
		"User":               session,
		"Project":            project,
		"ScoreSettings":      scoreSettings,
		"ExcludedExtensions": excludedExtensions,
		"ExcludedFolders":    excludedFolders,
		"UpdateSettings":     updateSettings,
	}

	c.HTML(http.StatusOK, "project_settings", data)
}

// UpdateProjectName handles project name updates
func (h *ProjectHandler) UpdateProjectName(c *gin.Context) {
	session := middleware.GetSession(c)
	if session == nil {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	projectID := c.Param("id")
	name := strings.TrimSpace(c.PostForm("name"))

	if name == "" {
		c.HTML(http.StatusBadRequest, "error", gin.H{
			"Title": "Error",
			"User":  session,
			"Error": "Project name is required",
		})
		return
	}

	// Get project and verify ownership
	project, err := h.projectService.GetProjectByID(projectID)
	if err != nil {
		c.HTML(http.StatusNotFound, "error", gin.H{
			"Title": "Project Not Found",
			"User":  session,
			"Error": "The requested project could not be found.",
		})
		return
	}

	userID, err := uuid.Parse(session.UserID)
	if err != nil || project.OwnerID != userID {
		c.HTML(http.StatusForbidden, "error", gin.H{
			"Title": "Access Denied",
			"User":  session,
			"Error": "You don't have permission to modify this project.",
		})
		return
	}

	// Update project name
	project.Name = name
	if err := h.projectService.UpdateProject(project); err != nil {
		c.HTML(http.StatusInternalServerError, "error", gin.H{
			"Title": "Error",
			"User":  session,
			"Error": "Failed to update project name: " + err.Error(),
		})
		return
	}

	c.Redirect(http.StatusFound, "/projects/"+projectID+"/settings")
}

// UpdateScoreSettings handles score settings updates
func (h *ProjectHandler) UpdateScoreSettings(c *gin.Context) {
	session := middleware.GetSession(c)
	if session == nil {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	projectID := c.Param("id")

	// Parse form values
	additions, _ := strconv.Atoi(c.PostForm("additions"))
	deletions, _ := strconv.Atoi(c.PostForm("deletions"))
	commits, _ := strconv.Atoi(c.PostForm("commits"))
	pullRequests, _ := strconv.Atoi(c.PostForm("pull_requests"))
	comments, _ := strconv.Atoi(c.PostForm("comments"))

	// Validate project ownership
	project, err := h.projectService.GetProjectByID(projectID)
	if err != nil {
		c.HTML(http.StatusNotFound, "error", gin.H{
			"Title": "Project Not Found",
			"User":  session,
			"Error": "The requested project could not be found.",
		})
		return
	}

	userID, err := uuid.Parse(session.UserID)
	if err != nil || project.OwnerID != userID {
		c.HTML(http.StatusForbidden, "error", gin.H{
			"Title": "Access Denied",
			"User":  session,
			"Error": "You don't have permission to modify this project.",
		})
		return
	}

	// Get or create score settings
	scoreSettings, err := h.scoreSettingsService.GetScoreSettingsByProjectID(projectID)
	if err != nil {
		scoreSettings = models.NewScoreSettings(projectID)
	}

	// Update values
	scoreSettings.Additions = additions
	scoreSettings.Deletions = deletions
	scoreSettings.Commits = commits
	scoreSettings.PullRequests = pullRequests
	scoreSettings.Comments = comments

	if err := h.scoreSettingsService.UpdateScoreSettings(scoreSettings); err != nil {
		c.HTML(http.StatusInternalServerError, "error", gin.H{
			"Title": "Error",
			"User":  session,
			"Error": "Failed to update score settings: " + err.Error(),
		})
		return
	}

	c.Redirect(http.StatusFound, "/projects/"+projectID+"/settings")
}

// AddExcludedExtension handles adding excluded extensions
func (h *ProjectHandler) AddExcludedExtension(c *gin.Context) {
	session := middleware.GetSession(c)
	if session == nil {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	projectID := c.Param("id")
	extension := strings.TrimSpace(c.PostForm("extension"))

	if extension == "" {
		c.HTML(http.StatusBadRequest, "error", gin.H{
			"Title": "Error",
			"User":  session,
			"Error": "Extension is required",
		})
		return
	}

	// Validate project ownership
	project, err := h.projectService.GetProjectByID(projectID)
	if err != nil {
		c.HTML(http.StatusNotFound, "error", gin.H{
			"Title": "Project Not Found",
			"User":  session,
			"Error": "The requested project could not be found.",
		})
		return
	}

	userID, err := uuid.Parse(session.UserID)
	if err != nil || project.OwnerID != userID {
		c.HTML(http.StatusForbidden, "error", gin.H{
			"Title": "Access Denied",
			"User":  session,
			"Error": "You don't have permission to modify this project.",
		})
		return
	}

	if err := h.excludedExtensionService.CreateExcludedExtension(projectID, extension); err != nil {
		c.HTML(http.StatusInternalServerError, "error", gin.H{
			"Title": "Error",
			"User":  session,
			"Error": "Failed to add excluded extension: " + err.Error(),
		})
		return
	}

	c.Redirect(http.StatusFound, "/projects/"+projectID+"/settings")
}

// DeleteExcludedExtension handles removing excluded extensions
func (h *ProjectHandler) DeleteExcludedExtension(c *gin.Context) {
	session := middleware.GetSession(c)
	if session == nil {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	projectID := c.Param("id")
	extensionID := c.Param("extension_id")

	// Validate project ownership
	project, err := h.projectService.GetProjectByID(projectID)
	if err != nil {
		c.HTML(http.StatusNotFound, "error", gin.H{
			"Title": "Project Not Found",
			"User":  session,
			"Error": "The requested project could not be found.",
		})
		return
	}

	userID, err := uuid.Parse(session.UserID)
	if err != nil || project.OwnerID != userID {
		c.HTML(http.StatusForbidden, "error", gin.H{
			"Title": "Access Denied",
			"User":  session,
			"Error": "You don't have permission to modify this project.",
		})
		return
	}

	if err := h.excludedExtensionService.DeleteExcludedExtension(extensionID); err != nil {
		c.HTML(http.StatusInternalServerError, "error", gin.H{
			"Title": "Error",
			"User":  session,
			"Error": "Failed to delete excluded extension: " + err.Error(),
		})
		return
	}

	c.Redirect(http.StatusFound, "/projects/"+projectID+"/settings")
}

// DeleteProject handles project deletion
func (h *ProjectHandler) DeleteProject(c *gin.Context) {
	session := middleware.GetSession(c)
	if session == nil {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	projectID := c.Param("id")

	// Validate project ownership
	project, err := h.projectService.GetProjectByID(projectID)
	if err != nil {
		c.HTML(http.StatusNotFound, "error", gin.H{
			"Title": "Project Not Found",
			"User":  session,
			"Error": "The requested project could not be found.",
		})
		return
	}

	userID, err := uuid.Parse(session.UserID)
	if err != nil || project.OwnerID != userID {
		c.HTML(http.StatusForbidden, "error", gin.H{
			"Title": "Access Denied",
			"User":  session,
			"Error": "You don't have permission to delete this project.",
		})
		return
	}

	// Delete the project (this will cascade delete score settings and excluded extensions)
	if err := h.projectService.DeleteProject(projectID); err != nil {
		c.HTML(http.StatusInternalServerError, "error", gin.H{
			"Title": "Error",
			"User":  session,
			"Error": "Failed to delete project: " + err.Error(),
		})
		return
	}

	c.Redirect(http.StatusFound, "/dashboard")
}

// FetchRepositories fetches repositories from GitHub and associates them with the project
func (h *ProjectHandler) FetchRepositories(c *gin.Context) {
	session := middleware.GetSession(c)
	if session == nil {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	projectID := c.Param("id")

	// Validate project ownership
	project, err := h.projectService.GetProjectByID(projectID)
	if err != nil {
		c.HTML(http.StatusNotFound, "error", gin.H{
			"Title": "Project Not Found",
			"User":  session,
			"Error": "The requested project could not be found.",
		})
		return
	}

	userID, err := uuid.Parse(session.UserID)
	if err != nil || project.OwnerID != userID {
		c.HTML(http.StatusForbidden, "error", gin.H{
			"Title": "Access Denied",
			"User":  session,
			"Error": "You don't have permission to modify this project.",
		})
		return
	}

	// Get GitHub token from database
	user, err := h.userService.GetUserByID(session.UserID)
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error", gin.H{
			"Title": "Error",
			"User":  session,
			"Error": "Failed to retrieve user data.",
		})
		return
	}

	if user.GitHubAccessToken == "" {
		c.HTML(http.StatusBadRequest, "error", gin.H{
			"Title": "GitHub Token Required",
			"User":  session,
			"Error": "GitHub token not found. Please login with GitHub again.",
		})
		return
	}

	// Fetch repositories from GitHub
	if err := h.githubRepoService.FetchUserRepositories(projectID, user.GitHubAccessToken); err != nil {
		c.HTML(http.StatusInternalServerError, "error", gin.H{
			"Title": "Error Fetching Repositories",
			"User":  session,
			"Error": "Failed to fetch repositories from GitHub: " + err.Error(),
		})
		return
	}

	c.Redirect(http.StatusFound, "/projects/"+projectID)
}

// AddExcludedFolder handles adding excluded folders
func (h *ProjectHandler) AddExcludedFolder(c *gin.Context) {
	session := middleware.GetSession(c)
	if session == nil {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	projectID := c.Param("id")
	folderPath := strings.TrimSpace(c.PostForm("folder_path"))

	if folderPath == "" {
		c.HTML(http.StatusBadRequest, "error", gin.H{
			"Title": "Error",
			"User":  session,
			"Error": "Folder path is required",
		})
		return
	}

	// Validate project ownership
	project, err := h.projectService.GetProjectByID(projectID)
	if err != nil {
		c.HTML(http.StatusNotFound, "error", gin.H{
			"Title": "Project Not Found",
			"User":  session,
			"Error": "The requested project could not be found.",
		})
		return
	}

	userID, err := uuid.Parse(session.UserID)
	if err != nil || project.OwnerID != userID {
		c.HTML(http.StatusForbidden, "error", gin.H{
			"Title": "Access Denied",
			"User":  session,
			"Error": "You don't have permission to modify this project.",
		})
		return
	}

	_, err = h.excludedFolderService.CreateExcludedFolder(projectID, folderPath)
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error", gin.H{
			"Title": "Error",
			"User":  session,
			"Error": "Failed to add excluded folder: " + err.Error(),
		})
		return
	}

	c.Redirect(http.StatusFound, "/projects/"+projectID+"/settings")
}

// DeleteExcludedFolder handles removing excluded folders
func (h *ProjectHandler) DeleteExcludedFolder(c *gin.Context) {
	session := middleware.GetSession(c)
	if session == nil {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	projectID := c.Param("id")
	folderID := c.Param("folder_id")

	// Validate project ownership
	project, err := h.projectService.GetProjectByID(projectID)
	if err != nil {
		c.HTML(http.StatusNotFound, "error", gin.H{
			"Title": "Project Not Found",
			"User":  session,
			"Error": "The requested project could not be found.",
		})
		return
	}

	userID, err := uuid.Parse(session.UserID)
	if err != nil || project.OwnerID != userID {
		c.HTML(http.StatusForbidden, "error", gin.H{
			"Title": "Access Denied",
			"User":  session,
			"Error": "You don't have permission to modify this project.",
		})
		return
	}

	if err := h.excludedFolderService.DeleteExcludedFolder(folderID); err != nil {
		c.HTML(http.StatusInternalServerError, "error", gin.H{
			"Title": "Error",
			"User":  session,
			"Error": "Failed to delete excluded folder: " + err.Error(),
		})
		return
	}

	c.Redirect(http.StatusFound, "/projects/"+projectID+"/settings")
}

// ToggleRepositoryTracking toggles the tracking status of a project repository
func (h *ProjectHandler) ToggleRepositoryTracking(c *gin.Context) {
	session := middleware.GetSession(c)
	if session == nil {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	projectID := c.Param("id")
	repositoryID := c.Param("repository_id")

	// Validate project ownership
	project, err := h.projectService.GetProjectByID(projectID)
	if err != nil {
		c.HTML(http.StatusNotFound, "error", gin.H{
			"Title": "Project Not Found",
			"User":  session,
			"Error": "The requested project could not be found.",
		})
		return
	}

	userID, err := uuid.Parse(session.UserID)
	if err != nil || project.OwnerID != userID {
		c.HTML(http.StatusForbidden, "error", gin.H{
			"Title": "Access Denied",
			"User":  session,
			"Error": "You don't have permission to modify this project.",
		})
		return
	}

	// Get the project repository
	projectRepo, err := h.githubRepoService.GetProjectRepository(repositoryID)
	if err != nil {
		c.HTML(http.StatusNotFound, "error", gin.H{
			"Title": "Repository Not Found",
			"User":  session,
			"Error": "The requested repository could not be found.",
		})
		return
	}

	// Verify the repository belongs to the project
	if projectRepo.ProjectID != projectID {
		c.HTML(http.StatusForbidden, "error", gin.H{
			"Title": "Access Denied",
			"User":  session,
			"Error": "This repository does not belong to the specified project.",
		})
		return
	}

	// Toggle the tracking status
	projectRepo.IsTracked = !projectRepo.IsTracked

	if err := h.githubRepoService.UpdateProjectRepository(projectRepo); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to update repository tracking status: " + err.Error(),
		})
		return
	}

	// Return success response
	message := "Repository untracked successfully"
	if projectRepo.IsTracked {
		message = "Repository tracked successfully"
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": message,
	})
}

// CreateCloneJob creates a clone job for a specific repository
func (h *ProjectHandler) CreateCloneJob(c *gin.Context) {
	session := middleware.GetSession(c)
	if session == nil {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	projectID := c.Param("id")
	projectRepositoryID := c.Param("repository_id")

	// Validate project ID
	if _, err := uuid.Parse(projectID); err != nil {
		c.HTML(http.StatusBadRequest, "error", gin.H{
			"Title": "Error",
			"User":  session,
			"Error": "Invalid project ID",
		})
		return
	}

	// Validate project repository ID
	if _, err := uuid.Parse(projectRepositoryID); err != nil {
		c.HTML(http.StatusBadRequest, "error", gin.H{
			"Title": "Error",
			"User":  session,
			"Error": "Invalid project repository ID",
		})
		return
	}

	// Check if user owns the project
	project, err := h.projectService.GetProjectByID(projectID)
	if err != nil {
		c.HTML(http.StatusNotFound, "error", gin.H{
			"Title": "Error",
			"User":  session,
			"Error": "Project not found",
		})
		return
	}

	userID, err := uuid.Parse(session.UserID)
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error", gin.H{
			"Title": "Error",
			"User":  session,
			"Error": "Invalid user session",
		})
		return
	}

	if project.OwnerID != userID {
		c.HTML(http.StatusForbidden, "error", gin.H{
			"Title": "Error",
			"User":  session,
			"Error": "Access denied",
		})
		return
	}

	// Create the clone and commit jobs
	err = h.jobService.CreateCloneAndCommitJobs(projectID, projectRepositoryID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to create clone and commit jobs: " + err.Error(),
		})
		return
	}

	// Return success response
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Clone and commit jobs created successfully",
	})
}

// CreateFetchGithubJob creates only a pull request job for a specific repository
func (h *ProjectHandler) CreateFetchGithubJob(c *gin.Context) {
	session := middleware.GetSession(c)
	if session == nil {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	projectID := c.Param("id")
	projectRepositoryID := c.Param("repository_id")

	// Validate project ID
	if _, err := uuid.Parse(projectID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid project ID",
		})
		return
	}

	// Validate project repository ID
	if projectRepositoryID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Project repository ID is required",
		})
		return
	}

	if _, err := uuid.Parse(projectRepositoryID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid project repository ID",
		})
		return
	}

	// Check if user owns the project
	project, err := h.projectService.GetProjectByID(projectID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"message": "Project not found",
		})
		return
	}

	userID, err := uuid.Parse(session.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Invalid user session",
		})
		return
	}

	if project.OwnerID != userID {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"message": "Access denied",
		})
		return
	}

	// Create only the pull_request job
	err = h.jobService.CreatePullRequestJob(projectID, projectRepositoryID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to create pull request job: " + err.Error(),
		})
		return
	}

	// Return success response
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Pull request job created successfully",
	})
}

// CreateAnalyzeJobs creates only a stats job for a specific repository
func (h *ProjectHandler) CreateAnalyzeJobs(c *gin.Context) {
	session := middleware.GetSession(c)
	if session == nil {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	projectID := c.Param("id")
	projectRepositoryID := c.Param("repository_id") // This might be empty for project-level calls

	// Validate project ID
	if _, err := uuid.Parse(projectID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid project ID",
		})
		return
	}

	// Validate project repository ID only if it's provided (for repository-level calls)
	if projectRepositoryID != "" {
		if _, err := uuid.Parse(projectRepositoryID); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": "Invalid project repository ID",
			})
			return
		}
	}

	// Check if user owns the project
	project, err := h.projectService.GetProjectByID(projectID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"message": "Project not found",
		})
		return
	}

	userID, err := uuid.Parse(session.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Invalid user session",
		})
		return
	}

	if project.OwnerID != userID {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"message": "Access denied",
		})
		return
	}

	// Create only the stats job
	if projectRepositoryID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Project repository ID is required",
		})
		return
	}

	// Check if there are GitHub people with email associations for this project
	emailAssociations, err := h.githubPersonEmailService.GetGitHubPersonEmailsByProjectID(projectID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to check GitHub people associations: " + err.Error(),
		})
		return
	}

	if len(emailAssociations) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "You need at least 1 GitHub account with an email associated to run analyze",
		})
		return
	}

	err = h.jobService.CreateStatsJob(projectID, projectRepositoryID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to create stats job: " + err.Error(),
		})
		return
	}

	// Return success response
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Stats job created successfully",
	})
}

// CloneAllRepositories creates clone jobs for all tracked repositories in a project
func (h *ProjectHandler) CloneAllRepositories(c *gin.Context) {
	session := middleware.GetSession(c)
	if session == nil {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	projectID := c.Param("id")

	// Validate project ID
	if _, err := uuid.Parse(projectID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid project ID",
		})
		return
	}

	// Check if user owns the project
	project, err := h.projectService.GetProjectByID(projectID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"message": "Project not found",
		})
		return
	}

	userID, err := uuid.Parse(session.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Invalid user session",
		})
		return
	}

	if project.OwnerID != userID {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"message": "Access denied",
		})
		return
	}

	// Get all project repositories for this project
	projectRepos, err := h.githubRepoService.GetProjectRepositories(projectID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to get project repositories: " + err.Error(),
		})
		return
	}

	createdCount := 0
	for _, projectRepo := range projectRepos {
		// Only create jobs for tracked repositories
		if !projectRepo.IsTracked {
			continue
		}

		// Create clone and commit jobs for this repository
		if err := h.jobService.CreateCloneAndCommitJobs(projectID, projectRepo.ID); err != nil {
			continue // Skip if we can't create the job
		}

		createdCount++
	}

	// Return success response
	message := fmt.Sprintf("Created %d clone and commit job pairs for tracked repositories", createdCount)
	if createdCount == 0 {
		message = "No jobs created (no tracked repositories)"
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": message,
	})
}

// TrackAllRepositories tracks all repositories in a project
func (h *ProjectHandler) TrackAllRepositories(c *gin.Context) {
	session := middleware.GetSession(c)
	if session == nil {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	projectID := c.Param("id")

	// Validate project ID
	if _, err := uuid.Parse(projectID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid project ID",
		})
		return
	}

	// Check if user owns the project
	project, err := h.projectService.GetProjectByID(projectID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"message": "Project not found",
		})
		return
	}

	userID, err := uuid.Parse(session.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Invalid user session",
		})
		return
	}

	if project.OwnerID != userID {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"message": "Access denied",
		})
		return
	}

	// Get all project repositories for this project
	projectRepos, err := h.githubRepoService.GetProjectRepositories(projectID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to get project repositories: " + err.Error(),
		})
		return
	}

	trackedCount := 0
	for _, projectRepo := range projectRepos {
		// Skip if already tracked
		if projectRepo.IsTracked {
			continue
		}

		// Track this repository
		projectRepo.IsTracked = true
		if err := h.githubRepoService.UpdateProjectRepository(projectRepo); err != nil {
			continue // Skip if we can't track the repository
		}

		trackedCount++
	}

	// Return success response
	message := fmt.Sprintf("Tracked %d repositories", trackedCount)
	if trackedCount == 0 {
		message = "No repositories to track (all repositories are already tracked)"
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": message,
	})
}

// ViewProjectEmails displays the emails page for a project
func (h *ProjectHandler) ViewProjectEmails(c *gin.Context) {
	session := middleware.GetSession(c)
	if session == nil {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	projectID := c.Param("id")
	if projectID == "" {
		c.HTML(http.StatusBadRequest, "error", gin.H{
			"Title": "Error",
			"User":  session,
			"Error": "Project ID is required",
		})
		return
	}

	project, err := h.projectService.GetProjectByID(projectID)
	if err != nil {
		c.HTML(http.StatusNotFound, "error", gin.H{
			"Title": "Project Not Found",
			"User":  session,
			"Error": "The requested project could not be found.",
		})
		return
	}

	// Check if the project belongs to the current owner
	userID, err := uuid.Parse(session.UserID)
	if err != nil || project.OwnerID != userID {
		c.HTML(http.StatusForbidden, "error", gin.H{
			"Title": "Access Denied",
			"User":  session,
			"Error": "You don't have permission to view this project.",
		})
		return
	}

	// Get merged emails for this project
	mergedEmails, err := h.emailMergeService.GetMergedEmailsForProject(projectID)
	if err != nil {
		mergedEmails = make(map[string]string)
	}

	// Get email statistics (with merges applied)
	emails, err := h.commitRepo.GetEmailStatsByProjectID(projectID, mergedEmails)
	if err != nil {
		emails = []*models.EmailStats{}
	}

	// Create sorted emails for each row's dropdown
	emailSortedEmails := make(map[string][]string)

	// Create a set of target emails (emails that have other emails merged into them)
	targetEmails := make(map[string]bool)
	for _, merge := range mergedEmails {
		targetEmails[merge] = true
	}

	for _, email := range emails {
		// Extract all email addresses for this dropdown (excluding current email)
		var emailAddresses []string
		for _, e := range emails {
			if e.Email != email.Email {
				emailAddresses = append(emailAddresses, e.Email)
			}
		}

		// Sort emails by similarity to current email (using email username part)
		emailUsername := strings.Split(email.Email, "@")[0]
		sortedEmails := h.textSimilarityService.SortEmailsBySimilarity(emailAddresses, emailUsername)

		// Extract just the email addresses in sorted order, excluding target emails
		var sortedEmailAddresses []string
		for _, emailSimilarity := range sortedEmails {
			// Skip if this email is a target email (has other emails merged into it)
			if !targetEmails[emailSimilarity.Email] {
				sortedEmailAddresses = append(sortedEmailAddresses, emailSimilarity.Email)
			}
		}

		emailSortedEmails[email.Email] = sortedEmailAddresses
	}

	// Create a map of merged emails to exclude them from dropdowns
	associatedEmails := make(map[string]bool)
	for sourceEmail := range mergedEmails {
		associatedEmails[sourceEmail] = true
	}

	data := gin.H{
		"Title":             "Emails - " + project.Name,
		"User":              session,
		"Project":           project,
		"Emails":            emails,
		"EmailSortedEmails": emailSortedEmails,
		"AssociatedEmails":  associatedEmails,
	}

	c.HTML(http.StatusOK, "project_emails", data)
}

// ViewProjectPeople displays the people page for a project
func (h *ProjectHandler) ViewProjectPeople(c *gin.Context) {
	session := middleware.GetSession(c)
	if session == nil {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	projectID := c.Param("id")
	if projectID == "" {
		c.HTML(http.StatusBadRequest, "error", gin.H{
			"Title": "Error",
			"User":  session,
			"Error": "Project ID is required",
		})
		return
	}

	project, err := h.projectService.GetProjectByID(projectID)
	if err != nil {
		c.HTML(http.StatusNotFound, "error", gin.H{
			"Title": "Project Not Found",
			"User":  session,
			"Error": "The requested project could not be found.",
		})
		return
	}

	// Check if the project belongs to the current owner
	userID, err := uuid.Parse(session.UserID)
	if err != nil || project.OwnerID != userID {
		c.HTML(http.StatusForbidden, "error", gin.H{
			"Title": "Access Denied",
			"User":  session,
			"Error": "You don't have permission to view this project.",
		})
		return
	}

	// Get GitHub people
	people, err := h.githubPersonRepo.GetByProjectID(projectID)
	if err != nil {
		people = []*models.GithubPerson{}
	}

	// Get email associations for this project
	emailAssociations, err := h.githubPersonEmailService.GetGitHubPersonEmailsByProjectID(projectID)
	if err != nil {
		emailAssociations = []*models.GitHubPersonEmail{}
	}

	// Get merged emails for this project
	mergedEmails, err := h.emailMergeService.GetMergedEmailsForProject(projectID)
	if err != nil {
		mergedEmails = make(map[string]string)
	}

	// Get all emails for dropdown (with merges applied)
	emails, err := h.commitRepo.GetEmailStatsByProjectID(projectID, mergedEmails)
	if err != nil {
		emails = []*models.EmailStats{}
	}

	// Filter out merged emails from the dropdown options
	var filteredEmails []*models.EmailStats
	for _, email := range emails {
		// Check if this email is a source email (merged into another)
		if _, isMerged := mergedEmails[email.Email]; !isMerged {
			filteredEmails = append(filteredEmails, email)
		}
	}

	// Create a map of GitHub person ID to associated email
	personEmailMap := make(map[string]string)
	associatedEmails := make(map[string]bool)

	for _, assoc := range emailAssociations {
		// Get the person to get their email
		person, err := h.personRepo.GetByID(assoc.PersonID)
		if err == nil && person != nil {
			personEmailMap[assoc.GitHubPersonID] = person.PrimaryEmail
			associatedEmails[person.PrimaryEmail] = true
		}
	}

	// Create a map of GitHub person ID to sorted emails for dropdown
	personSortedEmails := make(map[string][]string)
	for _, person := range people {
		// Extract email addresses from filtered emails
		var emailAddresses []string
		for _, email := range filteredEmails {
			emailAddresses = append(emailAddresses, email.Email)
		}

		// Sort emails by similarity to GitHub username
		sortedEmails := h.textSimilarityService.SortEmailsBySimilarity(emailAddresses, person.Username)

		// Extract just the email addresses in sorted order, excluding emails associated with any GitHub person
		var sortedEmailAddresses []string
		for _, emailSimilarity := range sortedEmails {
			// Skip if this email is already associated with any GitHub person
			if !associatedEmails[emailSimilarity.Email] {
				sortedEmailAddresses = append(sortedEmailAddresses, emailSimilarity.Email)
			}
		}

		personSortedEmails[person.ID] = sortedEmailAddresses
	}

	data := gin.H{
		"Title":              "People - " + project.Name,
		"User":               session,
		"Project":            project,
		"People":             people,
		"Emails":             filteredEmails,
		"PersonEmailMap":     personEmailMap,
		"AssociatedEmails":   associatedEmails,
		"PersonSortedEmails": personSortedEmails,
	}

	c.HTML(http.StatusOK, "project_people", data)
}

// ViewProjectReports displays the reports page for a project
func (h *ProjectHandler) ViewProjectReports(c *gin.Context) {
	session := middleware.GetSession(c)
	if session == nil {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	projectID := c.Param("id")
	if projectID == "" {
		c.HTML(http.StatusBadRequest, "error", gin.H{
			"Title": "Error",
			"User":  session,
			"Error": "Project ID is required",
		})
		return
	}

	project, err := h.projectService.GetProjectByID(projectID)
	if err != nil {
		c.HTML(http.StatusNotFound, "error", gin.H{
			"Title": "Project Not Found",
			"User":  session,
			"Error": "The requested project could not be found.",
		})
		return
	}

	// Check if the project belongs to the current owner
	userID, err := uuid.Parse(session.UserID)
	if err != nil || project.OwnerID != userID {
		c.HTML(http.StatusForbidden, "error", gin.H{
			"Title": "Access Denied",
			"User":  session,
			"Error": "You don't have permission to view this project.",
		})
		return
	}

	// Get all-time statistics for all GitHub people in this project
	allTimeStats, err := h.peopleStatsService.GetAllTimeStatisticsByProject(projectID)
	if err != nil {
		allTimeStats = []*models.GitHubPersonStats{}
	}

	data := gin.H{
		"Title":        "Reports - " + project.Name,
		"User":         session,
		"Project":      project,
		"AllTimeStats": allTimeStats,
	}

	c.HTML(http.StatusOK, "project_reports", data)
}

// ViewProjectReportsDaily displays the daily reports page for a project
func (h *ProjectHandler) ViewProjectReportsDaily(c *gin.Context) {
	session := middleware.GetSession(c)
	if session == nil {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	projectID := c.Param("id")
	if projectID == "" {
		c.HTML(http.StatusBadRequest, "error", gin.H{
			"Title": "Error",
			"User":  session,
			"Error": "Project ID is required",
		})
		return
	}

	project, err := h.projectService.GetProjectByID(projectID)
	if err != nil {
		c.HTML(http.StatusNotFound, "error", gin.H{
			"Title": "Project Not Found",
			"User":  session,
			"Error": "The requested project could not be found.",
		})
		return
	}

	// Check if the project belongs to the current owner
	userID, err := uuid.Parse(session.UserID)
	if err != nil || project.OwnerID != userID {
		c.HTML(http.StatusForbidden, "error", gin.H{
			"Title": "Access Denied",
			"User":  session,
			"Error": "You don't have permission to view this project.",
		})
		return
	}

	// Get available days for this project
	availableDays, err := h.peopleStatsService.GetAvailableDaysForProject(projectID)
	if err != nil {
		availableDays = []string{time.Now().Format("2006-01-02")}
	}

	// Get selected day from query parameter, default to the first available day (most recent)
	selectedDay := time.Now().Format("2006-01-02")
	if dayStr := c.Query("day"); dayStr != "" {
		selectedDay = dayStr
	} else if len(availableDays) > 0 {
		// Use the first (most recent) available day as default
		selectedDay = availableDays[0]
	}

	// Parse selected day to get date
	selectedDate, err := time.Parse("2006-01-02", selectedDay)
	if err != nil {
		selectedDate = time.Now()
	}

	// Get daily statistics for the selected day
	dailyStats, err := h.peopleStatsService.GetDailyStatisticsByProject(projectID, selectedDate)
	if err != nil {
		dailyStats = []*models.GitHubPersonStats{}
	}

	data := gin.H{
		"Title":       "Daily Reports - " + project.Name,
		"User":        session,
		"Project":     project,
		"Days":        availableDays,
		"SelectedDay": selectedDay,
		"DailyStats":  dailyStats,
	}

	c.HTML(http.StatusOK, "project_reports_daily", data)
}

// ViewProjectReportsWeekly displays the weekly reports page for a project
func (h *ProjectHandler) ViewProjectReportsWeekly(c *gin.Context) {
	session := middleware.GetSession(c)
	if session == nil {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	projectID := c.Param("id")
	if projectID == "" {
		c.HTML(http.StatusBadRequest, "error", gin.H{
			"Title": "Error",
			"User":  session,
			"Error": "Project ID is required",
		})
		return
	}

	project, err := h.projectService.GetProjectByID(projectID)
	if err != nil {
		c.HTML(http.StatusNotFound, "error", gin.H{
			"Title": "Project Not Found",
			"User":  session,
			"Error": "The requested project could not be found.",
		})
		return
	}

	// Check if the project belongs to the current owner
	userID, err := uuid.Parse(session.UserID)
	if err != nil || project.OwnerID != userID {
		c.HTML(http.StatusForbidden, "error", gin.H{
			"Title": "Access Denied",
			"User":  session,
			"Error": "You don't have permission to view this project.",
		})
		return
	}

	// Get available weeks for this project
	availableWeeks, err := h.peopleStatsService.GetAvailableWeeksForProject(projectID)
	if err != nil {
		now := time.Now()
		year, week := now.ISOWeek()
		availableWeeks = []string{fmt.Sprintf("%d-W%02d", year, week)}
	}

	// Get selected week from query parameter, default to the first available week (most recent)
	now := time.Now()
	year, week := now.ISOWeek()
	selectedWeek := fmt.Sprintf("%d-W%02d", year, week)
	if weekStr := c.Query("week"); weekStr != "" {
		selectedWeek = weekStr
	} else if len(availableWeeks) > 0 {
		// Use the first (most recent) available week as default
		selectedWeek = availableWeeks[0]
	}

	// Parse selected week to get year and week
	var selectedYear, selectedWeekInt int
	if _, err := fmt.Sscanf(selectedWeek, "%d-W%d", &selectedYear, &selectedWeekInt); err != nil {
		selectedYear = year
		selectedWeekInt = week
	}

	// Get weekly statistics for the selected week
	weeklyStats, err := h.peopleStatsService.GetWeeklyStatisticsByProject(projectID, selectedYear, selectedWeekInt)
	if err != nil {
		weeklyStats = []*models.GitHubPersonStats{}
	}

	// Calculate date range for the selected week
	weekDateRange := h.calculateWeekDateRange(selectedYear, selectedWeekInt)

	data := gin.H{
		"Title":         "Weekly Reports - " + project.Name,
		"User":          session,
		"Project":       project,
		"Weeks":         availableWeeks,
		"SelectedWeek":  selectedWeek,
		"WeeklyStats":   weeklyStats,
		"WeekDateRange": weekDateRange,
	}

	c.HTML(http.StatusOK, "project_reports_weekly", data)
}

// calculateWeekDateRange calculates the start and end date for a given year and week
func (h *ProjectHandler) calculateWeekDateRange(year, week int) string {
	// Create a date for January 1st of the given year
	jan1 := time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC)

	// Find the first Monday of the year
	daysUntilMonday := (8 - int(jan1.Weekday())) % 7
	if daysUntilMonday == 0 {
		daysUntilMonday = 7
	}
	firstMonday := jan1.AddDate(0, 0, daysUntilMonday-1)

	// Calculate the start of the specified week
	weekStart := firstMonday.AddDate(0, 0, (week-1)*7)

	// Calculate the end of the week (Sunday)
	weekEnd := weekStart.AddDate(0, 0, 6)

	// Format the date range
	startStr := weekStart.Format("02.01.06")
	endStr := weekEnd.Format("02.01.06")

	return fmt.Sprintf("%s - %s", startStr, endStr)
}

// ViewProjectReportsMonthly displays the monthly reports page for a project
func (h *ProjectHandler) ViewProjectReportsMonthly(c *gin.Context) {
	session := middleware.GetSession(c)
	if session == nil {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	projectID := c.Param("id")
	if projectID == "" {
		c.HTML(http.StatusBadRequest, "error", gin.H{
			"Title": "Error",
			"User":  session,
			"Error": "Project ID is required",
		})
		return
	}

	project, err := h.projectService.GetProjectByID(projectID)
	if err != nil {
		c.HTML(http.StatusNotFound, "error", gin.H{
			"Title": "Project Not Found",
			"User":  session,
			"Error": "The requested project could not be found.",
		})
		return
	}

	// Check if the project belongs to the current owner
	userID, err := uuid.Parse(session.UserID)
	if err != nil || project.OwnerID != userID {
		c.HTML(http.StatusForbidden, "error", gin.H{
			"Title": "Access Denied",
			"User":  session,
			"Error": "You don't have permission to view this project.",
		})
		return
	}

	// Get available months for this project
	availableMonths, err := h.peopleStatsService.GetAvailableMonthsForProject(projectID)
	if err != nil {
		availableMonths = []string{fmt.Sprintf("%d-%02d", time.Now().Year(), time.Now().Month())}
	}

	// Get selected month from query parameter, default to the first available month (most recent)
	selectedMonth := fmt.Sprintf("%d-%02d", time.Now().Year(), time.Now().Month())
	if monthStr := c.Query("month"); monthStr != "" {
		selectedMonth = monthStr
	} else if len(availableMonths) > 0 {
		// Use the first (most recent) available month as default
		selectedMonth = availableMonths[0]
	}

	// Parse selected month to get year and month
	var selectedYear, selectedMonthInt int
	if _, err := fmt.Sscanf(selectedMonth, "%d-%d", &selectedYear, &selectedMonthInt); err != nil {
		selectedYear = time.Now().Year()
		selectedMonthInt = int(time.Now().Month())
	}

	// Get monthly statistics for the selected month
	monthlyStats, err := h.peopleStatsService.GetMonthlyStatisticsByProject(projectID, selectedYear, selectedMonthInt)
	if err != nil {
		monthlyStats = []*models.GitHubPersonStats{}
	}

	data := gin.H{
		"Title":         "Monthly Reports - " + project.Name,
		"User":          session,
		"Project":       project,
		"Months":        availableMonths,
		"SelectedMonth": selectedMonth,
		"MonthlyStats":  monthlyStats,
	}

	c.HTML(http.StatusOK, "project_reports_monthly", data)
}

// ViewProjectReportsYearly displays the yearly reports page for a project
func (h *ProjectHandler) ViewProjectReportsYearly(c *gin.Context) {
	session := middleware.GetSession(c)
	if session == nil {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	projectID := c.Param("id")
	if projectID == "" {
		c.HTML(http.StatusBadRequest, "error", gin.H{
			"Title": "Error",
			"User":  session,
			"Error": "Project ID is required",
		})
		return
	}

	project, err := h.projectService.GetProjectByID(projectID)
	if err != nil {
		c.HTML(http.StatusNotFound, "error", gin.H{
			"Title": "Project Not Found",
			"User":  session,
			"Error": "The requested project could not be found.",
		})
		return
	}

	// Check if the project belongs to the current owner
	userID, err := uuid.Parse(session.UserID)
	if err != nil || project.OwnerID != userID {
		c.HTML(http.StatusForbidden, "error", gin.H{
			"Title": "Access Denied",
			"User":  session,
			"Error": "You don't have permission to view this project.",
		})
		return
	}

	// Get available years for this project
	availableYears, err := h.peopleStatsService.GetAvailableYearsForProject(projectID)
	if err != nil {
		availableYears = []int{time.Now().Year()}
	}

	// Get selected year from query parameter, default to the first available year (most recent)
	selectedYear := time.Now().Year()
	if yearStr := c.Query("year"); yearStr != "" {
		if year, err := strconv.Atoi(yearStr); err == nil {
			selectedYear = year
		}
	} else if len(availableYears) > 0 {
		// Use the first (most recent) available year as default
		selectedYear = availableYears[0]
	}

	// Get yearly statistics for the selected year
	yearlyStats, err := h.peopleStatsService.GetYearlyStatisticsByProject(projectID, selectedYear)
	if err != nil {
		yearlyStats = []*models.GitHubPersonStats{}
	}

	data := gin.H{
		"Title":        "Yearly Reports - " + project.Name,
		"User":         session,
		"Project":      project,
		"Years":        availableYears,
		"SelectedYear": selectedYear,
		"YearlyStats":  yearlyStats,
	}

	c.HTML(http.StatusOK, "project_reports_yearly", data)
}

// CreateEmailMerge creates an email merge
func (h *ProjectHandler) CreateEmailMerge(c *gin.Context) {
	session := middleware.GetSession(c)
	if session == nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "Authentication required",
		})
		return
	}

	projectID := c.Param("id")
	sourceEmail := c.PostForm("source_email")
	targetEmail := c.PostForm("target_email")

	if projectID == "" || sourceEmail == "" || targetEmail == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Project ID, source email, and target email are required",
		})
		return
	}

	// Validate project ownership
	project, err := h.projectService.GetProjectByID(projectID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"message": "Project not found",
		})
		return
	}

	userID, err := uuid.Parse(session.UserID)
	if err != nil || project.OwnerID != userID {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"message": "Access denied",
		})
		return
	}

	// Check if source email is already merged
	isMerged, _, err := h.emailMergeService.IsEmailMerged(projectID, sourceEmail)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to check email merge status: " + err.Error(),
		})
		return
	}

	if isMerged {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Email is already merged into another email",
		})
		return
	}

	// Create the email merge (swap source and target so the current row stays)
	err = h.emailMergeService.CreateEmailMerge(projectID, targetEmail, sourceEmail)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to create email merge: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Email merge created successfully",
	})
}

// DetachEmailMerge detaches all emails merged into a target email
func (h *ProjectHandler) DetachEmailMerge(c *gin.Context) {
	session := middleware.GetSession(c)
	if session == nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "Authentication required",
		})
		return
	}

	projectID := c.Param("id")

	var request struct {
		TargetEmail string `json:"target_email"`
		SourceEmail string `json:"source_email"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid request body",
		})
		return
	}

	if projectID == "" || request.TargetEmail == "" || request.SourceEmail == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Project ID, target email, and source email are required",
		})
		return
	}

	// Validate project ownership
	project, err := h.projectService.GetProjectByID(projectID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"message": "Project not found",
		})
		return
	}

	userID, err := uuid.Parse(session.UserID)
	if err != nil || project.OwnerID != userID {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"message": "Access denied",
		})
		return
	}

	// Detach the specific email merge
	err = h.emailMergeService.DeleteEmailMergeBySourceAndTarget(projectID, request.SourceEmail, request.TargetEmail)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to detach email merge: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Email merge detached successfully",
	})
}

// CreateGitHubPersonEmailAssociation creates an association between a GitHub person and an email
func (h *ProjectHandler) CreateGitHubPersonEmailAssociation(c *gin.Context) {
	session := middleware.GetSession(c)
	if session == nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "Authentication required",
		})
		return
	}

	projectID := c.Param("id")
	githubPersonID := c.PostForm("github_person_id")
	email := c.PostForm("email")

	if projectID == "" || githubPersonID == "" || email == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Project ID, GitHub person ID, and email are required",
		})
		return
	}

	// Validate project ownership
	project, err := h.projectService.GetProjectByID(projectID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"message": "Project not found",
		})
		return
	}

	userID, err := uuid.Parse(session.UserID)
	if err != nil || project.OwnerID != userID {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"message": "Access denied",
		})
		return
	}

	// Get or create person by email
	person, err := h.personRepo.GetOrCreateByEmail("", email) // Empty name, will be updated if needed
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Error: " + err.Error(),
		})
		return
	}

	// Create the association
	_, err = h.githubPersonEmailService.CreateGitHubPersonEmail(projectID, githubPersonID, person.ID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Error: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Email associated successfully",
	})
}

// DeleteGitHubPersonEmailAssociation deletes an association between a GitHub person and an email
func (h *ProjectHandler) DeleteGitHubPersonEmailAssociation(c *gin.Context) {
	session := middleware.GetSession(c)
	if session == nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "Authentication required",
		})
		return
	}

	projectID := c.Param("id")
	email := c.PostForm("email")

	if projectID == "" || email == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Project ID and email are required",
		})
		return
	}

	// Validate project ownership
	project, err := h.projectService.GetProjectByID(projectID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"message": "Project not found",
		})
		return
	}

	userID, err := uuid.Parse(session.UserID)
	if err != nil || project.OwnerID != userID {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"message": "Access denied",
		})
		return
	}

	// Get person by email
	person, err := h.personRepo.GetByEmail(email)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Person not found",
		})
		return
	}

	// Delete the association
	err = h.githubPersonEmailService.DeleteGitHubPersonEmailByPersonID(projectID, person.ID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Error: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Email association removed successfully",
	})
}

// FetchAllRepositories creates pull_request jobs for all tracked repositories in a project
func (h *ProjectHandler) FetchAllRepositories(c *gin.Context) {
	session := middleware.GetSession(c)
	if session == nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "Authentication required",
		})
		return
	}

	projectID := c.Param("id")
	if projectID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Project ID is required",
		})
		return
	}

	// Validate project ownership
	project, err := h.projectService.GetProjectByID(projectID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"message": "Project not found",
		})
		return
	}

	userID, err := uuid.Parse(session.UserID)
	if err != nil || project.OwnerID != userID {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"message": "Access denied",
		})
		return
	}

	// Get all tracked repositories for this project
	repositories, err := h.githubRepoService.GetProjectRepositories(projectID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Error retrieving repositories: " + err.Error(),
		})
		return
	}

	// Filter to only tracked repositories
	var trackedRepos []*models.ProjectRepository
	for _, repo := range repositories {
		if repo.IsTracked {
			trackedRepos = append(trackedRepos, repo)
		}
	}

	if len(trackedRepos) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "No tracked repositories found",
		})
		return
	}

	// Create pull_request jobs for all tracked repositories
	createdJobs := 0
	for _, repo := range trackedRepos {
		err := h.jobService.CreatePullRequestJob(projectID, repo.ID)
		if err != nil {
			// Continue with other repos even if one fails
			continue
		}
		createdJobs++
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": fmt.Sprintf("Created %d fetch jobs for tracked repositories", createdJobs),
	})
}

// AnalyzeAllRepositories creates stats jobs for all tracked repositories in a project
func (h *ProjectHandler) AnalyzeAllRepositories(c *gin.Context) {
	session := middleware.GetSession(c)
	if session == nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "Authentication required",
		})
		return
	}

	projectID := c.Param("id")
	if projectID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Project ID is required",
		})
		return
	}

	// Validate project ownership
	project, err := h.projectService.GetProjectByID(projectID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"message": "Project not found",
		})
		return
	}

	userID, err := uuid.Parse(session.UserID)
	if err != nil || project.OwnerID != userID {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"message": "Access denied",
		})
		return
	}

	// Get all tracked repositories for this project
	repositories, err := h.githubRepoService.GetProjectRepositories(projectID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Error retrieving repositories: " + err.Error(),
		})
		return
	}

	// Filter to only tracked repositories
	var trackedRepos []*models.ProjectRepository
	for _, repo := range repositories {
		if repo.IsTracked {
			trackedRepos = append(trackedRepos, repo)
		}
	}

	if len(trackedRepos) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "No tracked repositories found",
		})
		return
	}

	// Create stats jobs for all tracked repositories
	createdJobs := 0
	for _, repo := range trackedRepos {
		err := h.jobService.CreateStatsJob(projectID, repo.ID)
		if err != nil {
			// Continue with other repos even if one fails
			continue
		}
		createdJobs++
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": fmt.Sprintf("Created %d analyze jobs for tracked repositories", createdJobs),
	})
}

// UpdateAllRepositories handles updating all tracked repositories in the correct order
func (h *ProjectHandler) UpdateAllRepositories(c *gin.Context) {
	session := middleware.GetSession(c)
	if session == nil {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	projectID := c.Param("id")

	// Validate project ownership
	project, err := h.projectService.GetProjectByID(projectID)
	if err != nil {
		c.HTML(http.StatusNotFound, "error", gin.H{
			"Title": "Project Not Found",
			"User":  session,
			"Error": "The requested project could not be found.",
		})
		return
	}

	userID, err := uuid.Parse(session.UserID)
	if err != nil || project.OwnerID != userID {
		c.HTML(http.StatusForbidden, "error", gin.H{
			"Title": "Access Denied",
			"User":  session,
			"Error": "You don't have permission to modify this project.",
		})
		return
	}

	// Get all tracked repositories for this project
	repositories, err := h.githubRepoService.GetProjectRepositories(projectID)
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error", gin.H{
			"Title": "Error",
			"User":  session,
			"Error": "Failed to get tracked repositories: " + err.Error(),
		})
		return
	}

	// Filter to only tracked repositories
	var trackedRepos []*models.ProjectRepository
	for _, repo := range repositories {
		if repo.IsTracked {
			trackedRepos = append(trackedRepos, repo)
		}
	}

	if len(trackedRepos) == 0 {
		c.HTML(http.StatusBadRequest, "error", gin.H{
			"Title": "No Tracked Repositories",
			"User":  session,
			"Error": "No tracked repositories found. Please track some repositories first.",
		})
		return
	}

	// Create jobs in the correct order with dependencies
	// Each repository gets its own chain: clone -> commit -> pull_request -> stats

	// Step 1: Create clone jobs for all tracked repositories
	for _, repo := range trackedRepos {
		cloneJob := models.NewJob(projectID, models.JobTypeClone)
		cloneJob.ProjectRepositoryID = &repo.ID
		if err := h.jobRepo.Create(cloneJob); err != nil {
			c.HTML(http.StatusInternalServerError, "error", gin.H{
				"Title": "Error",
				"User":  session,
				"Error": "Failed to create clone job: " + err.Error(),
			})
			return
		}
		log.Printf("Created clone job %s for repository %s", cloneJob.ID, repo.ID)

		// Step 2: Create commit job that depends on this clone job
		commitJob := models.NewJob(projectID, models.JobTypeCommit)
		commitJob.ProjectRepositoryID = &repo.ID
		commitJob.DependsOn = &cloneJob.ID
		if err := h.jobRepo.Create(commitJob); err != nil {
			c.HTML(http.StatusInternalServerError, "error", gin.H{
				"Title": "Error",
				"User":  session,
				"Error": "Failed to create commit job: " + err.Error(),
			})
			return
		}
		log.Printf("Created commit job %s for repository %s (depends on %s)", commitJob.ID, repo.ID, cloneJob.ID)

		// Step 3: Create pull request job that depends on this commit job
		pullRequestJob := models.NewJob(projectID, models.JobTypePullRequest)
		pullRequestJob.ProjectRepositoryID = &repo.ID
		pullRequestJob.DependsOn = &commitJob.ID
		if err := h.jobRepo.Create(pullRequestJob); err != nil {
			c.HTML(http.StatusInternalServerError, "error", gin.H{
				"Title": "Error",
				"User":  session,
				"Error": "Failed to create pull request job: " + err.Error(),
			})
			return
		}
		log.Printf("Created pull request job %s for repository %s (depends on %s)", pullRequestJob.ID, repo.ID, commitJob.ID)

		// Step 4: Create stats job that depends on this pull request job
		statsJob := models.NewJob(projectID, models.JobTypeStats)
		statsJob.ProjectRepositoryID = &repo.ID
		statsJob.DependsOn = &pullRequestJob.ID
		if err := h.jobRepo.Create(statsJob); err != nil {
			c.HTML(http.StatusInternalServerError, "error", gin.H{
				"Title": "Error",
				"User":  session,
				"Error": "Failed to create stats job: " + err.Error(),
			})
			return
		}
		log.Printf("Created stats job %s for repository %s (depends on %s)", statsJob.ID, repo.ID, pullRequestJob.ID)
	}

	c.Redirect(http.StatusFound, "/projects/"+projectID)
}

// UpdateProjectUpdateSettings handles updating project update settings
func (h *ProjectHandler) UpdateProjectUpdateSettings(c *gin.Context) {
	session := middleware.GetSession(c)
	if session == nil {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	projectID := c.Param("id")

	// Validate project ownership
	project, err := h.projectService.GetProjectByID(projectID)
	if err != nil {
		c.HTML(http.StatusNotFound, "error", gin.H{
			"Title": "Project Not Found",
			"User":  session,
			"Error": "The requested project could not be found.",
		})
		return
	}

	userID, err := uuid.Parse(session.UserID)
	if err != nil || project.OwnerID != userID {
		c.HTML(http.StatusForbidden, "error", gin.H{
			"Title": "Access Denied",
			"User":  session,
			"Error": "You don't have permission to modify this project.",
		})
		return
	}

	// Parse form data
	isEnabled := c.PostForm("auto_update_enabled") == "on"
	hourStr := c.PostForm("auto_update_hour")

	var hour int
	if hourStr != "" {
		if _, err := fmt.Sscanf(hourStr, "%d", &hour); err != nil {
			c.HTML(http.StatusBadRequest, "error", gin.H{
				"Title": "Invalid Input",
				"User":  session,
				"Error": "Invalid hour value. Please enter a number between 0 and 23.",
			})
			return
		}
	} else {
		hour = 0 // Default to midnight
	}

	// Validate hour
	if hour < 0 || hour > 23 {
		c.HTML(http.StatusBadRequest, "error", gin.H{
			"Title": "Invalid Input",
			"User":  session,
			"Error": "Hour must be between 0 and 23.",
		})
		return
	}

	// Update project update settings
	_, err = h.projectUpdateSettingsService.UpsertProjectUpdateSettings(projectID, isEnabled, hour)
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error", gin.H{
			"Title": "Error",
			"User":  session,
			"Error": "Failed to update project update settings: " + err.Error(),
		})
		return
	}

	c.Redirect(http.StatusFound, "/projects/"+projectID+"/settings")
}

// RetryFailedJob retries a specific failed job
func (h *ProjectHandler) RetryFailedJob(c *gin.Context) {
	session := middleware.GetSession(c)
	if session == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "Unauthorized"})
		return
	}

	jobID := c.Param("job_id")
	if jobID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Job ID is required"})
		return
	}

	// Get the job
	job, err := h.jobRepo.GetByID(jobID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "Job not found"})
		return
	}

	// Check if the job belongs to a project owned by the current user
	userID, err := uuid.Parse(session.UserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Invalid user session"})
		return
	}

	project, err := h.projectService.GetProjectByID(job.ProjectID)
	if err != nil || project.OwnerID != userID {
		c.JSON(http.StatusForbidden, gin.H{"success": false, "message": "Access denied"})
		return
	}

	// Check if the job is actually failed
	if job.Status != models.JobStatusFailed {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Job is not in failed state"})
		return
	}

	// Create a new job with the same parameters but reset status
	newJob := models.NewJob(job.ProjectID, job.JobType)
	newJob.ProjectRepositoryID = job.ProjectRepositoryID
	newJob.DependsOn = job.DependsOn

	// Save the new job
	if err := h.jobRepo.Create(newJob); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Failed to create retry job"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": fmt.Sprintf("Retry job created for %s job", job.JobType),
		"job_id":  newJob.ID,
	})
}
