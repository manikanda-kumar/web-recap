package readinglist

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gocolly/colly/v2"
	"github.com/gocolly/colly/v2/extensions"
	"github.com/rzolkos/web-recap/internal/database"
	"github.com/rzolkos/web-recap/internal/models"
)

// MediumPublicWebStrategy uses web scraping to fetch public Medium reading lists
type MediumPublicWebStrategy struct {
	config *Config
}

// NewMediumPublicWebStrategy creates a new Medium public web scraping strategy
func NewMediumPublicWebStrategy(config *Config) *MediumPublicWebStrategy {
	return &MediumPublicWebStrategy{config: config}
}

// Name returns the strategy name
func (s *MediumPublicWebStrategy) Name() string {
	return "public-web"
}

// IsAvailable checks if we have a public URL configured
func (s *MediumPublicWebStrategy) IsAvailable() bool {
	return s.config.PublicURL != ""
}

// Fetch scrapes a public Medium reading list page and returns entries
func (s *MediumPublicWebStrategy) Fetch(config *Config, startDate, endDate time.Time) ([]models.ReadingListEntry, error) {
	if config.PublicURL == "" {
		return nil, fmt.Errorf("public URL not configured")
	}

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

	extensions.RandomUserAgent(c)
	extensions.Referer(c)

	var entries []models.ReadingListEntry
	var scrapeErr error
	seenURLs := make(map[string]bool)

	// Medium public reading list HTML structure
	// Each article is in an <article> element with links to the story
	c.OnHTML("article", func(e *colly.HTMLElement) {
		// Extract the article URL from the main link
		articleURL := ""
		e.ForEach("a[href*='medium.com'], a[href^='/@'], a[href^='/']", func(_ int, link *colly.HTMLElement) {
			href := link.Attr("href")
			// Skip author profile links and bookmark links
			if strings.Contains(href, "/m/signin") || strings.Contains(href, "bookmark") {
				return
			}
			// Look for article links (contain alphanumeric ID at the end)
			if articleURL == "" && (strings.Contains(href, "?source=") || regexp.MustCompile(`-[a-f0-9]{10,}`).MatchString(href)) {
				articleURL = e.Request.AbsoluteURL(href)
				// Clean up the URL by removing source tracking parameters
				if idx := strings.Index(articleURL, "?source="); idx != -1 {
					articleURL = articleURL[:idx]
				}
			}
		})

		if articleURL == "" || seenURLs[articleURL] {
			return
		}
		seenURLs[articleURL] = true

		// Extract title from h2
		title := strings.TrimSpace(e.ChildText("h2"))
		if title == "" {
			return
		}

		// Extract subtitle/description from h3
		subtitle := strings.TrimSpace(e.ChildText("h3"))

		// Extract author - look for author links
		author := ""
		e.ForEach("a[href*='/@'], a[href*='source=list']", func(_ int, link *colly.HTMLElement) {
			href := link.Attr("href")
			// Author links contain /@ but not article content
			if strings.Contains(href, "/@") && !strings.Contains(href, "-") && author == "" {
				text := strings.TrimSpace(link.Text)
				// Filter out "In" prefix and publication names
				if text != "" && text != "In" && !strings.Contains(text, "by") {
					author = text
				}
			}
		})

		// Try to get author from paragraph if not found
		if author == "" {
			e.ForEach("p", func(_ int, p *colly.HTMLElement) {
				text := strings.TrimSpace(p.Text)
				if author == "" && text != "" && text != "In" && text != "by" && len(text) < 50 {
					author = text
				}
			})
		}

		// Extract publication if available
		publication := ""
		e.ForEach("a[href*='medium.com/']", func(_ int, link *colly.HTMLElement) {
			href := link.Attr("href")
			// Publication links don't have @ and aren't article links
			if !strings.Contains(href, "/@") && !strings.Contains(href, "?source=") && publication == "" {
				text := strings.TrimSpace(link.Text)
				if text != "" && text != "In" && text != author {
					publication = text
				}
			}
		})

		// Extract date - look for date text patterns
		dateStr := ""
		savedAt := time.Time{}
		e.ForEach("*", func(_ int, el *colly.HTMLElement) {
			text := strings.TrimSpace(el.Text)
			// Match patterns like "Dec 12", "May 31", "Nov 12, 2024", "Jan 16"
			if dateStr == "" {
				if matched, _ := regexp.MatchString(`^(Jan|Feb|Mar|Apr|May|Jun|Jul|Aug|Sep|Oct|Nov|Dec)\s+\d{1,2}(,\s*\d{4})?$`, text); matched {
					dateStr = text
				}
			}
		})

		if dateStr != "" {
			savedAt = parseMedianDate(dateStr)
		}

		// Extract claps count if available
		clapsStr := ""
		e.ForEach("*", func(_ int, el *colly.HTMLElement) {
			text := strings.TrimSpace(el.Text)
			// Match patterns like "333", "1.4K", "6.6K"
			if clapsStr == "" && (regexp.MustCompile(`^\d+(\.\d+)?K?$`).MatchString(text)) {
				clapsStr = text
			}
		})

		entry := models.ReadingListEntry{
			Platform:    "medium",
			Title:       title,
			URL:         articleURL,
			Author:      author,
			Publication: publication,
			Excerpt:     subtitle,
			Domain:      database.ExtractDomain(articleURL),
			SavedAt:     savedAt,
		}

		entries = append(entries, entry)
	})

	c.OnError(func(r *colly.Response, err error) {
		scrapeErr = fmt.Errorf("scraping error (status %d): %w", r.StatusCode, err)
	})

	// Visit the public reading list URL
	err := c.Visit(config.PublicURL)
	if err != nil {
		return nil, fmt.Errorf("Medium blocked the request - use the export script instead:\n"+
			"  1. Open your reading list in a browser: %s\n"+
			"  2. Open DevTools (F12) and go to Console\n"+
			"  3. Paste the script from scripts/export-medium-public.js\n"+
			"  4. Run: web-recap reading-list --platform medium --file medium-reading-list.json", config.PublicURL)
	}

	if scrapeErr != nil {
		return nil, scrapeErr
	}

	if len(entries) == 0 {
		return nil, fmt.Errorf("no entries found - Medium blocks automated access. Use the export script instead:\n" +
			"  1. Open your reading list in a browser: %s\n" +
			"  2. Open DevTools (F12) and go to Console\n" +
			"  3. Paste the script from scripts/export-medium-public.js\n" +
			"  4. Run: web-recap reading-list --platform medium --file medium-reading-list.json", config.PublicURL)
	}

	// Filter by date range if specified
	if !startDate.IsZero() || !endDate.IsZero() {
		entries = filterByDateRange(entries, startDate, endDate)
	}

	return entries, nil
}

// parseMedianDate parses Medium's date format
func parseMedianDate(dateStr string) time.Time {
	// Try formats with year first
	formats := []string{
		"Jan 2, 2006",
		"Jan 2 2006",
		"January 2, 2006",
	}

	for _, format := range formats {
		if t, err := time.Parse(format, dateStr); err == nil {
			return t
		}
	}

	// Try without year - assume current year, or previous year if date is in the future
	shortFormats := []string{
		"Jan 2",
		"January 2",
	}

	now := time.Now()
	for _, format := range shortFormats {
		if t, err := time.Parse(format, dateStr); err == nil {
			t = time.Date(now.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
			// If the date is in the future, assume it's from last year
			if t.After(now) {
				t = t.AddDate(-1, 0, 0)
			}
			return t
		}
	}

	return time.Time{}
}

// parseClaps converts clap strings like "1.4K" to integers
func parseClaps(s string) int {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0
	}

	multiplier := 1
	if strings.HasSuffix(s, "K") {
		multiplier = 1000
		s = strings.TrimSuffix(s, "K")
	}

	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0
	}

	return int(f * float64(multiplier))
}
