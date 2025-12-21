package database

import (
	"database/sql"
	"io"
	"os"
	"time"

	"github.com/rzolkos/web-recap/internal/models"
	_ "modernc.org/sqlite"
)

// FirefoxBookmarkHandler handles Firefox bookmark extraction
type FirefoxBookmarkHandler struct {
	dbPath string
}

// NewFirefoxBookmarkHandler creates a new Firefox bookmark handler
func NewFirefoxBookmarkHandler(dbPath string) *FirefoxBookmarkHandler {
	return &FirefoxBookmarkHandler{
		dbPath: dbPath,
	}
}

// GetBookmarks retrieves all bookmarks from Firefox
func (h *FirefoxBookmarkHandler) GetBookmarks(startTime, endTime time.Time) ([]models.BookmarkEntry, error) {
	// Copy database to temp location to avoid locking issues
	tempDB, err := h.copyDatabase()
	if err != nil {
		return nil, err
	}
	defer os.Remove(tempDB)

	db, err := sql.Open("sqlite", tempDB)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	// Firefox stores bookmarks in moz_bookmarks and moz_places tables
	// Type 1 = bookmark, Type 2 = folder, Type 3 = separator
	query := `
		SELECT
			b.dateAdded,
			b.lastModified,
			p.url,
			b.title,
			b.parent,
			p.id
		FROM moz_bookmarks b
		JOIN moz_places p ON b.fk = p.id
		WHERE b.type = 1
		AND p.url IS NOT NULL
		ORDER BY b.dateAdded DESC
	`

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var bookmarks []models.BookmarkEntry

	for rows.Next() {
		var dateAdded, dateModified int64
		var url string
		var title sql.NullString
		var parent, placeID int64

		if err := rows.Scan(&dateAdded, &dateModified, &url, &title, &parent, &placeID); err != nil {
			continue
		}

		// Convert timestamp
		dateAddedTime := ConvertFirefoxTimestamp(dateAdded)

		// Filter by date if time range is specified
		if !startTime.IsZero() && dateAddedTime.Before(startTime) {
			continue
		}
		if !endTime.IsZero() && dateAddedTime.After(endTime) {
			continue
		}

		// Get folder path
		folderPath := h.getFolderPath(db, parent)

		// Get tags
		tags := h.getTags(db, placeID)

		titleStr := ""
		if title.Valid {
			titleStr = title.String
		}

		bookmarks = append(bookmarks, models.BookmarkEntry{
			DateAdded:    dateAddedTime,
			DateModified: ConvertFirefoxTimestamp(dateModified),
			URL:          url,
			Title:        titleStr,
			Folder:       folderPath,
			Domain:       ExtractDomain(url),
			Browser:      "firefox",
			Tags:         tags,
		})
	}

	return bookmarks, rows.Err()
}

// getFolderPath builds the folder path for a bookmark
func (h *FirefoxBookmarkHandler) getFolderPath(db *sql.DB, parentID int64) string {
	var path []string

	for parentID > 0 {
		var title sql.NullString
		var newParent int64

		err := db.QueryRow(`
			SELECT title, parent
			FROM moz_bookmarks
			WHERE id = ? AND type = 2
		`, parentID).Scan(&title, &newParent)

		if err != nil {
			break
		}

		if title.Valid && title.String != "" {
			// Skip root folders
			if title.String != "root" && title.String != "menu" &&
			   title.String != "toolbar" && title.String != "unfiled" {
				path = append([]string{title.String}, path...)
			}
		}

		parentID = newParent
	}

	folderPath := ""
	for i, p := range path {
		if i > 0 {
			folderPath += "/"
		}
		folderPath += p
	}

	return folderPath
}

// getTags gets tags for a bookmark
func (h *FirefoxBookmarkHandler) getTags(db *sql.DB, placeID int64) []string {
	query := `
		SELECT b.title
		FROM moz_bookmarks b
		JOIN moz_bookmarks p ON b.parent = p.id
		WHERE b.fk = ? AND p.title = 'tags'
	`

	rows, err := db.Query(query, placeID)
	if err != nil {
		return nil
	}
	defer rows.Close()

	var tags []string
	for rows.Next() {
		var tag sql.NullString
		if err := rows.Scan(&tag); err != nil {
			continue
		}
		if tag.Valid && tag.String != "" {
			tags = append(tags, tag.String)
		}
	}

	return tags
}

// copyDatabase copies the Firefox database to a temporary file
func (h *FirefoxBookmarkHandler) copyDatabase() (string, error) {
	src, err := os.Open(h.dbPath)
	if err != nil {
		return "", err
	}
	defer src.Close()

	dst, err := os.CreateTemp("", "web-recap-firefox-bookmarks-*.db")
	if err != nil {
		return "", err
	}
	tmpFile := dst.Name()
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		os.Remove(tmpFile)
		return "", err
	}

	return tmpFile, nil
}
