package readinglist

import (
	"fmt"
	"time"

	"github.com/gocolly/colly/v2"
	"github.com/gocolly/colly/v2/extensions"
	"github.com/rzolkos/web-recap/internal/database"
	"github.com/rzolkos/web-recap/internal/models"
)

// MediumWebStrategy uses web scraping to fetch Medium reading list
type MediumWebStrategy struct {
	config *Config
}

// NewMediumWebStrategy creates a new Medium web scraping strategy
func NewMediumWebStrategy(config *Config) *MediumWebStrategy {
	return &MediumWebStrategy{config: config}
}

// Name returns the strategy name
func (s *MediumWebStrategy) Name() string {
	return "web"
}

// IsAvailable checks if we have session token or cookie for authentication
func (s *MediumWebStrategy) IsAvailable() bool {
	return s.config.SessionToken != "" || s.config.Cookie != ""
}

// Fetch scrapes the Medium reading list page and returns entries
func (s *MediumWebStrategy) Fetch(config *Config, startDate, endDate time.Time) ([]models.ReadingListEntry, error) {
	c := colly.NewCollector(
		colly.UserAgent("Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36"),
	)

	// Set rate limiting to avoid being blocked
	if config.RateLimitDelay == 0 {
		config.RateLimitDelay = 2 * time.Second
	}

	c.Limit(&colly.LimitRule{
		DomainGlob:  "*medium.com*",
		Delay:       config.RateLimitDelay,
		RandomDelay: 1 * time.Second,
	})

	// Add cookies for authentication
	if config.Cookie != "" {
		c.OnRequest(func(r *colly.Request) {
			r.Headers.Set("Cookie", config.Cookie)
		})
	}

	extensions.RandomUserAgent(c)
	extensions.Referer(c)

	var entries []models.ReadingListEntry
	var scrapeErr error

	// Medium reading list HTML structure
	// Note: This selector may need adjustment based on Medium's actual HTML structure
	c.OnHTML("article, div[data-post-id]", func(e *colly.HTMLElement) {
		entry := models.ReadingListEntry{
			Platform: "medium",
			Title:    e.ChildText("h2, h3, [data-testid='storyTitle']"),
			URL:      e.Request.AbsoluteURL(e.ChildAttr("a", "href")),
			Author:   e.ChildText(".author-name, [data-testid='authorName'], a[rel='author']"),
			Excerpt:  e.ChildText("p, .excerpt, [data-testid='storyDescription']"),
			Domain:   database.ExtractDomain(e.Request.AbsoluteURL(e.ChildAttr("a", "href"))),
		}

		// Parse saved date if available
		dateStr := e.ChildAttr("time", "datetime")
		if dateStr != "" {
			if savedAt, err := time.Parse(time.RFC3339, dateStr); err == nil {
				entry.SavedAt = savedAt
			}
		}

		// Only add if we have at least a title and URL
		if entry.Title != "" && entry.URL != "" {
			entries = append(entries, entry)
		}
	})

	c.OnError(func(r *colly.Response, err error) {
		scrapeErr = fmt.Errorf("scraping error (status %d): %w", r.StatusCode, err)
	})

	// Visit Medium reading list page
	err := c.Visit("https://medium.com/me/list/reading-list")
	if err != nil {
		return nil, fmt.Errorf("failed to visit Medium: %w", err)
	}

	if scrapeErr != nil {
		return nil, scrapeErr
	}

	if len(entries) == 0 {
		return nil, fmt.Errorf("no entries found - check authentication or page structure")
	}

	// Filter by date range if specified
	if !startDate.IsZero() || !endDate.IsZero() {
		entries = filterByDateRange(entries, startDate, endDate)
	}

	return entries, nil
}
