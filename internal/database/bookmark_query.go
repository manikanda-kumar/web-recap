package database

import (
	"fmt"
	"sort"
	"time"

	"github.com/rzolkos/web-recap/internal/browser"
	"github.com/rzolkos/web-recap/internal/models"
)

// BookmarkQuerier defines the interface for querying browser bookmarks
type BookmarkQuerier interface {
	GetBookmarks(startTime, endTime time.Time) ([]models.BookmarkEntry, error)
}

// NewBookmarkQuerier creates a new bookmark querier for the given browser
func NewBookmarkQuerier(b *browser.Browser, bookmarkPath string) (BookmarkQuerier, error) {
	switch b.Type {
	case browser.Chrome, browser.Chromium, browser.Edge, browser.Brave, browser.Vivaldi:
		return NewChromeBookmarkHandler(bookmarkPath, string(b.Type)), nil
	case browser.Firefox:
		return NewFirefoxBookmarkHandler(bookmarkPath), nil
	case browser.Safari:
		return NewSafariBookmarkHandler(bookmarkPath), nil
	default:
		return nil, ErrUnsupportedBrowser
	}
}

// QueryBookmarks retrieves bookmark entries from a specific browser
func QueryBookmarks(b *browser.Browser, bookmarkPath string, startTime, endTime time.Time) ([]models.BookmarkEntry, error) {
	querier, err := NewBookmarkQuerier(b, bookmarkPath)
	if err != nil {
		return nil, err
	}

	entries, err := querier.GetBookmarks(startTime, endTime)
	if err != nil {
		return nil, err
	}

	// Sort by date added descending (most recent first)
	sort.Slice(entries, func(i, j int) bool {
		return bookmarkEntryLess(entries[i], entries[j])
	})

	return entries, nil
}

// QueryMultipleBrowsersBookmarks retrieves bookmarks from all detected browsers
func QueryMultipleBrowsersBookmarks(detector *browser.Detector, startTime, endTime time.Time) ([]models.BookmarkEntry, []string) {
	var allEntries []models.BookmarkEntry
	var warnings []string

	detectedBrowsers := detector.Detect()
	for _, b := range detectedBrowsers {
		br := b // Copy to avoid pointer issues

		// Get bookmark path for this browser
		bookmarkPath, err := browser.GetBookmarkPath(br.Type)
		if err != nil {
			warnings = append(warnings, fmt.Sprintf("%s: failed to resolve bookmark path: %v", br.Type, err))
			continue
		}

		// Check if bookmark file exists
		if bookmarkPath == "" {
			warnings = append(warnings, fmt.Sprintf("%s: bookmark path is empty", br.Type))
			continue
		}

		// For Firefox, we need to find the profile
		if br.Type == browser.Firefox {
			bookmarkPath, err = browser.GetFirefoxProfilePath(bookmarkPath)
			if err != nil {
				warnings = append(warnings, fmt.Sprintf("%s: failed to resolve profile path: %v", br.Type, err))
				continue
			}
		}

		entries, err := QueryBookmarks(&br, bookmarkPath, startTime, endTime)
		if err != nil {
			warnings = append(warnings, fmt.Sprintf("%s: failed to query bookmarks: %v", br.Type, err))
			continue
		}
		allEntries = append(allEntries, entries...)
	}

	// Sort all entries by date added descending
	sort.Slice(allEntries, func(i, j int) bool {
		return bookmarkEntryLess(allEntries[i], allEntries[j])
	})

	return allEntries, warnings
}

func bookmarkEntryLess(a, b models.BookmarkEntry) bool {
	aHasDate := !a.DateAdded.IsZero()
	bHasDate := !b.DateAdded.IsZero()

	if aHasDate && !bHasDate {
		return true
	}
	if !aHasDate && bHasDate {
		return false
	}
	if aHasDate && bHasDate {
		return a.DateAdded.After(b.DateAdded)
	}

	if a.Title != b.Title {
		return a.Title < b.Title
	}
	return a.URL < b.URL
}
