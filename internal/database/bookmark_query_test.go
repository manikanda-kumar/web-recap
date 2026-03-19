package database

import (
	"testing"
	"time"

	"github.com/rzolkos/web-recap/internal/models"
)

func TestBookmarkEntryLess(t *testing.T) {
	now := time.Date(2026, 1, 20, 10, 0, 0, 0, time.UTC)
	later := now.Add(2 * time.Hour)

	tests := []struct {
		name string
		a    models.BookmarkEntry
		b    models.BookmarkEntry
		want bool
	}{
		{
			name: "dated entries sort desc",
			a:    models.BookmarkEntry{DateAdded: later, URL: "https://a", Title: "A"},
			b:    models.BookmarkEntry{DateAdded: now, URL: "https://b", Title: "B"},
			want: true,
		},
		{
			name: "dated entry before undated",
			a:    models.BookmarkEntry{DateAdded: now, URL: "https://a", Title: "A"},
			b:    models.BookmarkEntry{URL: "https://b", Title: "B"},
			want: true,
		},
		{
			name: "undated entry after dated",
			a:    models.BookmarkEntry{URL: "https://a", Title: "A"},
			b:    models.BookmarkEntry{DateAdded: now, URL: "https://b", Title: "B"},
			want: false,
		},
		{
			name: "undated tie-breaks by title then url",
			a:    models.BookmarkEntry{URL: "https://z", Title: "A"},
			b:    models.BookmarkEntry{URL: "https://a", Title: "B"},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := bookmarkEntryLess(tt.a, tt.b); got != tt.want {
				t.Fatalf("bookmarkEntryLess() = %v, want %v", got, tt.want)
			}
		})
	}
}
