package readinglist

import (
	"encoding/csv"
	"fmt"
	"os"
	"time"

	"github.com/rzolkos/web-recap/internal/database"
	"github.com/rzolkos/web-recap/internal/models"
)

// MediumCSVStrategy parses manually exported CSV files from Medium
type MediumCSVStrategy struct {
	config *Config
}

// NewMediumCSVStrategy creates a new Medium CSV parsing strategy
func NewMediumCSVStrategy(config *Config) *MediumCSVStrategy {
	return &MediumCSVStrategy{config: config}
}

// Name returns the strategy name
func (s *MediumCSVStrategy) Name() string {
	return "csv"
}

// IsAvailable checks if the CSV file exists and is accessible
func (s *MediumCSVStrategy) IsAvailable() bool {
	if s.config.FilePath == "" {
		return false
	}

	// Check if file exists
	_, err := os.Stat(s.config.FilePath)
	return err == nil
}

// Fetch parses the CSV file and returns reading list entries
func (s *MediumCSVStrategy) Fetch(config *Config, startDate, endDate time.Time) ([]models.ReadingListEntry, error) {
	file, err := os.Open(config.FilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open CSV: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to parse CSV: %w", err)
	}

	if len(records) < 2 {
		return nil, fmt.Errorf("CSV file is empty or has no data rows")
	}

	// Parse header to find column indices
	header := records[0]
	columnMap := make(map[string]int)
	for i, col := range header {
		columnMap[col] = i
	}

	var entries []models.ReadingListEntry

	for _, record := range records[1:] {
		entry := models.ReadingListEntry{
			Platform: "medium",
		}

		if idx, ok := columnMap["title"]; ok && idx < len(record) {
			entry.Title = record[idx]
		}
		if idx, ok := columnMap["url"]; ok && idx < len(record) {
			entry.URL = record[idx]
			entry.Domain = database.ExtractDomain(entry.URL)
		}
		if idx, ok := columnMap["author"]; ok && idx < len(record) {
			entry.Author = record[idx]
		}
		if idx, ok := columnMap["publication"]; ok && idx < len(record) {
			entry.Publication = record[idx]
		}
		if idx, ok := columnMap["saved_at"]; ok && idx < len(record) {
			// Try multiple date formats
			dateStr := record[idx]
			if t, err := time.Parse(time.RFC3339, dateStr); err == nil {
				entry.SavedAt = t
			} else if t, err := time.Parse("2006-01-02", dateStr); err == nil {
				entry.SavedAt = t
			} else if t, err := time.Parse("2006-01-02 15:04:05", dateStr); err == nil {
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
