package models

// GitHubPersonStats represents aggregated statistics for a GitHub person
type GitHubPersonStats struct {
	GitHubPerson      *GithubPerson `json:"github_person"`
	TotalCommits      int           `json:"total_commits"`
	TotalAdditions    int           `json:"total_additions"`
	TotalDeletions    int           `json:"total_deletions"`
	TotalComments     int           `json:"total_comments"`
	TotalPullRequests int           `json:"total_pull_requests"`
	TotalScore        int           `json:"total_score"`
}
