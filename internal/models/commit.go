package models

import (
	"time"

	"github.com/google/uuid"
)

// Commit represents a Git commit
type Commit struct {
	ID                 string    `json:"id"`
	GithubRepositoryID string    `json:"github_repository_id"`
	CommitSHA          string    `json:"commit_sha"`
	Message            string    `json:"message"`
	AuthorName         string    `json:"author_name"`
	AuthorEmail        *string   `json:"author_email"`
	CommitDate         time.Time `json:"commit_date"`
	IsMergeCommit      bool      `json:"is_merge_commit"`
	MergeCommitSHA     *string   `json:"merge_commit_sha"`
	Additions          int       `json:"additions"`
	Deletions          int       `json:"deletions"`
	Changes            int       `json:"changes"`
	CreatedAt          time.Time `json:"created_at"`
}

// NewCommit creates a new Commit with a generated UUID
func NewCommit(githubRepositoryID, commitSHA, message, authorName string, authorEmail *string, commitDate time.Time) *Commit {
	return &Commit{
		ID:                 uuid.New().String(),
		GithubRepositoryID: githubRepositoryID,
		CommitSHA:          commitSHA,
		Message:            message,
		AuthorName:         authorName,
		AuthorEmail:        authorEmail,
		CommitDate:         commitDate,
		IsMergeCommit:      false,
		Additions:          0,
		Deletions:          0,
		Changes:            0,
	}
}

// SetMergeCommit marks this commit as a merge commit
func (c *Commit) SetMergeCommit(mergeCommitSHA string) {
	c.IsMergeCommit = true
	c.MergeCommitSHA = &mergeCommitSHA
}

// SetStats sets the commit statistics
func (c *Commit) SetStats(additions, deletions, changes int) {
	c.Additions = additions
	c.Deletions = deletions
	c.Changes = changes
}
