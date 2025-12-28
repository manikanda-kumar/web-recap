package readinglist

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/rzolkos/web-recap/internal/database"
	"github.com/rzolkos/web-recap/internal/models"
)

// MediumJSONStrategy parses JSON files exported from Medium reading lists
type MediumJSONStrategy struct {
	config *Config
}

// NewMediumJSONStrategy creates a new Medium JSON parsing strategy
func NewMediumJSONStrategy(config *Config) *MediumJSONStrategy {
	return &MediumJSONStrategy{config: config}
}

// Name returns the strategy name
func (s *MediumJSONStrategy) Name() string {
	return "json"
}

// IsAvailable checks if the JSON file exists and is accessible
func (s *MediumJSONStrategy) IsAvailable() bool {
	if s.config.FilePath == "" {
		return false
	}

	// Only handle JSON files
	if !strings.HasSuffix(strings.ToLower(s.config.FilePath), ".json") {
		return false
	}

	// Check if file exists
	_, err := os.Stat(s.config.FilePath)
	return err == nil
}

// mediumJSONExport represents the structure of the exported JSON file
type mediumJSONExport struct {
	Platform     string              `json:"platform"`
	ExportedAt   string              `json:"exported_at"`
	TotalEntries int                 `json:"total_entries"`
	Entries      []mediumJSONEntry   `json:"entries"`
}

// mediumJSONEntry represents a single article in the JSON export
type mediumJSONEntry struct {
	URL         string `json:"url"`
	Title       string `json:"title"`
	Author      string `json:"author"`
	Publication string `json:"publication"`
	Excerpt     string `json:"excerpt"`
	SavedAt     string `json:"saved_at"`
	Platform    string `json:"platform"`
}

// Fetch parses the JSON file and returns reading list entries
func (s *MediumJSONStrategy) Fetch(config *Config, startDate, endDate time.Time) ([]models.ReadingListEntry, error) {
	file, err := os.Open(config.FilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open JSON file: %w", err)
	}
	defer file.Close()

	var export mediumJSONExport
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&export); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	if len(export.Entries) == 0 {
		return nil, fmt.Errorf("JSON file contains no entries")
	}

	var entries []models.ReadingListEntry

	for _, item := range export.Entries {
		entry := models.ReadingListEntry{
			Platform:    "medium",
			Title:       item.Title,
			URL:         item.URL,
			Author:      item.Author,
			Publication: item.Publication,
			Excerpt:     item.Excerpt,
			Domain:      database.ExtractDomain(item.URL),
		}

		// Parse saved date if available
		if item.SavedAt != "" {
			// Try RFC3339 first (ISO 8601)
			if t, err := time.Parse(time.RFC3339, item.SavedAt); err == nil {
				entry.SavedAt = t
			} else if t, err := time.Parse("2006-01-02", item.SavedAt); err == nil {
				entry.SavedAt = t
			} else if t, err := time.Parse("2006-01-02 15:04:05", item.SavedAt); err == nil {
				entry.SavedAt = t
			} else if t, err := time.Parse("Jan 2, 2006", item.SavedAt); err == nil {
				entry.SavedAt = t
			} else if t, err := time.Parse("Jan 2", item.SavedAt); err == nil {
				// Assume current year
				now := time.Now()
				t = time.Date(now.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
				// If date is in the future, assume previous year
				if t.After(now) {
					t = t.AddDate(-1, 0, 0)
				}
				entry.SavedAt = t
			}
		}

		entries = append(entries, entry)
	}

	// Filter by date range
	if !startDate.IsZero() || !endDate.IsZero() {
		entries = filterByDateRange(entries, startDate, endDate)
	}

	return entries, nil
}
