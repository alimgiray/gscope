package handlers

import (
	"log"
	"net/http"

	"github.com/alimgiray/gscope/internal/middleware"
	"github.com/alimgiray/gscope/internal/models"
	"github.com/alimgiray/gscope/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type AuthHandler struct {
	userService   *services.UserService
	githubService *services.GitHubService
}

func NewAuthHandler(userService *services.UserService) *AuthHandler {
	return &AuthHandler{
		userService:   userService,
		githubService: services.NewGitHubService(),
	}
}

// Login handles the login page
func (h *AuthHandler) Login(c *gin.Context) {
	session := middleware.GetSession(c)
	errorMsg := c.Query("error")

	data := gin.H{
		"Title": "Login",
		"User":  session,
		"Error": errorMsg,
	}

	c.HTML(http.StatusOK, "login", data)
}

// Logout handles user logout
func (h *AuthHandler) Logout(c *gin.Context) {
	middleware.ClearSession(c)
	c.Redirect(http.StatusFound, "/")
}

// GitHubLogin initiates GitHub OAuth flow
func (h *AuthHandler) GitHubLogin(c *gin.Context) {
	authURL := h.githubService.GetAuthURL()
	c.Redirect(http.StatusTemporaryRedirect, authURL)
}

// GitHubCallback handles GitHub OAuth callback
func (h *AuthHandler) GitHubCallback(c *gin.Context) {
	code := c.Query("code")
	if code == "" {
		c.Redirect(http.StatusFound, "/login?error=no_code")
		return
	}

	// Exchange code for token
	token, err := h.githubService.ExchangeCodeForToken(code)
	if err != nil {
		c.Redirect(http.StatusFound, "/login?error=token_exchange_failed")
		return
	}

	// Get user info from GitHub
	githubUser, err := h.githubService.GetUserInfo(token)
	if err != nil {
		c.Redirect(http.StatusFound, "/login?error=user_info_failed")
		return
	}

	// Check if user exists in our database
	user, err := h.userService.GetUserByEmail(githubUser.Email)
	log.Printf("GitHub callback - Email: %s, User found: %v, Error: %v", githubUser.Email, user != nil, err)
	if err != nil || user == nil {
		// User doesn't exist, create new user
		user = &models.User{
			ID:                uuid.New(),
			Name:              githubUser.Name,
			Username:          githubUser.Login,
			Email:             githubUser.Email,
			ProfilePicture:    githubUser.AvatarURL,
			GitHubAccessToken: token.AccessToken,
		}

		if err := h.userService.CreateUser(user); err != nil {
			c.Redirect(http.StatusFound, "/login?error=user_creation_failed")
			return
		}
	} else {
		// Update existing user's GitHub token
		user.GitHubAccessToken = token.AccessToken
		if err := h.userService.UpdateUser(user); err != nil {
			c.Redirect(http.StatusFound, "/login?error=user_update_failed")
			return
		}
	}

	// Create session
	if user == nil {
		c.Redirect(http.StatusFound, "/login?error=user_data_missing")
		return
	}

	if err := middleware.SetSession(c, user.ID.String(), user.Username, user.Email); err != nil {
		c.Redirect(http.StatusFound, "/login?error=session_creation_failed")
		return
	}

	// Redirect to dashboard
	c.Redirect(http.StatusFound, "/dashboard")
}
