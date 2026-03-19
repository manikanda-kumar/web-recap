package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/rzolkos/web-recap/internal/browser"
	"github.com/rzolkos/web-recap/internal/database"
	"github.com/rzolkos/web-recap/internal/models"
	"github.com/rzolkos/web-recap/internal/output"
	"github.com/rzolkos/web-recap/internal/readinglist"
	"github.com/rzolkos/web-recap/internal/twitter"
	"github.com/rzolkos/web-recap/internal/youtube"
	"github.com/spf13/cobra"
	"google.golang.org/api/option"
)

var (
	browserType string
	date        string
	startDate   string
	endDate     string
	startTime   string
	endTime     string
	timeHour    string
	timezone    string
	utcMode     bool
	outputFile  string
	dbPath      string
	allBrowsers bool
	version     = "0.1.0-alpha"
	// Reading list flags
	platform     string
	sessionToken string
	cookie       string
	username     string
	filePath     string
	publicURL    string
	allPlatforms bool

	// YouTube flags
	youtubeClientSecret string
	youtubeTokenPath    string
	youtubeDataPath     string
	youtubePlaylistID   string
	youtubeChannelID    string
	youtubeDebug        bool

	// YouTube copy-playlist flags
	copySourceData     string
	copyTargetPlaylist string
	copyPlaylistTitle  string
	copyPrivacyStatus  string

	// Twitter flags
	twitterDataPath     string
	twitterAuthToken    string
	twitterCt0          string
	twitterProvider     string
	composioAPIKey      string
	composioMCPURL      string
	composioUserID      string
	composioTwitterTool string
)

var rootCmd = &cobra.Command{
	Use:   "web-recap",
	Short: "Extract browser history in LLM-friendly JSON format",
	Long: `web-recap extracts browser history from Chrome, Chromium, Firefox, Safari, Edge, Brave, and Vivaldi
and outputs it in JSON format suitable for analysis by LLMs and other tools.

Date and time inputs are interpreted in your local timezone by default.

Examples:
  web-recap                          # Extract today's history from default browser
  web-recap --browser chrome         # Extract from Chrome specifically
  web-recap --date 2025-12-15        # Extract history from specific date (local time)
  web-recap --date 2025-12-15 --time 12           # Extract 12pm hour (12:00-12:59)
  web-recap --date 2025-12-15 --start-time 12:00 --end-time 13:00  # Time range
  web-recap --tz America/New_York --date 2025-12-15  # Explicit timezone
  web-recap --start-date 2025-12-01 --end-date 2025-12-15  # Date range
  web-recap --all-browsers -o history.json  # All browsers to file
`,
	RunE: runWeb,
}

func init() {
	// Persistent flags available to all subcommands
	rootCmd.PersistentFlags().StringVarP(&browserType, "browser", "b", "auto", "Browser type: auto, chrome, chromium, edge, brave, vivaldi, firefox, or safari")
	rootCmd.PersistentFlags().StringVar(&date, "date", "", "Specific date (YYYY-MM-DD, interpreted in local timezone)")
	rootCmd.PersistentFlags().StringVar(&startDate, "start-date", "", "Start date (YYYY-MM-DD, interpreted in local timezone)")
	rootCmd.PersistentFlags().StringVar(&endDate, "end-date", "", "End date (YYYY-MM-DD, interpreted in local timezone)")
	rootCmd.PersistentFlags().StringVar(&startTime, "start-time", "", "Start time (HH:MM format)")
	rootCmd.PersistentFlags().StringVar(&endTime, "end-time", "", "End time (HH:MM format)")
	rootCmd.PersistentFlags().StringVar(&timeHour, "time", "", "Time hour shorthand (e.g., '12' for 12:00-12:59)")
	rootCmd.PersistentFlags().StringVar(&timezone, "tz", "", "Timezone (e.g., America/New_York, UTC, local for system timezone)")
	rootCmd.PersistentFlags().BoolVar(&utcMode, "utc", false, "Treat all dates/times as UTC instead of local timezone")
	rootCmd.PersistentFlags().StringVarP(&outputFile, "output", "o", "", "Output file (default: stdout)")
	rootCmd.PersistentFlags().StringVar(&dbPath, "db-path", "", "Custom database path")
	rootCmd.PersistentFlags().BoolVar(&allBrowsers, "all-browsers", false, "Extract from all detected browsers")

	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(bookmarksCmd)
	rootCmd.AddCommand(tabsCmd)
	rootCmd.AddCommand(readingListCmd)
	rootCmd.AddCommand(youtubeWatchLaterCmd)
	rootCmd.AddCommand(youtubeCopyPlaylistCmd)
	rootCmd.AddCommand(twitterBookmarksCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// getTimezone returns the appropriate timezone based on flags
func getTimezone(tzFlag string, utcFlag bool) (*time.Location, error) {
	if utcFlag {
		return time.UTC, nil
	}

	if tzFlag != "" {
		if tzFlag == "local" {
			return time.Local, nil
		}
		loc, err := time.LoadLocation(tzFlag)
		if err != nil {
			return nil, fmt.Errorf("invalid timezone %q: %v", tzFlag, err)
		}
		return loc, nil
	}

	// Default to system local timezone
	return time.Local, nil
}

// parseDateTimeInLocation parses a date and optional time in a specific timezone
func parseDateTimeInLocation(dateStr, timeStr string, loc *time.Location) (time.Time, error) {
	if dateStr == "" {
		return time.Time{}, nil
	}

	// Parse date
	dateTime, err := time.ParseInLocation("2006-01-02", dateStr, loc)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid date format: %v", err)
	}

	if timeStr == "" {
		// No time specified, use start of day
		return dateTime, nil
	}

	// Parse time
	t, err := time.Parse("15:04", timeStr)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid time format (use HH:MM): %v", err)
	}

	// Combine date + time in the specified timezone
	return time.Date(dateTime.Year(), dateTime.Month(), dateTime.Day(),
		t.Hour(), t.Minute(), 0, 0, loc), nil
}

