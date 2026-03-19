package twitter

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/rzolkos/web-recap/internal/models"
	"github.com/rzolkos/web-recap/internal/urlutil"
)

type FetchProvider string

const (
	ProviderAuto     FetchProvider = "auto"
	ProviderBird     FetchProvider = "bird"
	ProviderComposio FetchProvider = "composio"
)

const defaultComposioTool = "TWITTER_BOOKMARKS_BY_USER"

// ComposioConfig contains the minimum configuration required to execute
// the Twitter bookmarks tool via Composio's MCP endpoint.
type ComposioConfig struct {
	APIKey string
	MCPURL string
	UserID string
	Tool   string
}

func (c ComposioConfig) toolName() string {
	if c.Tool != "" {
		return c.Tool
	}
	return defaultComposioTool
}

func (c ComposioConfig) IsConfigured() bool {
	return c.APIKey != "" && c.MCPURL != "" && c.UserID != ""
}

// birdTweet represents the raw JSON structure returned by bird --json.
type birdTweet struct {
	ID        string `json:"id"`
	Text      string `json:"text"`
	CreatedAt string `json:"createdAt"`
	Author    struct {
		Name     string `json:"name"`
		Username string `json:"username"`
	} `json:"author"`
	URL string `json:"url"`
}

// FetchBookmarks fetches Twitter bookmarks using a provider strategy.
// If since is non-zero, it returns only items where SavedAt > since.
func FetchBookmarks(since time.Time, provider FetchProvider, authToken, ct0 string, composioConfig ComposioConfig) ([]models.TwitterBookmark, error) {
	switch provider {
	case ProviderBird:
		return fetchBookmarksWithBird(since, authToken, ct0)
	case ProviderComposio:
		if !composioConfig.IsConfigured() {
			return nil, fmt.Errorf("composio provider requires COMPOSIO_API_KEY, COMPOSIO_MCP_URL, and COMPOSIO_USER_ID")
		}
		return fetchBookmarksWithComposio(since, composioConfig)
	case ProviderAuto:
		if composioConfig.IsConfigured() {
			items, err := fetchBookmarksWithComposio(since, composioConfig)
			if err == nil {
				return items, nil
			}
			// Fall back to bird when composio fails in auto mode.
			birdItems, birdErr := fetchBookmarksWithBird(since, authToken, ct0)
			if birdErr != nil {
				return nil, fmt.Errorf("composio fetch failed: %v; bird fallback failed: %v", err, birdErr)
			}
			return birdItems, nil
		}
		return fetchBookmarksWithBird(since, authToken, ct0)
	default:
		return nil, fmt.Errorf("unsupported twitter provider %q (use auto, composio, or bird)", provider)
	}
}

// fetchBookmarksWithBird fetches Twitter bookmarks using the bird CLI.
// authToken and ct0 are optional; if provided, they're passed to bird directly.
func fetchBookmarksWithBird(since time.Time, authToken, ct0 string) ([]models.TwitterBookmark, error) {
	// Check if bird is available
	_, err := exec.LookPath("bird")
	if err != nil {
		return nil, fmt.Errorf("bird CLI not found in PATH. Install it from https://github.com/steipete/bird")
	}

	// Build command args
	args := []string{"bookmarks", "--json", "--cookie-source", "chrome", "--timeout", "30000"}

	cmd := exec.Command("bird", args...)
	cmd.Stderr = os.Stderr

	// Pass credentials via environment variables to avoid exposure in process list
	if authToken != "" || ct0 != "" {
		cmd.Env = os.Environ()
		if authToken != "" {
			cmd.Env = append(cmd.Env, "AUTH_TOKEN="+authToken)
		}
		if ct0 != "" {
			cmd.Env = append(cmd.Env, "CT0="+ct0)
		}
	}

	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return nil, fmt.Errorf("bird command failed: %s", string(exitErr.Stderr))
		}
		return nil, fmt.Errorf("bird command failed: %w", err)
	}

	var tweets []birdTweet
	if err := json.Unmarshal(output, &tweets); err != nil {
		return nil, fmt.Errorf("failed to parse bird output: %w", err)
	}

	var items []models.TwitterBookmark
	for _, t := range tweets {
		createdAt, _ := parseTwitterTime(t.CreatedAt)

		// Use createdAt as savedAt approximation (Twitter API doesn't expose bookmark time)
		savedAt := createdAt

		if !since.IsZero() && !savedAt.After(since) {
			continue
		}

		url := t.URL
		if url == "" {
			url = fmt.Sprintf("https://x.com/%s/status/%s", t.Author.Username, t.ID)
		}

		// Expand t.co URLs in tweet text
		expandedURLs := urlutil.ExpandTcoURLsInText(t.Text)

		items = append(items, models.TwitterBookmark{
			TweetID:      t.ID,
			URL:          url,
			Text:         t.Text,
			AuthorName:   t.Author.Name,
			AuthorHandle: t.Author.Username,
			CreatedAt:    createdAt,
			SavedAt:      savedAt,
			ExpandedURLs: expandedURLs,
		})
	}

	// Sort by saved time
	sort.Slice(items, func(i, j int) bool {
		return items[i].SavedAt.Before(items[j].SavedAt)
	})

	return items, nil
}

