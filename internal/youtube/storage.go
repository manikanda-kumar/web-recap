package youtube

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"time"

	"github.com/rzolkos/web-recap/internal/models"
)

func LoadWatchLaterFile(path string) (*models.YouTubeWatchLaterReport, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var report models.YouTubeWatchLaterReport
	if err := json.Unmarshal(b, &report); err != nil {
		return nil, fmt.Errorf("parse watch later file: %w", err)
	}

	return &report, nil
}

func SaveWatchLaterFile(path string, report models.YouTubeWatchLaterReport) error {
	b, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal watch later report: %w", err)
	}
	return os.WriteFile(path, b, 0o644)
}

func MaxAddedAt(items []models.YouTubePlaylistItem) time.Time {
	var max time.Time
	for _, it := range items {
		if it.AddedAt.After(max) {
			max = it.AddedAt
		}
	}
	return max
}

func MergeByVideoID(existing, incoming []models.YouTubePlaylistItem) []models.YouTubePlaylistItem {
	byID := make(map[string]models.YouTubePlaylistItem, len(existing)+len(incoming))
	for _, it := range existing {
		if it.VideoID == "" {
			continue
		}
		byID[it.VideoID] = it
	}
	for _, it := range incoming {
		if it.VideoID == "" {
			continue
		}
		if _, ok := byID[it.VideoID]; ok {
			continue
		}
		byID[it.VideoID] = it
	}

	out := make([]models.YouTubePlaylistItem, 0, len(byID))
	for _, it := range byID {
		out = append(out, it)
	}

	sort.Slice(out, func(i, j int) bool { return out[i].AddedAt.Before(out[j].AddedAt) })
	return out
}
