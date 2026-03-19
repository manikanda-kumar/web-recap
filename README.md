# web-recap

Extract browser history and bookmarks from Chrome, Chromium, Brave, Vivaldi, Firefox, Safari, and Edge browsers and output them in JSON format suitable for analysis by LLMs and other tools.

> **Privacy:** This tool runs entirely on your machine and never transmits data. Your browser history and bookmarks stay local unless you explicitly pipe them to an external service (like an LLM API).

## Features

- **Multi-browser support**: Chrome, Chromium, Edge, Brave, Vivaldi, Firefox, and Safari
- **Cross-platform**: Works on Linux, macOS, and Windows
- **History, Bookmarks & Open Tabs**: Extract browsing history, bookmarks, and currently open tabs
- **Reading Lists**: Extract saved articles from Medium and Substack (with hybrid file export + web scraping)
- **YouTube Watch Later**: Extract your YouTube Watch Later playlist (requires OAuth2)
- **Twitter/X Bookmarks**: Extract your Twitter/X bookmarks using Composio (preferred) or bird CLI fallback
- **Automatic detection**: Auto-detects installed browsers or specify manually
- **Date filtering**: Extract history and bookmarks for specific dates or date ranges
- **Timezone support**: Parse dates in your local timezone or specify any timezone
- **Time filtering**: Extract history for specific hours or time ranges
- **Folder structure**: Preserves bookmark folder hierarchy
- **Tags support**: Extracts Firefox bookmark tags
- **LLM-friendly output**: JSON format optimized for consumption by language models
- **Minimal dependencies**: Pure Go implementation with no CGO required for better cross-platform compilation
- **Privacy-first**: Runs entirely on your machine, no data transmission (unless you explicitly pipe to external services)

## Installation

### Download Binary

