package models

import "time"

// ReadingListEntry represents a single saved article from any platform
type ReadingListEntry struct {
	SavedAt     time.Time `json:"saved_at"`
	URL         string    `json:"url"`
	Title       string    `json:"title"`
	Author      string    `json:"author,omitempty"`
	Publication string    `json:"publication,omitempty"`
	Excerpt     string    `json:"excerpt,omitempty"`
	Domain      string    `json:"domain"`
	Platform    string    `json:"platform"` // "medium", "substack", "readwise", "raindrop", etc.
	ReadStatus  string    `json:"read_status,omitempty"` // "read", "unread", "archived"
}

// ReadingListReport represents a collection of reading list entries
type ReadingListReport struct {
	Platform     string              `json:"platform"`
	StartDate    *time.Time          `json:"start_date,omitempty"`
	EndDate      *time.Time          `json:"end_date,omitempty"`
	Timezone     string              `json:"timezone,omitempty"`
	TotalEntries int                 `json:"total_entries"`
	Entries      []ReadingListEntry  `json:"entries"`
}
