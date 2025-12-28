package readinglist

import (
	"time"

	"github.com/rzolkos/web-recap/internal/models"
)

// filterByDateRange filters reading list entries by date range
func filterByDateRange(entries []models.ReadingListEntry, startDate, endDate time.Time) []models.ReadingListEntry {
	if startDate.IsZero() && endDate.IsZero() {
		return entries
	}

	var filtered []models.ReadingListEntry

	for _, entry := range entries {
		if entry.SavedAt.IsZero() {
			continue
		}

		if !startDate.IsZero() && entry.SavedAt.Before(startDate) {
			continue
		}

		if !endDate.IsZero() && entry.SavedAt.After(endDate) {
			continue
		}

		filtered = append(filtered, entry)
	}

	return filtered
}
