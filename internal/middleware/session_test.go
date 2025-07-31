package middleware

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/alimgiray/gscope/pkg/config"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestSessionExtension(t *testing.T) {
	// Load config for testing
	config.Load()

	// Create a new Gin router
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(SessionMiddleware())

	// Add a test route that returns success
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// Create a session
	sessionData := SessionData{
		UserID:    "test-user",
		Username:  "testuser",
		Email:     "test@example.com",
		ExpiresAt: time.Now().Add(1 * time.Hour), // 1 hour from now
	}

	// Encode session data
	data, _ := json.Marshal(sessionData)
	encodedData := base64.URLEncoding.EncodeToString(data)
	signature := createSignature(encodedData)
	cookieValue := signature + "." + encodedData

	// Create request with session cookie
	req, _ := http.NewRequest("GET", "/test", nil)
	req.AddCookie(&http.Cookie{
		Name:  "session",
		Value: cookieValue,
	})

	// Create response recorder
	w := httptest.NewRecorder()

	// Process request
	router.ServeHTTP(w, req)

	// Check that response is successful
	assert.Equal(t, http.StatusOK, w.Code)

	// Check that session cookie was updated by looking at Set-Cookie header
	setCookieHeader := w.Header().Get("Set-Cookie")
	assert.NotEmpty(t, setCookieHeader, "Set-Cookie header should be present")
	assert.Contains(t, setCookieHeader, "session=", "Set-Cookie header should contain session cookie")

	// Extract the session cookie value from the header
	if strings.Contains(setCookieHeader, "session=") {
		// Parse the cookie value from the header
		cookieParts := strings.Split(setCookieHeader, ";")
		sessionPart := cookieParts[0]
		sessionValue := strings.TrimPrefix(sessionPart, "session=")

		// URL-decode the session value since it's URL-encoded in the header
		decodedSessionValue, err := url.QueryUnescape(sessionValue)
		assert.NoError(t, err, "Should be able to URL-decode session value")

		// Verify that the session value is different (extended)
		assert.NotEqual(t, cookieValue, decodedSessionValue, "Session cookie should be updated with extended expiry")

		// Verify the extended session data
		parts := strings.Split(decodedSessionValue, ".")
		assert.Equal(t, 2, len(parts), "Cookie should have signature and data parts")

		// Verify signature
		isValid := verifySignature(parts[1], parts[0])
		assert.True(t, isValid, "Cookie signature should be valid")

		// Decode and verify session data
		decodedData, err := base64.URLEncoding.DecodeString(parts[1])
		assert.NoError(t, err, "Should be able to decode session data")

		var extendedSession SessionData
		err = json.Unmarshal(decodedData, &extendedSession)
		assert.NoError(t, err, "Should be able to unmarshal session data")

		// Verify session data
		assert.Equal(t, sessionData.UserID, extendedSession.UserID)
		assert.Equal(t, sessionData.Username, extendedSession.Username)
		assert.Equal(t, sessionData.Email, extendedSession.Email)

		// Verify that expiry time was extended (should be around 24 hours from now)
		expectedMinExpiry := time.Now().Add(23 * time.Hour) // Allow 1 hour tolerance
		assert.True(t, extendedSession.ExpiresAt.After(expectedMinExpiry), "Session expiry should be extended")
	}
}

func TestSessionExtensionOnError(t *testing.T) {
	// Load config for testing
	config.Load()

	// Create a new Gin router
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(SessionMiddleware())

	// Add a test route that returns an error
	router.GET("/error", func(c *gin.Context) {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "test error"})
	})

	// Create a session
	sessionData := SessionData{
		UserID:    "test-user",
		Username:  "testuser",
		Email:     "test@example.com",
		ExpiresAt: time.Now().Add(1 * time.Hour),
	}

	// Encode session data
	data, _ := json.Marshal(sessionData)
	encodedData := base64.URLEncoding.EncodeToString(data)
	signature := createSignature(encodedData)
	cookieValue := signature + "." + encodedData

	// Create request with session cookie
	req, _ := http.NewRequest("GET", "/error", nil)
	req.AddCookie(&http.Cookie{
		Name:  "session",
		Value: cookieValue,
	})

	// Create response recorder
	w := httptest.NewRecorder()

	// Process request
	router.ServeHTTP(w, req)

	// Check that response is an error
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	// Check that session cookie was NOT updated (because of error response)
	setCookieHeader := w.Header().Get("Set-Cookie")
	assert.Empty(t, setCookieHeader, "Set-Cookie header should not be present on error responses")
}

func TestSessionExtensionWithoutSession(t *testing.T) {
	// Load config for testing
	config.Load()

	// Create a new Gin router
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(SessionMiddleware())

	// Add a test route that returns success
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// Create request without session cookie
	req, _ := http.NewRequest("GET", "/test", nil)

	// Create response recorder
	w := httptest.NewRecorder()

	// Process request
	router.ServeHTTP(w, req)

	// Check that response is successful
	assert.Equal(t, http.StatusOK, w.Code)

	// Check that no session cookie was set (since there was no session)
	setCookieHeader := w.Header().Get("Set-Cookie")
	assert.Empty(t, setCookieHeader, "Set-Cookie header should not be present when no session exists")
}