Download the latest binary from [GitHub Releases](https://github.com/manikanda-kumar/web-recap/releases):

| Platform | Binary |
|----------|--------|
| Linux | `web-recap-linux-amd64` |
| macOS (Intel) | `web-recap-darwin-amd64` |
| macOS (Apple Silicon) | `web-recap-darwin-arm64` |
| Windows | `web-recap-windows-amd64.exe` |

```bash
# Linux
curl -L https://github.com/manikanda-kumar/web-recap/releases/latest/download/web-recap-linux-amd64 -o ~/.local/bin/web-recap
chmod +x ~/.local/bin/web-recap

# macOS (Apple Silicon)
curl -L https://github.com/manikanda-kumar/web-recap/releases/latest/download/web-recap-darwin-arm64 -o ~/.local/bin/web-recap
chmod +x ~/.local/bin/web-recap

# macOS (Intel)
curl -L https://github.com/manikanda-kumar/web-recap/releases/latest/download/web-recap-darwin-amd64 -o ~/.local/bin/web-recap
chmod +x ~/.local/bin/web-recap
```

> **Note:** Ensure `~/.local/bin` is in your PATH. Add `export PATH="$HOME/.local/bin:$PATH"` to your shell config if needed.

### Build from Source

Requires Go 1.21+

```bash
git clone https://github.com/manikanda-kumar/web-recap.git
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

### Extract Open Tabs

Extract currently open tabs from Chromium-based browsers (Chrome, Chromium, Edge, Brave, Vivaldi).

```bash
# Extract open tabs from all Chromium browsers
web-recap tabs

# Extract from specific browser
web-recap tabs --browser chrome
web-recap tabs --browser vivaldi

# Extract from all detected Chromium browsers
web-recap tabs --all-browsers

# Save to file
web-recap tabs -o tabs.json

# Custom session path
web-recap tabs --db-path /path/to/Sessions
```

> **Note:** Open tabs extraction only works with Chromium-based browsers. Firefox and Safari are not yet supported. There may be a slight delay between actual browser state and what is reported, as browsers don't immediately flush session data to disk.

### Extract Reading Lists (Medium, Substack)

Extract saved articles from Medium reading lists and Substack saved posts.

**Supports three methods:**
1. **Public URL export** (for public Medium reading lists, no auth needed)
2. **File export** (recommended for private lists, more reliable)
3. **Web scraping** (requires authentication, may break with platform changes)

#### Method 1: Public Medium Reading Lists

For public Medium reading lists (like `https://medium.com/@username/list/reading-list`):

```bash
# Export using the browser script, then import
# 1. Open the public reading list URL in your browser
# 2. Open DevTools (F12) → Console
# 3. Paste contents of scripts/export-medium-public.js
# 4. A JSON file will be downloaded

# Import the exported JSON file
web-recap reading-list --platform medium --file medium-reading-list-2025-12-28.json

# With date filtering
web-recap reading-list --platform medium --file medium-reading-list.json --start-date 2025-01-01
```

#### Method 2: File Export (Recommended for Private Lists)

Export your reading list to a file first, then parse it:

```bash
# Medium - Export to CSV or JSON first (see "Getting Export Files" below)
web-recap reading-list --platform medium --file medium-reading-list.csv
web-recap reading-list --platform medium --file medium-reading-list.json

# Substack - Export to JSON first
web-recap reading-list --platform substack --file substack-saves.json

# With date filtering
web-recap reading-list --platform medium --file medium.json --start-date 2025-01-01
```

#### Method 3: Web Scraping (Authentication Required)

Use cookies/session tokens from your browser:

```bash
# Medium - Using cookie
export MEDIUM_COOKIE="sid=YOUR_COOKIE_VALUE"
web-recap reading-list --platform medium

# Or pass directly
web-recap reading-list --platform medium --cookie "sid=YOUR_VALUE"

# Substack - Using session token
export SUBSTACK_SESSION_TOKEN="YOUR_TOKEN"
web-recap reading-list --platform substack

# Or using cookie
export SUBSTACK_COOKIE="substack.sid=VALUE; substack.lli=VALUE"
web-recap reading-list --platform substack
```

#### All Platforms

```bash
# Query all configured platforms
web-recap reading-list --all-platforms

# With date range
web-recap reading-list --all-platforms --start-date 2025-01-01 --end-date 2025-12-31

# Save to file
web-recap reading-list --platform medium -o reading-list.json
```

#### Getting Authentication Credentials

**For Medium:**

1. Open Medium and log in to your account
2. Go to your reading list: https://medium.com/me/list/reading-list
3. Open DevTools (F12 or Cmd+Option+I)
4. Go to **Application** → **Cookies** → **https://medium.com**
5. Copy the `sid` cookie value:
   ```bash
   export MEDIUM_COOKIE="sid=YOUR_SID_VALUE"
   ```

**For Substack:**

1. Open Substack and log in
2. Go to your inbox: https://substack.com/inbox
3. Open DevTools (F12) → **Network** tab
4. Refresh the page (F5)
5. Click any API request → **Headers** tab
6. Copy cookies from **Request Headers**:
   ```bash
   export SUBSTACK_COOKIE="substack.sid=VALUE; substack.lli=VALUE"
   ```

**Alternative: Copy full cookie string**
- In Network tab, right-click any request → **Copy as cURL**
- Extract the Cookie header from the cURL command

> **Security Note:** Never commit cookies to git. Cookies expire after weeks/months. Store them in environment variables or use the `--cookie` flag.

#### Getting Export Files

**Medium Public Reading List Export:**

For any public Medium reading list:
1. Open the public reading list URL (e.g., `https://medium.com/@username/list/reading-list`)
2. Open DevTools (F12) → Console
3. Paste contents of `scripts/export-medium-public.js`
4. Press Enter - a JSON file will download

**Medium Private Reading List Export:**

For your own reading list:
1. Open https://medium.com/me/list/reading-list (must be logged in)
2. Open DevTools (F12) → Console
3. Paste contents of `scripts/export-medium.js`
4. Press Enter - a CSV file will download

**Substack Export:**

1. Open https://substack.com/inbox → "Saved" tab
2. Open DevTools (F12) → Console
3. Paste contents of `scripts/export-substack.js`
4. Press Enter - a JSON file will download

See [scripts/README.md](./scripts/README.md) for detailed instructions.

**Quick Start**: See [QUICK_START_READING_LISTS.md](./QUICK_START_READING_LISTS.md) for a 5-minute setup guide.

**Full Documentation**: See [READING_LIST.md](./READING_LIST.md) for complete instructions.

### Extract YouTube Watch Later

Extract your private YouTube Watch Later playlist. This requires OAuth2 authentication.

#### Setup (One-time)

1. Go to [Google Cloud Console](https://console.cloud.google.com/)
2. Create a new project (or use existing)
3. Enable the **YouTube Data API v3**
4. Go to **Credentials** → **Create Credentials** → **OAuth client ID**
5. Select **Desktop app**, name it, and click **Create**
6. Download the JSON file (recommended path: `data/youtube_client.json`)

#### Usage

```bash
# First run - will open browser for OAuth authorization
web-recap youtube-watch-later --client-secret data/youtube_client.json

# Subsequent runs use cached token and fetch only new items
web-recap youtube-watch-later --client-secret data/youtube_client.json

# Save to specific output file
web-recap youtube-watch-later --client-secret data/youtube_client.json -o data/watch_later.json

# Specify custom data file location (for delta tracking)
web-recap youtube-watch-later --client-secret data/youtube_client.json --data data/my_watch_later.json

# Fetch any playlist by ID (not just Watch Later)
web-recap youtube-watch-later --client-secret data/youtube_client.json --playlist-id PLxxxxxxxx
```

#### How It Works

- On first run, opens browser for Google OAuth consent
- Saves OAuth token locally (default: `<client-secret>.token.json`)
- Maintains a local data file for delta sync (only fetches new items)
- Outputs video URLs, titles, channels, and timestamps

### Extract Twitter/X Bookmarks

Extract your Twitter/X bookmarks using **Composio** (preferred) or [bird CLI](https://github.com/steipete/bird) fallback.

#### Setup (One-time)

Option A: Composio (recommended)

1. Create a Composio account and connect Twitter/X
2. Get your API key and Tool Router MCP URL from Composio
3. Set environment variables:

```bash
export COMPOSIO_API_KEY="..."
export COMPOSIO_MCP_URL="https://..."
export COMPOSIO_USER_ID="your-user-id"
```

Option B: bird fallback

1. Install bird CLI from https://github.com/steipete/bird
2. Bird uses cookie-based authentication from your browser (Safari, Chrome, or Firefox)
3. Make sure you're logged into Twitter/X in your browser

#### Usage

```bash
# Fetch bookmarks (saves to data/twitter_bookmarks.json by default)
web-recap twitter-bookmarks

# Force Composio provider
web-recap twitter-bookmarks --provider composio

# Force bird provider
web-recap twitter-bookmarks --provider bird

# Save to specific output file
web-recap twitter-bookmarks -o bookmarks.json

# Specify custom data file location (for delta tracking)
web-recap twitter-bookmarks --data data/my_bookmarks.json
```

#### How It Works

- Uses Composio MCP `TWITTER_BOOKMARKS_BY_USER` when configured
- Falls back to bird CLI in `--provider auto` mode if Composio is unavailable
- Maintains a local data file for delta sync (only fetches new items)
- Outputs tweet URLs, text, authors, and timestamps

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

### Tabs Output Format

The tool outputs open tabs in the following JSON format:

```json
{
  "browser": "Google Chrome",
  "total_tabs": 15,
  "total_windows": 2,
  "entries": [
    {
      "url": "https://example.com/page",
      "title": "Example Page Title",
      "domain": "example.com",
      "active": true,
      "group": "Research",
      "window_id": 1,
      "browser": "Google Chrome"
    },
    ...
  ]
}
```

### Reading List Output Format

The tool outputs reading lists in the following JSON format:

```json
{
  "platform": "medium",
  "start_date": "2025-01-01T00:00:00Z",
  "end_date": "2025-12-31T23:59:59Z",
  "timezone": "America/New_York",
  "total_entries": 42,
  "entries": [
    {
      "saved_at": "2025-12-15T14:30:00Z",
      "url": "https://medium.com/@author/article-title",
      "title": "Interesting Article Title",
      "author": "Author Name",
      "publication": "Publication Name",
      "excerpt": "Article excerpt or description...",
      "domain": "medium.com",
      "platform": "medium",
      "read_status": "unread"
    },
    ...
  ]
}
```

Note: `start_date`, `end_date`, and `timezone` fields are only included when date filtering is used.

### Twitter Bookmarks Output Format

The tool outputs Twitter bookmarks in the following JSON format:

```json
{
  "fetched_at": "2025-12-28T10:30:00Z",
  "total_items": 25,
  "delta_added": 5,
  "items": [
    {
      "tweet_id": "1234567890123456789",
      "url": "https://x.com/username/status/1234567890123456789",
      "text": "This is the tweet content...",
      "author_name": "Display Name",
      "author_handle": "username",
      "created_at": "2025-12-15T14:30:00Z",
      "saved_at": "2025-12-15T14:30:00Z",
      "expanded_urls": {
        "https://t.co/abc123": "https://github.com/user/repo",
        "https://t.co/xyz789": "https://example.com/article"
      }
    },
    ...
  ],
  "source": "twitter",
  "description": "Twitter/X bookmarks snapshot"
}
```

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

### Tabs Fields

- **browser**: Browser name (chrome, vivaldi, edge, brave)
- **total_tabs**: Number of open tabs
- **total_windows**: Number of browser windows
- **entries**: Array of tab entries, each containing:
  - **url**: Current URL of the tab
  - **title**: Page title
  - **domain**: Extracted domain name
  - **active**: Whether this is the active tab in the window
  - **group**: Tab group name (if grouped, Chromium feature)
  - **window_id**: Window identifier
  - **browser**: Browser source

### Reading List Fields

- **platform**: Platform name (medium, substack, or "all")
- **start_date**: Filter period start (ISO 8601 UTC format, only when date filtering is used)
- **end_date**: Filter period end (ISO 8601 UTC format, only when date filtering is used)
- **timezone**: Timezone used for date interpretation (only when date filtering is used)
- **total_entries**: Number of saved articles in the report
- **entries**: Array of reading list entries, each containing:
  - **saved_at**: When article was saved (ISO 8601 UTC format)
  - **url**: Full URL of the article
  - **title**: Article title
  - **author**: Article author (optional)
  - **publication**: Publication name (optional)
  - **excerpt**: Article excerpt or description (optional)
  - **domain**: Extracted domain name
  - **platform**: Source platform (medium, substack, etc.)
  - **read_status**: Read status (unread, read, archived - optional)

### Twitter Bookmarks Fields

- **fetched_at**: When the bookmarks were fetched (ISO 8601 UTC format)
- **total_items**: Total number of bookmarks in the report
- **delta_added**: Number of new bookmarks added since last fetch
- **source**: Always "twitter"
- **description**: Human-readable description
- **items**: Array of bookmark entries, each containing:
  - **tweet_id**: Unique tweet identifier
  - **url**: Full URL to the tweet
  - **text**: Tweet content
  - **author_name**: Author's display name
  - **author_handle**: Author's Twitter handle (without @)
  - **created_at**: When the tweet was created (ISO 8601 UTC format)
  - **saved_at**: Approximate time when bookmarked (ISO 8601 UTC format)
  - **expanded_urls**: Map of t.co shortened URLs to their expanded destinations (optional)

## LLM Usage

The JSON output is designed to be easily consumed by language models.

### Using with Claude Code CLI

If you have [Claude Code](https://github.com/anthropics/claude-code) installed:

```bash
# Save output to files and start Claude Code with them
web-recap tabs -o tabs.json
web-recap bookmarks --start-date 2025-12-20 -o recent-bookmarks.json
web-recap --date 2025-12-27 -o today-history.json

# Start Claude Code in the current directory
claude-code

# In the chat, ask Claude to analyze the files:
# "Analyze my tabs.json and tell me what topics I'm researching"
# "Based on today-history.json, summarize my browsing activity for today"
# "Review recent-bookmarks.json and categorize them"
```

### Direct Analysis with Saved Files

```bash
# Save history to file
web-recap --browser chrome --date 2025-12-15 -o history.json

# Save bookmarks to file
web-recap bookmarks --all-browsers -o bookmarks.json

# Save open tabs to file
web-recap tabs -o tabs.json

# Use with Claude.ai: Upload the JSON files when starting a conversation
# Use with Anthropic API: Include the content in your messages

# Or pipe to other tools for processing
web-recap tabs | jq '.entries[] | select(.active==true) | .url'

# Combine multiple sources for comprehensive analysis
web-recap --date 2025-12-27 -o today-history.json
web-recap bookmarks --date 2025-12-27 -o today-bookmarks.json
web-recap tabs -o current-tabs.json
# Then reference all three files in your LLM conversation
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

**Sessions (Open Tabs):**
- Chrome: `~/.config/google-chrome/Default/Sessions/`
- Chromium: `~/.config/chromium/Default/Sessions/`
- Edge: `~/.config/microsoft-edge/Default/Sessions/`
- Brave: `~/.config/BraveSoftware/Brave-Browser/Default/Sessions/`
- Vivaldi: `~/.config/vivaldi/Default/Sessions/`

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

**Sessions (Open Tabs):**
- Chrome: `~/Library/Application Support/Google/Chrome/Default/Sessions/`
- Chromium: `~/Library/Application Support/Chromium/Default/Sessions/`
- Edge: `~/Library/Application Support/Microsoft Edge/Default/Sessions/`
- Brave: `~/Library/Application Support/BraveSoftware/Brave-Browser/Default/Sessions/`
- Vivaldi: `~/Library/Application Support/Vivaldi/Default/Sessions/`

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

**Sessions (Open Tabs):**
- Chrome: `%LOCALAPPDATA%\Google\Chrome\User Data\Default\Sessions\`
- Chromium: `%LOCALAPPDATA%\Chromium\User Data\Default\Sessions\`
- Edge: `%LOCALAPPDATA%\Microsoft\Edge\User Data\Default\Sessions\`
- Brave: `%LOCALAPPDATA%\BraveSoftware\Brave-Browser\User Data\Default\Sessions\`
- Vivaldi: `%LOCALAPPDATA%\Vivaldi\User Data\Default\Sessions\`

## Technical Details

### Database Locking
The tool automatically handles browser database locking by copying the database to a temporary file before reading it. This allows you to extract history and bookmarks while your browser is running.

### Bookmark Formats
Different browsers use different formats for storing bookmarks:
- **Chrome/Chromium/Edge/Brave/Vivaldi**: JSON file format with hierarchical folder structure
- **Firefox**: SQLite database (places.sqlite) with bookmarks in `moz_bookmarks` table, supports tags
- **Safari**: Property list (plist) format

### Session Files (Open Tabs)
Chromium-based browsers store session data in SNSS (Session Storage) binary format:
- Session files are located in the `Sessions/` directory
- Files named `Session_*` or `Tabs_*` contain the current session state
- The parser reads the most recently modified session file
- Tab groups and active tab state are preserved

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
LMs.
tions with Claude and other LLMs.
r LLMs.
LMs.
tions with Claude and other LLMs.
