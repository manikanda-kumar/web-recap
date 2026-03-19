/**
 * YouTube Watch Later Exporter
 *
 * Instructions:
 * 1. Go to https://www.youtube.com/playlist?list=WL (make sure you're logged in)
 * 2. Open DevTools (F12 or Cmd+Option+I)
 * 3. Go to the Console tab
 * 4. Paste this entire script and press Enter
 * 5. Wait for the export to complete (it scrolls to load all videos)
 * 6. A JSON file will be downloaded automatically
 *
 * Then use with web-recap:
 *   web-recap youtube-copy-playlist --client-secret data/youtube_client.json --data data/watch_later_full.json
 */

(async function exportYouTubeWatchLater() {
  console.log('🚀 Starting YouTube Watch Later Export...');

  const videos = [];
  const seenIds = new Set();

  function extractVideos() {
    // Each video row in the playlist
    const rows = document.querySelectorAll('ytd-playlist-video-renderer');

    rows.forEach(row => {
      const linkEl = row.querySelector('a#video-title');
      if (!linkEl) return;

      const href = linkEl.getAttribute('href') || '';
      const match = href.match(/[?&]v=([^&]+)/);
      if (!match) return;

      const videoId = match[1];
      if (seenIds.has(videoId)) return;
      seenIds.add(videoId);

      const title = (linkEl.textContent || '').trim();

      // Channel name
      const channelEl = row.querySelector('ytd-channel-name a, ytd-channel-name yt-formatted-string');
      const channelTitle = channelEl ? channelEl.textContent.trim() : '';

      // Thumbnail timestamp can hint at video length but not added_at
      // YouTube playlist page doesn't expose added_at in the DOM

      videos.push({
        playlist_item_id: '',
        video_id: videoId,
        url: 'https://www.youtube.com/watch?v=' + videoId,
        title: title,
        channel_title: channelTitle,
        added_at: new Date().toISOString()  // best-effort; real date not in DOM
      });
    });
  }

  // Scroll to load all videos
  const scrollContainer = document.scrollingElement || document.documentElement;
  let lastHeight = scrollContainer.scrollHeight;
  let staleCount = 0;
  const maxStale = 8;

  console.log('📜 Scrolling to load all videos...');

  while (staleCount < maxStale) {
    extractVideos();
    scrollContainer.scrollTop = scrollContainer.scrollHeight;

    await new Promise(r => setTimeout(r, 1500));

    const newHeight = scrollContainer.scrollHeight;
    if (newHeight === lastHeight) {
      staleCount++;
    } else {
      staleCount = 0;
      lastHeight = newHeight;
    }

    console.log(`Found ${videos.length} videos so far...`);
  }

  // Final pass
  extractVideos();

  console.log(`✅ Found ${videos.length} total videos`);

  if (videos.length === 0) {
    console.error('❌ No videos found! Make sure you\'re on https://www.youtube.com/playlist?list=WL and logged in.');
    return;
  }

  // Build report in the same format as watch_later.json
  const report = {
    fetched_at: new Date().toISOString(),
    playlist_id: 'WL',
    total_items: videos.length,
    delta_added: videos.length,
    items: videos,
    source: 'youtube',
    description: 'YouTube Watch Later playlist (browser export)'
  };

  const jsonContent = JSON.stringify(report, null, 2);

  // Download JSON
  const blob = new Blob([jsonContent], { type: 'application/json' });
  const link = document.createElement('a');
  const url = URL.createObjectURL(blob);

  const timestamp = new Date().toISOString().split('T')[0];
  link.setAttribute('href', url);
  link.setAttribute('download', `watch-later-${timestamp}.json`);
  link.style.visibility = 'hidden';

  document.body.appendChild(link);
  link.click();
  document.body.removeChild(link);

  console.log(`💾 Downloaded watch-later-${timestamp}.json`);
  console.log('🎉 Export complete!');

  console.log('\nFirst 5 videos:');
  console.table(videos.slice(0, 5).map(v => ({ title: v.title, channel: v.channel_title, url: v.url })));
})();
