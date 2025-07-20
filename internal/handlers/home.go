package handlers

import (
	"net/http"

	"github.com/alimgiray/gscope/internal/middleware"
	"github.com/alimgiray/gscope/internal/services"
	"github.com/gin-gonic/gin"
)

type HomeHandler struct {
	userService *services.UserService
}

func NewHomeHandler(userService *services.UserService) *HomeHandler {
	return &HomeHandler{
		userService: userService,
	}
}

// Index handles the home page
func (h *HomeHandler) Index(c *gin.Context) {
	session := middleware.GetSession(c)

	data := gin.H{
		"Title":   "Home",
		"Message": "Your Go web application is running!",
		"User":    session,
	}

	c.HTML(http.StatusOK, "index", data)
}
