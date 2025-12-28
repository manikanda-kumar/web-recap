package readinglist

import (
	"fmt"
	"time"

	"github.com/rzolkos/web-recap/internal/models"
)

// MediumHandler manages fetching from Medium reading list
type MediumHandler struct {
	config     *Config
	strategies []FetchStrategy
}

// NewMediumHandler creates a Medium reading list handler
func NewMediumHandler(config *Config) *MediumHandler {
	h := &MediumHandler{
		config: config,
	}

	// Register strategies in priority order
	// Try public web scraping first if URL is provided,
	// then authenticated web scraping, then fall back to file parsing
	h.strategies = []FetchStrategy{
		NewMediumPublicWebStrategy(config),
		NewMediumWebStrategy(config),
		NewMediumJSONStrategy(config),
		NewMediumCSVStrategy(config),
	}

	return h
}

// GetReadingList retrieves Medium reading list entries
// Tries each strategy in order until one succeeds
func (h *MediumHandler) GetReadingList(startDate, endDate time.Time) ([]models.ReadingListEntry, error) {
	var lastErr error
	var attemptedStrategies []string

	for _, strategy := range h.strategies {
		if !strategy.IsAvailable() {
			continue
		}

		attemptedStrategies = append(attemptedStrategies, strategy.Name())

		entries, err := strategy.Fetch(h.config, startDate, endDate)
		if err == nil {
			return entries, nil
		}

		lastErr = err
	}

	if lastErr != nil {
		return nil, fmt.Errorf("all Medium strategies failed (tried: %v): %w", attemptedStrategies, lastErr)
	}

	return nil, fmt.Errorf("no available Medium strategies (provide --cookie for web scraping or --file for CSV parsing)")
}
