package database

// Chrome SNSS session file parser
// Adapted from https://github.com/lemnos/chrome-session-dump

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"unicode/utf16"

	"github.com/rzolkos/web-recap/internal/browser"
	"github.com/rzolkos/web-recap/internal/models"
)

// SNSS command types
const (
	kCommandSetTabWindow               = 0
	kCommandSetTabIndexInWindow        = 2
	kCommandUpdateTabNavigation        = 6
	kCommandSetSelectedNavigationIndex = 7
	kCommandSetSelectedTabInIndex      = 8
	kCommandTabClosed                  = 16
	kCommandWindowClosed               = 17
	kCommandSetActiveWindow            = 20
	kCommandSetTabGroup                = 25
	kCommandSetTabGroupMetadata2       = 27
)

// Internal structures for parsing
type tabGroup struct {
	high uint64
	low  uint64
	name string
}

type sessionWindow struct {
	activeTabIdx uint32
	id           uint32
	deleted      bool
	tabs         []*sessionTab
}

type historyItem struct {
	idx   uint32
	url   string
	title string
}

type sessionTab struct {
	id                uint32
	history           []*historyItem
	idx               uint32
	win               uint32
	deleted           bool
	currentHistoryIdx uint32
	group             *tabGroup
}

// SessionParser holds the state for parsing a session file
type SessionParser struct {
	tabs         map[uint32]*sessionTab
	windows      map[uint32]*sessionWindow
	groups       map[string]*tabGroup
	activeWindow *sessionWindow
}

func newSessionParser() *SessionParser {
	return &SessionParser{
		tabs:    make(map[uint32]*sessionTab),
		windows: make(map[uint32]*sessionWindow),
		groups:  make(map[string]*tabGroup),
	}
}

func (p *SessionParser) getWindow(id uint32) *sessionWindow {
	if _, ok := p.windows[id]; !ok {
		p.windows[id] = &sessionWindow{id: id}
	}
	return p.windows[id]
}

func (p *SessionParser) getGroup(high, low uint64) *tabGroup {
	key := fmt.Sprintf("%x%x", high, low)
	if _, ok := p.groups[key]; !ok {
		p.groups[key] = &tabGroup{high, low, ""}
	}
	return p.groups[key]
}

func (p *SessionParser) getTab(id uint32) *sessionTab {
	if _, ok := p.tabs[id]; !ok {
		p.tabs[id] = &sessionTab{id: id}
	}
	return p.tabs[id]
}

// Binary reading helpers
func readUint8(r io.Reader) (uint8, error) {
	var b [1]byte
	if _, err := io.ReadFull(r, b[:]); err != nil {
		return 0, err
	}
	return uint8(b[0]), nil
}

func readUint16(r io.Reader) (uint16, error) {
	var b [2]byte
	if _, err := io.ReadFull(r, b[:]); err != nil {
		return 0, err
	}
	return uint16(b[0]) | uint16(b[1])<<8, nil
}

func readUint32(r io.Reader) (uint32, error) {
	var b [4]byte
	if _, err := io.ReadFull(r, b[:]); err != nil {
		return 0, err
	}
	return uint32(b[3])<<24 | uint32(b[2])<<16 | uint32(b[1])<<8 | uint32(b[0]), nil
}

func readUint64(r io.Reader) (uint64, error) {
	var b [8]byte
	if _, err := io.ReadFull(r, b[:]); err != nil {
		return 0, err
	}
	return uint64(b[7])<<56 | uint64(b[6])<<48 | uint64(b[5])<<40 | uint64(b[4])<<32 |
		uint64(b[3])<<24 | uint64(b[2])<<16 | uint64(b[1])<<8 | uint64(b[0]), nil
}

func readString(r io.Reader) (string, error) {
	sz, err := readUint32(r)
	if err != nil {
		return "", err
	}

	// Chrome 32-bit aligns pickled data
	rsz := sz
	if rsz%4 != 0 {
		rsz += 4 - (rsz % 4)
	}

	b := make([]byte, rsz)
	if _, err := io.ReadFull(r, b); err != nil {
		return "", err
	}

	return string(b[:sz]), nil
}

