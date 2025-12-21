# web-recap

Extract browser history and bookmarks from Chrome, Chromium, Brave, Vivaldi, Firefox, Safari, and Edge browsers and output them in JSON format suitable for analysis by LLMs and other tools.

> **Privacy:** This tool runs entirely on your machine and never transmits data. Your browser history and bookmarks stay local unless you explicitly pipe them to an external service (like an LLM API).

## Features

- **Multi-browser support**: Chrome, Chromium, Edge, Brave, Vivaldi, Firefox, and Safari
- **Cross-platform**: Works on Linux, macOS, and Windows
- **History & Bookmarks**: Extract both browsing history and bookmarks
- **Automatic detection**: Auto-detects installed browsers or specify manually
- **Date filtering**: Extract history and bookmarks for specific dates or date ranges
- **Timezone support**: Parse dates in your local timezone or specify any timezone
- **Time filtering**: Extract history for specific hours or time ranges
- **Folder structure**: Preserves bookmark folder hierarchy
- **Tags support**: Extracts Firefox bookmark tags
- **LLM-friendly output**: JSON format optimized for consumption by language models
- **Minimal dependencies**: Pure Go implementation with no CGO required for better cross-platform compilation

## Installation

### Download Binary

Download the latest binary from [GitHub Releases](https://github.com/robzolkos/web-recap/releases):

| Platform | Binary |
|----------|--------|
| Linux | `web-recap-linux-amd64` |
| macOS (Intel) | `web-recap-darwin-amd64` |
| macOS (Apple Silicon) | `web-recap-darwin-arm64` |
| Windows | `web-recap-windows-amd64.exe` |

```bash
# Linux
curl -L https://github.com/robzolkos/web-recap/releases/latest/download/web-recap-linux-amd64 -o ~/.local/bin/web-recap
chmod +x ~/.local/bin/web-recap

# macOS (Apple Silicon)
curl -L https://github.com/robzolkos/web-recap/releases/latest/download/web-recap-darwin-arm64 -o ~/.local/bin/web-recap
chmod +x ~/.local/bin/web-recap

# macOS (Intel)
curl -L https://github.com/robzolkos/web-recap/releases/latest/download/web-recap-darwin-amd64 -o ~/.local/bin/web-recap
chmod +x ~/.local/bin/web-recap
```

> **Note:** Ensure `~/.local/bin` is in your PATH. Add `export PATH="$HOME/.local/bin:$PATH"` to your shell config if needed.

### Build from Source

Requires Go 1.21+

```bash
git clone https://github.com/robzolkos/web-recap.git
cd web-recap
go build ./cmd/web-recap
./web-recap --help
```

## Usage

### Basic Commands

```bash
# Show help
web-recap --help

# List detected browsers
web-recap list

# Show version
web-recap version
```

### Extract Bookmarks

```bash
# Extract bookmarks from default browser
web-recap bookmarks

# Extract from specific browser
web-recap bookmarks --browser chrome
web-recap bookmarks --browser firefox
web-recap bookmarks --browser safari

# Extract from all browsers
web-recap bookmarks --all-browsers

# Save to file
web-recap bookmarks -o bookmarks.json

# Custom bookmark path
web-recap bookmarks --db-path /path/to/Bookmarks

# Filter by date - bookmarks added on specific date
web-recap bookmarks --date 2025-12-15

# Filter by date range - bookmarks added between dates
web-recap bookmarks --start-date 2025-12-01 --end-date 2025-12-15

# Combine with timezone support
web-recap bookmarks --date 2025-12-15 --tz America/New_York

# Get bookmarks from last week
web-recap bookmarks --start-date 2025-12-09 --end-date 2025-12-16
```

### Extract History

```bash
# Extract today's history from default browser
web-recap

# Extract from specific browser
web-recap --browser chrome
web-recap --browser firefox
web-recap --browser safari

# Extract from specific date
web-recap --date 2025-12-15

# Extract date range
web-recap --start-date 2025-12-01 --end-date 2025-12-15

# Extract from all browsers
web-recap --all-browsers

# Save to file
web-recap -o history.json

# Custom database path
web-recap --db-path /path/to/History

# Timezone support (dates interpreted in your timezone)
web-recap --date 2025-12-15 --tz America/New_York

# Explicit UTC mode
web-recap --date 2025-12-15 --utc

# Time filtering - extract history for specific hours
web-recap --date 2025-12-15 --start-time 12:00 --end-time 13:00

# Shorthand for single hour
web-recap --date 2025-12-15 --time 12  # Extracts 12:00-12:59
```

### Command Examples

```bash
# Get Chrome history from last 7 days
web-recap --browser chrome --start-date 2025-12-09 --end-date 2025-12-16

# Export all browser history to file
web-recap --all-browsers --output all-history.json

# Check what browsers are available
web-recap list

# Extract yesterday's activity between 12pm and 1pm (in your local timezone)
web-recap --date "$(date -d yesterday +%Y-%m-%d)" --start-time 12:00 --end-time 13:00

# Extract from a specific timezone
web-recap --date 2025-12-15 --tz Europe/London --time 14

# Explicitly use UTC (useful in scripts or CI)
web-recap --start-date 2025-12-09 --end-date 2025-12-15 --utc --all-browsers

# Get bookmarks about Claude added in the last week
web-recap bookmarks --start-date 2025-12-09 --end-date 2025-12-16 | grep -i claude

# Extract all bookmarks from Firefox added in December 2025
web-recap bookmarks --browser firefox --start-date 2025-12-01 --end-date 2025-12-31
```

## JSON Output Formats

### History Output Format

The tool outputs history in the following JSON format:

```json
{
  "browser": "chrome",
  "start_date": "2025-12-15T00:00:00Z",
  "end_date": "2025-12-15T23:59:59Z",
  "timezone": "America/New_York",
  "total_entries": 343,
  "entries": [
    {
      "timestamp": "2025-12-15T09:15:23Z",
      "url": "https://example.com/page",
      "title": "Example Page Title",
      "visit_count": 3,
      "domain": "example.com",
      "browser": "chrome"
    },
    ...
  ]
}
```

### Bookmark Output Format

The tool outputs bookmarks in the following JSON format:

```json
{
  "browser": "chrome",
  "start_date": "2025-12-01T00:00:00Z",
  "end_date": "2025-12-31T23:59:59Z",
  "timezone": "America/New_York",
  "total_entries": 150,
  "entries": [
    {
      "date_added": "2025-12-15T10:30:00Z",
      "date_modified": "2025-12-16T14:20:00Z",
      "url": "https://example.com/article",
      "title": "Interesting Article",
      "folder": "Bookmarks Bar/Tech",
      "domain": "example.com",
      "browser": "chrome",
      "tags": ["golang", "tutorial"]
    },
    ...
  ]
}
```

Note: `start_date`, `end_date`, and `timezone` fields are only included when date filtering is used.

## Output Fields

### History Fields

- **browser**: Browser name (chrome, firefox, safari, edge)
- **start_date**: Report period start (ISO 8601 UTC format)
- **end_date**: Report period end (ISO 8601 UTC format)
- **timezone**: Timezone used for date interpretation (e.g., "America/New_York", "UTC")
- **total_entries**: Number of history entries in the report
- **entries**: Array of history entries, each containing:
  - **timestamp**: Visit time in ISO 8601 UTC format
  - **url**: Full URL visited
  - **title**: Page title
  - **visit_count**: Total visits to this URL
  - **domain**: Extracted domain name
  - **browser**: Browser source

### Bookmark Fields

- **browser**: Browser name (chrome, firefox, safari, edge, brave)
- **start_date**: Filter period start (ISO 8601 UTC format, only when date filtering is used)
- **end_date**: Filter period end (ISO 8601 UTC format, only when date filtering is used)
- **timezone**: Timezone used for date interpretation (only when date filtering is used)
- **total_entries**: Number of bookmark entries in the report
- **entries**: Array of bookmark entries, each containing:
  - **date_added**: When bookmark was created (ISO 8601 UTC format)
  - **date_modified**: When bookmark was last modified (ISO 8601 UTC format, optional)
  - **url**: Full URL of the bookmark
  - **title**: Bookmark title
  - **folder**: Folder path (e.g., "Bookmarks Bar/Work/Projects")
  - **domain**: Extracted domain name
  - **browser**: Browser source
  - **tags**: Array of tags (Firefox only)

## LLM Usage

The JSON output is designed to be easily consumed by language models. You can pipe the output directly to your LLM:

```bash
# Get chrome history and pass to Claude
web-recap --browser chrome --date 2025-12-15 | claude --prompt "Summarize my web activity"

# Extract bookmarks for analysis
web-recap bookmarks --all-browsers | claude --prompt "Organize my bookmarks by category"

# Find bookmarks about a specific topic from last week
web-recap bookmarks --start-date 2025-12-09 --end-date 2025-12-16 | \
  claude --prompt "Show me all my bookmarks about Claude and AI agents"

# Or save to file for later analysis
web-recap --all-browsers --output history.json
web-recap bookmarks --all-browsers --output bookmarks.json
# Then use with your LLM as context
```

## Supported Browsers

### Chrome/Chromium/Edge/Brave/Vivaldi
- **Platforms**: Linux, macOS, Windows
- **Database**: SQLite (`History` file)
- **Timestamp format**: Microseconds since 1601-01-01

### Firefox
- **Platforms**: Linux, macOS, Windows
- **Database**: SQLite (`places.sqlite`)
- **Timestamp format**: Microseconds since Unix epoch

### Safari
- **Platforms**: macOS only
- **Database**: SQLite (`History.db`)
- **Timestamp format**: Seconds since 2001-01-01

## Database Locations

### Linux

**History:**
- Chrome: `~/.config/google-chrome/Default/History`
- Chromium: `~/.config/chromium/Default/History`
- Edge: `~/.config/microsoft-edge/Default/History`
- Brave: `~/.config/BraveSoftware/Brave-Browser/Default/History`
- Vivaldi: `~/.config/vivaldi/Default/History`
- Firefox: `~/.mozilla/firefox/*/places.sqlite`

**Bookmarks:**
- Chrome: `~/.config/google-chrome/Default/Bookmarks`
- Chromium: `~/.config/chromium/Default/Bookmarks`
- Edge: `~/.config/microsoft-edge/Default/Bookmarks`
- Brave: `~/.config/BraveSoftware/Brave-Browser/Default/Bookmarks`
- Vivaldi: `~/.config/vivaldi/Default/Bookmarks`
- Firefox: `~/.mozilla/firefox/*/places.sqlite` (same as history)

### macOS

**History:**
- Chrome: `~/Library/Application Support/Google/Chrome/Default/History`
- Chromium: `~/Library/Application Support/Chromium/Default/History`
- Edge: `~/Library/Application Support/Microsoft Edge/Default/History`
- Brave: `~/Library/Application Support/BraveSoftware/Brave-Browser/Default/History`
- Vivaldi: `~/Library/Application Support/Vivaldi/Default/History`
- Firefox: `~/Library/Application Support/Firefox/*/places.sqlite`
- Safari: `~/Library/Safari/History.db`

**Bookmarks:**
- Chrome: `~/Library/Application Support/Google/Chrome/Default/Bookmarks`
- Chromium: `~/Library/Application Support/Chromium/Default/Bookmarks`
- Edge: `~/Library/Application Support/Microsoft Edge/Default/Bookmarks`
- Brave: `~/Library/Application Support/BraveSoftware/Brave-Browser/Default/Bookmarks`
- Vivaldi: `~/Library/Application Support/Vivaldi/Default/Bookmarks`
- Firefox: `~/Library/Application Support/Firefox/*/places.sqlite` (same as history)
- Safari: `~/Library/Safari/Bookmarks.plist`

### Windows

**History:**
- Chrome: `%LOCALAPPDATA%\Google\Chrome\User Data\Default\History`
- Chromium: `%LOCALAPPDATA%\Chromium\User Data\Default\History`
- Edge: `%LOCALAPPDATA%\Microsoft\Edge\User Data\Default\History`
- Brave: `%LOCALAPPDATA%\BraveSoftware\Brave-Browser\User Data\Default\History`
- Vivaldi: `%LOCALAPPDATA%\Vivaldi\User Data\Default\History`
- Firefox: `%LOCALAPPDATA%\Mozilla\Firefox\*/places.sqlite`

**Bookmarks:**
- Chrome: `%LOCALAPPDATA%\Google\Chrome\User Data\Default\Bookmarks`
- Chromium: `%LOCALAPPDATA%\Chromium\User Data\Default\Bookmarks`
- Edge: `%LOCALAPPDATA%\Microsoft\Edge\User Data\Default\Bookmarks`
- Brave: `%LOCALAPPDATA%\BraveSoftware\Brave-Browser\User Data\Default\Bookmarks`
- Vivaldi: `%LOCALAPPDATA%\Vivaldi\User Data\Default\Bookmarks`
- Firefox: `%LOCALAPPDATA%\Mozilla\Firefox\*/places.sqlite` (same as history)

## Technical Details

### Database Locking
The tool automatically handles browser database locking by copying the database to a temporary file before reading it. This allows you to extract history and bookmarks while your browser is running.

### Bookmark Formats
Different browsers use different formats for storing bookmarks:
- **Chrome/Chromium/Edge/Brave/Vivaldi**: JSON file format with hierarchical folder structure
- **Firefox**: SQLite database (places.sqlite) with bookmarks in `moz_bookmarks` table, supports tags
- **Safari**: Property list (plist) format

### Timestamp Conversion
Each browser uses a different timestamp format which is automatically converted to ISO 8601 UTC format for consistency.

### Timezone Support
By default, web-recap interprets dates in your system's local timezone. You can:
- Explicitly specify a timezone with `--tz America/New_York`
- Force UTC interpretation with `--utc` (useful for scripts and CI/CD)
- Extract by time of day with `--start-time` and `--end-time` (in 24-hour format)
- Use the `--time` shorthand for single-hour extraction (e.g., `--time 12` extracts 12:00-12:59)

All dates are converted to UTC for database queries, and timestamps in the JSON output are always in UTC format.

### Date Filtering
When using `--date`, it extracts history for the entire 24-hour period in the specified timezone. When using `--start-date` and `--end-date`, both dates are inclusive and cover the full 24-hour period. Use `--start-time` and `--end-time` to narrow results to specific hours of the day.

## Development

### Build
```bash
go build ./cmd/web-recap
```

### Test
```bash
go test ./...
```

### Cross-compile for all platforms
```bash
make build-all
```

### Release Process

Releases are created manually by pushing a version tag. The GitHub Actions workflow will automatically build binaries for all platforms and create a GitHub release.

```bash
# Create and push a version tag
git tag v0.1.0
git push origin v0.1.0
```

This triggers the release workflow which:
1. Builds binaries for Linux (amd64), macOS (amd64, arm64), and Windows (amd64)
2. Creates a GitHub release with the tag name
3. Uploads all binaries as release assets

## Troubleshooting

### "No browsers detected"
Ensure your browser is installed at the default location. You can specify a custom database path:
```bash
web-recap --db-path /path/to/History
```

### Permission errors
Make sure you have read access to the browser database. On macOS, you may need to grant permissions:
```bash
xattr -d com.apple.quarantine ./web-recap
```

### Browser running errors
The tool should handle locked databases automatically, but if you get errors, close your browser or try again later.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

MIT License - see LICENSE file for details

## SKILL.md

See [SKILL.md](./SKILL.md) for integration instructions with Claude and other LLMs.
