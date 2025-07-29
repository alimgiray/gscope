package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	GitHub   GitHubConfig
	Session  SessionConfig
}

type ServerConfig struct {
	Port         string
	Mode         string
	ReadTimeout  int
	WriteTimeout int
}

type DatabaseConfig struct {
	Path string
}

type GitHubConfig struct {
	ClientID     string
	ClientSecret string
	CallbackURL  string
}

type SessionConfig struct {
	Secret string
}

var AppConfig *Config

// Load loads configuration from .env file and environment variables
func Load() error {
	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	AppConfig = &Config{
		Server: ServerConfig{
			Port:         getEnv("PORT", "8080"),
			Mode:         getEnv("GIN_MODE", "release"),
			ReadTimeout:  getEnvAsInt("READ_TIMEOUT", 15),
			WriteTimeout: getEnvAsInt("WRITE_TIMEOUT", 15),
		},
		Database: DatabaseConfig{
			Path: getEnv("DB_PATH", "./gscope.db"),
		},
		GitHub: GitHubConfig{
			ClientID:     getEnv("GITHUB_CLIENT_ID", ""),
			ClientSecret: getEnv("GITHUB_CLIENT_SECRET", ""),
			CallbackURL:  getEnv("GITHUB_CALLBACK_URL", "http://localhost:8080/auth/github/callback"),
		},
		Session: SessionConfig{
			Secret: getEnv("SESSION_SECRET", "default-secret-key-change-in-production"),
		},
	}

	// Log configuration (without sensitive data)
	log.Printf("Configuration loaded:")
	log.Printf("  Server Port: %s", AppConfig.Server.Port)
	log.Printf("  Gin Mode: %s", AppConfig.Server.Mode)
	log.Printf("  Database Path: %s", AppConfig.Database.Path)
	log.Printf("  GitHub Client ID: %s", maskString(AppConfig.GitHub.ClientID))
	log.Printf("  GitHub Callback URL: %s", AppConfig.GitHub.CallbackURL)

	return nil
}

// getEnv gets an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvAsInt gets an environment variable as integer or returns a default value
func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// maskString masks sensitive strings for logging
func maskString(s string) string {
	if s == "" {
		return "<not-set>"
	}
	if len(s) <= 4 {
		return "****"
	}
	return s[:4] + "****"
}
