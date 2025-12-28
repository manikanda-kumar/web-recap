/**
 * Medium Public Reading List Exporter
 *
 * Instructions:
 * 1. Go to any PUBLIC Medium reading list (e.g., https://medium.com/@username/list/reading-list)
 * 2. Open DevTools (F12 or Cmd+Option+I)
 * 3. Go to the Console tab
 * 4. Paste this entire script and press Enter
 * 5. Wait for the export to complete
 * 6. A JSON file will be downloaded automatically
 *
 * The script will scroll through the reading list to load all articles.
 */

(async function exportMediumPublicReadingList() {
  console.log('🚀 Starting Medium Public Reading List Export...');

  // Array to store all articles
  const articles = [];
  const seenUrls = new Set();

  // Function to extract article URL from href
  function getArticleUrl(href) {
    if (!href) return null;
    // Article URLs have a hash ID pattern like -628ce7b4de76
    const hashMatch = href.match(/-([a-f0-9]{8,12})(\?|$)/);
    if (!hashMatch) return null;
    if (href.includes('/m/signin') || href.includes('bookmark')) return null;

    let url = href.startsWith('http') ? href : 'https://medium.com' + href;
    // Remove query params
    const qIdx = url.indexOf('?');
    if (qIdx > -1) url = url.substring(0, qIdx);
    return url;
  }

  // Function to extract articles from the current page
  function extractArticles() {
    const articleElements = document.querySelectorAll('article');

    articleElements.forEach(element => {
      // Find all links in the article
      const links = Array.from(element.querySelectorAll('a[href]'));

      // Find the article URL
      let articleUrl = null;
      for (const link of links) {
        const url = getArticleUrl(link.getAttribute('href'));
        if (url) {
          articleUrl = url;
          break;
        }
      }

      if (!articleUrl || seenUrls.has(articleUrl)) return;
      seenUrls.add(articleUrl);

      // Get title from h2
      const h2 = element.querySelector('h2');
      const title = h2 ? h2.textContent.trim() : '';
      if (!title) return;

      // Get subtitle/excerpt from h3
      const h3 = element.querySelector('h3');
      const excerpt = h3 ? h3.textContent.trim() : '';

      // Get author from links with /@username pattern (not article links)
      let author = '';
      for (const link of links) {
        const href = link.getAttribute('href') || '';
        // Author links contain /@ but don't have the article hash
        if (href.includes('/@') && !/-[a-f0-9]{8,}/.test(href)) {
          const text = link.textContent.trim();
          if (text && text !== 'In' && text !== 'by' && text.length > 1 && text.length < 50) {
            author = text;
            break;
          }
        }
      }

      // Get publication from links
      let publication = '';
      for (const link of links) {
        const href = link.getAttribute('href') || '';
        if (href.includes('medium.com/') && !href.includes('/@') &&
            !href.includes('?source=') && !/-[a-f0-9]{8,}/.test(href)) {
          const text = link.textContent.trim();
          if (text && text !== 'In' && text !== author && text.length > 1 && text.length < 50) {
            publication = text;
            break;
          }
        }
      }

      // Get date
      let savedAt = '';
      const allText = element.textContent;
      const dateMatch = allText.match(/(Jan|Feb|Mar|Apr|May|Jun|Jul|Aug|Sep|Oct|Nov|Dec)\s+\d{1,2}(,\s*\d{4})?/);
      if (dateMatch) {
        savedAt = dateMatch[0];
        // Parse the date
        const parsedDate = new Date(savedAt);
        if (!isNaN(parsedDate.getTime())) {
          savedAt = parsedDate.toISOString();
        } else {
          // Try adding current year
          const withYear = savedAt + ', ' + new Date().getFullYear();
          const parsedWithYear = new Date(withYear);
          if (!isNaN(parsedWithYear.getTime())) {
            // Check if date is in the future
            if (parsedWithYear > new Date()) {
              parsedWithYear.setFullYear(parsedWithYear.getFullYear() - 1);
            }
            savedAt = parsedWithYear.toISOString();
          }
        }
      }

      articles.push({
        url: articleUrl,
        title: title,
        author: author,
        publication: publication,
        excerpt: excerpt,
        saved_at: savedAt,
        platform: 'medium'
      });
    });
  }

  // Function to scroll to bottom of page
  function scrollToBottom() {
    window.scrollTo(0, document.body.scrollHeight);
  }

  // Get initial height
  let lastHeight = document.body.scrollHeight;
  let scrollAttempts = 0;
  const maxScrollAttempts = 50;

  console.log('📜 Scrolling to load all articles...');

  // Scroll and load more articles
  while (scrollAttempts < maxScrollAttempts) {
    extractArticles();
    scrollToBottom();

    // Wait for content to load
    await new Promise(resolve => setTimeout(resolve, 1500));

    const newHeight = document.body.scrollHeight;

    if (newHeight === lastHeight) {
      scrollAttempts++;
    } else {
      scrollAttempts = 0;
      lastHeight = newHeight;
    }

    console.log(`Found ${articles.length} articles so far...`);
  }

  // Final extraction
  extractArticles();

  console.log(`✅ Found ${articles.length} total articles`);

  if (articles.length === 0) {
    console.error('❌ No articles found! Make sure you\'re on a Medium reading list page.');
    return;
  }

  // Create JSON output
  const output = {
    platform: 'medium',
    exported_at: new Date().toISOString(),
    total_entries: articles.length,
    entries: articles
  };

  const jsonContent = JSON.stringify(output, null, 2);

  // Download JSON
  const blob = new Blob([jsonContent], { type: 'application/json' });
  const link = document.createElement('a');
  const url = URL.createObjectURL(blob);

  const timestamp = new Date().toISOString().split('T')[0];
  link.setAttribute('href', url);
  link.setAttribute('download', `medium-reading-list-${timestamp}.json`);
  link.style.visibility = 'hidden';

  document.body.appendChild(link);
  link.click();
  document.body.removeChild(link);

  console.log(`💾 Downloaded medium-reading-list-${timestamp}.json`);
  console.log('🎉 Export complete!');

  // Log sample for verification
  console.log('\nFirst 3 articles:');
  console.table(articles.slice(0, 3));

  return articles;
})();
