package youtube

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
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
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil && filepath.Dir(path) != "." {
		return fmt.Errorf("create watch later dir: %w", err)
	}
	return os.WriteFile(path, b, 0o644)
}

// LoadTakeoutCSV reads a Google Takeout "Watch later videos.csv" file and
// returns items in the same format as the JSON report.
func LoadTakeoutCSV(path string) (*models.YouTubeWatchLaterReport, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	reader := csv.NewReader(f)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("parse takeout csv: %w", err)
	}

	var items []models.YouTubePlaylistItem
	for i, row := range records {
		if i == 0 {
			continue // skip header
		}
		if len(row) < 2 {
			continue
		}

		videoID := strings.TrimSpace(row[0])
		if videoID == "" {
			continue
		}

		addedAt, err := time.Parse(time.RFC3339, strings.TrimSpace(row[1]))
		if err != nil {
			addedAt = time.Time{}
		}

		items = append(items, models.YouTubePlaylistItem{
			VideoID: videoID,
			URL:     "https://www.youtube.com/watch?v=" + videoID,
			AddedAt: addedAt.UTC(),
		})
	}

	sort.Slice(items, func(i, j int) bool { return items[i].AddedAt.Before(items[j].AddedAt) })

	return &models.YouTubeWatchLaterReport{
		FetchedAt:   time.Now().UTC(),
		PlaylistID:  "WL",
		TotalItems:  len(items),
		DeltaAdded:  len(items),
		Items:       items,
		Source:      "youtube",
		Description: "YouTube Watch Later (Google Takeout import)",
	}, nil
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
