package handlers

import (
	"net/http"

	"github.com/alimgiray/gscope/internal/middleware"
	"github.com/alimgiray/gscope/internal/models"
	"github.com/alimgiray/gscope/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type LLMAPIKeyHandler struct {
	llmAPIKeyService           *services.LLMAPIKeyService
	projectService             *services.ProjectService
	projectCollaboratorService *services.ProjectCollaboratorService
}

func NewLLMAPIKeyHandler(
	llmAPIKeyService *services.LLMAPIKeyService,
	projectService *services.ProjectService,
	projectCollaboratorService *services.ProjectCollaboratorService,
) *LLMAPIKeyHandler {
	return &LLMAPIKeyHandler{
		llmAPIKeyService:           llmAPIKeyService,
		projectService:             projectService,
		projectCollaboratorService: projectCollaboratorService,
	}
}

// ViewLLMSettings displays the LLM API key settings page
func (h *LLMAPIKeyHandler) ViewLLMSettings(c *gin.Context) {
	session := middleware.GetSession(c)
	if session == nil {
		c.Redirect(http.StatusSeeOther, "/login")
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

	// Get project
	project, err := h.projectService.GetProjectByID(projectID)
	if err != nil {
		c.HTML(http.StatusNotFound, "error", gin.H{
			"Title": "Error",
			"User":  session,
			"Error": "Project not found",
		})
		return
	}

	// Check access
	accessType, err := h.projectCollaboratorService.GetProjectAccessType(projectID, session.UserID)
	if err != nil || accessType == "none" {
		c.HTML(http.StatusForbidden, "error", gin.H{
			"Title": "Access Denied",
			"User":  session,
			"Error": "You don't have permission to access this project",
		})
		return
	}

	// Get existing API key
	userUUID, _ := uuid.Parse(session.UserID)
	projectUUID, _ := uuid.Parse(projectID)

	apiKey, err := h.llmAPIKeyService.GetAPIKey(projectUUID, userUUID)
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error", gin.H{
			"Title": "Error",
			"User":  session,
			"Error": "Failed to get API key settings: " + err.Error(),
		})
		return
	}

	data := gin.H{
		"Title":      "LLM Settings - " + project.Name,
		"User":       session,
		"Project":    project,
		"AccessType": accessType,
		"APIKey":     apiKey,
	}

	c.HTML(http.StatusOK, "project_llm_settings", data)
}

// CreateOrUpdateAPIKey creates or updates an LLM API key (owner-only)
func (h *LLMAPIKeyHandler) CreateOrUpdateAPIKey(c *gin.Context) {
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

	// Parse and validate request
	var request models.LLMAPIKeyRequest
	if err := c.ShouldBind(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid request data: " + err.Error(),
		})
		return
	}

	// Validate request
	if err := request.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	userUUID, err := uuid.Parse(session.UserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid user ID",
		})
		return
	}

	projectUUID, err := uuid.Parse(projectID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid project ID",
		})
		return
	}

	// Create or update API key (service validates ownership)
	apiKey, err := h.llmAPIKeyService.CreateOrUpdateAPIKey(projectUUID, userUUID, &request)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "API key saved successfully",
		"data": gin.H{
			"id":             apiKey.ID,
			"provider":       apiKey.GetDisplayProvider(),
			"masked_api_key": apiKey.MaskAPIKey(),
			"created_at":     apiKey.CreatedAt,
			"updated_at":     apiKey.UpdatedAt,
		},
	})
}

// DeleteAPIKey deletes an LLM API key (owner-only)
func (h *LLMAPIKeyHandler) DeleteAPIKey(c *gin.Context) {
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

	userUUID, err := uuid.Parse(session.UserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid user ID",
		})
		return
	}

	projectUUID, err := uuid.Parse(projectID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid project ID",
		})
		return
	}

	// Delete API key (service validates ownership)
	err = h.llmAPIKeyService.DeleteAPIKey(projectUUID, userUUID)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "API key deleted successfully",
	})
}
