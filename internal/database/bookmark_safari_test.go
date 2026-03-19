package database

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestSafariBookmarkHandlerRejectsDateFiltering(t *testing.T) {
	h := NewSafariBookmarkHandler("/does/not/matter.plist")

	_, err := h.GetBookmarks(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC), time.Time{})
	if err == nil {
		t.Fatalf("expected date filtering to be rejected for Safari bookmarks")
	}
}

func TestSafariBookmarkHandlerExtractsBookmarks(t *testing.T) {
	tempDir := t.TempDir()
	plistPath := filepath.Join(tempDir, "Bookmarks.plist")

	plist := `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
  <key>Children</key>
  <array>
    <dict>
      <key>WebBookmarkType</key><string>WebBookmarkTypeList</string>
      <key>Title</key><string>Favorites</string>
      <key>Children</key>
      <array>
        <dict>
          <key>WebBookmarkType</key><string>WebBookmarkTypeLeaf</string>
          <key>URLString</key><string>https://example.com/article</string>
          <key>URIDictionary</key>
          <dict>
            <key>title</key><string>Example Article</string>
          </dict>
        </dict>
      </array>
    </dict>
    <dict>
      <key>WebBookmarkType</key><string>WebBookmarkTypeProxy</string>
      <key>Title</key><string>Read Later</string>
      <key>URLString</key><string>https://news.ycombinator.com/</string>
    </dict>
  </array>
</dict>
</plist>`

	if err := os.WriteFile(plistPath, []byte(plist), 0o600); err != nil {
		t.Fatalf("write plist: %v", err)
	}

	h := NewSafariBookmarkHandler(plistPath)
	entries, err := h.GetBookmarks(time.Time{}, time.Time{})
	if err != nil {
		t.Fatalf("GetBookmarks() error = %v", err)
	}

	if len(entries) != 2 {
		t.Fatalf("expected 2 bookmarks, got %d", len(entries))
	}

	if entries[0].Title != "Example Article" {
		t.Fatalf("expected title from URIDictionary, got %q", entries[0].Title)
	}
	if entries[0].Folder != "Favorites" {
		t.Fatalf("expected Favorites folder, got %q", entries[0].Folder)
	}
	if entries[0].Browser != "safari" {
		t.Fatalf("expected safari browser, got %q", entries[0].Browser)
	}
	if !entries[0].DateAdded.IsZero() {
		t.Fatalf("expected Safari bookmark date_added to be zero/unset")
	}

	if entries[1].Folder != "/Reading List" {
		t.Fatalf("expected proxy bookmark to go to Reading List, got %q", entries[1].Folder)
	}
}
