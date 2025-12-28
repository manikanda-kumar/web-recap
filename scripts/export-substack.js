/**
 * Substack Saved Posts Exporter
 *
 * Instructions:
 * 1. Go to https://substack.com/inbox (make sure you're logged in)
 * 2. Click on "Saved" tab to view your saved posts
 * 3. Open DevTools (F12 or Cmd+Option+I)
 * 4. Go to the Console tab
 * 5. Paste this entire script and press Enter
 * 6. Wait for the export to complete
 * 7. A JSON file will be downloaded automatically
 *
 * The script will scroll through your saved posts to load them all.
 */

(async function exportSubstackSaves() {
  console.log('🚀 Starting Substack Saved Posts Export...');

  // Array to store all saved posts
  const savedPosts = [];

  // Function to extract saved posts from the current page
  function extractPosts() {
    // Try multiple selectors as Substack's DOM structure may vary
    const postElements = document.querySelectorAll('[class*="post"], [class*="Post"], article, [data-testid*="post"]');

    postElements.forEach(element => {
      // Look for the post link and title
      const linkElement = element.querySelector('a[href*="/p/"], a h2, a h3');
      const titleElement = element.querySelector('h2, h3, [class*="title"]');
      const authorElement = element.querySelector('[class*="author"], [class*="byline"]');
      const publicationElement = element.querySelector('[class*="publication"]');
      const excerptElement = element.querySelector('[class*="subtitle"], [class*="description"], p');
      const timeElement = element.querySelector('time');

      if (linkElement && titleElement) {
        const url = linkElement.href || linkElement.closest('a')?.href || '';
        const title = titleElement.textContent.trim();

        const post = {
          title: title,
          url: url,
          author: authorElement ? authorElement.textContent.trim() : '',
          publication: publicationElement ? publicationElement.textContent.trim() : '',
          excerpt: excerptElement ? excerptElement.textContent.trim().substring(0, 200) : '',
          saved_at: timeElement ? timeElement.getAttribute('datetime') || new Date().toISOString() : new Date().toISOString()
        };

        // Only add if not already in the list and has valid data
        if (url && !savedPosts.find(p => p.url === url)) {
          savedPosts.push(post);
        }
      }
    });
  }

  // Function to scroll to bottom of page
  function scrollToBottom() {
    window.scrollTo(0, document.body.scrollHeight);
  }

  // Get initial height
  let lastHeight = document.body.scrollHeight;
  let scrollAttempts = 0;
  const maxScrollAttempts = 30; // Maximum number of scroll attempts

  console.log('📜 Scrolling to load all saved posts...');

  // Scroll and load more posts
  while (scrollAttempts < maxScrollAttempts) {
    extractPosts();
    scrollToBottom();

    // Wait for content to load
    await new Promise(resolve => setTimeout(resolve, 2000));

    const newHeight = document.body.scrollHeight;

    if (newHeight === lastHeight) {
      // No more content loaded, try a few more times to be sure
      scrollAttempts++;
    } else {
      scrollAttempts = 0; // Reset counter if we loaded more content
      lastHeight = newHeight;
    }

    console.log(`Found ${savedPosts.length} posts so far...`);
  }

  // Final extraction
  extractPosts();

  console.log(`✅ Found ${savedPosts.length} total saved posts`);

  if (savedPosts.length === 0) {
    console.error('❌ No saved posts found!');
    console.error('Make sure you:');
    console.error('  1. Are logged into Substack');
    console.error('  2. Are on the "Saved" tab in your inbox');
    console.error('  3. Have some saved posts');
    return;
  }

  // Create JSON structure matching web-recap format
  const exportData = {
    saved_posts: savedPosts,
    export_date: new Date().toISOString(),
    total_count: savedPosts.length
  };

  // Convert to JSON
  const jsonContent = JSON.stringify(exportData, null, 2);

  // Download JSON
  const blob = new Blob([jsonContent], { type: 'application/json;charset=utf-8;' });
  const link = document.createElement('a');
  const url = URL.createObjectURL(blob);

  const timestamp = new Date().toISOString().split('T')[0];
  link.setAttribute('href', url);
  link.setAttribute('download', `substack-saves-${timestamp}.json`);
  link.style.visibility = 'hidden';

  document.body.appendChild(link);
  link.click();
  document.body.removeChild(link);

  console.log(`💾 Downloaded substack-saves-${timestamp}.json`);
  console.log('🎉 Export complete!');

  // Also log a sample for verification
  console.log('\nFirst 3 posts:');
  console.table(savedPosts.slice(0, 3));
})();