type mcpToolCallRequest struct {
	JSONRPC string `json:"jsonrpc"`
	ID      int    `json:"id"`
	Method  string `json:"method"`
	Params  struct {
		Name      string                 `json:"name"`
		Arguments map[string]interface{} `json:"arguments"`
	} `json:"params"`
}

type mcpError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type mcpToolCallResponse struct {
	JSONRPC string                 `json:"jsonrpc"`
	ID      interface{}            `json:"id"`
	Result  map[string]interface{} `json:"result"`
	Error   *mcpError              `json:"error"`
}

func fetchBookmarksWithComposio(since time.Time, cfg ComposioConfig) ([]models.TwitterBookmark, error) {
	reqBody := mcpToolCallRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "tools/call",
	}
	reqBody.Params.Name = cfg.toolName()
	reqBody.Params.Arguments = map[string]interface{}{
		"id": cfg.UserID,
	}

	payload, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to encode composio request: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, cfg.MCPURL, bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("failed to create composio request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json, text/event-stream")
	req.Header.Set("x-api-key", cfg.APIKey)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("composio request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read composio response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("composio request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var mcpResp mcpToolCallResponse
	if err := parseMCPResponseBody(body, &mcpResp); err != nil {
		return nil, fmt.Errorf("failed to decode composio response: %w", err)
	}

	if mcpResp.Error != nil {
		return nil, fmt.Errorf("composio tool call error (%d): %s", mcpResp.Error.Code, mcpResp.Error.Message)
	}

	rawTweets := extractTweetMapsFromMCPResult(mcpResp.Result)
	items := make([]models.TwitterBookmark, 0, len(rawTweets))
	for _, t := range rawTweets {
		item := mapRawTweetToBookmark(t)
		if item.TweetID == "" {
			continue
		}
		if !since.IsZero() && !item.SavedAt.IsZero() && !item.SavedAt.After(since) {
			continue
		}
		items = append(items, item)
	}

	sort.Slice(items, func(i, j int) bool {
		return items[i].SavedAt.Before(items[j].SavedAt)
	})

	return dedupeByTweetID(items), nil
}

func parseMCPResponseBody(body []byte, out *mcpToolCallResponse) error {
	trimmed := strings.TrimSpace(string(body))
	if trimmed == "" {
		return fmt.Errorf("empty response body")
	}

	if err := json.Unmarshal([]byte(trimmed), out); err == nil {
		return nil
	}

	// Some MCP servers stream as SSE. Extract JSON payloads from data: lines and
	// keep the last JSON-RPC object.
	var lastJSON string
	for _, line := range strings.Split(trimmed, "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "data:") {
			continue
		}
		candidate := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
		if candidate == "" || candidate == "[DONE]" {
			continue
		}
		var probe map[string]interface{}
		if err := json.Unmarshal([]byte(candidate), &probe); err == nil {
			lastJSON = candidate
		}
	}

	if lastJSON != "" {
		if err := json.Unmarshal([]byte(lastJSON), out); err == nil {
			return nil
		}
	}

	maxLen := 300
	if len(trimmed) > maxLen {
		trimmed = trimmed[:maxLen] + "..."
	}
	return fmt.Errorf("unrecognized response format: %s", trimmed)
}

func extractTweetMapsFromMCPResult(result map[string]interface{}) []map[string]interface{} {
	var out []map[string]interface{}
	seen := make(map[string]bool)

	var walk func(interface{})
	walk = func(value interface{}) {
		switch v := value.(type) {
		case map[string]interface{}:
			if tweetID := extractString(v, "id", "id_str", "tweet_id", "tweetId", "rest_id"); tweetID != "" {
				if text := extractString(v, "text", "full_text"); text != "" || hasTimeFields(v) || hasAuthorFields(v) {
					if !seen[tweetID] {
						seen[tweetID] = true
						out = append(out, v)
					}
				}
			}

			for key, child := range v {
				if key == "text" {
					if s, ok := child.(string); ok {
						trimmed := strings.TrimSpace(s)
						if strings.HasPrefix(trimmed, "{") || strings.HasPrefix(trimmed, "[") {
							var parsed interface{}
							if err := json.Unmarshal([]byte(trimmed), &parsed); err == nil {
								walk(parsed)
							}
						}
					}
				}
				walk(child)
			}
		case []interface{}:
			for _, child := range v {
				walk(child)
			}
		case string:
			trimmed := strings.TrimSpace(v)
			if strings.HasPrefix(trimmed, "{") || strings.HasPrefix(trimmed, "[") {
				var parsed interface{}
				if err := json.Unmarshal([]byte(trimmed), &parsed); err == nil {
					walk(parsed)
				}
			}
		}
	}

	walk(result)
	return out
}

func hasTimeFields(raw map[string]interface{}) bool {
	return extractString(raw, "createdAt", "created_at", "timestamp") != ""
}

