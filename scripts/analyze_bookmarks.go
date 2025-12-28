package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
)

type Bookmark struct {
	DateAdded    string `json:"date_added"`
	DateModified string `json:"date_modified"`
	URL          string `json:"url"`
	Title        string `json:"title"`
	Folder       string `json:"folder"`
	Domain       string `json:"domain"`
	Browser      string `json:"browser"`
}

type BookmarkData struct {
	Browser      string     `json:"browser"`
	StartDate    string     `json:"start_date"`
	EndDate      string     `json:"end_date"`
	TotalEntries int        `json:"total_entries"`
	Entries      []Bookmark `json:"entries"`
}

type CategoryInfo struct {
	Count     int
	Examples  []string
	Keywords  []string
}

func categorizeBookmark(title, domain, url string) []string {
	text := strings.ToLower(title + " " + domain + " " + url)
	categories := []string{}

	categoryKeywords := map[string][]string{
		"AI/ML/LLM": {"ai", "llm", "gpt", "claude", "anthropic", "openai", "machine learning", "neural", "agent", "prompt", "embedding", "vector", "inference", "model", "training", "chatgpt", "gemini", "huggingface", "langchain", "ollama", "letta"},
		"Software Development": {"code", "coding", "programming", "developer", "software", "debugging", "git", "github", "gitlab", "api", "sdk", "framework", "library", "typescript", "javascript", "python", "go", "rust", "java", "react", "vue", "node"},
		"DevOps/Infrastructure": {"docker", "kubernetes", "k8s", "deploy", "infrastructure", "cloud", "aws", "azure", "gcp", "ci/cd", "devops", "terraform", "ansible", "monitoring", "observability"},
		"Data/Analytics": {"data", "analytics", "database", "sql", "nosql", "postgres", "mongodb", "elasticsearch", "bigquery", "datawarehouse", "etl", "pipeline"},
		"Security": {"security", "vulnerability", "exploit", "crypto", "encryption", "authentication", "authorization", "oauth", "jwt", "ssl", "tls", "penetration", "hacking"},
		"Web Development": {"web", "frontend", "backend", "fullstack", "html", "css", "responsive", "ux", "ui", "design system", "tailwind", "bootstrap"},
		"Mobile Development": {"mobile", "ios", "android", "swift", "kotlin", "flutter", "react native", "app store", "play store"},
		"Architecture/Design": {"architecture", "design pattern", "microservices", "monolith", "scalability", "distributed", "system design", "architecture"},
		"Productivity/Tools": {"productivity", "tool", "workflow", "automation", "cli", "terminal", "vim", "vscode", "ide", "editor"},
		"Business/Startup": {"startup", "business", "founder", "entrepreneurship", "marketing", "sales", "growth", "revenue", "funding", "venture"},
		"Writing/Content": {"writing", "blog", "article", "documentation", "technical writing", "content", "newsletter", "medium"},
		"Research/Papers": {"research", "paper", "arxiv", "study", "academic", "journal", "scientific"},
		"Finance/Crypto": {"finance", "crypto", "cryptocurrency", "bitcoin", "ethereum", "blockchain", "trading", "investment", "stock"},
		"Gaming": {"game", "gaming", "unity", "unreal", "steam", "playstation", "xbox"},
		"Tutorials/Learning": {"tutorial", "learn", "course", "guide", "how to", "introduction", "beginner", "workshop", "lesson"},
		"News/Tech News": {"news", "techcrunch", "hacker news", "ycombinator", "tech news", "announcement"},
		"Open Source": {"open source", "oss", "contribution", "community", "license", "mit", "apache"},
	}

	for category, keywords := range categoryKeywords {
		for _, keyword := range keywords {
			if strings.Contains(text, keyword) {
				categories = append(categories, category)
				break
			}
		}
	}

	// If no category matched, assign to "Other"
	if len(categories) == 0 {
		categories = append(categories, "Other")
	}

	return categories
}

func main() {
	file, err := os.Open("bookmarks_last_year.json")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening file: %v\n", err)
		os.Exit(1)
	}
	defer file.Close()

	var data BookmarkData
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&data); err != nil {
		fmt.Fprintf(os.Stderr, "Error decoding JSON: %v\n", err)
		os.Exit(1)
	}

	categoryMap := make(map[string]*CategoryInfo)

	for _, bookmark := range data.Entries {
		categories := categorizeBookmark(bookmark.Title, bookmark.Domain, bookmark.URL)

		for _, category := range categories {
			if categoryMap[category] == nil {
				categoryMap[category] = &CategoryInfo{
					Examples: []string{},
				}
			}

			categoryMap[category].Count++

			// Store up to 5 examples per category
			if len(categoryMap[category].Examples) < 5 {
				example := fmt.Sprintf("%s (%s)", bookmark.Title, bookmark.Domain)
				categoryMap[category].Examples = append(categoryMap[category].Examples, example)
			}
		}
	}

	// Sort categories by count
	type CategoryCount struct {
		Name  string
		Info  *CategoryInfo
	}

	var categories []CategoryCount
	for name, info := range categoryMap {
		categories = append(categories, CategoryCount{Name: name, Info: info})
	}

	sort.Slice(categories, func(i, j int) bool {
		return categories[i].Info.Count > categories[j].Info.Count
	})

	fmt.Printf("\n=== BOOKMARK ANALYSIS FOR LAST YEAR ===\n")
	fmt.Printf("Total Bookmarks: %d\n", data.TotalEntries)
	fmt.Printf("Date Range: %s to %s\n\n", data.StartDate, data.EndDate)

	fmt.Printf("=== CATEGORIES (sorted by count) ===\n\n")
	for _, cat := range categories {
		fmt.Printf("📁 %s: %d bookmarks\n", cat.Name, cat.Info.Count)
		fmt.Printf("   Examples:\n")
		for _, example := range cat.Info.Examples {
			// Truncate long titles
			if len(example) > 100 {
				example = example[:97] + "..."
			}
			fmt.Printf("   - %s\n", example)
		}
		fmt.Println()
	}
}
