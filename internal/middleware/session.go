package middleware

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"strings"
	"time"

	"github.com/alimgiray/gscope/pkg/config"
	"github.com/gin-gonic/gin"
)

type SessionData struct {
	UserID    string    `json:"user_id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	ExpiresAt time.Time `json:"expires_at"`
}

// SessionMiddleware handles session management using cookies
func SessionMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get session from cookie
		sessionData := getSessionFromCookie(c)

		// Set session data in context
		c.Set("session", sessionData)

		c.Next()
	}
}

// getSessionFromCookie extracts and validates session data from cookie
func getSessionFromCookie(c *gin.Context) *SessionData {
	cookie, err := c.Cookie("session")
	if err != nil {
		return nil
	}

	// Split cookie value (signature.data)
	parts := strings.Split(cookie, ".")
	if len(parts) != 2 {
		return nil
	}

	signature, data := parts[0], parts[1]

	// Verify signature
	if !verifySignature(data, signature) {
		return nil
	}

	// Decode data
	decodedData, err := base64.URLEncoding.DecodeString(data)
	if err != nil {
		return nil
	}

	var sessionData SessionData
	if err := json.Unmarshal(decodedData, &sessionData); err != nil {
		return nil
	}

	// Check if session is expired
	if time.Now().After(sessionData.ExpiresAt) {
		return nil
	}

	return &sessionData
}

// SetSession creates a new session cookie
func SetSession(c *gin.Context, userID, username, email string) error {
	sessionData := SessionData{
		UserID:    userID,
		Username:  username,
		Email:     email,
		ExpiresAt: time.Now().Add(24 * time.Hour), // 24 hour expiry
	}

	// Encode session data
	data, err := json.Marshal(sessionData)
	if err != nil {
		return err
	}

	encodedData := base64.URLEncoding.EncodeToString(data)

	// Create signature
	signature := createSignature(encodedData)

	// Set cookie
	c.SetCookie("session", signature+"."+encodedData, 86400, "/", "", false, true)

	return nil
}

// ClearSession removes the session cookie
func ClearSession(c *gin.Context) {
	c.SetCookie("session", "", -1, "/", "", false, true)
}

// createSignature creates HMAC signature for data
func createSignature(data string) string {
	h := hmac.New(sha256.New, []byte(config.AppConfig.Session.Secret))
	h.Write([]byte(data))
	return base64.URLEncoding.EncodeToString(h.Sum(nil))
}

// verifySignature verifies HMAC signature
func verifySignature(data, signature string) bool {
	expectedSignature := createSignature(data)
	return hmac.Equal([]byte(signature), []byte(expectedSignature))
}

// GetSession retrieves session data from context
func GetSession(c *gin.Context) *SessionData {
	session, exists := c.Get("session")
	if !exists {
		return nil
	}

	if sessionData, ok := session.(*SessionData); ok {
		return sessionData
	}

	return nil
}
