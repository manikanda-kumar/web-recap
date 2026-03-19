# Reading List Integration Guide

This guide explains how to extract reading lists and saved articles from Medium and Substack using web-recap.

## Table of Contents

- [Quick Start](#quick-start)
- [Method 1: File Export (Recommended)](#method-1-file-export-recommended)
- [Method 2: Web Scraping](#method-2-web-scraping)
- [Troubleshooting](#troubleshooting)
- [Future Integrations](#future-integrations)

## Quick Start

**Easiest approach - File export:**

1. Use browser console scripts to export your data
2. Parse with web-recap

```bash
# Medium
web-recap reading-list --platform medium --file medium-reading-list-2025-12-27.csv

# Substack
web-recap reading-list --platform substack --file substack-saves-2025-12-27.json
```

## Method 1: File Export (Recommended)

File export is more reliable than web scraping because:
- ✅ Platform HTML structure changes don't break it
- ✅ No authentication issues
- ✅ Works offline once exported
- ✅ Can be versioned and backed up

### Medium Export

#### Step 1: Run Export Script

1. **Go to your Medium reading list**: https://medium.com/me/list/reading-list
2. **Make sure you're logged in**
3. **Open DevTools**:
   - Press `F12` (Windows/Linux)
   - Press `Cmd+Option+I` (Mac)
4. **Go to Console tab**
5. **Copy and paste** the entire contents of `scripts/export-medium.js`
6. **Press Enter**
7. **Wait** for the script to scroll and load all articles
8. **Download** will start automatically

The script will:
- Scroll through your entire reading list
- Extract: title, URL, author, publication, excerpt, saved date
- Download a CSV file named `medium-reading-list-YYYY-MM-DD.csv`

#### Step 2: Parse with web-recap

```bash
# Parse the exported CSV
web-recap reading-list --platform medium --file medium-reading-list-2025-12-27.csv

# With date filtering
web-recap reading-list --platform medium --file medium-reading-list-2025-12-27.csv \
  --start-date 2025-01-01 --end-date 2025-12-31

# Save to JSON
web-recap reading-list --platform medium --file medium-reading-list-2025-12-27.csv \
  -o medium-parsed.json
```

### Substack Export

#### Step 1: Run Export Script

1. **Go to Substack inbox**: https://substack.com/inbox
2. **Click "Saved" tab** to view saved posts
3. **Open DevTools** (F12 or Cmd+Option+I)
4. **Go to Console tab**
5. **Copy and paste** the entire contents of `scripts/export-substack.js`
6. **Press Enter**
7. **Wait** for the script to scroll and load all saved posts
8. **Download** will start automatically

The script will:
- Scroll through your saved posts
- Extract: title, URL, author, publication, excerpt, saved date
- Download a JSON file named `substack-saves-YYYY-MM-DD.json`

#### Step 2: Parse with web-recap

```bash
# Parse the exported JSON
web-recap reading-list --platform substack --file substack-saves-2025-12-27.json

# With date filtering
web-recap reading-list --platform substack --file substack-saves-2025-12-27.json \
  --start-date 2025-01-01

# Save to JSON
web-recap reading-list --platform substack --file substack-saves-2025-12-27.json \
  -o substack-parsed.json
```

### CSV Format (Medium)

The exported CSV should have these columns:

```csv
title,url,author,publication,excerpt,saved_at
"Article Title","https://medium.com/@author/article","Author Name","Publication","Excerpt...","2025-12-27T10:30:00Z"
```

### JSON Format (Substack)

The exported JSON should have this structure:

```json
{
  "saved_posts": [
    {
      "title": "Article Title",
      "url": "https://example.substack.com/p/article",
      "author": "Author Name",
      "publication": "Publication Name",
      "excerpt": "Article excerpt...",
      "saved_at": "2025-12-27T10:30:00Z"
    }
  ],
  "export_date": "2025-12-27T15:00:00Z",
  "total_count": 42
}
```

## Method 2: Web Scraping

Web scraping requires authentication and may break if platforms change their HTML structure.

### Getting Cookies

#### Medium Cookies

1. **Open Medium** and log in
2. **Go to reading list**: https://medium.com/me/list/reading-list
3. **Open DevTools** (F12)
4. **Application tab** → **Cookies** → **https://medium.com**
5. **Find and copy** the `sid` cookie value

```bash
# Set environment variable
export MEDIUM_COOKIE="sid=YOUR_SID_VALUE_HERE"

# Or pass directly
web-recap reading-list --platform medium --cookie "sid=YOUR_VALUE"
```

#### Substack Cookies

1. **Open Substack** and log in
2. **Go to inbox**: https://substack.com/inbox
3. **Open DevTools** (F12) → **Network tab**
4. **Refresh page** (F5)
5. **Click any API request** → **Headers** tab
6. **Find Cookie header** in Request Headers

```bash
# Set environment variable
export SUBSTACK_COOKIE="substack.sid=VALUE; substack.lli=VALUE"

# Or pass directly
web-recap reading-list --platform substack --cookie "substack.sid=VALUE; substack.lli=VALUE"
```

### Using Web Scraping

```bash
# Medium with cookie
export MEDIUM_COOKIE="sid=YOUR_VALUE"
web-recap reading-list --platform medium

# Substack with cookie
export SUBSTACK_COOKIE="substack.sid=VALUE; substack.lli=VALUE"
web-recap reading-list --platform substack

# All platforms
web-recap reading-list --all-platforms
```

### Cookie Expiration

⚠️ **Important:**
- Cookies expire after weeks or months
- You'll need to re-extract them periodically
- Never commit cookies to version control
- Use environment variables or `.env` files (add to `.gitignore`)

## Troubleshooting

### Export Script Issues

**"No articles/posts found"**
- Make sure you're logged in
- Verify you're on the correct page (reading list or saved posts)
- Try scrolling manually first to trigger loading
- Check browser console for errors

**Script doesn't scroll**
- Try running it again
- Manually scroll a bit first, then run the script
- Some browser extensions may interfere - try in incognito mode

**Missing fields in export**
- Platforms change their HTML frequently
- The scripts use multiple selectors as fallbacks
- Empty fields are okay - web-recap handles them

### Web Scraping Issues

**"Authentication failed"**
- Cookie expired - get a fresh one
- Wrong cookie format - make sure to include cookie name (e.g., `sid=...`)
- Platform changed authentication - switch to file export

**"No entries found"**
- HTML structure changed - web scraping is fragile
- Cookie doesn't have proper permissions
- **Solution**: Use file export method instead

**Rate limiting**
- Web scraping has built-in 2-second delays
- If you see 429 errors, wait a few minutes
- Consider using file export to avoid rate limits

## Future Integrations

The hybrid architecture supports adding more platforms:

### Planned Integrations
- **Readwise Reader** (API available, $7.99/month)
- **Raindrop.io** (API available, free tier)
- **Pocket** (shutdown July 2025 ❌)
- **Instapaper** (API has 500-item limit ⚠️)

### Adding New Platforms

To add a new platform, create handlers following the existing pattern:

```
internal/readinglist/
├── newplatform.go           # Platform handler
├── newplatform_web.go       # Web scraping strategy
└── newplatform_file.go      # File parsing strategy
```

See the implementation plan in `.claude/plans/` for details.

## Best Practices

1. **Start with file export** - It's more reliable
2. **Export regularly** - Don't lose your reading list
3. **Version your exports** - Keep dated backups
4. **Combine with LLMs**:
   ```bash
   web-recap reading-list --all-platforms -o reading-list.json
   # Then analyze with Claude, ChatGPT, etc.
   ```
5. **Use date filters** to focus on recent saves:
   ```bash
   web-recap reading-list --platform medium --file export.csv --start-date 2025-12-01
   ```

## Security Notes

- 🔒 Never commit cookies or tokens to git
- 🔒 Store credentials in environment variables
- 🔒 Add `.env` files to `.gitignore`
- 🔒 Cookies expire - don't hardcode them
- 🔒 File exports don't contain credentials - safe to backup

## Support

Having issues? Check:
1. This guide's troubleshooting section
2. Main README.md for general web-recap help
3. GitHub issues: https://github.com/manikanda-kumar/web-recap/issues
