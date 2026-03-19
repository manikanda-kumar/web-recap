package models

import (
	"encoding/json"
	"testing"
	"time"
)

func TestBookmarkEntryMarshalJSON_OmitsZeroTimestamps(t *testing.T) {
	entry := BookmarkEntry{
		URL:     "https://example.com",
		Title:   "Example",
		Domain:  "example.com",
		Browser: "safari",
	}

	data, err := json.Marshal(entry)
	if err != nil {
		t.Fatalf("marshal bookmark entry: %v", err)
	}

	var payload map[string]any
	if err := json.Unmarshal(data, &payload); err != nil {
		t.Fatalf("unmarshal bookmark payload: %v", err)
	}

	if _, ok := payload["date_added"]; ok {
		t.Fatalf("expected date_added to be omitted when zero")
	}
	if _, ok := payload["date_modified"]; ok {
		t.Fatalf("expected date_modified to be omitted when zero")
	}
}

func TestBookmarkEntryMarshalJSON_IncludesSetTimestamps(t *testing.T) {
	dateAdded := time.Date(2026, 1, 10, 9, 30, 0, 0, time.UTC)
	dateModified := dateAdded.Add(2 * time.Hour)

	entry := BookmarkEntry{
		DateAdded:    dateAdded,
		DateModified: dateModified,
		URL:          "https://example.com",
		Title:        "Example",
		Domain:       "example.com",
		Browser:      "chrome",
	}

	data, err := json.Marshal(entry)
	if err != nil {
		t.Fatalf("marshal bookmark entry: %v", err)
	}

	var payload map[string]any
	if err := json.Unmarshal(data, &payload); err != nil {
		t.Fatalf("unmarshal bookmark payload: %v", err)
	}

	if _, ok := payload["date_added"]; !ok {
		t.Fatalf("expected date_added to be included when set")
	}
	if _, ok := payload["date_modified"]; !ok {
		t.Fatalf("expected date_modified to be included when set")
	}
}
