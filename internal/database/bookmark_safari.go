package database

import (
	"os"
	"time"

	"github.com/rzolkos/web-recap/internal/models"
	"howett.net/plist"
)

// SafariBookmarkHandler handles Safari bookmark extraction
type SafariBookmarkHandler struct {
	plistPath string
	startTime time.Time
	endTime   time.Time
}

// NewSafariBookmarkHandler creates a new Safari bookmark handler
func NewSafariBookmarkHandler(plistPath string) *SafariBookmarkHandler {
	return &SafariBookmarkHandler{
		plistPath: plistPath,
	}
}

// Safari plist structures
type safariBookmarkPlist struct {
	Children []safariBookmarkNode `plist:"Children"`
}

type safariBookmarkNode struct {
	WebBookmarkType        string                 `plist:"WebBookmarkType"`
	Title                  string                 `plist:"Title"`
	URLString              string                 `plist:"URLString"`
	Children               []safariBookmarkNode   `plist:"Children"`
	ReadingListNonSync     map[string]interface{} `plist:"ReadingListNonSync"`
	ReadingList            map[string]interface{} `plist:"ReadingList"`
	URIDictionary          map[string]interface{} `plist:"URIDictionary"`
	WebBookmarkUUID        string                 `plist:"WebBookmarkUUID"`
	WebBookmarkFileVersion int                    `plist:"WebBookmarkFileVersion"`
}

// GetBookmarks retrieves all bookmarks from Safari
func (h *SafariBookmarkHandler) GetBookmarks(startTime, endTime time.Time) ([]models.BookmarkEntry, error) {
	h.startTime = startTime
	h.endTime = endTime

	data, err := os.ReadFile(h.plistPath)
	if err != nil {
		return nil, err
	}

	var bookmarkPlist safariBookmarkPlist
	if _, err := plist.Unmarshal(data, &bookmarkPlist); err != nil {
		return nil, err
	}

	var bookmarks []models.BookmarkEntry

	// Extract from all root nodes
	for _, node := range bookmarkPlist.Children {
		bookmarks = append(bookmarks, h.extractFromNode(node, "")...)
	}

	return bookmarks, nil
}

// extractFromNode recursively extracts bookmarks from a Safari node
func (h *SafariBookmarkHandler) extractFromNode(node safariBookmarkNode, folderPath string) []models.BookmarkEntry {
	var bookmarks []models.BookmarkEntry

	switch node.WebBookmarkType {
	case "WebBookmarkTypeLeaf":
		// This is a bookmark
		if node.URLString != "" {
			// Try to get date from URIDictionary
			var dateAdded time.Time
			if node.URIDictionary != nil {
				if title, ok := node.URIDictionary["title"].(string); ok && title != "" && node.Title == "" {
					node.Title = title
				}
			}

			// Filter by date if time range is specified and date is available
			if !dateAdded.IsZero() {
				if !h.startTime.IsZero() && dateAdded.Before(h.startTime) {
					return bookmarks
				}
				if !h.endTime.IsZero() && dateAdded.After(h.endTime) {
					return bookmarks
				}
			}

			bookmarks = append(bookmarks, models.BookmarkEntry{
				DateAdded: dateAdded,
				URL:       node.URLString,
				Title:     node.Title,
				Folder:    folderPath,
				Domain:    ExtractDomain(node.URLString),
				Browser:   "safari",
			})
		}

	case "WebBookmarkTypeList":
		// This is a folder
		newFolderPath := folderPath
		if node.Title != "" {
			if folderPath != "" {
				newFolderPath = folderPath + "/" + node.Title
			} else {
				newFolderPath = node.Title
			}
		}

		for _, child := range node.Children {
			bookmarks = append(bookmarks, h.extractFromNode(child, newFolderPath)...)
		}

	case "WebBookmarkTypeProxy":
		// Reading list or other special items - extract if they have URLs
		if node.URLString != "" {
			bookmarks = append(bookmarks, models.BookmarkEntry{
				URL:     node.URLString,
				Title:   node.Title,
				Folder:  folderPath + "/Reading List",
				Domain:  ExtractDomain(node.URLString),
				Browser: "safari",
			})
		}

		// Also check children
		for _, child := range node.Children {
			bookmarks = append(bookmarks, h.extractFromNode(child, folderPath)...)
		}
	}

	return bookmarks
}
