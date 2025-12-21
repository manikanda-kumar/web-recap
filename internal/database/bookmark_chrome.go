package database

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/rzolkos/web-recap/internal/models"
)

// ChromeBookmarkHandler handles Chrome/Chromium/Edge/Brave bookmark extraction
type ChromeBookmarkHandler struct {
	bookmarkPath string
	browserName  string
}

// NewChromeBookmarkHandler creates a new Chrome bookmark handler
func NewChromeBookmarkHandler(bookmarkPath, browserName string) *ChromeBookmarkHandler {
	return &ChromeBookmarkHandler{
		bookmarkPath: bookmarkPath,
		browserName:  browserName,
	}
}

// Chrome bookmark JSON structure
type chromeBookmarkFile struct {
	Checksum string                 `json:"checksum"`
	Roots    chromeBookmarkRoots    `json:"roots"`
	Version  int                    `json:"version"`
}

type chromeBookmarkRoots struct {
	BookmarkBar chromeBookmarkNode `json:"bookmark_bar"`
	Other       chromeBookmarkNode `json:"other"`
	Synced      chromeBookmarkNode `json:"synced"`
}

type chromeBookmarkNode struct {
	DateAdded      string               `json:"date_added"`
	DateModified   string               `json:"date_modified,omitempty"`
	GUID           string               `json:"guid"`
	ID             string               `json:"id"`
	Name           string               `json:"name"`
	Type           string               `json:"type"` // "folder" or "url"
	URL            string               `json:"url,omitempty"`
	Children       []chromeBookmarkNode `json:"children,omitempty"`
	MetaInfo       map[string]string    `json:"meta_info,omitempty"`
}

// GetBookmarks retrieves all bookmarks from Chrome
func (h *ChromeBookmarkHandler) GetBookmarks() ([]models.BookmarkEntry, error) {
	data, err := os.ReadFile(h.bookmarkPath)
	if err != nil {
		return nil, err
	}

	var bookmarkFile chromeBookmarkFile
	if err := json.Unmarshal(data, &bookmarkFile); err != nil {
		return nil, err
	}

	var bookmarks []models.BookmarkEntry

	// Extract from all root folders
	bookmarks = append(bookmarks, h.extractFromNode(bookmarkFile.Roots.BookmarkBar, "Bookmarks Bar")...)
	bookmarks = append(bookmarks, h.extractFromNode(bookmarkFile.Roots.Other, "Other Bookmarks")...)
	bookmarks = append(bookmarks, h.extractFromNode(bookmarkFile.Roots.Synced, "Synced Bookmarks")...)

	return bookmarks, nil
}

// extractFromNode recursively extracts bookmarks from a node
func (h *ChromeBookmarkHandler) extractFromNode(node chromeBookmarkNode, folderPath string) []models.BookmarkEntry {
	var bookmarks []models.BookmarkEntry

	if node.Type == "url" {
		// This is a bookmark
		dateAdded := h.convertChromeTimestamp(node.DateAdded)
		dateModified := h.convertChromeTimestamp(node.DateModified)

		bookmarks = append(bookmarks, models.BookmarkEntry{
			DateAdded:    dateAdded,
			DateModified: dateModified,
			URL:          node.URL,
			Title:        node.Name,
			Folder:       folderPath,
			Domain:       ExtractDomain(node.URL),
			Browser:      h.browserName,
		})
	} else if node.Type == "folder" {
		// Recursively extract from children
		newFolderPath := folderPath
		if node.Name != "" {
			if folderPath != "" {
				newFolderPath = folderPath + "/" + node.Name
			} else {
				newFolderPath = node.Name
			}
		}

		for _, child := range node.Children {
			bookmarks = append(bookmarks, h.extractFromNode(child, newFolderPath)...)
		}
	}

	return bookmarks
}

// convertChromeTimestamp converts Chrome's timestamp to time.Time
// Chrome uses microseconds since 1601-01-01 00:00:00 UTC
func (h *ChromeBookmarkHandler) convertChromeTimestamp(timestampStr string) time.Time {
	if timestampStr == "" {
		return time.Time{}
	}

	var timestamp int64
	if _, err := fmt.Sscanf(timestampStr, "%d", &timestamp); err != nil {
		return time.Time{}
	}

	return ConvertChromeTimestamp(timestamp)
}
