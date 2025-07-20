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
			CallbackURL:  getEnv("GITHUB_CALLBACK_URL", ""),
		},
		Session: SessionConfig{
			Secret: getEnv("SESSION_SECRET", "default-secret-key"),
		},
	}

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
