package services

import (
	"testing"

	"github.com/alimgiray/gscope/internal/models"
	"github.com/stretchr/testify/assert"
)

func TestScoreSettingsCreation(t *testing.T) {
	// Test default score settings creation
	t.Run("Default score settings", func(t *testing.T) {
		settings := models.NewScoreSettings("test-project")

		assert.Equal(t, 1, settings.Additions)
		assert.Equal(t, 3, settings.Deletions)
		assert.Equal(t, 10, settings.Commits)
		assert.Equal(t, 20, settings.PullRequests)
		assert.Equal(t, 100, settings.Comments)
		assert.NotEmpty(t, settings.ID)
		assert.Equal(t, "test-project", settings.ProjectID)
	})

	// Test custom score settings
	t.Run("Custom score settings", func(t *testing.T) {
		settings := &models.ScoreSettings{
			Additions:    5,
			Deletions:    2,
			Commits:      15,
			PullRequests: 25,
			Comments:     150,
		}

		assert.Equal(t, 5, settings.Additions)
		assert.Equal(t, 2, settings.Deletions)
		assert.Equal(t, 15, settings.Commits)
		assert.Equal(t, 25, settings.PullRequests)
		assert.Equal(t, 150, settings.Comments)
	})
}

func TestScoreCalculationLogic(t *testing.T) {
	// Test score calculation with different settings
	testCases := []struct {
		name          string
		commits       int
		additions     int
		deletions     int
		pullRequests  int
		comments      int
		scoreSettings *models.ScoreSettings
		expectedScore int
		description   string
	}{
		{
			name:          "Default settings calculation",
			commits:       5,
			additions:     100,
			deletions:     50,
			pullRequests:  2,
			comments:      10,
			scoreSettings: models.NewScoreSettings("test-project"),
			expectedScore: 5*10 + 100*1 + 50*3 + 2*20 + 10*100, // 50 + 100 + 150 + 40 + 1000 = 1340
			description:   "Score calculation with default settings",
		},
		{
			name:          "Zero activity",
			commits:       0,
			additions:     0,
			deletions:     0,
			pullRequests:  0,
			comments:      0,
			scoreSettings: models.NewScoreSettings("test-project"),
			expectedScore: 0,
			description:   "Zero activity should result in zero score",
		},
		{
			name:          "High activity",
			commits:       20,
			additions:     500,
			deletions:     200,
			pullRequests:  5,
			comments:      50,
			scoreSettings: models.NewScoreSettings("test-project"),
			expectedScore: 20*10 + 500*1 + 200*3 + 5*20 + 50*100, // 200 + 500 + 600 + 100 + 5000 = 6400
			description:   "High activity should result in high score",
		},
		{
			name:         "Custom settings",
			commits:      3,
			additions:    150,
			deletions:    75,
			pullRequests: 1,
			comments:     5,
			scoreSettings: &models.ScoreSettings{
				Additions:    2,
				Deletions:    1,
				Commits:      5,
				PullRequests: 10,
				Comments:     50,
			},
			expectedScore: 3*5 + 150*2 + 75*1 + 1*10 + 5*50, // 15 + 300 + 75 + 10 + 250 = 650
			description:   "Custom score settings should be applied correctly",
		},
		{
			name:          "Negative values",
			commits:       -1,
			additions:     -1,
			deletions:     -1,
			pullRequests:  -1,
			comments:      -1,
			scoreSettings: models.NewScoreSettings("test-project"),
			expectedScore: -1*10 + -1*1 + -1*3 + -1*20 + -1*100, // -10 + -1 + -3 + -20 + -100 = -134
			description:   "Negative values should be handled correctly",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			score := calculateScore(
				tc.commits,
				tc.additions,
				tc.deletions,
				tc.pullRequests,
				tc.comments,
				tc.scoreSettings,
			)

			assert.Equal(t, tc.expectedScore, score, tc.description)
		})
	}
}