// parseHour parses a single hour value (0-23)
func parseHour(hourStr string) (int, error) {
	var hour int
	_, err := fmt.Sscanf(hourStr, "%d", &hour)
	if err != nil || hour < 0 || hour > 23 {
		return 0, fmt.Errorf("invalid hour (must be 0-23): %s", hourStr)
	}
	return hour, nil
}

func runWeb(cmd *cobra.Command, args []string) error {
	// Get timezone
	loc, err := getTimezone(timezone, utcMode)
	if err != nil {
		return err
	}

	// Parse dates with timezone
	var startTimeValue, endTimeValue time.Time
	var err2 error

	if date != "" {
		// Single date mode
		start, err := parseDateTimeInLocation(date, "", loc)
		if err != nil {
			return err
		}

		if timeHour != "" {
			// --time 12 means 12:00-12:59
			hour, err := parseHour(timeHour)
			if err != nil {
				return err
			}
			startTimeValue = time.Date(start.Year(), start.Month(), start.Day(),
				hour, 0, 0, 0, loc)
			endTimeValue = startTimeValue.Add(1 * time.Hour)
		} else if startTime != "" || endTime != "" {
			// Explicit time range
			var st, et string
			if startTime != "" {
				st = startTime
			} else {
				st = "00:00"
			}
			if endTime != "" {
				et = endTime
			} else {
				et = "23:59"
			}

			startTimeValue, err = parseDateTimeInLocation(date, st, loc)
			if err != nil {
				return err
			}
			endTimeValue, err = parseDateTimeInLocation(date, et, loc)
			if err != nil {
				return err
			}
		} else {
			// Full day
			startTimeValue = start
			endTimeValue = start.Add(24 * time.Hour)
		}
	} else if startDate != "" || endDate != "" {
		// Date range mode (existing logic, updated to use timezone)
		if startDate != "" {
			startTimeValue, err2 = parseDateTimeInLocation(startDate, "", loc)
			if err2 != nil {
				return err2
			}
		}

		if endDate != "" {
			endTimeValue, err2 = parseDateTimeInLocation(endDate, "", loc)
			if err2 != nil {
				return err2
			}
			endTimeValue = endTimeValue.Add(24 * time.Hour)
		}
	} else {
		// No date specified - default to today
		now := time.Now().In(loc)
		startTimeValue = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)
		endTimeValue = startTimeValue.Add(24 * time.Hour)
	}

	// Convert to UTC for database query (important!)
	startTimeValue = startTimeValue.UTC()
	endTimeValue = endTimeValue.UTC()

	// Get browser
	detector := browser.NewDetector()
	var b *browser.Browser

	// Default to all browsers if no specific browser and no --all-browsers flag
	useAllBrowsers := allBrowsers || browserType == "auto"

	if useAllBrowsers {
		// Handle multiple browsers
		entries, err := database.QueryMultipleBrowsers(detector, startTimeValue, endTimeValue)
		if err != nil {
			return fmt.Errorf("failed to query browsers: %v", err)
		}

		// Write output
		out := os.Stdout
		if outputFile != "" {
			f, err := os.Create(outputFile)
			if err != nil {
				return fmt.Errorf("failed to create output file: %v", err)
			}
			defer f.Close()
			out = f
		}

		return output.FormatJSON(out, entries, "all", startTimeValue, endTimeValue, timezone)
	}

	// Get specific browser
	bType := browser.Type(browserType)
	if dbPath != "" {
		// Validate custom path
		info, err := os.Stat(dbPath)
		if err != nil {
			if os.IsNotExist(err) {
				return fmt.Errorf("database file not found: %s", dbPath)
			}
			return fmt.Errorf("cannot access database file: %v", err)
		}
		if info.IsDir() {
			return fmt.Errorf("path is a directory, not a file: %s", dbPath)
		}

		// Use custom path
		b = &browser.Browser{
			Type: bType,
			Name: string(bType),
			Path: dbPath,
		}
	} else {
		var err error
		b, err = detector.GetBrowser(bType)
		if err != nil {
			return fmt.Errorf("failed to get browser: %v", err)
		}
	}

	// Query history
	entries, err := database.Query(b, startTimeValue, endTimeValue)
	if err != nil {
		return fmt.Errorf("failed to query history: %v", err)
	}

	// Write output
	out := os.Stdout
	if outputFile != "" {
		f, err := os.Create(outputFile)
		if err != nil {
			return fmt.Errorf("failed to create output file: %v", err)
		}
		defer f.Close()
		out = f
	}

	return output.FormatJSON(out, entries, b.Name, startTimeValue, endTimeValue, timezone)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("web-recap version %s\n", version)
	},
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List detected browsers",
	RunE: func(cmd *cobra.Command, args []string) error {
		detector := browser.NewDetector()
		browsers := detector.Detect()

		if len(browsers) == 0 {
			fmt.Println("No browsers detected")
			return nil
		}

		fmt.Println("Detected browsers:")
		for _, b := range browsers {
			fmt.Printf("  - %s (%s): %s\n", b.Name, b.Type, b.Path)
		}

		return nil
	},
}

