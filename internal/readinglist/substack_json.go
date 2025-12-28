package readinglist

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/rzolkos/web-recap/internal/database"
	"github.com/rzolkos/web-recap/internal/models"
)

// SubstackJSONStrategy parses manually exported JSON files from Substack
type SubstackJSONStrategy struct {
	config *Config
}

// NewSubstackJSONStrategy creates a new Substack JSON parsing strategy
func NewSubstackJSONStrategy(config *Config) *SubstackJSONStrategy {
	return &SubstackJSONStrategy{config: config}
}

// Name returns the strategy name
func (s *SubstackJSONStrategy) Name() string {
	return "json"
}

// IsAvailable checks if the JSON file exists and is accessible
func (s *SubstackJSONStrategy) IsAvailable() bool {
	if s.config.FilePath == "" {
		return false
	}

	_, err := os.Stat(s.config.FilePath)
	return err == nil
}

// Fetch parses the JSON file and returns reading list entries
func (s *SubstackJSONStrategy) Fetch(config *Config, startDate, endDate time.Time) ([]models.ReadingListEntry, error) {
	file, err := os.Open(config.FilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open JSON: %w", err)
	}
	defer file.Close()

	var data struct {
		SavedPosts []struct {
			Title       string    `json:"title"`
			URL         string    `json:"url"`
			Author      string    `json:"author"`
			Publication string    `json:"publication"`
			Excerpt     string    `json:"excerpt"`
			SavedAt     time.Time `json:"saved_at"`
		} `json:"saved_posts"`
	}

	if err := json.NewDecoder(file).Decode(&data); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	var entries []models.ReadingListEntry

	for _, post := range data.SavedPosts {
		entry := models.ReadingListEntry{
			Platform:    "substack",
			Title:       post.Title,
			URL:         post.URL,
			Author:      post.Author,
			Publication: post.Publication,
			Excerpt:     post.Excerpt,
			SavedAt:     post.SavedAt,
			Domain:      database.ExtractDomain(post.URL),
		}

		entries = append(entries, entry)
	}

	// Filter by date range
	if !startDate.IsZero() || !endDate.IsZero() {
		entries = filterByDateRange(entries, startDate, endDate)
	}

	return entries, nil
}
