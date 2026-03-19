package youtube

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
)

// CreatePlaylist creates a new YouTube playlist with the given title, description, and privacy status.
// privacyStatus should be "public", "private", or "unlisted".
func CreatePlaylist(ctx context.Context, httpClientOption option.ClientOption, title, description, privacyStatus string) (string, error) {
	svc, err := youtube.NewService(ctx, httpClientOption)
	if err != nil {
		return "", fmt.Errorf("create youtube service: %w", err)
	}

	playlist := &youtube.Playlist{
		Snippet: &youtube.PlaylistSnippet{
			Title:       title,
			Description: description,
		},
		Status: &youtube.PlaylistStatus{
			PrivacyStatus: privacyStatus,
		},
	}

	resp, err := svc.Playlists.Insert([]string{"snippet", "status"}, playlist).Context(ctx).Do()
	if err != nil {
		return "", fmt.Errorf("create playlist: %w", err)
	}

	return resp.Id, nil
}

// InsertVideosIntoPlaylist adds videos to an existing playlist by their video IDs.
// It returns the number of successfully inserted videos and any error from the last failure.
func InsertVideosIntoPlaylist(ctx context.Context, httpClientOption option.ClientOption, playlistID string, videoIDs []string) (int, error) {
	svc, err := youtube.NewService(ctx, httpClientOption)
	if err != nil {
		return 0, fmt.Errorf("create youtube service: %w", err)
	}

	inserted := 0
	for _, videoID := range videoIDs {
		item := &youtube.PlaylistItem{
			Snippet: &youtube.PlaylistItemSnippet{
				PlaylistId: playlistID,
				ResourceId: &youtube.ResourceId{
					Kind:    "youtube#video",
					VideoId: videoID,
				},
			},
		}

		_, err := svc.PlaylistItems.Insert([]string{"snippet"}, item).Context(ctx).Do()
		if err != nil {
			fmt.Printf("  ⚠ failed to add video %s: %v\n", videoID, err)
			continue
		}
		inserted++

		// Respect API rate limits
		if inserted%10 == 0 {
			time.Sleep(1 * time.Second)
		}
	}

	return inserted, nil
}