var bookmarksCmd = &cobra.Command{
	Use:   "bookmarks",
	Short: "Extract browser bookmarks in JSON format",
	Long: `Extract bookmarks from Chrome, Chromium, Firefox, Safari, Edge, Brave, and Vivaldi browsers
and output them in JSON format.

Examples:
  web-recap bookmarks                          # Extract all bookmarks from default browser
  web-recap bookmarks --browser chrome         # Extract from Chrome specifically
  web-recap bookmarks --all-browsers           # Extract from all detected browsers
  web-recap bookmarks -o bookmarks.json        # Save to file
  web-recap bookmarks --date 2025-12-15        # Extract bookmarks added on specific date
  web-recap bookmarks --start-date 2025-12-01 --end-date 2025-12-15  # Date range
`,
	RunE: runBookmarks,
}

var tabsCmd = &cobra.Command{
	Use:   "tabs",
	Short: "Extract open browser tabs in JSON format",
	Long: `Extract open tabs from Chromium-based browsers (Chrome, Chromium, Edge, Brave, Vivaldi)
and output them in JSON format.

Note: This feature only works with Chromium-based browsers. Firefox and Safari are not supported yet.
Also note that the browser's session files may not be immediately updated, so there may be
a slight delay between actual browser state and what is reported.

Examples:
  web-recap tabs                          # Extract open tabs from default Chromium browser
  web-recap tabs --browser chrome         # Extract from Chrome specifically
  web-recap tabs --browser vivaldi        # Extract from Vivaldi
  web-recap tabs --all-browsers           # Extract from all detected Chromium browsers
  web-recap tabs -o tabs.json             # Save to file
`,
	RunE: runTabs,
}

func runTabs(cmd *cobra.Command, args []string) error {
	detector := browser.NewDetector()

	// Determine if we should query all browsers
	useAllBrowsers := allBrowsers || browserType == "auto"

	if useAllBrowsers {
		// Query all Chromium-based browsers
		entries, err := database.QueryMultipleBrowsersTabs(detector)
		if err != nil {
			return fmt.Errorf("failed to query tabs: %v", err)
		}

		if len(entries) == 0 {
			return fmt.Errorf("no open tabs found (only Chromium-based browsers are supported)")
		}

		// Write output
		out := os.Stdout
		if outputFile != "" {
			f, err := os.Create(outputFile)
			if err != nil {
				return fmt.Errorf("failed to create output file: %v", err)
			}
			defer f.Close()
			out = f
		}

		return output.FormatTabsJSON(out, entries, "all")
	}

	// Get specific browser
	bType := browser.Type(browserType)

	// Check if it's a Chromium-based browser
	if !browser.IsChromiumBased(bType) {
		return fmt.Errorf("tabs extraction only supported for Chromium-based browsers (chrome, chromium, edge, brave, vivaldi)")
	}

	var b *browser.Browser
	var sessionPath string

	if dbPath != "" {
		// Custom session path provided
		info, err := os.Stat(dbPath)
		if err != nil {
			if os.IsNotExist(err) {
				return fmt.Errorf("session path not found: %s", dbPath)
			}
			return fmt.Errorf("cannot access session path: %v", err)
		}

		if !info.IsDir() {
			return fmt.Errorf("session path must be a directory: %s", dbPath)
		}

		b = &browser.Browser{
			Type: bType,
			Name: string(bType),
			Path: dbPath,
		}
		sessionPath = dbPath
	} else {
		// Auto-detect browser
		var err error
		b, err = detector.GetBrowser(bType)
		if err != nil {
			return fmt.Errorf("failed to get browser: %v", err)
		}

		// Get session path
		sessionPath, err = browser.GetSessionPath(b.Type)
		if err != nil {
			return fmt.Errorf("failed to get session path: %v", err)
		}
	}

	// Query tabs
	entries, err := database.QueryTabs(b, sessionPath)
	if err != nil {
		return fmt.Errorf("failed to query tabs: %v", err)
	}

	if len(entries) == 0 {
		return fmt.Errorf("no open tabs found")
	}

	// Write output
	out := os.Stdout
	if outputFile != "" {
		f, err := os.Create(outputFile)
		if err != nil {
			return fmt.Errorf("failed to create output file: %v", err)
		}
		defer f.Close()
		out = f
	}

	return output.FormatTabsJSON(out, entries, b.Name)
}

