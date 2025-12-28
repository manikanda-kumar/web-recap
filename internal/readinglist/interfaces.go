package readinglist

import (
	"time"

	"github.com/rzolkos/web-recap/internal/models"
)

// ReadingListQuerier defines the interface for fetching reading list entries
type ReadingListQuerier interface {
	GetReadingList(startDate, endDate time.Time) ([]models.ReadingListEntry, error)
}

// FetchStrategy defines the strategy for fetching reading list data
type FetchStrategy interface {
	Name() string
	Fetch(config *Config, startDate, endDate time.Time) ([]models.ReadingListEntry, error)
	IsAvailable() bool // Check if this strategy can be used
}

// Config holds authentication and configuration data
type Config struct {
	Platform    string
	Strategy    string // "web", "file", "auto"

	// Authentication (from env vars or flags)
	SessionToken string
	Cookie       string
	Username     string
	Password     string

	// File-based config
	FilePath string

	// Public reading list URL (for Medium public lists)
	PublicURL string

	// Rate limiting
	RateLimitDelay time.Duration
}