func readString16(r io.Reader) (string, error) {
	sz, err := readUint32(r)
	if err != nil {
		return "", err
	}

	rsz := sz * 2
	if rsz%4 != 0 {
		rsz += 4 - (rsz % 4)
	}

	b := make([]byte, rsz)
	if _, err := io.ReadFull(r, b); err != nil {
		return "", err
	}

	var s []uint16
	for i := 0; i < int(sz*2); i += 2 {
		s = append(s, uint16(b[i+1])<<8|uint16(b[i]))
	}

	return string(utf16.Decode(s)), nil
}

// parseSessionFile parses a Chrome SNSS session file and returns tab entries
func parseSessionFile(path string, browserName string) ([]models.TabEntry, error) {
	fh, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open session file: %w", err)
	}
	defer fh.Close()

	// Check magic header
	var magic [4]byte
	if _, err := io.ReadFull(fh, magic[:]); err != nil {
		return nil, fmt.Errorf("failed to read magic header: %w", err)
	}

	if magic != [4]byte{0x53, 0x4E, 0x53, 0x53} { // "SNSS"
		return nil, fmt.Errorf("invalid SNSS file: bad magic header")
	}

	ver, err := readUint32(fh)
	if err != nil {
		return nil, fmt.Errorf("failed to read version: %w", err)
	}

	if ver != 1 && ver != 3 {
		return nil, fmt.Errorf("unsupported SNSS version: %d", ver)
	}

	parser := newSessionParser()

	// Read commands
	for {
		sz, err := readUint16(fh)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to read command size: %w", err)
		}

		typ, err := readUint8(fh)
		if err != nil {
			return nil, fmt.Errorf("failed to read command type: %w", err)
		}

		buf := make([]byte, int(sz)-1)
		if _, err := io.ReadFull(fh, buf); err != nil {
			return nil, fmt.Errorf("failed to read command payload: %w", err)
		}

		data := bytes.NewBuffer(buf)
		parser.processCommand(typ, data)
	}

	return parser.buildTabEntries(browserName), nil
}

func (p *SessionParser) processCommand(typ uint8, data *bytes.Buffer) {
	switch typ {
	case kCommandUpdateTabNavigation:
		readUint32(data) // size of the data (again)
		id, _ := readUint32(data)
		histIdx, _ := readUint32(data)
		urlStr, _ := readString(data)
		title, _ := readString16(data)

		t := p.getTab(id)

		var item *historyItem
		for _, h := range t.history {
			if h.idx == histIdx {
				item = h
				break
			}
		}

		if item == nil {
			item = &historyItem{idx: histIdx}
			t.history = append(t.history, item)
		}

		item.url = urlStr
		item.title = title

	case kCommandSetSelectedTabInIndex:
		id, _ := readUint32(data)
		idx, _ := readUint32(data)
		p.getWindow(id).activeTabIdx = idx

	case kCommandSetTabGroupMetadata2:
		readUint32(data) // Size
		high, _ := readUint64(data)
		low, _ := readUint64(data)
		name, _ := readString16(data)
		p.getGroup(high, low).name = name

	case kCommandSetTabGroup:
		id, _ := readUint32(data)
		readUint32(data) // Struct padding
		high, _ := readUint64(data)
		low, _ := readUint64(data)
		p.getTab(id).group = p.getGroup(high, low)

	case kCommandSetTabWindow:
		win, _ := readUint32(data)
		id, _ := readUint32(data)
		p.getTab(id).win = win

	case kCommandWindowClosed:
		id, _ := readUint32(data)
		p.getWindow(id).deleted = true

	case kCommandTabClosed:
		id, _ := readUint32(data)
		p.getTab(id).deleted = true

	case kCommandSetTabIndexInWindow:
		id, _ := readUint32(data)
		index, _ := readUint32(data)
		p.getTab(id).idx = index

	case kCommandSetActiveWindow:
		id, _ := readUint32(data)
		p.activeWindow = p.getWindow(id)

	case kCommandSetSelectedNavigationIndex:
		id, _ := readUint32(data)
		idx, _ := readUint32(data)
		p.getTab(id).currentHistoryIdx = idx
	}
}