func runBookmarks(cmd *cobra.Command, args []string) error {
	// Get timezone
	loc, err := getTimezone(timezone, utcMode)
	if err != nil {
		return err
	}

	// Parse dates with timezone (same logic as history)
	var startTimeValue, endTimeValue time.Time
	var err2 error

	if date != "" {
		// Single date mode
		start, err := parseDateTimeInLocation(date, "", loc)
		if err != nil {
			return err
		}

		if timeHour != "" {
			// --time 12 means 12:00-12:59
			hour, err := parseHour(timeHour)
			if err != nil {
				return err
			}
			startTimeValue = time.Date(start.Year(), start.Month(), start.Day(),
				hour, 0, 0, 0, loc)
			endTimeValue = startTimeValue.Add(1 * time.Hour)
		} else if startTime != "" || endTime != "" {
			// Explicit time range
			var st, et string
			if startTime != "" {
				st = startTime
			} else {
				st = "00:00"
			}
			if endTime != "" {
				et = endTime
			} else {
				et = "23:59"
			}

			startTimeValue, err = parseDateTimeInLocation(date, st, loc)
			if err != nil {
				return err
			}
			endTimeValue, err = parseDateTimeInLocation(date, et, loc)
			if err != nil {
				return err
			}
		} else {
			// Full day
			startTimeValue = start
			endTimeValue = start.Add(24 * time.Hour)
		}
	} else if startDate != "" || endDate != "" {
		// Date range mode
		if startDate != "" {
			startTimeValue, err2 = parseDateTimeInLocation(startDate, "", loc)
			if err2 != nil {
				return err2
			}
		}

		if endDate != "" {
			endTimeValue, err2 = parseDateTimeInLocation(endDate, "", loc)
			if err2 != nil {
				return err2
			}
			endTimeValue = endTimeValue.Add(24 * time.Hour)
		}
	}
	// If no date specified, leave as zero values to return all bookmarks

	// Convert to UTC for database query (important!)
	if !startTimeValue.IsZero() {
		startTimeValue = startTimeValue.UTC()
	}
	if !endTimeValue.IsZero() {
		endTimeValue = endTimeValue.UTC()
	}

	// Get browser detector
	detector := browser.NewDetector()

	// Determine if we should query all browsers
	useAllBrowsers := allBrowsers || browserType == "auto"

	if useAllBrowsers {
		// Query all browsers
		entries, err := database.QueryMultipleBrowsersBookmarks(detector, startTimeValue, endTimeValue)
		if err != nil {
			return fmt.Errorf("failed to query bookmarks: %v", err)
		}

		// Write output
		out := os.Stdout
		if outputFile != "" {
			f, err := os.Create(outputFile)
			if err != nil {
				return fmt.Errorf("failed to create output file: %v", err)
			}
			defer f.Close()
			out = f
		}

		return output.FormatBookmarksJSON(out, entries, "all", startTimeValue, endTimeValue, timezone)
	}

	// Get specific browser
	bType := browser.Type(browserType)
	var b *browser.Browser
	var bookmarkPath string

	if dbPath != "" {
		// Custom bookmark path provided
		info, err := os.Stat(dbPath)
		if err != nil {
			if os.IsNotExist(err) {
				return fmt.Errorf("bookmark file not found: %s", dbPath)
			}
			return fmt.Errorf("cannot access bookmark file: %v", err)
		}

		// For Firefox, dbPath might be a directory (profile path)
		if info.IsDir() && bType != browser.Firefox {
			return fmt.Errorf("path is a directory, not a file: %s", dbPath)
		}

		b = &browser.Browser{
			Type: bType,
			Name: string(bType),
			Path: dbPath,
		}
		bookmarkPath = dbPath
	} else {
		// Auto-detect browser
		var err error
		b, err = detector.GetBrowser(bType)
		if err != nil {
			return fmt.Errorf("failed to get browser: %v", err)
		}

		// Get bookmark path
		bookmarkPath, err = browser.GetBookmarkPath(b.Type)
		if err != nil {
			return fmt.Errorf("failed to get bookmark path: %v", err)
		}

		// For Firefox, find the profile
		if b.Type == browser.Firefox {
			bookmarkPath, err = browser.GetFirefoxProfilePath(bookmarkPath)
			if err != nil {
				return fmt.Errorf("failed to find Firefox profile: %v", err)
			}
		}
	}

	// Query bookmarks
	entries, err := database.QueryBookmarks(b, bookmarkPath, startTimeValue, endTimeValue)
	if err != nil {
		return fmt.Errorf("failed to query bookmarks: %v", err)
	}

	// Write output
	out := os.Stdout
	if outputFile != "" {
		f, err := os.Create(outputFile)
		if err != nil {
			return fmt.Errorf("failed to create output file: %v", err)
		}
		defer f.Close()
		out = f
	}

	return output.FormatBookmarksJSON(out, entries, b.Name, startTimeValue, endTimeValue, timezone)
}

