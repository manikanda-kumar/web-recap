package readinglist

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/rzolkos/web-recap/internal/database"
	"github.com/rzolkos/web-recap/internal/models"
)

// SubstackWebStrategy uses authenticated HTTP requests to fetch saved posts
type SubstackWebStrategy struct {
	config *Config
}

// NewSubstackWebStrategy creates a new Substack web API strategy
func NewSubstackWebStrategy(config *Config) *SubstackWebStrategy {
	return &SubstackWebStrategy{config: config}
}

// Name returns the strategy name
func (s *SubstackWebStrategy) Name() string {
	return "web"
}

// IsAvailable checks if we have authentication credentials
func (s *SubstackWebStrategy) IsAvailable() bool {
	return s.config.Cookie != "" || s.config.SessionToken != ""
}

// SubstackAPIResponse represents the JSON response from Substack API
type SubstackAPIResponse struct {
	Posts []struct {
		ID          int64     `json:"id"`
		Title       string    `json:"title"`
		URL         string    `json:"canonical_url"`
		Subtitle    string    `json:"subtitle"`
		Author      string    `json:"author_name"`
		Publication string    `json:"publication_name"`
		SavedAt     time.Time `json:"saved_at"`
	} `json:"posts"`
}

// Fetch makes authenticated API requests to fetch Substack saved posts
func (s *SubstackWebStrategy) Fetch(config *Config, startDate, endDate time.Time) ([]models.ReadingListEntry, error) {
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	// Substack API endpoint for saved posts
	// Note: This endpoint may need adjustment based on actual Substack API structure
	url := "https://substack.com/api/v1/saved-posts"

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add authentication
	if config.Cookie != "" {
		req.Header.Set("Cookie", config.Cookie)
	}
	if config.SessionToken != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", config.SessionToken))
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch Substack saves: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Substack API returned status %d: %s", resp.StatusCode, string(body))
	}

	var apiResp SubstackAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	var entries []models.ReadingListEntry

	for _, post := range apiResp.Posts {
		entry := models.ReadingListEntry{
			Platform:    "substack",
			Title:       post.Title,
			URL:         post.URL,
			Author:      post.Author,
			Publication: post.Publication,
			Excerpt:     post.Subtitle,
			SavedAt:     post.SavedAt,
			Domain:      database.ExtractDomain(post.URL),
		}

		entries = append(entries, entry)
	}

	// Filter by date range
	if !startDate.IsZero() || !endDate.IsZero() {
		entries = filterByDateRange(entries, startDate, endDate)
	}

	return entries, nil
}
