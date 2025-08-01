package handlers

import (
	"net/http"
	"strconv"

	"github.com/alimgiray/gscope/internal/models"
	"github.com/alimgiray/gscope/internal/services"
	"github.com/gin-gonic/gin"
)

type WorkingHoursSettingsHandler struct {
	workingHoursSettingsService *services.WorkingHoursSettingsService
}

func NewWorkingHoursSettingsHandler(workingHoursSettingsService *services.WorkingHoursSettingsService) *WorkingHoursSettingsHandler {
	return &WorkingHoursSettingsHandler{
		workingHoursSettingsService: workingHoursSettingsService,
	}
}

// GetWorkingHoursSettings retrieves working hours settings for a project
func (h *WorkingHoursSettingsHandler) GetWorkingHoursSettings(c *gin.Context) {
	projectID := c.Param("id")
	if projectID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Project ID is required"})
		return
	}

	settings, err := h.workingHoursSettingsService.GetByProjectID(projectID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get working hours settings"})
		return
	}

	c.JSON(http.StatusOK, settings)
}

// UpdateWorkingHoursSettings updates working hours settings for a project
func (h *WorkingHoursSettingsHandler) UpdateWorkingHoursSettings(c *gin.Context) {
	projectID := c.Param("id")
	if projectID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Project ID is required"})
		return
	}

	// Parse form data
	startHourStr := c.PostForm("start_hour")
	endHourStr := c.PostForm("end_hour")

	startHour, err := strconv.Atoi(startHourStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid start hour"})
		return
	}

	endHour, err := strconv.Atoi(endHourStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid end hour"})
		return
	}

	// Parse boolean values for days
	monday := c.PostForm("monday") == "on"
	tuesday := c.PostForm("tuesday") == "on"
	wednesday := c.PostForm("wednesday") == "on"
	thursday := c.PostForm("thursday") == "on"
	friday := c.PostForm("friday") == "on"
	saturday := c.PostForm("saturday") == "on"
	sunday := c.PostForm("sunday") == "on"

	settings := &models.WorkingHoursSettings{
		ProjectID: projectID,
		StartHour: startHour,
		EndHour:   endHour,
		Monday:    monday,
		Tuesday:   tuesday,
		Wednesday: wednesday,
		Thursday:  thursday,
		Friday:    friday,
		Saturday:  saturday,
		Sunday:    sunday,
	}

	err = h.workingHoursSettingsService.CreateOrUpdate(settings)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Redirect back to project settings
	c.Redirect(http.StatusSeeOther, "/projects/"+projectID+"/settings")
}

// WorkingHoursSettingsForm renders the working hours settings form
func (h *WorkingHoursSettingsHandler) WorkingHoursSettingsForm(c *gin.Context) {
	projectID := c.Param("id")
	if projectID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Project ID is required"})
		return
	}

	settings, err := h.workingHoursSettingsService.GetByProjectID(projectID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get working hours settings"})
		return
	}

	c.HTML(http.StatusOK, "projects/working_hours_settings.html", gin.H{
		"ProjectID": projectID,
		"Settings":  settings,
	})
}