var youtubeWatchLaterCmd = &cobra.Command{
	Use:   "youtube-watch-later",
	Short: "Fetch YouTube Watch later playlist URLs",
	Long: `Fetch your private YouTube Watch later playlist and output all video URLs.

This requires OAuth2 (not just an API key). Provide the OAuth client secret JSON
(downloaded from Google Cloud Console) via --client-secret.

By default, it writes a local JSON snapshot and on subsequent runs fetches only
new items based on the latest added_at timestamp in that file.

Examples:
  web-recap youtube-watch-later --client-secret data/youtube/client.json --data data/youtube/watch_later.json
  web-recap youtube-watch-later --client-secret data/youtube/client.json --token data/youtube/token.json --data data/youtube/watch_later.json -o data/youtube/watch_later.json
`,

	RunE: runYouTubeWatchLater,
}

func init() {
	youtubeWatchLaterCmd.Flags().StringVar(&youtubeClientSecret, "client-secret", "", "Path to Google OAuth client secret JSON")
	youtubeWatchLaterCmd.Flags().StringVar(&youtubeTokenPath, "token", "", "Path to cached OAuth token JSON (default: <client-secret>.token.json)")
	youtubeWatchLaterCmd.Flags().StringVar(&youtubeDataPath, "data", "data/youtube/watch_later.json", "Path to local Watch later data file")
	youtubeWatchLaterCmd.Flags().StringVar(&youtubePlaylistID, "playlist-id", "WL", "Playlist ID to fetch (default: WL for Watch Later)")
	youtubeWatchLaterCmd.Flags().StringVar(&youtubeChannelID, "channel-id", "", "Channel ID to use (debug/override; default: mine=true first channel)")
	youtubeWatchLaterCmd.Flags().BoolVar(&youtubeDebug, "debug", false, "Print debug info about discovered channels")
	_ = youtubeWatchLaterCmd.MarkFlagRequired("client-secret")
}

func runYouTubeWatchLater(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()

	client, err := youtube.GetClient(ctx, youtubeClientSecret, youtubeTokenPath)
	if err != nil {
		return err
	}

	var existingItems []models.YouTubePlaylistItem
	var since time.Time
	if youtubeDataPath != "" {
		if existing, err := youtube.LoadWatchLaterFile(youtubeDataPath); err == nil {
			existingItems = existing.Items
			since = youtube.MaxAddedAt(existing.Items)
		}
	}

	playlistID, newItems, err := youtube.FetchWatchLaterItemsWithOptions(ctx, option.WithHTTPClient(client), youtubePlaylistID, youtubeChannelID, youtubeDebug, since)
	if err != nil {
		return err
	}

	merged := youtube.MergeByVideoID(existingItems, newItems)

	report := models.YouTubeWatchLaterReport{
		FetchedAt:   time.Now().UTC(),
		PlaylistID:  playlistID,
		TotalItems:  len(merged),
		DeltaAdded:  len(newItems),
		Items:       merged,
		Source:      "youtube",
		Description: "YouTube Watch later playlist snapshot",
	}

	// Always update local data file if provided.
	if youtubeDataPath != "" {
		if err := youtube.SaveWatchLaterFile(youtubeDataPath, report); err != nil {
			return err
		}
	}

	out := os.Stdout
	if outputFile != "" {
		f, err := os.Create(outputFile)
		if err != nil {
			return fmt.Errorf("failed to create output file: %v", err)
		}
		defer f.Close()
		out = f
	}

	return output.FormatYouTubeWatchLaterJSON(out, report)
}

var youtubeCopyPlaylistCmd = &cobra.Command{
	Use:   "youtube-copy-playlist",
	Short: "Copy videos from Watch Later data to a new or existing public playlist",
	Long: `Read videos from a local data/youtube/watch_later.json file and insert them into
a YouTube playlist. If --target-playlist is not provided, a new playlist is created.

This requires OAuth2 with read-write access. On first run it will open a browser
for authorization (a separate token from the readonly one).

Examples:
  # Create a new public playlist from data/youtube/watch_later.json
  web-recap youtube-copy-playlist --client-secret data/youtube/client.json

  # Create with a custom title
  web-recap youtube-copy-playlist --client-secret data/youtube/client.json --title "My Watch Later Archive"

  # Add to an existing playlist
  web-recap youtube-copy-playlist --client-secret data/youtube/client.json --target-playlist PLxxxxxxxx

  # Create an unlisted playlist
  web-recap youtube-copy-playlist --client-secret data/youtube/client.json --privacy unlisted
`,

	RunE: runYouTubeCopyPlaylist,
}

