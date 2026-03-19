# Bookmark Extraction Feature

This document summarizes the bookmark extraction feature added to web-recap.

## Overview

The web-recap tool now supports extracting bookmarks in addition to browser history. The implementation maintains the original history functionality while adding comprehensive bookmark support across all major browsers.

## New Features

### 1. Bookmark Extraction Command
```bash
web-recap bookmarks [flags]
```

Supports:
- Single browser extraction: `--browser chrome`
- All browsers: `--all-browsers`
- Custom paths: `--db-path /path/to/Bookmarks`
- Output to file: `-o bookmarks.json`

### 2. Multi-Browser Support

#### Chrome/Chromium/Edge/Brave
- Format: JSON file
- Location: `Default/Bookmarks`
- Features: Hierarchical folder structure, date added/modified

#### Firefox
- Format: SQLite database (places.sqlite)
- Location: Firefox profile directory
- Features: Hierarchical folders, tags support

#### Safari
- Format: Property list (plist)
- Location: `~/Library/Safari/Bookmarks.plist`
- Features: Folder structure, reading list items

### 3. JSON Output Format

```json
{
  "browser": "chrome",
  "total_entries": 150,
  "entries": [
    {
      "date_added": "2025-12-15T10:30:00Z",
      "date_modified": "2025-12-16T14:20:00Z",
      "url": "https://example.com",
      "title": "Example",
      "folder": "Bookmarks Bar/Tech",
      "domain": "example.com",
      "browser": "chrome",
      "tags": ["golang", "tutorial"]
    }
  ]
}
```

## Implementation Details

### New Files Created

1. **internal/models/bookmark.go**
   - `BookmarkEntry` struct
   - `BookmarkReport` struct
   - `BookmarkFolder` and `BookmarkNode` for hierarchical structures

2. **internal/database/bookmark_chrome.go**
   - ChromeBookmarkHandler for JSON-based bookmarks
   - Recursive folder traversal
   - Chrome timestamp conversion

3. **internal/database/bookmark_firefox.go**
   - FirefoxBookmarkHandler for SQLite-based bookmarks
   - Folder path reconstruction
   - Tag extraction from moz_bookmarks table

4. **internal/database/bookmark_safari.go**
   - SafariBookmarkHandler for plist-based bookmarks
   - Property list parsing using howett.net/plist
   - Reading list support

5. **internal/database/bookmark_query.go**
   - BookmarkQuerier interface
   - Factory pattern for creating browser-specific handlers
   - Multi-browser query support

### Modified Files

1. **internal/browser/paths.go**
   - Added `GetBookmarkPath()` function
   - Platform-specific bookmark path resolution
   - Support for all browsers across Linux, macOS, Windows

2. **internal/output/json.go**
   - Added `FormatBookmarksJSON()`
   - Added `FormatBookmarksJSONCompact()`
   - Added `FormatBookmarksJSONLines()`

3. **cmd/web-recap/main.go**
   - Added `bookmarksCmd` subcommand
   - Added `runBookmarks()` handler
   - Integrated with existing flag system

4. **README.md**
   - Added bookmark extraction documentation
   - Updated usage examples
   - Added bookmark output format specification
   - Updated database locations for bookmarks
   - Added technical details about bookmark formats

5. **go.mod**
   - Added dependency: `howett.net/plist v1.0.1` for Safari bookmark parsing

## Architecture Decisions

### 1. Separation of Concerns
- Bookmark extraction is completely separate from history extraction
- Each browser has its own bookmark handler
- Common interface (`BookmarkQuerier`) for all browsers

### 2. Format-Specific Handling
- Chrome/Brave/Edge: Direct JSON parsing (no database locking needed)
- Firefox: SQLite with database copying for lock handling
- Safari: Plist parsing for macOS-specific format

### 3. Folder Path Preservation
- Full folder paths stored as strings (e.g., "Bookmarks Bar/Tech/Go")
- Hierarchical structure maintained from source
- Root folders filtered appropriately

### 4. Tag Support
- Firefox tags extracted from special "tags" folder
- Other browsers don't support tags natively
- Tags array is optional and only populated for Firefox

## Testing Checklist

- [x] Build succeeds without errors
- [x] CLI help shows bookmark command
- [x] Command accepts all standard flags
- [ ] Chrome bookmark extraction works
- [ ] Firefox bookmark extraction works
- [ ] Safari bookmark extraction works (macOS only)
- [ ] All-browsers mode works
- [ ] Folder paths are correctly preserved
- [ ] Firefox tags are extracted
- [ ] JSON output is valid
- [ ] Browser detection works correctly

## Usage Examples

### Extract from default browser
```bash
./web-recap bookmarks
```

### Extract from Chrome
```bash
./web-recap bookmarks --browser chrome
```

### Extract from all browsers to file
```bash
./web-recap bookmarks --all-browsers -o my-bookmarks.json
```

### Use with LLM
```bash
./web-recap bookmarks --all-browsers | claude --prompt "Organize these bookmarks by category"
```

## Future Enhancements

Potential improvements:
1. Bookmark search/filter by folder
2. Bookmark search by date range
3. Export to different formats (HTML, Markdown)
4. Bookmark de-duplication across browsers
5. Bookmark validation (check for dead links)
6. Import bookmarks from one browser to another
7. Bookmark organization suggestions using LLMs

## Compatibility

- **Go Version**: 1.21+
- **Platforms**: Linux, macOS, Windows
- **Browsers**: Chrome, Chromium, Brave, Edge, Firefox, Safari
- **No Breaking Changes**: Original history extraction functionality unchanged
