package services

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCalculateSimilarity(t *testing.T) {
	service := NewTextSimilarityService()

	testCases := []struct {
		name        string
		str1        string
		str2        string
		expectedMin float64
		expectedMax float64
		description string
	}{
		{
			name:        "Identical strings",
			str1:        "testuser",
			str2:        "testuser",
			expectedMin: 1.0,
			expectedMax: 1.0,
			description: "Identical strings should have similarity of 1.0",
		},
		{
			name:        "Empty strings",
			str1:        "",
			str2:        "",
			expectedMin: 1.0, // Empty strings are identical
			expectedMax: 1.0,
			description: "Empty strings should have similarity of 1.0",
		},
		{
			name:        "One empty string",
			str1:        "testuser",
			str2:        "",
			expectedMin: 0.0,
			expectedMax: 0.0,
			description: "One empty string should have similarity of 0.0",
		},
		{
			name:        "Similar strings",
			str1:        "testuser",
			str2:        "testuser123",
			expectedMin: 0.5,
			expectedMax: 1.0, // Allow for exact match after normalization
			description: "Similar strings should have high similarity",
		},
		{
			name:        "Very different strings",
			str1:        "testuser",
			str2:        "completelydifferent",
			expectedMin: 0.0,
			expectedMax: 0.3,
			description: "Very different strings should have low similarity",
		},
		{
			name:        "Case insensitive",
			str1:        "TestUser",
			str2:        "testuser",
			expectedMin: 0.8,
			expectedMax: 1.0,
			description: "Case differences should be normalized",
		},
		{
			name:        "Special characters",
			str1:        "test-user",
			str2:        "testuser",
			expectedMin: 0.7,
			expectedMax: 1.0,
			description: "Special characters should be handled",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			similarity := service.CalculateSimilarity(tc.str1, tc.str2)

			assert.GreaterOrEqual(t, similarity, tc.expectedMin,
				"%s: similarity should be >= %f, got %f", tc.description, tc.expectedMin, similarity)
			assert.LessOrEqual(t, similarity, tc.expectedMax,
				"%s: similarity should be <= %f, got %f", tc.description, tc.expectedMax, similarity)
		})
	}
}

func TestCalculateEmailUsernameSimilarity(t *testing.T) {
	service := NewTextSimilarityService()

	testCases := []struct {
		name        string
		email       string
		username    string
		expectedMin float64
		expectedMax float64
		description string
	}{
		{
			name:        "Exact match",
			email:       "testuser@example.com",
			username:    "testuser",
			expectedMin: 0.9,
			expectedMax: 1.0,
			description: "Exact email username match should have high similarity",
		},
		{
			name:        "Similar with numbers",
			email:       "testuser123@example.com",
			username:    "testuser",
			expectedMin: 0.6,
			expectedMax: 1.0, // Allow for exact match after normalization
			description: "Username with numbers should have good similarity",
		},
		{
			name:        "Different case",
			email:       "TestUser@example.com",
			username:    "testuser",
			expectedMin: 0.8,
			expectedMax: 1.0,
			description: "Case differences should be handled",
		},
		{
			name:        "With dots",
			email:       "test.user@example.com",
			username:    "testuser",
			expectedMin: 0.7,
			expectedMax: 1.0, // Allow for exact match after normalization
			description: "Dots in email should be handled",
		},
		{
			name:        "With underscores",
			email:       "test_user@example.com",
			username:    "testuser",
			expectedMin: 0.7,
			expectedMax: 1.0, // Allow for exact match after normalization
			description: "Underscores in email should be handled",
		},
		{
			name:        "Very different",
			email:       "completely@example.com",
			username:    "testuser",
			expectedMin: 0.0,
			expectedMax: 0.3,
			description: "Very different email and username should have low similarity",
		},
		{
			name:        "Empty email",
			email:       "@example.com",
			username:    "testuser",
			expectedMin: 0.0,
			expectedMax: 0.0,
			description: "Empty email username should have 0 similarity",
		},
		{
			name:        "Empty username",
			email:       "testuser@example.com",
			username:    "",
			expectedMin: 0.0,
			expectedMax: 0.0,
			description: "Empty username should have 0 similarity",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			similarity := service.CalculateEmailUsernameSimilarity(tc.email, tc.username)

			assert.GreaterOrEqual(t, similarity, tc.expectedMin,
				"%s: similarity should be >= %f, got %f", tc.description, tc.expectedMin, similarity)
			assert.LessOrEqual(t, similarity, tc.expectedMax,
				"%s: similarity should be <= %f, got %f", tc.description, tc.expectedMax, similarity)
		})
	}
}

