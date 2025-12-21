package models

import "time"

// BookmarkEntry represents a single browser bookmark entry
type BookmarkEntry struct {
	DateAdded   time.Time `json:"date_added"`
	DateModified time.Time `json:"date_modified,omitempty"`
	URL         string    `json:"url"`
	Title       string    `json:"title"`
	Folder      string    `json:"folder,omitempty"`
	Domain      string    `json:"domain"`
	Browser     string    `json:"browser"`
	Tags        []string  `json:"tags,omitempty"`
}

// BookmarkReport represents a collection of bookmark entries
type BookmarkReport struct {
	Browser      string           `json:"browser"`
	TotalEntries int              `json:"total_entries"`
	Entries      []BookmarkEntry  `json:"entries"`
}

// BookmarkFolder represents a folder/directory structure in bookmarks
type BookmarkFolder struct {
	Name     string            `json:"name"`
	Children []BookmarkNode    `json:"children,omitempty"`
}

// BookmarkNode can be either a bookmark or a folder
type BookmarkNode struct {
	Type     string          `json:"type"` // "url" or "folder"
	Bookmark *BookmarkEntry  `json:"bookmark,omitempty"`
	Folder   *BookmarkFolder `json:"folder,omitempty"`
}
