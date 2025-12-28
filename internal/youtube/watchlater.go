package youtube

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/rzolkos/web-recap/internal/models"
	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
)

// FetchWatchLaterItems fetches Watch Later playlist items from YouTube.
//
// If since is non-zero, it returns only items where AddedAt > since.
func FetchWatchLaterItems(ctx context.Context, httpClientOption option.ClientOption, since time.Time) (playlistID string, items []models.YouTubePlaylistItem, err error) {
	return FetchWatchLaterItemsWithOptions(ctx, httpClientOption, "WL", "", false, since)
}

// FetchWatchLaterItemsWithOptions fetches Watch Later playlist items from YouTube.
//
// If playlistID is non-empty, it is used directly (e.g. "WL" for Watch Later).
// channelID can be used to force a specific channel when calling channels.list.
// If debug is true, discovered channels are printed to stderr.
func FetchWatchLaterItemsWithOptions(ctx context.Context, httpClientOption option.ClientOption, playlistID string, channelID string, debug bool, since time.Time) (playlistIDOut string, items []models.YouTubePlaylistItem, err error) {
	svc, err := youtube.NewService(ctx, httpClientOption)
	if err != nil {
		return "", nil, fmt.Errorf("create youtube service: %w", err)
	}

	watchLaterID := strings.TrimSpace(playlistID)
	if watchLaterID == "" {
		watchLaterID, err = findWatchLaterPlaylistID(ctx, svc)
		if err != nil {
			watchLaterID, err = findWatchLaterByChannel(ctx, svc, channelID, debug)
			if err != nil {
				return "", nil, err
			}
		}
	}

	var out []models.YouTubePlaylistItem
	pageToken := ""

	for {
		call := svc.PlaylistItems.List([]string{"snippet", "contentDetails", "status"}).PlaylistId(watchLaterID).MaxResults(50)
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}

		resp, err := call.Context(ctx).Do()
		if err != nil {
			return "", nil, fmt.Errorf("list playlist items: %w", err)
		}

		for _, it := range resp.Items {
			if it == nil || it.Snippet == nil || it.ContentDetails == nil {
				continue
			}

			if debug {
				privacy := ""
				if it.Status != nil {
					privacy = it.Status.PrivacyStatus
				}
				fmt.Fprintf(os.Stderr, "item: playlistItemId=%s videoId=%s title=%q publishedAt=%s privacy=%q\n", it.Id, it.ContentDetails.VideoId, it.Snippet.Title, it.Snippet.PublishedAt, privacy)
			}

			addedAt, err := time.Parse(time.RFC3339, it.Snippet.PublishedAt)
			if err != nil {
				continue
			}
			if !since.IsZero() && !addedAt.After(since) {
				continue
			}

			videoID := strings.TrimSpace(it.ContentDetails.VideoId)
			if videoID == "" {
				continue
			}

			out = append(out, models.YouTubePlaylistItem{
				PlaylistItemID: it.Id,
				VideoID:        videoID,
				URL:            "https://www.youtube.com/watch?v=" + url.QueryEscape(videoID),
				Title:          strings.TrimSpace(it.Snippet.Title),
				ChannelTitle:   strings.TrimSpace(it.Snippet.VideoOwnerChannelTitle),
				AddedAt:        addedAt.UTC(),
			})
		}

		pageToken = resp.NextPageToken
		if pageToken == "" {
			break
		}
	}

	sort.Slice(out, func(i, j int) bool { return out[i].AddedAt.Before(out[j].AddedAt) })
	return watchLaterID, out, nil
}

func findWatchLaterPlaylistID(ctx context.Context, svc *youtube.Service) (string, error) {
	pageToken := ""
	for {
		call := svc.Playlists.List([]string{"snippet"}).Mine(true).MaxResults(50)
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Context(ctx).Do()
		if err != nil {
			return "", fmt.Errorf("list playlists: %w", err)
		}

		for _, pl := range resp.Items {
			if pl == nil || pl.Snippet == nil {
				continue
			}
			// Localized. In practice it is "Watch later" in English, but we also
			// accept variants to reduce friction.
			title := strings.ToLower(strings.TrimSpace(pl.Snippet.Title))
			if strings.Contains(title, "watch") && strings.Contains(title, "later") {
				return pl.Id, nil
			}
		}

		pageToken = resp.NextPageToken
		if pageToken == "" {
			break
		}
	}
	return "", fmt.Errorf("watch later playlist not found via playlists.list")
}

func findWatchLaterByChannel(ctx context.Context, svc *youtube.Service, channelID string, debug bool) (string, error) {
	// Fallback: some accounts/locales do not surface Watch Later via playlists.list(mine=true).
	// The Channels endpoint includes a dedicated watchLater playlist ID.
	call := svc.Channels.List([]string{"snippet", "contentDetails"}).Mine(true).MaxResults(50)
	resp, err := call.Context(ctx).Do()
	if err != nil {
		return "", fmt.Errorf("get channel content details: %w", err)
	}
	if len(resp.Items) == 0 {
		return "", fmt.Errorf("no channels returned for mine=true")
	}

	if debug {
		for _, ch := range resp.Items {
			if ch == nil {
				continue
			}
			title := ""
			if ch.Snippet != nil {
				title = ch.Snippet.Title
			}
			watchLater := ""
			if ch.ContentDetails != nil && ch.ContentDetails.RelatedPlaylists != nil {
				watchLater = ch.ContentDetails.RelatedPlaylists.WatchLater
			}
			fmt.Fprintf(os.Stderr, "channel: id=%s title=%q watchLater=%q\n", ch.Id, title, watchLater)
		}
	}

	pick := func(ch *youtube.Channel) (string, bool) {
		if ch == nil || ch.ContentDetails == nil || ch.ContentDetails.RelatedPlaylists == nil {
			return "", false
		}
		id := strings.TrimSpace(ch.ContentDetails.RelatedPlaylists.WatchLater)
		if id == "" {
			return "", false
		}
		return id, true
	}

	if strings.TrimSpace(channelID) != "" {
		for _, ch := range resp.Items {
			if ch != nil && ch.Id == channelID {
				if wl, ok := pick(ch); ok {
					return wl, nil
				}
				return "", fmt.Errorf("specified channel-id has no watchLater playlist id")
			}
		}
		return "", fmt.Errorf("specified channel-id not found in channels.list(mine=true)")
	}

	for _, ch := range resp.Items {
		if wl, ok := pick(ch); ok {
			return wl, nil
		}
	}

	return "", fmt.Errorf("watch later playlist id missing from all returned channels")
}
