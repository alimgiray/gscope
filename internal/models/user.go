package models

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID                uuid.UUID
	Name              string
	Username          string
	Email             string
	ProfilePicture    string
	AccessToken       string
	GitHubAccessToken string
	CreatedAt         time.Time
}