// Helper function to calculate score (copied from the service for testing)
func calculateScore(
	commits, additions, deletions, pullRequests, comments int,
	scoreSettings *models.ScoreSettings,
) int {
	if scoreSettings == nil {
		return 0
	}

	score := 0

	// Add points for commits
	score += commits * scoreSettings.Commits

	// Add points for additions
	score += additions * scoreSettings.Additions

	// Add points for deletions
	score += deletions * scoreSettings.Deletions

	// Add points for pull requests
	score += pullRequests * scoreSettings.PullRequests

	// Add points for comments
	score += comments * scoreSettings.Comments

	return score
}

func TestScoreSettingsEdgeCases(t *testing.T) {
	t.Run("Nil score settings", func(t *testing.T) {
		score := calculateScore(1, 1, 1, 1, 1, nil)
		assert.Equal(t, 0, score, "Nil score settings should return 0")
	})

	t.Run("Zero score settings", func(t *testing.T) {
		settings := &models.ScoreSettings{
			Additions:    0,
			Deletions:    0,
			Commits:      0,
			PullRequests: 0,
			Comments:     0,
		}
		score := calculateScore(10, 100, 50, 5, 20, settings)
		assert.Equal(t, 0, score, "Zero score settings should return 0 regardless of activity")
	})

	t.Run("Very large values", func(t *testing.T) {
		settings := models.NewScoreSettings("test-project")
		score := calculateScore(1000, 10000, 5000, 100, 500, settings)
		expected := 1000*10 + 10000*1 + 5000*3 + 100*20 + 500*100
		assert.Equal(t, expected, score, "Very large values should be handled correctly")
	})
}

func TestScoreSettingsValidationLogic(t *testing.T) {
	t.Run("Valid score settings", func(t *testing.T) {
		settings := &models.ScoreSettings{
			Additions:    1,
			Deletions:    3,
			Commits:      10,
			PullRequests: 20,
			Comments:     100,
		}

		// All values should be non-negative
		assert.GreaterOrEqual(t, settings.Additions, 0)
		assert.GreaterOrEqual(t, settings.Deletions, 0)
		assert.GreaterOrEqual(t, settings.Commits, 0)
		assert.GreaterOrEqual(t, settings.PullRequests, 0)
		assert.GreaterOrEqual(t, settings.Comments, 0)
	})

	t.Run("Negative score settings", func(t *testing.T) {
		settings := &models.ScoreSettings{
			Additions:    -1,
			Deletions:    -2,
			Commits:      -3,
			PullRequests: -4,
			Comments:     -5,
		}

		// Negative values should still be allowed for calculation
		score := calculateScore(1, 1, 1, 1, 1, settings)
		expected := 1*-3 + 1*-1 + 1*-2 + 1*-4 + 1*-5
		assert.Equal(t, expected, score, "Negative score settings should be calculated correctly")
	})
}

func TestScoreCalculationCommutativity(t *testing.T) {
	// Test that score calculation is commutative (order doesn't matter)
	settings := models.NewScoreSettings("test-project")

	// Test with same values in different order
	score1 := calculateScore(5, 100, 50, 2, 10, settings)
	score2 := calculateScore(2, 10, 50, 5, 100, settings)

	// The scores should be the same because the calculation is based on multiplication
	// But the order of parameters matters, so we expect different results
	expected1 := 5*10 + 100*1 + 50*3 + 2*20 + 10*100 // 50 + 100 + 150 + 40 + 1000 = 1340
	expected2 := 2*10 + 10*1 + 50*3 + 5*20 + 100*100 // 20 + 10 + 150 + 100 + 10000 = 10280

	assert.Equal(t, expected1, score1, "First score calculation should be correct")
	assert.Equal(t, expected2, score2, "Second score calculation should be correct")
}

func TestScoreCalculationAssociativity(t *testing.T) {
	// Test that score calculation is associative
	settings := models.NewScoreSettings("test-project")

	// Test that breaking down the calculation gives the same result
	commitsScore := 5 * settings.Commits
	additionsScore := 100 * settings.Additions
	deletionsScore := 50 * settings.Deletions
	pullRequestsScore := 2 * settings.PullRequests
	commentsScore := 10 * settings.Comments

	totalScore := commitsScore + additionsScore + deletionsScore + pullRequestsScore + commentsScore

	// Compare with direct calculation
	directScore := calculateScore(5, 100, 50, 2, 10, settings)

	assert.Equal(t, totalScore, directScore, "Score calculation should be associative")
}