func init() {
	youtubeCopyPlaylistCmd.Flags().StringVar(&youtubeClientSecret, "client-secret", "", "Path to Google OAuth client secret JSON")
	youtubeCopyPlaylistCmd.Flags().StringVar(&youtubeTokenPath, "token", "", "Path to cached OAuth token JSON (default: <client-secret>.rw-token.json)")
	youtubeCopyPlaylistCmd.Flags().StringVar(&copySourceData, "data", "data/youtube/watch_later.json", "Path to local Watch Later data file")
	youtubeCopyPlaylistCmd.Flags().StringVar(&copyTargetPlaylist, "target-playlist", "", "Existing playlist ID to add videos to (if empty, creates a new one)")
	youtubeCopyPlaylistCmd.Flags().StringVar(&copyPlaylistTitle, "title", "Watch Later Archive", "Title for the new playlist (ignored if --target-playlist is set)")
	youtubeCopyPlaylistCmd.Flags().StringVar(&copyPrivacyStatus, "privacy", "public", "Privacy status: public, unlisted, or private")
	_ = youtubeCopyPlaylistCmd.MarkFlagRequired("client-secret")
}

func runYouTubeCopyPlaylist(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()

	// Load videos from data file (auto-detect CSV vs JSON)
	var report *models.YouTubeWatchLaterReport
	var err error
	if strings.HasSuffix(strings.ToLower(copySourceData), ".csv") {
		report, err = youtube.LoadTakeoutCSV(copySourceData)
	} else {
		report, err = youtube.LoadWatchLaterFile(copySourceData)
	}
	if err != nil {
		return fmt.Errorf("load data file %s: %w", copySourceData, err)
	}

	if len(report.Items) == 0 {
		fmt.Println("No videos found in data file.")
		return nil
	}

	fmt.Printf("Found %d videos in %s\n", len(report.Items), copySourceData)

	// Get read-write OAuth client
	client, err := youtube.GetClientReadWrite(ctx, youtubeClientSecret, youtubeTokenPath)
	if err != nil {
		return err
	}

	targetID := copyTargetPlaylist

	// Create new playlist if no target specified
	if targetID == "" {
		fmt.Printf("Creating new %s playlist: %q\n", copyPrivacyStatus, copyPlaylistTitle)
		targetID, err = youtube.CreatePlaylist(ctx, option.WithHTTPClient(client), copyPlaylistTitle, "Archived from Watch Later", copyPrivacyStatus)
		if err != nil {
			return err
		}
		fmt.Printf("Created playlist: https://www.youtube.com/playlist?list=%s\n", targetID)
	}

	// Insert videos
	fmt.Printf("Inserting %d videos into playlist %s...\n", len(report.Items), targetID)

	videoIDs := make([]string, len(report.Items))
	for i, item := range report.Items {
		videoIDs[i] = item.VideoID
	}

	inserted, err := youtube.InsertVideosIntoPlaylist(ctx, option.WithHTTPClient(client), targetID, videoIDs)
	if err != nil {
		return err
	}

	fmt.Printf("Done! Inserted %d/%d videos.\n", inserted, len(videoIDs))
	return nil
}

var readingListCmd = &cobra.Command{
	Use:   "reading-list",
	Short: "Extract reading list/saved articles from Medium, Substack, etc.",
	Long: `Extract saved articles from platforms like Medium and Substack.

Supports multiple fetching strategies:
  1. Public URL scraping (for public Medium reading lists, no auth needed)
  2. Web scraping (requires authentication via cookies/session tokens)
  3. Manual file parsing (CSV for Medium, JSON for Substack)

The tool tries strategies in order until one succeeds.

Authentication can be provided via:
  - Command-line flags (--cookie, --session-token, --username)
  - Environment variables (MEDIUM_COOKIE, SUBSTACK_SESSION_TOKEN, etc.)
  - File path for manual exports (--file)

Examples:
  # Medium public reading list (no authentication needed!)
  web-recap reading-list --platform medium --url https://medium.com/@username/list/reading-list

  # Medium reading list (web scraping with cookie)
  export MEDIUM_COOKIE="your-cookie-string"
  web-recap reading-list --platform medium

  # Medium from CSV export
  web-recap reading-list --platform medium --file medium-export.csv

  # Substack saved posts (with session token)
  export SUBSTACK_SESSION_TOKEN="your-token"
  web-recap reading-list --platform substack

  # Substack from JSON export
  web-recap reading-list --platform substack --file substack-saves.json

  # All platforms with date range
  web-recap reading-list --all-platforms --start-date 2025-01-01 --end-date 2025-12-31

  # Save to file
  web-recap reading-list --platform medium -o reading-list.json
`,
	RunE: runReadingList,
}

