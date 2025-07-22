package models

import (
	"time"

	"github.com/google/uuid"
)

// FileStatus represents the status of a file in a commit
type FileStatus string

const (
	FileStatusAdded    FileStatus = "added"
	FileStatusModified FileStatus = "modified"
	FileStatusRemoved  FileStatus = "removed"
	FileStatusRenamed  FileStatus = "renamed"
)

// CommitFile represents a file change in a commit
type CommitFile struct {
	ID        string     `json:"id"`
	CommitID  string     `json:"commit_id"`
	Filename  string     `json:"filename"`
	Status    FileStatus `json:"status"`
	Additions int        `json:"additions"`
	Deletions int        `json:"deletions"`
	Changes   int        `json:"changes"`
	CreatedAt time.Time  `json:"created_at"`
}

// NewCommitFile creates a new CommitFile with a generated UUID
func NewCommitFile(commitID, filename string, status FileStatus) *CommitFile {
	return &CommitFile{
		ID:        uuid.New().String(),
		CommitID:  commitID,
		Filename:  filename,
		Status:    status,
		Additions: 0,
		Deletions: 0,
		Changes:   0,
	}
}

// SetStats sets the file change statistics
func (cf *CommitFile) SetStats(additions, deletions, changes int) {
	cf.Additions = additions
	cf.Deletions = deletions
	cf.Changes = changes
}
