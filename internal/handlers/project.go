package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/alimgiray/gscope/internal/middleware"
	"github.com/alimgiray/gscope/internal/models"
	"github.com/alimgiray/gscope/internal/repositories"
	"github.com/alimgiray/gscope/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ProjectHandler struct {
	projectService           *services.ProjectService
	userService              *services.UserService
	scoreSettingsService     *services.ScoreSettingsService
	excludedExtensionService *services.ExcludedExtensionService
	githubRepoService        *services.GitHubRepositoryService
	jobService               *services.JobService
	commitRepo               *repositories.CommitRepository
	githubPersonRepo         *repositories.GithubPersonRepository
	personRepo               *repositories.PersonRepository
	emailMergeService        *services.EmailMergeService
	githubPersonEmailService *services.GitHubPersonEmailService
	textSimilarityService    *services.TextSimilarityService
}

func NewProjectHandler(projectService *services.ProjectService, userService *services.UserService,
	scoreSettingsService *services.ScoreSettingsService, excludedExtensionService *services.ExcludedExtensionService,
	githubRepoService *services.GitHubRepositoryService, jobService *services.JobService,
	commitRepo *repositories.CommitRepository, githubPersonRepo *repositories.GithubPersonRepository,
	personRepo *repositories.PersonRepository, emailMergeService *services.EmailMergeService,
	githubPersonEmailService *services.GitHubPersonEmailService, textSimilarityService *services.TextSimilarityService) *ProjectHandler {
	return &ProjectHandler{
		projectService:           projectService,
		userService:              userService,
		scoreSettingsService:     scoreSettingsService,
		excludedExtensionService: excludedExtensionService,
		githubRepoService:        githubRepoService,
		jobService:               jobService,
		commitRepo:               commitRepo,
		githubPersonRepo:         githubPersonRepo,
		personRepo:               personRepo,
		emailMergeService:        emailMergeService,
		githubPersonEmailService: githubPersonEmailService,
		textSimilarityService:    textSimilarityService,
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

		// Check for active clone or commit jobs (commit depends on clone)
		for _, job := range allJobs {
			if job.JobType == models.JobTypeClone || job.JobType == models.JobTypeCommit {
				if job.Status == models.JobStatusPending || job.Status == models.JobStatusInProgress {
					hasActiveCloneJobs = true
					break
				}
				if job.JobType == models.JobTypeClone && job.Status == models.JobStatusFailed {
					latestCloneJobFailed = true
					if job.ErrorMessage != nil {
						latestCloneJobError = *job.ErrorMessage
					}
				}
			}
		}

		// Get project-level jobs for analyze status
		projectJobs, err := h.jobService.GetProjectJobs(projectID)
		if err != nil {
			projectJobs = []*models.Job{}
		}

		// Calculate analyze job status (project-level jobs)
		hasActiveAnalyzeJobs := false
		latestAnalyzeJobFailed := false
		latestAnalyzeJobError := ""

		// Check for active pull_request or stats jobs (stats depends on pull_request)
		for _, job := range projectJobs {
			if job.JobType == models.JobTypePullRequest || job.JobType == models.JobTypeStats {
				if job.Status == models.JobStatusPending || job.Status == models.JobStatusInProgress {
					hasActiveAnalyzeJobs = true
					break
				}
				if job.JobType == models.JobTypePullRequest && job.Status == models.JobStatusFailed {
					latestAnalyzeJobFailed = true
					if job.ErrorMessage != nil {
						latestAnalyzeJobError = *job.ErrorMessage
					}
				}
			}
		}

		repositories = append(repositories, map[string]interface{}{
			"ProjectRepo":            projectRepo,
			"GitHubRepo":             githubRepo,
			"HasActiveCloneJobs":     hasActiveCloneJobs,
			"LatestCloneJobFailed":   latestCloneJobFailed,
			"LatestCloneJobError":    latestCloneJobError,
			"HasActiveAnalyzeJobs":   hasActiveAnalyzeJobs,
			"LatestAnalyzeJobFailed": latestAnalyzeJobFailed,
			"LatestAnalyzeJobError":  latestAnalyzeJobError,
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

	data := gin.H{
		"Title":              "Project Settings",
		"User":               session,
		"Project":            project,
		"ScoreSettings":      scoreSettings,
		"ExcludedExtensions": excludedExtensions,
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

// CreateAnalyzeJobs creates analysis jobs for a project
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

	// Create the pull_request and stats jobs
	err = h.jobService.CreatePullRequestAndStatsJobs(projectID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to create pull request and stats jobs: " + err.Error(),
		})
		return
	}

	// Return success response
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Pull request and stats jobs created successfully",
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

		// Extract just the email addresses in sorted order
		var sortedEmailAddresses []string
		for _, emailSimilarity := range sortedEmails {
			sortedEmailAddresses = append(sortedEmailAddresses, emailSimilarity.Email)
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

		// Extract just the email addresses in sorted order, excluding already associated emails
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