func init() {
	readingListCmd.Flags().StringVarP(&platform, "platform", "p", "medium", "Platform: medium, substack, or all")
	readingListCmd.Flags().StringVar(&sessionToken, "session-token", "", "Session token for authentication")
	readingListCmd.Flags().StringVar(&cookie, "cookie", "", "Cookie string for authentication")
	readingListCmd.Flags().StringVar(&username, "username", "", "Username (for platform-specific features)")
	readingListCmd.Flags().StringVarP(&filePath, "file", "f", "", "Path to exported file (CSV for Medium, JSON for Substack)")
	readingListCmd.Flags().StringVar(&publicURL, "url", "", "Public reading list URL (e.g., https://medium.com/@username/list/reading-list)")
	readingListCmd.Flags().BoolVar(&allPlatforms, "all-platforms", false, "Fetch from all configured platforms")
}

func runReadingList(cmd *cobra.Command, args []string) error {
	// Get timezone
	loc, err := getTimezone(timezone, utcMode)
	if err != nil {
		return err
	}

	// Parse dates with timezone (same logic as history/bookmarks)
	var startTimeValue, endTimeValue time.Time
	var err2 error

	if date != "" {
		// Single date mode
		start, err := parseDateTimeInLocation(date, "", loc)
		if err != nil {
			return err
		}

		if timeHour != "" {
			hour, err := parseHour(timeHour)
			if err != nil {
				return err
			}
			startTimeValue = time.Date(start.Year(), start.Month(), start.Day(),
				hour, 0, 0, 0, loc)
			endTimeValue = startTimeValue.Add(1 * time.Hour)
		} else if startTime != "" || endTime != "" {
			var st, et string
			if startTime != "" {
				st = startTime
			} else {
				st = "00:00"
			}
			if endTime != "" {
				et = endTime
			} else {
				et = "23:59"
			}

			startTimeValue, err = parseDateTimeInLocation(date, st, loc)
			if err != nil {
				return err
			}
			endTimeValue, err = parseDateTimeInLocation(date, et, loc)
			if err != nil {
				return err
			}
		} else {
			startTimeValue = start
			endTimeValue = start.Add(24 * time.Hour)
		}
	} else if startDate != "" || endDate != "" {
		// Date range mode
		if startDate != "" {
			startTimeValue, err2 = parseDateTimeInLocation(startDate, "", loc)
			if err2 != nil {
				return err2
			}
		}

		if endDate != "" {
			endTimeValue, err2 = parseDateTimeInLocation(endDate, "", loc)
			if err2 != nil {
				return err2
			}
			endTimeValue = endTimeValue.Add(24 * time.Hour)
		}
	}
	// If no date specified, leave as zero values to return all entries

	// Convert to UTC for querying
	if !startTimeValue.IsZero() {
		startTimeValue = startTimeValue.UTC()
	}
	if !endTimeValue.IsZero() {
		endTimeValue = endTimeValue.UTC()
	}

	var entries []models.ReadingListEntry
	var platformName string

	if allPlatforms {
		// Query all platforms
		platforms := []readinglist.PlatformType{
			readinglist.PlatformMedium,
			readinglist.PlatformSubstack,
		}

		configs := make(map[readinglist.PlatformType]*readinglist.Config)

		for _, p := range platforms {
			// Load from env vars first
			envConfig, err := readinglist.LoadConfigFromEnv(p)
			if err != nil {
				continue
			}

			// Create flag config
			flagConfig := readinglist.LoadConfigFromFlags(p, sessionToken, cookie, username, filePath, publicURL)

			// Merge configs (flags take precedence)
			config := readinglist.MergeConfigs(flagConfig, envConfig)

			configs[p] = config
		}

		entries, err = readinglist.QueryMultiplePlatforms(platforms, configs, startTimeValue, endTimeValue)
		if err != nil {
			return fmt.Errorf("failed to query reading lists: %v", err)
		}

		platformName = "all"
	} else {
		// Query single platform
		platformType := readinglist.PlatformType(platform)

		// Load from env vars first
		envConfig, err := readinglist.LoadConfigFromEnv(platformType)
		if err != nil {
			return fmt.Errorf("unsupported platform: %s", platform)
		}

		// Create flag config
		flagConfig := readinglist.LoadConfigFromFlags(platformType, sessionToken, cookie, username, filePath, publicURL)

		// Merge configs (flags take precedence)
		config := readinglist.MergeConfigs(flagConfig, envConfig)

		entries, err = readinglist.Query(platformType, config, startTimeValue, endTimeValue)
		if err != nil {
			return fmt.Errorf("failed to query %s reading list: %v", platform, err)
		}

		platformName = platform
	}

	// Write output
	out := os.Stdout
	if outputFile != "" {
		f, err := os.Create(outputFile)
		if err != nil {
			return fmt.Errorf("failed to create output file: %v", err)
		}
		defer f.Close()
		out = f
	}

	return output.FormatReadingListJSON(out, entries, platformName, startTimeValue, endTimeValue, timezone)
}