func hasAuthorFields(raw map[string]interface{}) bool {
	if _, ok := raw["author"]; ok {
		return true
	}
	if _, ok := raw["user"]; ok {
		return true
	}
	return false
}

func mapRawTweetToBookmark(raw map[string]interface{}) models.TwitterBookmark {
	id := extractString(raw, "id", "id_str", "tweet_id", "tweetId", "rest_id")
	text := extractString(raw, "text", "full_text")
	createdAt := parseFirstTime(raw,
		"createdAt",
		"created_at",
		"timestamp",
	)

	authorName := ""
	authorHandle := ""

	if author, ok := raw["author"].(map[string]interface{}); ok {
		authorName = extractString(author, "name")
		authorHandle = extractString(author, "username", "screen_name", "handle")
	}
	if authorHandle == "" {
		if user, ok := raw["user"].(map[string]interface{}); ok {
			if authorName == "" {
				authorName = extractString(user, "name")
			}
			authorHandle = extractString(user, "username", "screen_name", "handle")
		}
	}
	authorHandle = strings.TrimPrefix(authorHandle, "@")

	url := extractString(raw, "url", "permalink")
	if url == "" && id != "" {
		if authorHandle != "" {
			url = fmt.Sprintf("https://x.com/%s/status/%s", authorHandle, id)
		} else {
			url = fmt.Sprintf("https://x.com/i/web/status/%s", id)
		}
	}

	savedAt := createdAt

	return models.TwitterBookmark{
		TweetID:      id,
		URL:          url,
		Text:         text,
		AuthorName:   authorName,
		AuthorHandle: authorHandle,
		CreatedAt:    createdAt,
		SavedAt:      savedAt,
		ExpandedURLs: urlutil.ExpandTcoURLsInText(text),
	}
}

func extractString(raw map[string]interface{}, keys ...string) string {
	for _, key := range keys {
		v, ok := raw[key]
		if !ok || v == nil {
			continue
		}
		if s, ok := v.(string); ok {
			s = strings.TrimSpace(s)
			if s != "" {
				return s
			}
		}
	}
	return ""
}

func parseFirstTime(raw map[string]interface{}, keys ...string) time.Time {
	for _, key := range keys {
		s := extractString(raw, key)
		if s == "" {
			continue
		}
		if t, err := parseTwitterTime(s); err == nil {
			return t
		}
	}
	return time.Time{}
}

func dedupeByTweetID(items []models.TwitterBookmark) []models.TwitterBookmark {
	seen := make(map[string]bool)
	out := make([]models.TwitterBookmark, 0, len(items))
	for _, item := range items {
		if item.TweetID == "" || seen[item.TweetID] {
			continue
		}
		seen[item.TweetID] = true
		out = append(out, item)
	}
	return out
}

// parseTwitterTime parses Twitter's date format.
func parseTwitterTime(s string) (time.Time, error) {
	// Twitter uses format like "Mon Jan 02 15:04:05 +0000 2006"
	layouts := []string{
		"Mon Jan 02 15:04:05 -0700 2006",
		"Mon Jan 02 15:04:05 +0000 2006",
		time.RFC3339,
		"2006-01-02T15:04:05Z",
		"2006-01-02T15:04:05.000Z",
	}

	for _, layout := range layouts {
		if t, err := time.Parse(layout, s); err == nil {
			return t.UTC(), nil
		}
	}

	return time.Time{}, fmt.Errorf("failed to parse time: %s", s)
}

// LoadBookmarksFile loads a previously saved bookmarks file.
func LoadBookmarksFile(path string) (models.TwitterBookmarksReport, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return models.TwitterBookmarksReport{}, err
	}

	var report models.TwitterBookmarksReport
	if err := json.Unmarshal(data, &report); err != nil {
		return models.TwitterBookmarksReport{}, err
	}

	return report, nil
}

// SaveBookmarksFile saves bookmarks report to a file.
func SaveBookmarksFile(path string, report models.TwitterBookmarksReport) error {
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil && filepath.Dir(path) != "." {
		return fmt.Errorf("create twitter bookmarks dir: %w", err)
	}

	return os.WriteFile(path, data, 0600)
}

// MaxSavedAt returns the latest SavedAt time from items.
func MaxSavedAt(items []models.TwitterBookmark) time.Time {
	var max time.Time
	for _, item := range items {
		if item.SavedAt.After(max) {
			max = item.SavedAt
		}
	}
	return max
}

// MergeByTweetID merges old and new items, deduplicating by TweetID.
func MergeByTweetID(old, new []models.TwitterBookmark) []models.TwitterBookmark {
	seen := make(map[string]bool)
	var merged []models.TwitterBookmark

	for _, item := range old {
		if !seen[item.TweetID] {
			seen[item.TweetID] = true
			merged = append(merged, item)
		}
	}

	for _, item := range new {
		if !seen[item.TweetID] {
			seen[item.TweetID] = true
			merged = append(merged, item)
		}
	}

	sort.Slice(merged, func(i, j int) bool {
		return merged[i].SavedAt.Before(merged[j].SavedAt)
	})

	return merged
}
