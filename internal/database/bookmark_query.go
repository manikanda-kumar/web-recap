package database

import (
	"sort"

	"github.com/rzolkos/web-recap/internal/browser"
	"github.com/rzolkos/web-recap/internal/models"
)

// BookmarkQuerier defines the interface for querying browser bookmarks
type BookmarkQuerier interface {
	GetBookmarks() ([]models.BookmarkEntry, error)
}

// NewBookmarkQuerier creates a new bookmark querier for the given browser
func NewBookmarkQuerier(b *browser.Browser, bookmarkPath string) (BookmarkQuerier, error) {
	switch b.Type {
	case browser.Chrome, browser.Chromium, browser.Edge, browser.Brave, browser.Vivaldi:
		return NewChromeBookmarkHandler(bookmarkPath, b.Name), nil
	case browser.Firefox:
		return NewFirefoxBookmarkHandler(bookmarkPath), nil
	case browser.Safari:
		return NewSafariBookmarkHandler(bookmarkPath), nil
	default:
		return nil, ErrUnsupportedBrowser
	}
}

// QueryBookmarks retrieves bookmark entries from a specific browser
func QueryBookmarks(b *browser.Browser, bookmarkPath string) ([]models.BookmarkEntry, error) {
	querier, err := NewBookmarkQuerier(b, bookmarkPath)
	if err != nil {
		return nil, err
	}

	entries, err := querier.GetBookmarks()
	if err != nil {
		return nil, err
	}

	// Sort by date added descending (most recent first)
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].DateAdded.After(entries[j].DateAdded)
	})

	return entries, nil
}

// QueryMultipleBrowsersBookmarks retrieves bookmarks from all detected browsers
func QueryMultipleBrowsersBookmarks(detector *browser.Detector) ([]models.BookmarkEntry, error) {
	var allEntries []models.BookmarkEntry

	detectedBrowsers := detector.Detect()
	for _, b := range detectedBrowsers {
		br := b // Copy to avoid pointer issues

		// Get bookmark path for this browser
		bookmarkPath, err := browser.GetBookmarkPath(br.Type)
		if err != nil {
			continue
		}

		// Check if bookmark file exists
		if bookmarkPath == "" {
			continue
		}

		// For Firefox, we need to find the profile
		if br.Type == browser.Firefox {
			bookmarkPath, err = browser.GetFirefoxProfilePath(bookmarkPath)
			if err != nil {
				continue
			}
		}

		entries, err := QueryBookmarks(&br, bookmarkPath)
		if err != nil {
			// Log error but continue with other browsers
			continue
		}
		allEntries = append(allEntries, entries...)
	}

	// Sort all entries by date added descending
	sort.Slice(allEntries, func(i, j int) bool {
		return allEntries[i].DateAdded.After(allEntries[j].DateAdded)
	})

	return allEntries, nil
}
