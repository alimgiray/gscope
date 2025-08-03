package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type NotFoundHandler struct{}

func NewNotFoundHandler() *NotFoundHandler {
	return &NotFoundHandler{}
}

// NotFound handles 404 errors for non-existent routes
func (h *NotFoundHandler) NotFound(c *gin.Context) {
	data := gin.H{
		"Title":         "404 - Page Not Found",
		"User":          nil, // No user context needed for 404
		"RequestedPath": c.Request.URL.Path,
		"Timestamp":     time.Now().Format("2006-01-02 15:04:05"),
	}

	c.HTML(http.StatusNotFound, "404", data)
}
