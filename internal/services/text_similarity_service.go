package services

import (
	"math"
	"strings"
	"unicode"
)

type TextSimilarityService struct{}

func NewTextSimilarityService() *TextSimilarityService {
	return &TextSimilarityService{}
}

// EmailSimilarity represents the similarity between an email and a username
type EmailSimilarity struct {
	Email      string
	Similarity float64
}

// CalculateSimilarity calculates the similarity between two strings
// Returns a value between 0 (completely different) and 1 (identical)
func (s *TextSimilarityService) CalculateSimilarity(str1, str2 string) float64 {
	if str1 == str2 {
		return 1.0
	}

	if len(str1) == 0 || len(str2) == 0 {
		return 0.0
	}

	// Normalize strings (lowercase, remove special chars)
	normalized1 := s.normalizeString(str1)
	normalized2 := s.normalizeString(str2)

	// Calculate Levenshtein distance
	distance := s.levenshteinDistance(normalized1, normalized2)
	maxLen := float64(max(len(normalized1), len(normalized2)))

	// Convert distance to similarity (0 = identical, 1 = completely different)
	similarity := 1.0 - (float64(distance) / maxLen)

	// Boost similarity for partial matches
	partialBoost := s.calculatePartialMatchBoost(normalized1, normalized2)
	similarity = math.Min(1.0, similarity+partialBoost)

	return similarity
}

// CalculateEmailUsernameSimilarity calculates similarity between email and GitHub username
func (s *TextSimilarityService) CalculateEmailUsernameSimilarity(email, username string) float64 {
	// Extract email username (part before @)
	emailParts := strings.Split(email, "@")
	if len(emailParts) == 0 {
		return 0.0
	}
	emailUsername := emailParts[0]

	// Calculate base similarity
	baseSimilarity := s.CalculateSimilarity(emailUsername, username)

	// Additional checks for common patterns
	patternBonus := s.calculatePatternBonus(emailUsername, username)

	return math.Min(1.0, baseSimilarity+patternBonus)
}

// SortEmailsBySimilarity sorts emails by their similarity to a username
func (s *TextSimilarityService) SortEmailsBySimilarity(emails []string, username string) []EmailSimilarity {
	var similarities []EmailSimilarity

	for _, email := range emails {
		similarity := s.CalculateEmailUsernameSimilarity(email, username)
		similarities = append(similarities, EmailSimilarity{
			Email:      email,
			Similarity: similarity,
		})
	}

	// Sort by similarity (highest first)
	for i := 0; i < len(similarities)-1; i++ {
		for j := i + 1; j < len(similarities); j++ {
			if similarities[i].Similarity < similarities[j].Similarity {
				similarities[i], similarities[j] = similarities[j], similarities[i]
			}
		}
	}

	return similarities
}

// normalizeString normalizes a string for comparison
func (s *TextSimilarityService) normalizeString(str string) string {
	// Convert to lowercase
	str = strings.ToLower(str)

	// Remove special characters and keep only alphanumeric
	var result strings.Builder
	for _, r := range str {
		if unicode.IsLetter(r) || unicode.IsNumber(r) {
			result.WriteRune(r)
		}
	}

	return result.String()
}

// levenshteinDistance calculates the Levenshtein distance between two strings
func (s *TextSimilarityService) levenshteinDistance(str1, str2 string) int {
	len1, len2 := len(str1), len(str2)

	// Create matrix
	matrix := make([][]int, len1+1)
	for i := range matrix {
		matrix[i] = make([]int, len2+1)
	}

	// Initialize first row and column
	for i := 0; i <= len1; i++ {
		matrix[i][0] = i
	}
	for j := 0; j <= len2; j++ {
		matrix[0][j] = j
	}

	// Fill matrix
	for i := 1; i <= len1; i++ {
		for j := 1; j <= len2; j++ {
			if str1[i-1] == str2[j-1] {
				matrix[i][j] = matrix[i-1][j-1]
			} else {
				matrix[i][j] = min(
					matrix[i-1][j]+1,   // deletion
					matrix[i][j-1]+1,   // insertion
					matrix[i-1][j-1]+1, // substitution
				)
			}
		}
	}

	return matrix[len1][len2]
}

// calculatePartialMatchBoost gives bonus for partial matches
func (s *TextSimilarityService) calculatePartialMatchBoost(str1, str2 string) float64 {
	boost := 0.0

	// Check if one string contains the other
	if strings.Contains(str1, str2) || strings.Contains(str2, str1) {
		boost += 0.2
	}

	// Check for common prefixes
	minLen := min2(len(str1), len(str2))
	if minLen > 0 {
		commonPrefix := 0
		for i := 0; i < minLen && str1[i] == str2[i]; i++ {
			commonPrefix++
		}
		if commonPrefix > 0 {
			boost += float64(commonPrefix) / float64(max(len(str1), len(str2))) * 0.1
		}
	}

	return boost
}

// calculatePatternBonus gives bonus for common email patterns
func (s *TextSimilarityService) calculatePatternBonus(emailUsername, githubUsername string) float64 {
	bonus := 0.0

	// Check for exact username match
	if emailUsername == githubUsername {
		bonus += 0.3
	}

	// Check for username with common suffixes
	commonSuffixes := []string{"dev", "admin", "user", "test", "demo", "temp"}
	for _, suffix := range commonSuffixes {
		if strings.HasSuffix(emailUsername, suffix) {
			baseUsername := strings.TrimSuffix(emailUsername, suffix)
			if baseUsername == githubUsername {
				bonus += 0.2
			}
		}
	}

	// Check for username with common prefixes
	commonPrefixes := []string{"dev", "admin", "user", "test", "demo", "temp"}
	for _, prefix := range commonPrefixes {
		if strings.HasPrefix(emailUsername, prefix) {
			baseUsername := strings.TrimPrefix(emailUsername, prefix)
			if baseUsername == githubUsername {
				bonus += 0.2
			}
		}
	}

	// Check for username with dots or underscores
	if strings.Contains(emailUsername, ".") {
		parts := strings.Split(emailUsername, ".")
		for _, part := range parts {
			if part == githubUsername {
				bonus += 0.15
			}
		}
	}

	if strings.Contains(emailUsername, "_") {
		parts := strings.Split(emailUsername, "_")
		for _, part := range parts {
			if part == githubUsername {
				bonus += 0.15
			}
		}
	}

	return bonus
}

// Helper functions
func min(a, b, c int) int {
	if a <= b && a <= c {
		return a
	}
	if b <= a && b <= c {
		return b
	}
	return c
}

func min2(a, b int) int {
	if a <= b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
