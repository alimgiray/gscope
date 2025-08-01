package models

import (
	"time"
)

// WorkingHoursSettings represents the working hours configuration for a project
type WorkingHoursSettings struct {
	ID        string    `json:"id" db:"id"`
	ProjectID string    `json:"project_id" db:"project_id"`
	StartHour int       `json:"start_hour" db:"start_hour"`
	EndHour   int       `json:"end_hour" db:"end_hour"`
	Monday    bool      `json:"monday" db:"monday"`
	Tuesday   bool      `json:"tuesday" db:"tuesday"`
	Wednesday bool      `json:"wednesday" db:"wednesday"`
	Thursday  bool      `json:"thursday" db:"thursday"`
	Friday    bool      `json:"friday" db:"friday"`
	Saturday  bool      `json:"saturday" db:"saturday"`
	Sunday    bool      `json:"sunday" db:"sunday"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// IsWorkingDay checks if the given weekday is a working day
func (w *WorkingHoursSettings) IsWorkingDay(weekday time.Weekday) bool {
	switch weekday {
	case time.Monday:
		return w.Monday
	case time.Tuesday:
		return w.Tuesday
	case time.Wednesday:
		return w.Wednesday
	case time.Thursday:
		return w.Thursday
	case time.Friday:
		return w.Friday
	case time.Saturday:
		return w.Saturday
	case time.Sunday:
		return w.Sunday
	default:
		return false
	}
}
