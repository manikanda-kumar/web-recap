package urlutil

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

// tcoPattern matches t.co URLs
var tcoPattern = regexp.MustCompile(`https?://t\.co/[a-zA-Z0-9]+`)

// ExtractTcoURLs extracts all t.co URLs from text.
func ExtractTcoURLs(text string) []string {
	matches := tcoPattern.FindAllString(text, -1)
	seen := make(map[string]bool)
	var unique []string
	for _, m := range matches {
		if !seen[m] {
			seen[m] = true
			unique = append(unique, m)
		}
	}
	return unique
}

// ExpandURL follows HTTP redirects to resolve the final URL.
// Returns the expanded URL or the original if expansion fails.
func ExpandURL(shortURL string, timeout time.Duration) string {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// Create HTTP client that doesn't follow redirects automatically
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
		Timeout: timeout,
	}

	currentURL := shortURL
	maxRedirects := 10

	for i := 0; i < maxRedirects; i++ {
		req, err := http.NewRequestWithContext(ctx, "HEAD", currentURL, nil)
		if err != nil {
			return shortURL
		}

		req.Header.Set("User-Agent", "web-recap/1.0")

		resp, err := client.Do(req)
		if err != nil {
			return shortURL
		}
		resp.Body.Close()

		// Check if redirect
		if resp.StatusCode >= 300 && resp.StatusCode < 400 {
			location := resp.Header.Get("Location")
			if location == "" {
				break
			}

			// Handle relative redirects
			nextURL, err := url.Parse(location)
			if err != nil {
				break
			}

			if !nextURL.IsAbs() {
				baseURL, err := url.Parse(currentURL)
				if err != nil {
					break
				}
				nextURL = baseURL.ResolveReference(nextURL)
			}

			currentURL = nextURL.String()
		} else {
			// Not a redirect, we've reached the final destination
			break
		}
	}

	return currentURL
}

// ExpandTcoURLsInText finds all t.co URLs in text and expands them.
// Returns a map of original t.co URL -> expanded URL.
func ExpandTcoURLsInText(text string) map[string]string {
	tcoURLs := ExtractTcoURLs(text)
	if len(tcoURLs) == 0 {
		return nil
	}

	expanded := make(map[string]string)
	timeout := 5 * time.Second

	for _, tcoURL := range tcoURLs {
		expandedURL := ExpandURL(tcoURL, timeout)
		// Only include if expansion actually changed the URL
		if expandedURL != tcoURL {
			expanded[tcoURL] = expandedURL
		}
	}

	return expanded
}

// CleanURL removes tracking parameters from URLs.
func CleanURL(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}

	// Common tracking parameters to remove
	trackingParams := []string{
		"utm_source", "utm_medium", "utm_campaign", "utm_term", "utm_content",
		"fbclid", "gclid", "msclkid",
	}

	q := u.Query()
	for _, param := range trackingParams {
		q.Del(param)
	}
	u.RawQuery = q.Encode()

	return u.String()
}

// FormatExpandedURLs formats the expanded URLs map for display.
func FormatExpandedURLs(expanded map[string]string) string {
	if len(expanded) == 0 {
		return ""
	}

	var parts []string
	for short, long := range expanded {
		parts = append(parts, fmt.Sprintf("%s → %s", short, long))
	}
	return strings.Join(parts, "\n")
}