func TestSortEmailsBySimilarity(t *testing.T) {
	service := NewTextSimilarityService()

	t.Run("Sort emails by similarity to username", func(t *testing.T) {
		emails := []string{
			"testuser@example.com",
			"test.user@example.com",
			"testuser123@example.com",
			"completely@example.com",
			"test_user@example.com",
		}
		username := "testuser"

		similarities := service.SortEmailsBySimilarity(emails, username)

		// Should have same number of results as input
		assert.Equal(t, len(emails), len(similarities))

		// Should be sorted by similarity (highest first)
		for i := 0; i < len(similarities)-1; i++ {
			assert.GreaterOrEqual(t, similarities[i].Similarity, similarities[i+1].Similarity,
				"Emails should be sorted by similarity (highest first)")
		}

		// First email should be the most similar
		assert.Equal(t, "testuser@example.com", similarities[0].Email)
		assert.Greater(t, similarities[0].Similarity, 0.8, "Most similar email should have high similarity")
	})

	t.Run("Empty email list", func(t *testing.T) {
		emails := []string{}
		username := "testuser"

		similarities := service.SortEmailsBySimilarity(emails, username)

		assert.Equal(t, 0, len(similarities), "Empty email list should return empty result")
	})

	t.Run("Single email", func(t *testing.T) {
		emails := []string{"testuser@example.com"}
		username := "testuser"

		similarities := service.SortEmailsBySimilarity(emails, username)

		assert.Equal(t, 1, len(similarities), "Single email should return single result")
		assert.Equal(t, "testuser@example.com", similarities[0].Email)
		assert.Greater(t, similarities[0].Similarity, 0.8, "Single email should have high similarity")
	})
}

func TestNormalizeString(t *testing.T) {
	service := NewTextSimilarityService()

	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Lowercase conversion",
			input:    "TestUser",
			expected: "testuser",
		},
		{
			name:     "Remove special characters",
			input:    "test-user_123",
			expected: "testuser123",
		},
		{
			name:     "Keep alphanumeric",
			input:    "test123user",
			expected: "test123user",
		},
		{
			name:     "Remove all special chars",
			input:    "test@user#123$",
			expected: "testuser123",
		},
		{
			name:     "Empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "Only special characters",
			input:    "@#$%^&*()",
			expected: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := service.normalizeString(tc.input)
			assert.Equal(t, tc.expected, result, "String normalization should work correctly")
		})
	}
}

func TestLevenshteinDistance(t *testing.T) {
	service := NewTextSimilarityService()

	testCases := []struct {
		name     string
		str1     string
		str2     string
		expected int
	}{
		{
			name:     "Identical strings",
			str1:     "test",
			str2:     "test",
			expected: 0,
		},
		{
			name:     "One character difference",
			str1:     "test",
			str2:     "tost",
			expected: 1,
		},
		{
			name:     "Completely different",
			str1:     "test",
			str2:     "hello",
			expected: 4,
		},
		{
			name:     "Empty strings",
			str1:     "",
			str2:     "",
			expected: 0,
		},
		{
			name:     "One empty string",
			str1:     "test",
			str2:     "",
			expected: 4,
		},
		{
			name:     "Insertion",
			str1:     "test",
			str2:     "tests",
			expected: 1,
		},
		{
			name:     "Deletion",
			str1:     "tests",
			str2:     "test",
			expected: 1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := service.levenshteinDistance(tc.str1, tc.str2)
			assert.Equal(t, tc.expected, result, "Levenshtein distance should be calculated correctly")
		})
	}
}
