package readinglist

import (
	"fmt"
	"time"

	"github.com/rzolkos/web-recap/internal/models"
)

// SubstackHandler manages fetching from Substack saved posts
type SubstackHandler struct {
	config     *Config
	strategies []FetchStrategy
}

// NewSubstackHandler creates a Substack saved posts handler
func NewSubstackHandler(config *Config) *SubstackHandler {
	h := &SubstackHandler{
		config: config,
	}

	// Register strategies in priority order
	// Try web API first, then fall back to JSON file
	h.strategies = []FetchStrategy{
		NewSubstackWebStrategy(config),
		NewSubstackJSONStrategy(config),
	}

	return h
}

// GetReadingList retrieves Substack saved posts entries
// Tries each strategy in order until one succeeds
func (h *SubstackHandler) GetReadingList(startDate, endDate time.Time) ([]models.ReadingListEntry, error) {
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
		return nil, fmt.Errorf("all Substack strategies failed (tried: %v): %w", attemptedStrategies, lastErr)
	}

	return nil, fmt.Errorf("no available Substack strategies (provide --cookie or --session-token for web API or --file for JSON parsing)")
}
