package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/alimgiray/gscope/internal/middleware"
	"github.com/alimgiray/gscope/internal/models"
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
}

func NewProjectHandler(projectService *services.ProjectService, userService *services.UserService,
	scoreSettingsService *services.ScoreSettingsService, excludedExtensionService *services.ExcludedExtensionService,
	githubRepoService *services.GitHubRepositoryService, jobService *services.JobService) *ProjectHandler {
	return &ProjectHandler{
		projectService:           projectService,
		userService:              userService,
		scoreSettingsService:     scoreSettingsService,
		excludedExtensionService: excludedExtensionService,
		githubRepoService:        githubRepoService,
		jobService:               jobService,
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

		// Get the latest clone job for this repository
		var latestJob *models.Job
		cloneJobs, err := h.jobService.GetJobsByProjectRepository(projectRepo.ID)
		if err == nil && len(cloneJobs) > 0 {
			// Find the most recent clone job
			for _, job := range cloneJobs {
				if job.JobType == models.JobTypeClone {
					latestJob = job
					break
				}
			}
		}

		// Get the latest analyze job (commit, pull_request, or stats) for this repository
		var latestAnalyzeJob *models.Job
		if err == nil && len(cloneJobs) > 0 {
			// Find the most recent analyze job
			for _, job := range cloneJobs {
				if job.JobType == models.JobTypeCommit || job.JobType == models.JobTypePullRequest || job.JobType == models.JobTypeStats {
					latestAnalyzeJob = job
					break
				}
			}
		}

		repositories = append(repositories, map[string]interface{}{
			"ProjectRepo":      projectRepo,
			"GitHubRepo":       githubRepo,
			"LatestJob":        latestJob,
			"LatestAnalyzeJob": latestAnalyzeJob,
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

	// Create the clone job
	_, err = h.jobService.CreateCloneJob(projectID, projectRepositoryID)
	if err != nil {
		// Check if it's a duplicate job error
		if strings.Contains(err.Error(), "already in progress or pending") {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": "A clone job is already in progress for this repository",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": "Failed to create clone job: " + err.Error(),
			})
		}
		return
	}

	// Return success response
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Clone job created successfully",
	})
}

// CreateAnalyzeJobs creates analysis jobs for a specific repository
func (h *ProjectHandler) CreateAnalyzeJobs(c *gin.Context) {
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

	// Create the analyze jobs
	err = h.jobService.CreateAnalyzeJobs(projectID, projectRepositoryID)
	if err != nil {
		// Check if it's a duplicate job error
		if strings.Contains(err.Error(), "already in progress or pending") {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": err.Error(),
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": "Failed to create analysis jobs: " + err.Error(),
			})
		}
		return
	}

	// Return success response
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Analysis jobs created successfully",
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

	// Create clone jobs for all tracked repositories
	createdCount, err := h.jobService.CreateCloneJobsForAllTrackedRepositories(projectID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to create clone jobs: " + err.Error(),
		})
		return
	}

	// Return success response
	message := fmt.Sprintf("Created %d clone jobs for tracked repositories", createdCount)
	if createdCount == 0 {
		message = "No clone jobs created (no tracked repositories or all already have active jobs)"
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": message,
	})
}
