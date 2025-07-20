package handlers

import (
	"net/http"
	"strings"

	"github.com/alimgiray/gscope/internal/middleware"
	"github.com/alimgiray/gscope/internal/models"
	"github.com/alimgiray/gscope/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ProjectHandler struct {
	projectService *services.ProjectService
	userService    *services.UserService
}

func NewProjectHandler(projectService *services.ProjectService, userService *services.UserService) *ProjectHandler {
	return &ProjectHandler{
		projectService: projectService,
		userService:    userService,
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

// DeleteProject handles project deletion
func (h *ProjectHandler) DeleteProject(c *gin.Context) {
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

	// Check if project exists and belongs to user
	project, err := h.projectService.GetProjectByID(projectID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
		return
	}

	userID, err := uuid.Parse(session.UserID)
	if err != nil || project.OwnerID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	// Delete the project
	if err := h.projectService.DeleteProject(projectID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete project"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Project deleted successfully"})
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

	data := gin.H{
		"Title":   project.Name,
		"User":    session,
		"Project": project,
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

	data := gin.H{
		"Title":   "Project Settings",
		"User":    session,
		"Project": project,
	}

	c.HTML(http.StatusOK, "project_settings", data)
}
