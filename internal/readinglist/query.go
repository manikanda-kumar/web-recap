package readinglist

import (
	"fmt"
	"sort"
	"time"

	"github.com/rzolkos/web-recap/internal/models"
)

// PlatformType represents supported reading list platforms
type PlatformType string

const (
	// PlatformMedium represents Medium reading lists
	PlatformMedium PlatformType = "medium"
	// PlatformSubstack represents Substack saved posts
	PlatformSubstack PlatformType = "substack"
	// PlatformAll represents all platforms
	PlatformAll PlatformType = "all"
)

// NewQuerier creates a reading list querier for the given platform
func NewQuerier(platform PlatformType, config *Config) (ReadingListQuerier, error) {
	switch platform {
	case PlatformMedium:
		return NewMediumHandler(config), nil
	case PlatformSubstack:
		return NewSubstackHandler(config), nil
	default:
		return nil, fmt.Errorf("%w: %s", ErrUnsupportedPlatform, platform)
	}
}

// Query retrieves reading list entries from a specific platform
func Query(platform PlatformType, config *Config, startDate, endDate time.Time) ([]models.ReadingListEntry, error) {
	querier, err := NewQuerier(platform, config)
	if err != nil {
		return nil, err
	}

	entries, err := querier.GetReadingList(startDate, endDate)
	if err != nil {
		return nil, err
	}

	// Sort by saved date descending (most recent first)
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].SavedAt.After(entries[j].SavedAt)
	})

	return entries, nil
}

// QueryMultiplePlatforms retrieves reading lists from all specified platforms
func QueryMultiplePlatforms(platforms []PlatformType, configs map[PlatformType]*Config, startDate, endDate time.Time) ([]models.ReadingListEntry, error) {
	var allEntries []models.ReadingListEntry
	var errors []error

	for _, platform := range platforms {
		config, ok := configs[platform]
		if !ok {
			continue
		}

		entries, err := Query(platform, config, startDate, endDate)
		if err != nil {
			// Collect errors but continue with other platforms
			errors = append(errors, fmt.Errorf("%s: %w", platform, err))
			continue
		}

		allEntries = append(allEntries, entries...)
	}

	// If no entries were retrieved, return the errors
	if len(allEntries) == 0 && len(errors) > 0 {
		return nil, fmt.Errorf("failed to retrieve from any platform: %v", errors)
	}

	// Sort all entries by saved date descending
	sort.Slice(allEntries, func(i, j int) bool {
		return allEntries[i].SavedAt.After(allEntries[j].SavedAt)
	})

	return allEntries, nil
}
