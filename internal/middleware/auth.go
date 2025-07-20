package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// AuthRequired middleware checks if user is authenticated
func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := GetSession(c)

		if session == nil {
			// Redirect to home page if not authenticated
			c.Redirect(http.StatusFound, "/")
			c.Abort()
			return
		}

		// User is authenticated, continue
		c.Next()
	}
}