func (p *SessionParser) buildTabEntries(browserName string) []models.TabEntry {
	// Associate tabs with windows
	for _, t := range p.tabs {
		sort.Slice(t.history, func(i, j int) bool {
			return t.history[i].idx < t.history[j].idx
		})
		w := p.getWindow(t.win)
		w.tabs = append(w.tabs, t)
	}

	// Sort tabs within windows
	for _, w := range p.windows {
		sort.Slice(w.tabs, func(i, j int) bool {
			return w.tabs[i].idx < w.tabs[j].idx
		})
	}

	var entries []models.TabEntry
	windowID := 0

	for _, w := range p.windows {
		if w.deleted {
			continue
		}

		windowID++
		isActiveWindow := w == p.activeWindow
		idx := 0

		for _, t := range w.tabs {
			if t.deleted {
				continue
			}

			// Get current URL and title from history
			var tabURL, tabTitle string
			for _, h := range t.history {
				if h.idx == t.currentHistoryIdx {
					tabURL = h.url
					tabTitle = h.title
					break
				}
			}

			// Fallback to last history item if current index not found
			if tabURL == "" && len(t.history) > 0 {
				last := t.history[len(t.history)-1]
				tabURL = last.url
				tabTitle = last.title
			}

			// Skip empty tabs
			if tabURL == "" {
				continue
			}

			// Extract domain
			domain := ExtractDomain(tabURL)

			// Get group name
			groupName := ""
			if t.group != nil && t.group.name != "" {
				groupName = t.group.name
			}

			entry := models.TabEntry{
				URL:      tabURL,
				Title:    tabTitle,
				Domain:   domain,
				Active:   isActiveWindow && idx == int(w.activeTabIdx),
				Group:    groupName,
				WindowID: windowID,
				Browser:  browserName,
			}

			entries = append(entries, entry)
			idx++
		}
	}

	return entries
}

// findLatestSessionFile finds the most recently modified session file in the sessions directory
func findLatestSessionFile(sessionDir string) (string, error) {
	entries, err := os.ReadDir(sessionDir)
	if err != nil {
		return "", fmt.Errorf("failed to read session directory: %w", err)
	}

	var latestFile string
	var latestTime int64

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		// Look for Session_ files (current session) or Tabs_ files
		if !strings.HasPrefix(name, "Session_") && !strings.HasPrefix(name, "Tabs_") {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}

		modTime := info.ModTime().Unix()
		if modTime > latestTime {
			latestTime = modTime
			latestFile = filepath.Join(sessionDir, name)
		}
	}

	if latestFile == "" {
		return "", fmt.Errorf("no session file found in %s", sessionDir)
	}

	return latestFile, nil
}

// QueryTabs queries open tabs from a Chromium-based browser
func QueryTabs(b *browser.Browser, sessionPath string) ([]models.TabEntry, error) {
	if !browser.IsChromiumBased(b.Type) {
		return nil, fmt.Errorf("tabs extraction only supported for Chromium-based browsers")
	}

	sessionFile, err := findLatestSessionFile(sessionPath)
	if err != nil {
		return nil, err
	}

	return parseSessionFile(sessionFile, b.Name)
}

// QueryMultipleBrowsersTabs queries open tabs from all detected Chromium-based browsers
func QueryMultipleBrowsersTabs(detector *browser.Detector) ([]models.TabEntry, error) {
	browsers := detector.Detect()
	var allEntries []models.TabEntry

	for _, b := range browsers {
		if !browser.IsChromiumBased(b.Type) {
			continue
		}

		sessionPath, err := browser.GetSessionPath(b.Type)
		if err != nil {
			continue
		}

		entries, err := QueryTabs(&b, sessionPath)
		if err != nil {
			continue
		}

		allEntries = append(allEntries, entries...)
	}

	return allEntries, nil
}
