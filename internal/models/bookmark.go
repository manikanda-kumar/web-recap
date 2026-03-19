package models

import (
	"encoding/json"
	"time"
)

// BookmarkEntry represents a single browser bookmark entry
type BookmarkEntry struct {
	DateAdded    time.Time `json:"date_added"`
	DateModified time.Time `json:"date_modified,omitempty"`
	URL          string    `json:"url"`
	Title        string    `json:"title"`
	Folder       string    `json:"folder,omitempty"`
	Domain       string    `json:"domain"`
	Browser      string    `json:"browser"`
	Tags         []string  `json:"tags,omitempty"`
}

// MarshalJSON ensures unset bookmark timestamps are omitted from JSON output.
func (b BookmarkEntry) MarshalJSON() ([]byte, error) {
	type bookmarkEntryJSON struct {
		DateAdded    *time.Time `json:"date_added,omitempty"`
		DateModified *time.Time `json:"date_modified,omitempty"`
		URL          string     `json:"url"`
		Title        string     `json:"title"`
		Folder       string     `json:"folder,omitempty"`
		Domain       string     `json:"domain"`
		Browser      string     `json:"browser"`
		Tags         []string   `json:"tags,omitempty"`
	}

	var dateAdded *time.Time
	if !b.DateAdded.IsZero() {
		dateAdded = &b.DateAdded
	}

	var dateModified *time.Time
	if !b.DateModified.IsZero() {
		dateModified = &b.DateModified
	}

	return json.Marshal(bookmarkEntryJSON{
		DateAdded:    dateAdded,
		DateModified: dateModified,
		URL:          b.URL,
		Title:        b.Title,
		Folder:       b.Folder,
		Domain:       b.Domain,
		Browser:      b.Browser,
		Tags:         b.Tags,
	})
}

// BookmarkReport represents a collection of bookmark entries
type BookmarkReport struct {
	Browser      string          `json:"browser"`
	StartDate    *time.Time      `json:"start_date,omitempty"`
	EndDate      *time.Time      `json:"end_date,omitempty"`
	Timezone     string          `json:"timezone,omitempty"`
	TotalEntries int             `json:"total_entries"`
	Entries      []BookmarkEntry `json:"entries"`
}

// BookmarkFolder represents a folder/directory structure in bookmarks
type BookmarkFolder struct {
	Name     string         `json:"name"`
	Children []BookmarkNode `json:"children,omitempty"`
}

// BookmarkNode can be either a bookmark or a folder
type BookmarkNode struct {
	Type     string          `json:"type"` // "url" or "folder"
	Bookmark *BookmarkEntry  `json:"bookmark,omitempty"`
	Folder   *BookmarkFolder `json:"folder,omitempty"`
}