var twitterBookmarksCmd = &cobra.Command{
	Use:   "twitter-bookmarks",
	Short: "Fetch Twitter/X bookmarks using Composio or bird",
	Long: `Fetch your Twitter/X bookmarks using Composio (preferred) or bird CLI.

Provider behavior:
  - auto (default): uses Composio when configured, otherwise falls back to bird
  - composio: requires COMPOSIO_API_KEY, COMPOSIO_MCP_URL, COMPOSIO_USER_ID
  - bird: requires bird CLI installed and browser cookies/session

Install bird from: https://github.com/steipete/bird

By default, it writes a local JSON snapshot and on subsequent runs fetches only
new items based on the latest saved_at timestamp in that file.

Examples:
  web-recap twitter-bookmarks
  web-recap twitter-bookmarks --provider composio
  COMPOSIO_API_KEY=... COMPOSIO_MCP_URL=... COMPOSIO_USER_ID=... web-recap twitter-bookmarks --provider composio
  web-recap twitter-bookmarks --provider bird
  web-recap twitter-bookmarks --data data/twitter/bookmarks.json
  web-recap twitter-bookmarks -o bookmarks.json
`,
	RunE: runTwitterBookmarks,
}

func init() {
	twitterBookmarksCmd.Flags().StringVar(&twitterDataPath, "data", "data/twitter/bookmarks.json", "Path to local Twitter bookmarks data file")
	twitterBookmarksCmd.Flags().StringVar(&twitterProvider, "provider", "auto", "Provider: auto, composio, bird")
	twitterBookmarksCmd.Flags().StringVar(&twitterAuthToken, "auth-token", "", "Twitter auth_token (from browser cookies)")
	twitterBookmarksCmd.Flags().StringVar(&twitterCt0, "ct0", "", "Twitter ct0 token (from browser cookies)")
	twitterBookmarksCmd.Flags().StringVar(&composioAPIKey, "composio-api-key", "", "Composio API key (default: COMPOSIO_API_KEY)")
	twitterBookmarksCmd.Flags().StringVar(&composioMCPURL, "composio-mcp-url", "", "Composio MCP URL (default: COMPOSIO_MCP_URL)")
	twitterBookmarksCmd.Flags().StringVar(&composioUserID, "composio-user-id", "", "Composio user ID (default: COMPOSIO_USER_ID)")
	twitterBookmarksCmd.Flags().StringVar(&composioTwitterTool, "composio-tool", "", "Composio tool slug override (default: TWITTER_BOOKMARKS_BY_USER)")
}

func runTwitterBookmarks(cmd *cobra.Command, args []string) error {
	if composioAPIKey == "" {
		composioAPIKey = os.Getenv("COMPOSIO_API_KEY")
	}
	if composioMCPURL == "" {
		composioMCPURL = os.Getenv("COMPOSIO_MCP_URL")
	}
	if composioUserID == "" {
		composioUserID = os.Getenv("COMPOSIO_USER_ID")
	}

	var existingItems []models.TwitterBookmark
	var since time.Time
	if twitterDataPath != "" {
		if existing, err := twitter.LoadBookmarksFile(twitterDataPath); err == nil {
			existingItems = existing.Items
			since = twitter.MaxSavedAt(existing.Items)
		}
	}

	composioConfig := twitter.ComposioConfig{
		APIKey: composioAPIKey,
		MCPURL: composioMCPURL,
		UserID: composioUserID,
		Tool:   composioTwitterTool,
	}

	newItems, err := twitter.FetchBookmarks(since, twitter.FetchProvider(twitterProvider), twitterAuthToken, twitterCt0, composioConfig)
	if err != nil {
		return err
	}

	merged := twitter.MergeByTweetID(existingItems, newItems)

	report := models.TwitterBookmarksReport{
		FetchedAt:   time.Now().UTC(),
		TotalItems:  len(merged),
		DeltaAdded:  len(newItems),
		Items:       merged,
		Source:      "twitter",
		Description: "Twitter/X bookmarks snapshot",
	}

	// Always update local data file if provided.
	if twitterDataPath != "" {
		if err := twitter.SaveBookmarksFile(twitterDataPath, report); err != nil {
			return err
		}
	}

	out := os.Stdout
	if outputFile != "" {
		f, err := os.Create(outputFile)
		if err != nil {
			return fmt.Errorf("failed to create output file: %v", err)
		}
		defer f.Close()
		out = f
	}

	return output.FormatTwitterBookmarksJSON(out, report)
}
