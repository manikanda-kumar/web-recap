package models

import "time"

// YouTubePlaylistItem represents a single item from a YouTube playlist.
// For Watch Later, AddedAt corresponds to playlistItems.snippet.publishedAt.
//
// VideoID is the canonical stable identifier used for delta/reconciliation.
// URL is derived as https://www.youtube.com/watch?v=<videoId>.
//
// Note: PlaylistItemID is included to support future reconcile/deletion detection.
// It is not required for basic delta sync.
type YouTubePlaylistItem struct {
	PlaylistItemID string    `json:"playlist_item_id"`
	VideoID        string    `json:"video_id"`
	URL            string    `json:"url"`
	Title          string    `json:"title,omitempty"`
	ChannelTitle   string    `json:"channel_title,omitempty"`
	AddedAt        time.Time `json:"added_at"`
}

// YouTubeWatchLaterReport represents the Watch Later playlist snapshot.
type YouTubeWatchLaterReport struct {
	FetchedAt   time.Time             `json:"fetched_at"`
	PlaylistID  string                `json:"playlist_id"`
	TotalItems  int                   `json:"total_items"`
	DeltaAdded  int                   `json:"delta_added"`
	Items       []YouTubePlaylistItem `json:"items"`
	Source      string                `json:"source"` // "youtube"
	Description string                `json:"description,omitempty"`
}
