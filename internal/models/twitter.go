package models

import "time"

// TwitterBookmark represents a single saved tweet/bookmark from X/Twitter.
type TwitterBookmark struct {
	TweetID      string            `json:"tweet_id"`
	URL          string            `json:"url"`
	Text         string            `json:"text,omitempty"`
	AuthorName   string            `json:"author_name,omitempty"`
	AuthorHandle string            `json:"author_handle,omitempty"`
	CreatedAt    time.Time         `json:"created_at,omitempty"`
	SavedAt      time.Time         `json:"saved_at,omitempty"`
	ExpandedURLs map[string]string `json:"expanded_urls,omitempty"` // t.co -> real URL
}

// TwitterBookmarksReport represents the Twitter bookmarks snapshot.
type TwitterBookmarksReport struct {
	FetchedAt   time.Time         `json:"fetched_at"`
	TotalItems  int               `json:"total_items"`
	DeltaAdded  int               `json:"delta_added"`
	Items       []TwitterBookmark `json:"items"`
	Source      string            `json:"source"` // "twitter"
	Description string            `json:"description,omitempty"`
}
