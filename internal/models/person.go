package models

import (
	"time"

	"github.com/google/uuid"
)

// Person represents a unique person/author
type Person struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	PrimaryEmail string    `json:"primary_email"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// NewPerson creates a new Person with a generated UUID
func NewPerson(name, primaryEmail string) *Person {
	return &Person{
		ID:           uuid.New().String(),
		Name:         name,
		PrimaryEmail: primaryEmail,
	}
}
