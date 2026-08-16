package models

// TabEntry represents a single open browser tab
type TabEntry struct {
	URL      string   `json:"url"`
	Title    string   `json:"title"`
	Domain   string   `json:"domain"`
	Active   bool     `json:"active"`
	Pinned   bool     `json:"pinned,omitempty"`
	Group    string   `json:"group,omitempty"`
	WindowID int      `json:"window_id"`
	Browser  string   `json:"browser"`
	Device   string   `json:"device,omitempty"`
	Synced   bool     `json:"synced,omitempty"`
	Tags     []string `json:"tags,omitempty"`
}

// TabReport represents a collection of open tabs
type TabReport struct {
	Browser      string     `json:"browser"`
	TotalTabs    int        `json:"total_tabs"`
	TotalWindows int        `json:"total_windows"`
	Entries      []TabEntry `json:"entries"`
}
