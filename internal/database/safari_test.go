package database

import (
	"database/sql"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	_ "modernc.org/sqlite"
)

func TestSafariHandlerGetHistoryReadsVisits(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("Safari history is only supported on macOS")
	}

	dbPath := createSafariHistoryDB(t)
	h := NewSafariHandler(dbPath)

	entries, err := h.GetHistory(time.Time{}, time.Time{})
	if err != nil {
		t.Fatalf("GetHistory() error = %v", err)
	}

	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}

	if entries[0].URL != "https://example.com/newer" {
		t.Fatalf("expected newest visit first, got %q", entries[0].URL)
	}
	if entries[0].Title != "https://example.com/newer" {
		t.Fatalf("expected URL fallback title, got %q", entries[0].Title)
	}
	if entries[0].Browser != "safari" {
		t.Fatalf("expected safari browser, got %q", entries[0].Browser)
	}
	if entries[0].Domain != "example.com" {
		t.Fatalf("expected example.com domain, got %q", entries[0].Domain)
	}

	wantNewest := time.Date(2026, 1, 15, 12, 0, 0, 0, time.UTC)
	if !entries[0].Timestamp.Equal(wantNewest) {
		t.Fatalf("expected newest timestamp %s, got %s", wantNewest, entries[0].Timestamp)
	}
}

func TestSafariHandlerGetHistoryFiltersByDateRange(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("Safari history is only supported on macOS")
	}

	dbPath := createSafariHistoryDB(t)
	h := NewSafariHandler(dbPath)

	start := time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)

	entries, err := h.GetHistory(start, end)
	if err != nil {
		t.Fatalf("GetHistory() error = %v", err)
	}

	if len(entries) != 1 {
		t.Fatalf("expected 1 filtered entry, got %d", len(entries))
	}
	if entries[0].URL != "https://example.com/newer" {
		t.Fatalf("expected filtered result to be newer entry, got %q", entries[0].URL)
	}
}

func createSafariHistoryDB(t *testing.T) string {
	t.Helper()

	dbPath := filepath.Join(t.TempDir(), "History.db")
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("open sqlite db: %v", err)
	}
	defer db.Close()

	stmts := []string{
		`CREATE TABLE history_items (id INTEGER PRIMARY KEY, url TEXT NOT NULL, visit_count INTEGER NOT NULL);`,
		`CREATE TABLE history_visits (id INTEGER PRIMARY KEY, history_item INTEGER NOT NULL, visit_time INTEGER NOT NULL, title TEXT, FOREIGN KEY(history_item) REFERENCES history_items(id));`,
		`INSERT INTO history_items (id, url, visit_count) VALUES (1, 'https://example.com/older', 3);`,
		`INSERT INTO history_items (id, url, visit_count) VALUES (2, 'https://example.com/newer', 7);`,
		`INSERT INTO history_visits (id, history_item, visit_time, title) VALUES (1, 1, 789004800, 'Older Title');`,
		`INSERT INTO history_visits (id, history_item, visit_time, title) VALUES (2, 2, 790171200, NULL);`,
	}

	for _, stmt := range stmts {
		if _, err := db.Exec(stmt); err != nil {
			t.Fatalf("exec %q: %v", stmt, err)
		}
	}

	return dbPath
}
