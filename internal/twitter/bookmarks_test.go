package twitter

import (
	"testing"
	"time"
)

func TestExtractTweetMapsFromMCPResult(t *testing.T) {
	result := map[string]interface{}{
		"content": []interface{}{
			map[string]interface{}{
				"type": "text",
				"text": `{"data":[{"id":"123","text":"hello https://t.co/abc","created_at":"2025-02-01T10:00:00Z","author":{"name":"Jane","username":"jane"}}]}`,
			},
		},
	}

	tweets := extractTweetMapsFromMCPResult(result)
	if len(tweets) != 1 {
		t.Fatalf("expected 1 tweet map, got %d", len(tweets))
	}

	gotID := extractString(tweets[0], "id")
	if gotID != "123" {
		t.Fatalf("expected id 123, got %q", gotID)
	}
}

func TestMapRawTweetToBookmark(t *testing.T) {
	raw := map[string]interface{}{
		"id":         "999",
		"text":       "check this https://t.co/short",
		"created_at": "2025-02-01T10:00:00Z",
		"author": map[string]interface{}{
			"name":     "Alice",
			"username": "@alice",
		},
	}

	item := mapRawTweetToBookmark(raw)
	if item.TweetID != "999" {
		t.Fatalf("expected tweet id 999, got %q", item.TweetID)
	}
	if item.AuthorHandle != "alice" {
		t.Fatalf("expected author handle alice, got %q", item.AuthorHandle)
	}
	if item.URL != "https://x.com/alice/status/999" {
		t.Fatalf("unexpected url %q", item.URL)
	}

	expectedTime := time.Date(2025, 2, 1, 10, 0, 0, 0, time.UTC)
	if !item.CreatedAt.Equal(expectedTime) {
		t.Fatalf("unexpected created_at %v", item.CreatedAt)
	}
}

func TestParseMCPResponseBodySSE(t *testing.T) {
	body := []byte("event: message\ndata: {\"jsonrpc\":\"2.0\",\"id\":1,\"result\":{\"ok\":true}}\n\n")
	var resp mcpToolCallResponse

	if err := parseMCPResponseBody(body, &resp); err != nil {
		t.Fatalf("expected SSE response to parse, got error: %v", err)
	}

	if resp.Result["ok"] != true {
		t.Fatalf("expected result.ok true, got %#v", resp.Result["ok"])
	}
}
