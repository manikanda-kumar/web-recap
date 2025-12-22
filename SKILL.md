---
name: web-recap
description: Find articles from browser history or bookmarks by topic, date, or domain. Use when user asks about saved articles, browsing history, bookmarks, or wants to find URLs they've visited or saved.
---

# web-recap

Extracts browser history and bookmarks from Chrome, Chromium, Brave, Vivaldi, Firefox, Safari, Edge. Run `web-recap --help` for all flags.

## Commands

```
web-recap              Extract history (default)
web-recap bookmarks    Extract bookmarks
web-recap list         Show detected browsers
```

## Key Flags

```
--date YYYY-MM-DD        Specific date (local timezone)
--start-date YYYY-MM-DD  Start of range (inclusive)
--end-date YYYY-MM-DD    End of range (inclusive)
--time HH                Specific hour (e.g., --time 14 for 2pm-3pm)
--browser NAME           chrome|firefox|safari|edge|brave|vivaldi|auto
--all-browsers           Extract from all detected browsers
```

## Output Format

JSON with `entries` array containing:
- **History**: `timestamp`, `url`, `title`, `domain`, `visit_count`, `browser`
- **Bookmarks**: `date_added`, `url`, `title`, `folder`, `domain`, `browser`, `tags` (Firefox only)

## Usage Patterns

**Never dump raw output.** Always use jq to filter and reduce tokens.

### Find Articles by Topic

```bash
# Search bookmarks for a keyword
web-recap bookmarks | jq '[.entries[] | select(.title + .url + .folder | test("KEYWORD"; "i"))] | unique_by(.url) | map({title, url, folder})'

# Search history for a keyword
web-recap --start-date 2025-12-01 --end-date 2025-12-22 | jq '[.entries[] | select(.title + .url | test("KEYWORD"; "i"))] | unique_by(.url) | map({title, url, domain, visit_count})'

# Search across all browsers
web-recap bookmarks --all-browsers | jq '[.entries[] | select(.title + .url | test("KEYWORD"; "i"))] | map({title, url, browser, folder}) | unique_by(.url)'
```

### Find Recent Bookmarks

```bash
# Bookmarks from last 7 days
web-recap bookmarks --start-date 2025-12-15 --end-date 2025-12-22 | jq '.entries | map({title, url, date_added, folder})'

# Bookmarks from specific date
web-recap bookmarks --date 2025-12-15 | jq '.entries | map({title, url, folder})'
```

### Browse by Domain

```bash
# Find all bookmarks from a domain
web-recap bookmarks | jq '[.entries[] | select(.domain == "github.com")] | map({title, url, folder})'

# Most visited domains in history
web-recap | jq '[.entries[].domain] | group_by(.) | map({domain: .[0], count: length}) | sort_by(-.count) | .[0:10]'
```

### Browse by Folder

```bash
# List bookmark folders
web-recap bookmarks | jq '[.entries[].folder] | unique | sort'

# Get bookmarks from specific folder
web-recap bookmarks | jq '[.entries[] | select(.folder | contains("Tech"))] | map({title, url})'
```

### Quick Metadata

```bash
# Get summary
web-recap bookmarks | jq '{browser, total_entries, date_range: (.start_date // "all time")}'
```

## Examples

```bash
# Find AI-related bookmarks from last month
web-recap bookmarks --start-date 2025-11-22 --end-date 2025-12-22 | jq '[.entries[] | select(.title + .url | test("ai|claude|gpt|llm"; "i"))] | map({title, url, folder})'

# Find most visited URLs this week
web-recap --start-date 2025-12-16 --end-date 2025-12-22 | jq '[.entries[]] | sort_by(-.visit_count) | .[0:10] | map({title, url, visit_count})'

# Export Chrome bookmarks to review
web-recap bookmarks --browser chrome -o chrome-bookmarks.json

# Find Firefox bookmarks with specific tag
web-recap bookmarks --browser firefox | jq '[.entries[] | select(.tags | contains(["golang"]))] | map({title, url, tags})'
```
