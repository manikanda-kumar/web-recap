/**
 * Medium Reading List Exporter
 *
 * Instructions:
 * 1. Go to https://medium.com/me/list/reading-list (make sure you're logged in)
 * 2. Open DevTools (F12 or Cmd+Option+I)
 * 3. Go to the Console tab
 * 4. Paste this entire script and press Enter
 * 5. Wait for the export to complete
 * 6. A CSV file will be downloaded automatically
 *
 * The script will scroll through your entire reading list to load all articles.
 */

(async function exportMediumReadingList() {
  console.log('🚀 Starting Medium Reading List Export...');

  // Array to store all articles
  const articles = [];

  // Function to extract articles from the current page
  function extractArticles() {
    const articleElements = document.querySelectorAll('article, div[data-post-id], [data-testid="storyPreview"]');

    articleElements.forEach(element => {
      // Try different selectors to find the article link
      const linkElement = element.querySelector('a[data-testid="post-preview-title"], a h2, a h3, a[aria-label*="Post"]');
      const titleElement = element.querySelector('h2, h3, [data-testid="storyTitle"]');
      const authorElement = element.querySelector('[data-testid="authorName"], a[rel="author"], .author-name');
      const publicationElement = element.querySelector('[data-testid="publicationName"], .publication-name');
      const excerptElement = element.querySelector('p, [data-testid="storyDescription"], .excerpt');
      const timeElement = element.querySelector('time');

      if (linkElement && titleElement) {
        const article = {
          title: titleElement.textContent.trim(),
          url: linkElement.href || '',
          author: authorElement ? authorElement.textContent.trim() : '',
          publication: publicationElement ? publicationElement.textContent.trim() : '',
          excerpt: excerptElement ? excerptElement.textContent.trim().substring(0, 200) : '',
          saved_at: timeElement ? timeElement.getAttribute('datetime') || new Date().toISOString() : new Date().toISOString()
        };

        // Only add if not already in the list
        if (!articles.find(a => a.url === article.url)) {
          articles.push(article);
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
  const maxScrollAttempts = 50; // Maximum number of scroll attempts

  console.log('📜 Scrolling to load all articles...');

  // Scroll and load more articles
  while (scrollAttempts < maxScrollAttempts) {
    extractArticles();
    scrollToBottom();

    // Wait for content to load
    await new Promise(resolve => setTimeout(resolve, 1500));

    const newHeight = document.body.scrollHeight;

    if (newHeight === lastHeight) {
      // No more content loaded, try a few more times to be sure
      scrollAttempts++;
    } else {
      scrollAttempts = 0; // Reset counter if we loaded more content
      lastHeight = newHeight;
    }

    console.log(`Found ${articles.length} articles so far...`);
  }

  // Final extraction
  extractArticles();

  console.log(`✅ Found ${articles.length} total articles`);

  if (articles.length === 0) {
    console.error('❌ No articles found! Make sure you\'re on the reading list page and logged in.');
    return;
  }

  // Convert to CSV
  const headers = ['title', 'url', 'author', 'publication', 'excerpt', 'saved_at'];
  const csvRows = [headers.join(',')];

  articles.forEach(article => {
    const row = headers.map(header => {
      const value = article[header] || '';
      // Escape quotes and wrap in quotes if contains comma, quote, or newline
      const escaped = value.replace(/"/g, '""');
      return /[",\n]/.test(escaped) ? `"${escaped}"` : escaped;
    });
    csvRows.push(row.join(','));
  });

  const csvContent = csvRows.join('\n');

  // Download CSV
  const blob = new Blob([csvContent], { type: 'text/csv;charset=utf-8;' });
  const link = document.createElement('a');
  const url = URL.createObjectURL(blob);

  const timestamp = new Date().toISOString().split('T')[0];
  link.setAttribute('href', url);
  link.setAttribute('download', `medium-reading-list-${timestamp}.csv`);
  link.style.visibility = 'hidden';

  document.body.appendChild(link);
  link.click();
  document.body.removeChild(link);

  console.log(`💾 Downloaded medium-reading-list-${timestamp}.csv`);
  console.log('🎉 Export complete!');

  // Also log a sample for verification
  console.log('\nFirst 3 articles:');
  console.table(articles.slice(0, 3));
})();
