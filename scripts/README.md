# Export Scripts

This directory contains browser console scripts to export your reading lists from Medium and Substack.

## Quick Start

### Medium Export (Your Own Reading List)

1. Open https://medium.com/me/list/reading-list
2. Open DevTools (F12)
3. Go to Console tab
4. Copy/paste contents of `export-medium.js`
5. Press Enter
6. Wait for download (CSV file)
7. Use with web-recap:
   ```bash
   web-recap reading-list --platform medium --file medium-reading-list-2025-12-27.csv
   ```

### Medium Public Reading List Export

For public reading lists (e.g., `https://medium.com/@username/list/reading-list`):

1. Open the public reading list URL in your browser
2. Open DevTools (F12)
3. Go to Console tab
4. Copy/paste contents of `export-medium-public.js`
5. Press Enter
6. Wait for download (JSON file)
7. Use with web-recap:
   ```bash
   web-recap reading-list --platform medium --file medium-reading-list-2025-12-28.json
   ```

### Substack Export

1. Open https://substack.com/inbox
2. Click "Saved" tab
3. Open DevTools (F12)
4. Go to Console tab
5. Copy/paste contents of `export-substack.js`
6. Press Enter
7. Wait for download (JSON file)
8. Use with web-recap:
   ```bash
   web-recap reading-list --platform substack --file substack-saves-2025-12-27.json
   ```

### YouTube Watch Later Export

1. Open https://www.youtube.com/playlist?list=WL
2. Open DevTools (F12)
3. Go to Console tab
4. Copy/paste contents of `export-youtube-watch-later.js`
5. Press Enter
6. Wait for download (JSON file)
7. Copy videos to a public playlist:
   ```bash
   web-recap youtube-copy-playlist --client-secret data/youtube/client.json --data data/youtube/watch-later-2025-12-28.json
   ```

## Files

- **export-medium.js** - Exports your own Medium reading list to CSV
- **export-medium-public.js** - Exports any public Medium reading list to JSON
- **export-substack.js** - Exports Substack saved posts to JSON
- **export-youtube-watch-later.js** - Exports YouTube Watch Later playlist to JSON
- **README.md** - This file

## How It Works

The scripts:
1. Scroll through your reading list/saved posts
2. Extract article data (title, URL, author, etc.)
3. Format as CSV (Medium) or JSON (Substack)
4. Trigger browser download

## Troubleshooting

**No download?**
- Check browser's download settings
- Look in DevTools Console for errors
- Make sure you're logged in

**Missing articles?**
- Manually scroll first to load content
- Try running the script again
- Increase scroll wait time in the script

**Empty fields?**
- Normal - platforms don't always have all data
- web-recap handles missing fields gracefully

## Security

These scripts:
- ✅ Run entirely in your browser (client-side)
- ✅ Don't send data anywhere
- ✅ Don't require authentication credentials
- ✅ Are open source - review the code!

## For More Help

See [READING_LIST.md](../docs/READING_LIST.md) for complete documentation.
