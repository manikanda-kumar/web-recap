package output

import (
	"encoding/json"
	"io"
	"time"

	"github.com/rzolkos/web-recap/internal/models"
)

// FormatJSON writes history report as JSON to the given writer
func FormatJSON(w io.Writer, entries []models.HistoryEntry, browser string, startDate, endDate time.Time, tz string) error {
	if tz == "" {
		tz = "UTC"
	}

	report := models.HistoryReport{
		Browser:      browser,
		StartDate:    startDate,
		EndDate:      endDate,
		Timezone:     tz,
		TotalEntries: len(entries),
		Entries:      entries,
	}

	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	encoder.SetEscapeHTML(false)

	return encoder.Encode(report)
}

// FormatJSONCompact writes history report as compact JSON to the given writer
func FormatJSONCompact(w io.Writer, entries []models.HistoryEntry, browser string, startDate, endDate time.Time) error {
	report := models.HistoryReport{
		Browser:      browser,
		StartDate:    startDate,
		EndDate:      endDate,
		TotalEntries: len(entries),
		Entries:      entries,
	}

	encoder := json.NewEncoder(w)
	encoder.SetEscapeHTML(false)

	return encoder.Encode(report)
}

// FormatJSONLines writes history entries as JSON lines (one per line) to the given writer
func FormatJSONLines(w io.Writer, entries []models.HistoryEntry) error {
	encoder := json.NewEncoder(w)
	encoder.SetEscapeHTML(false)

	for _, entry := range entries {
		if err := encoder.Encode(entry); err != nil {
			return err
		}
	}

	return nil
}

// FormatBookmarksJSON writes bookmark report as JSON to the given writer
func FormatBookmarksJSON(w io.Writer, entries []models.BookmarkEntry, browser string) error {
	report := models.BookmarkReport{
		Browser:      browser,
		TotalEntries: len(entries),
		Entries:      entries,
	}

	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	encoder.SetEscapeHTML(false)

	return encoder.Encode(report)
}

// FormatBookmarksJSONCompact writes bookmark report as compact JSON to the given writer
func FormatBookmarksJSONCompact(w io.Writer, entries []models.BookmarkEntry, browser string) error {
	report := models.BookmarkReport{
		Browser:      browser,
		TotalEntries: len(entries),
		Entries:      entries,
	}

	encoder := json.NewEncoder(w)
	encoder.SetEscapeHTML(false)

	return encoder.Encode(report)
}

// FormatBookmarksJSONLines writes bookmark entries as JSON lines (one per line) to the given writer
func FormatBookmarksJSONLines(w io.Writer, entries []models.BookmarkEntry) error {
	encoder := json.NewEncoder(w)
	encoder.SetEscapeHTML(false)

	for _, entry := range entries {
		if err := encoder.Encode(entry); err != nil {
			return err
		}
	}

	return nil
}
