package services

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/alimgiray/gscope/pkg/config"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
)

type GitHubService struct {
	oauthConfig *oauth2.Config
}

type GitHubUser struct {
	ID        int    `json:"id"`
	Login     string `json:"login"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	AvatarURL string `json:"avatar_url"`
}

func NewGitHubService() *GitHubService {
	oauthConfig := &oauth2.Config{
		ClientID:     config.AppConfig.GitHub.ClientID,
		ClientSecret: config.AppConfig.GitHub.ClientSecret,
		RedirectURL:  config.AppConfig.GitHub.CallbackURL,
		Scopes: []string{
			"user:email", // Access to user's email addresses
			"read:user",  // Read access to user profile data
			"read:org",   // Read access to organization membership
			"repo",       // Full access to repositories (includes PRs, issues, etc.)
		},
		Endpoint: github.Endpoint,
	}

	return &GitHubService{
		oauthConfig: oauthConfig,
	}
}

// GetAuthURL returns the GitHub OAuth authorization URL
func (s *GitHubService) GetAuthURL() string {
	return s.oauthConfig.AuthCodeURL("state", oauth2.AccessTypeOffline)
}

// ExchangeCodeForToken exchanges authorization code for access token
func (s *GitHubService) ExchangeCodeForToken(code string) (*oauth2.Token, error) {
	ctx := context.Background()
	token, err := s.oauthConfig.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange code for token: %w", err)
	}
	return token, nil
}

// GetUserInfo retrieves user information from GitHub
func (s *GitHubService) GetUserInfo(token *oauth2.Token) (*GitHubUser, error) {
	ctx := context.Background()
	client := s.oauthConfig.Client(ctx, token)

	resp, err := client.Get("https://api.github.com/user")
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var user GitHubUser
	if err := json.Unmarshal(body, &user); err != nil {
		return nil, fmt.Errorf("failed to unmarshal user data: %w", err)
	}

	return &user, nil
}
