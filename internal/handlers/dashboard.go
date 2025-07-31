package handlers

import (
	"net/http"
	"sort"

	"github.com/alimgiray/gscope/internal/middleware"
	"github.com/alimgiray/gscope/internal/models"
	"github.com/alimgiray/gscope/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type DashboardHandler struct {
	userService                *services.UserService
	projectService             *services.ProjectService
	projectCollaboratorService *services.ProjectCollaboratorService
}

func NewDashboardHandler(userService *services.UserService, projectService *services.ProjectService, projectCollaboratorService *services.ProjectCollaboratorService) *DashboardHandler {
	return &DashboardHandler{
		userService:                userService,
		projectService:             projectService,
		projectCollaboratorService: projectCollaboratorService,
	}
}

// Dashboard handles the dashboard page
func (h *DashboardHandler) Dashboard(c *gin.Context) {
	session := middleware.GetSession(c)

	// Get user data from service
	user, err := h.userService.GetUserByID(session.UserID)
	if err != nil {
		// For now, use session data if user not found in database
		user = &models.User{
			ID:        uuid.MustParse(session.UserID),
			Name:      session.Username, // Using username as name for now
			Username:  session.Username,
			Email:     session.Email,
			CreatedAt: session.ExpiresAt, // Using session expiry as created date for now
		}
	}

	// Get user's accessible projects (owned + collaborated)
	projects, err := h.projectCollaboratorService.GetUserAccessibleProjects(session.UserID)
	if err != nil {
		// If error fetching projects, just show empty list
		projects = []*models.Project{}
	}

	// Create a map to store project access types
	projectAccessTypes := make(map[string]string)
	for _, project := range projects {
		accessType, err := h.projectCollaboratorService.GetProjectAccessType(project.ID.String(), session.UserID)
		if err != nil {
			accessType = "unknown"
		}
		projectAccessTypes[project.ID.String()] = accessType
	}

	// Sort projects alphabetically by name
	sort.Slice(projects, func(i, j int) bool {
		return projects[i].Name < projects[j].Name
	})

	data := gin.H{
		"Title":              "Dashboard",
		"User":               user,
		"Projects":           projects,
		"ProjectAccessTypes": projectAccessTypes,
	}

	c.HTML(http.StatusOK, "dashboard", data)
}
