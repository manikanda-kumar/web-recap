# Quick Start: Reading Lists

Get your Medium and Substack reading lists in 5 minutes.

## Method 1: File Export (Easiest)

### Medium

```bash
# 1. Export from browser
#    - Go to https://medium.com/me/list/reading-list
#    - Open DevTools (F12) → Console
#    - Copy/paste scripts/export-medium.js
#    - Press Enter, wait for download

# 2. Parse with web-recap
web-recap reading-list --platform medium --file medium-reading-list-2025-12-27.csv

# 3. (Optional) Save to file
web-recap reading-list --platform medium --file medium-reading-list-2025-12-27.csv -o output.json
```

### Substack

```bash
# 1. Export from browser
#    - Go to https://substack.com/inbox → "Saved" tab
#    - Open DevTools (F12) → Console
#    - Copy/paste scripts/export-substack.js
#    - Press Enter, wait for download

# 2. Parse with web-recap
web-recap reading-list --platform substack --file substack-saves-2025-12-27.json

# 3. (Optional) Save to file
web-recap reading-list --platform substack --file substack-saves-2025-12-27.json -o output.json
```

## Method 2: Web Scraping (Advanced)

### Get Cookies

**Medium:**
```bash
# DevTools → Application → Cookies → medium.com → copy 'sid'
export MEDIUM_COOKIE="sid=YOUR_VALUE_HERE"
```

**Substack:**
```bash
# DevTools → Network → Refresh page → Click API request → Copy Cookie header
export SUBSTACK_COOKIE="substack.sid=VALUE; substack.lli=VALUE"
```

### Use Web Scraping

```bash
# Medium
web-recap reading-list --platform medium

# Substack
web-recap reading-list --platform substack

# Both
web-recap reading-list --all-platforms
```

## Common Commands

```bash
# Filter by date range
web-recap reading-list --platform medium --file export.csv \
  --start-date 2025-01-01 --end-date 2025-12-31

# All platforms
web-recap reading-list --all-platforms -o all-reading-lists.json

# Recent saves only (last 30 days)
web-recap reading-list --platform medium --file export.csv \
  --start-date $(date -d '30 days ago' +%Y-%m-%d)
```

## Output Example

```json
{
  "platform": "medium",
  "total_entries": 42,
  "entries": [
    {
      "saved_at": "2025-12-27T10:30:00Z",
      "url": "https://medium.com/@author/article",
      "title": "Article Title",
      "author": "Author Name",
      "publication": "Publication",
      "excerpt": "Article preview...",
      "domain": "medium.com",
      "platform": "medium"
    }
  ]
}
```

## Use with LLMs

```bash
# Export reading list
web-recap reading-list --all-platforms -o reading-list.json

# Analyze with Claude Code
claude-code
# Then: "Analyze reading-list.json and summarize my interests"

# Or upload to Claude.ai / ChatGPT
```

## Troubleshooting

| Problem | Solution |
|---------|----------|
| Export script doesn't work | Make sure you're logged in and on correct page |
| No articles found | Manually scroll first, then run script again |
| Cookie authentication fails | Cookie expired - get a fresh one |
| Web scraping broken | Platform changed HTML - use file export instead |

## Full Documentation

- **Complete guide**: [READING_LIST.md](READING_LIST.md)
- **Export scripts**: [scripts/README.md](scripts/README.md)
- **Main docs**: [README.md](README.md)

## Security Reminder

- 🔒 Never commit cookies to git
- 🔒 Cookies expire (weeks/months)
- 🔒 File exports are safe to backup
- 🔒 Export scripts run client-side only
