package database

import (
	"database/sql"
	"fmt"
	"os"

	"github.com/rzolkos/web-recap/internal/models"
	_ "modernc.org/sqlite"
)

// SafariTabsHandler handles Safari open-tab extraction (macOS, Safari 15+).
//
// Since Safari 15 the open tabs no longer live in ~/Library/Safari/LastSession.plist
// (removed entirely in Safari 18). They are persisted in the SQLite database
// ~/Library/Containers/com.apple.Safari/Data/Library/Safari/SafariTabs.db.
//
// Schema (no public docs; verified against Safari 624 / macOS 26):
//   - `windows`: one row per window (date_closed IS NULL = currently open).
//   - `windows_tab_groups`: maps window_id → tab_group_id, plus active_tab_id.
//   - `bookmarks`: tree keyed by id/parent. Tab groups are type=1 nodes;
//     tabs are their children (type=0), ordered by order_index.
//     Note: `hidden` does NOT mean open — open groups are hidden=1 too.
//     The windows/windows_tab_groups join is the authoritative "open" set.
//   - Pinned tabs are stored under a separate special 'pinned' node, not under
//     the window's group; they are not extracted (unverified against live data).
type SafariTabsHandler struct {
	dbPath string
}

// NewSafariTabsHandler creates a new Safari tabs handler for the given SafariTabs.db path
func NewSafariTabsHandler(dbPath string) *SafariTabsHandler {
	return &SafariTabsHandler{dbPath: dbPath}
}

// safariInternalGroupTitles are tab group node titles Safari uses internally;
// they carry no user meaning and are omitted from the Group field.
var safariInternalGroupTitles = map[string]bool{
	"Local":    true, // default group for the normal (non-private) window
	"Private":  true, // private browsing group
	"Untitled": true, // unnamed tab group
}

// GetTabs retrieves currently open tabs from SafariTabs.db
func (h *SafariTabsHandler) GetTabs() ([]models.TabEntry, error) {
	// SafariTabs.db runs in WAL mode; copy -wal/-shm alongside so recent tab
	// changes are visible in the snapshot.
	tempDB, err := copySQLiteWithWAL(h.dbPath, "web-recap-safaritabs-*.db")
	if err != nil {
		return nil, err
	}
	defer os.Remove(tempDB)
	defer os.Remove(tempDB + "-wal")
	defer os.Remove(tempDB + "-shm")

	db, err := sql.Open("sqlite", tempDB)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	rows, err := db.Query(`
		SELECT
			w.id,
			COALESCE(g.title, ''),
			COALESCE(t.title, ''),
			t.url,
			CASE WHEN wtg.active_tab_id = t.id THEN 1 ELSE 0 END
		FROM windows w
		JOIN windows_tab_groups wtg ON wtg.window_id = w.id
		JOIN bookmarks g ON g.id = wtg.tab_group_id
		JOIN bookmarks t ON t.parent = g.id
		WHERE w.date_closed IS NULL
		  AND t.url IS NOT NULL AND t.url != ''
		  AND COALESCE(t.deleted, 0) = 0
		  AND t.title NOT IN ('TopScopedBookmarkList', 'Start Page')
		ORDER BY w.id, t.order_index ASC
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to query SafariTabs.db: %v", err)
	}
	defer rows.Close()

	var entries []models.TabEntry
	for rows.Next() {
		var windowID, active int
		var group, title, url string

		if err := rows.Scan(&windowID, &group, &title, &url, &active); err != nil {
			continue
		}

		entry := models.TabEntry{
			URL:      url,
			Title:    title,
			Domain:   ExtractDomain(url),
			Active:   active == 1,
			WindowID: windowID,
			Browser:  "safari",
		}
		if group != "" && !safariInternalGroupTitles[group] {
			entry.Group = group
		}

		entries = append(entries, entry)
	}

	return entries, nil
}

// copySQLiteWithWAL copies a SQLite database and its WAL/SHM sidecars (if present)
// to a temporary location so the live database is never opened directly.
func copySQLiteWithWAL(src, pattern string) (string, error) {
	data, err := os.ReadFile(src)
	if err != nil {
		return "", err
	}

	dst, err := os.CreateTemp("", pattern)
	if err != nil {
		return "", err
	}
	tmpFile := dst.Name()

	if _, err := dst.Write(data); err != nil {
		dst.Close()
		os.Remove(tmpFile)
		return "", err
	}
	if err := dst.Close(); err != nil {
		os.Remove(tmpFile)
		return "", err
	}

	for _, suffix := range []string{"-wal", "-shm"} {
		sidecar, err := os.ReadFile(src + suffix)
		if err != nil {
			continue // sidecar may not exist; that's fine
		}
		if err := os.WriteFile(tmpFile+suffix, sidecar, 0o600); err != nil {
			os.Remove(tmpFile)
			os.Remove(tmpFile + "-wal")
			os.Remove(tmpFile + "-shm")
			return "", err
		}
	}

	return tmpFile, nil
}

// QuerySafariTabs queries open tabs from a SafariTabs.db path
func QuerySafariTabs(dbPath string) ([]models.TabEntry, error) {
	if _, err := os.Stat(dbPath); err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("SafariTabs.db not found: %s", dbPath)
		}
		return nil, err
	}
	return NewSafariTabsHandler(dbPath).GetTabs()
}
