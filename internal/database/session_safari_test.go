package database

import (
	"database/sql"
	"path/filepath"
	"testing"

	_ "modernc.org/sqlite"
)

// createSafariTabsDB builds a minimal SafariTabs.db fixture mirroring the real
// schema (verified against Safari 624 / macOS 26): `windows` rows reference tab
// group nodes via `windows_tab_groups`; tabs are children of the group node.
func createSafariTabsDB(t *testing.T) string {
	t.Helper()

	dbPath := filepath.Join(t.TempDir(), "SafariTabs.db")
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("open fixture db: %v", err)
	}
	defer db.Close()

	_, err = db.Exec(`
		CREATE TABLE bookmarks (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			parent INTEGER,
			type INTEGER,
			title TEXT,
			url TEXT,
			num_children INTEGER DEFAULT 0,
			hidden INTEGER DEFAULT 0,
			deleted INTEGER DEFAULT 0,
			order_index INTEGER NOT NULL DEFAULT 0,
			subtype INTEGER DEFAULT 0
		);
		CREATE TABLE windows (
			id INTEGER PRIMARY KEY,
			active_tab_group_id INTEGER,
			date_closed REAL
		);
		CREATE TABLE windows_tab_groups (
			id INTEGER PRIMARY KEY,
			active_tab_id INTEGER,
			tab_group_id INTEGER NOT NULL,
			window_id INTEGER NOT NULL
		);
	`)
	if err != nil {
		t.Fatalf("create fixture schema: %v", err)
	}

	bookmarks := []struct {
		id          int
		parent      int
		typ         int
		title       string
		url         string
		numChildren int
		hidden      int
		deleted     int
		orderIndex  int
		subtype     int
	}{
		// Internal root container (hidden=0, must not be treated as open)
		{1, 0, 1, "Root", "", 0, 0, 0, 0, 0},
		// Open tab group of window 1 — note hidden=1, like real Safari data
		{100, 0, 1, "Local", "", 2, 1, 0, 0, 0},
		{101, 100, 0, "Example Article", "https://example.com/article", 0, 0, 0, 0, 0},
		{102, 100, 0, "Go.dev", "https://go.dev/doc/", 0, 0, 0, 1, 0},
		// Open named tab group of window 3 (e.g. second profile window)
		{200, 0, 1, "Research", "", 1, 1, 0, 0, 0},
		{201, 200, 0, "Tab Group Support", "https://support.apple.com/guide/safari/tab-groups", 0, 0, 0, 0, 0},
		// Saved-but-closed group (only referenced by a closed window)
		{300, 0, 1, "Old Group", "", 1, 1, 0, 0, 0},
		{301, 300, 0, "Stale Tab", "https://stale.example.com/", 0, 0, 0, 0, 0},
		// Junk rows inside an open group: placeholder title, empty URL, deleted
		{103, 100, 0, "TopScopedBookmarkList", "", 0, 0, 0, 2, 0},
		{104, 100, 0, "Start Page", "", 0, 0, 0, 3, 0},
		{105, 100, 0, "Deleted Tab", "https://deleted.example.com/", 0, 0, 1, 4, 0},
	}
	for _, r := range bookmarks {
		_, err := db.Exec(
			`INSERT INTO bookmarks (id, parent, type, title, url, num_children, hidden, deleted, order_index, subtype)
			 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			r.id, r.parent, r.typ, r.title, r.url, r.numChildren, r.hidden, r.deleted, r.orderIndex, r.subtype,
		)
		if err != nil {
			t.Fatalf("insert bookmark %d: %v", r.id, err)
		}
	}

	// Windows: 1 = open (active tab 102), 2 = closed, 3 = open with named group
	windows := []struct {
		id          int
		groupID     int
		dateClosed  interface{}
		activeTabID int
	}{
		{1, 100, nil, 102},
		{2, 300, 782000000.0, 0},
		{3, 200, nil, 201},
	}
	for _, w := range windows {
		if _, err := db.Exec(`INSERT INTO windows (id, active_tab_group_id, date_closed) VALUES (?, ?, ?)`,
			w.id, w.groupID, w.dateClosed); err != nil {
			t.Fatalf("insert window %d: %v", w.id, err)
		}
		if _, err := db.Exec(`INSERT INTO windows_tab_groups (window_id, tab_group_id, active_tab_id) VALUES (?, ?, ?)`,
			w.id, w.groupID, w.activeTabID); err != nil {
			t.Fatalf("insert window group %d: %v", w.id, err)
		}
	}

	return dbPath
}

func TestSafariTabsHandlerExtractsOpenTabs(t *testing.T) {
	dbPath := createSafariTabsDB(t)

	entries, err := NewSafariTabsHandler(dbPath).GetTabs()
	if err != nil {
		t.Fatalf("GetTabs() error = %v", err)
	}

	// 2 tabs in window 1 + 1 in window 3; closed window, placeholders, and
	// deleted rows excluded.
	if len(entries) != 3 {
		t.Fatalf("expected 3 open tabs, got %d", len(entries))
	}

	first := entries[0]
	if first.URL != "https://example.com/article" {
		t.Fatalf("expected first tab URL from window 1, got %q", first.URL)
	}
	if first.Title != "Example Article" {
		t.Fatalf("expected tab title, got %q", first.Title)
	}
	if first.Domain != "example.com" {
		t.Fatalf("expected example.com domain, got %q", first.Domain)
	}
	if first.Browser != "safari" {
		t.Fatalf("expected safari browser, got %q", first.Browser)
	}
	if first.WindowID != 1 {
		t.Fatalf("expected window id 1, got %d", first.WindowID)
	}
	if first.Group != "" {
		t.Fatalf("expected internal group title 'Local' to be suppressed, got %q", first.Group)
	}
	if first.Active {
		t.Fatalf("expected first tab to be inactive")
	}

	// Active tab comes from windows_tab_groups.active_tab_id
	if !entries[1].Active {
		t.Fatalf("expected second tab (id 102) to be active")
	}

	// Named tab group surfaces as Group
	if entries[2].Group != "Research" {
		t.Fatalf("expected group name 'Research', got %q", entries[2].Group)
	}
	if entries[2].WindowID != 3 {
		t.Fatalf("expected window id 3 for named group, got %d", entries[2].WindowID)
	}
}

func TestSafariTabsHandlerMissingDB(t *testing.T) {
	_, err := QuerySafariTabs(filepath.Join(t.TempDir(), "nope.db"))
	if err == nil {
		t.Fatalf("expected error for missing SafariTabs.db")
	}
}
