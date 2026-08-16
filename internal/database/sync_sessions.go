package database

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strings"

	"github.com/rzolkos/web-recap/internal/browser"
	"github.com/rzolkos/web-recap/internal/models"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/opt"
)

var deviceSlugRe = regexp.MustCompile(`[^a-z0-9]+`)

type syncDevice struct {
	guid  string
	name  string
	model string
	raw   string
}

type syncSessionTab struct {
	sessionTag string
	windowID   int
	index      int
	pinned     bool
	url        string
	title      string
}

type syncSessionHeader struct {
	sessionTag string
	clientName string
}

// QuerySyncedTabs reads Chromium/Vivaldi Sync Data LevelDB and returns open
// tabs from other devices. skipLocal drops the current machine's session when
// local SNSS tabs were already extracted.
func QuerySyncedTabs(b *browser.Browser, syncDataPath string, skipLocal bool) ([]models.TabEntry, error) {
	if b == nil || !browser.IsChromiumBased(b.Type) {
		return nil, fmt.Errorf("synced tabs only supported for Chromium-based browsers")
	}
	if syncDataPath == "" {
		return nil, fmt.Errorf("sync data path is empty")
	}
	if _, err := os.Stat(syncDataPath); err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("cannot access sync data: %w", err)
	}

	tmpDir, err := copyLevelDB(syncDataPath)
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(tmpDir)

	db, err := leveldb.OpenFile(tmpDir, &opt.Options{ReadOnly: true, ErrorIfMissing: true})
	if err != nil {
		return nil, fmt.Errorf("open sync leveldb: %w", err)
	}
	defer db.Close()

	devices := map[string]syncDevice{}
	headers := map[string]syncSessionHeader{}
	var tabs []syncSessionTab

	it := db.NewIterator(nil, nil)
	defer it.Release()
	for it.Next() {
		key := string(it.Key())
		value := it.Value()
		switch {
		case strings.Contains(key, "device_info-dt-"):
			if d, ok := parseDeviceInfo(value); ok {
				devices[d.guid] = d
			}
		case strings.Contains(key, "sessions-dt-"):
			header, tab, ok := parseSessionSpecifics(value)
			if !ok {
				continue
			}
			if header != nil {
				headers[header.sessionTag] = *header
				continue
			}
			if tab != nil && tab.url != "" {
				tabs = append(tabs, *tab)
			}
		}
	}
	if err := it.Error(); err != nil {
		return nil, fmt.Errorf("iterate sync leveldb: %w", err)
	}

	localTags := localSessionTags(devices, headers)
	windowIDs := map[string]int{}
	nextWindow := 1000
	var entries []models.TabEntry

	for _, tab := range tabs {
		if skipLocal && localTags[tab.sessionTag] {
			continue
		}
		deviceName := deviceNameFor(tab.sessionTag, headers, devices)
		win, ok := windowIDs[tab.sessionTag]
		if !ok {
			win = nextWindow
			nextWindow++
			windowIDs[tab.sessionTag] = win
		}
		entries = append(entries, models.TabEntry{
			URL:      tab.url,
			Title:    tab.title,
			Domain:   ExtractDomain(tab.url),
			Pinned:   tab.pinned,
			WindowID: win,
			Browser:  b.Name,
			Device:   deviceName,
			Synced:   true,
			Tags:     syncedTabTags(deviceName),
		})
	}

	sort.SliceStable(entries, func(i, j int) bool {
		if entries[i].Device != entries[j].Device {
			return entries[i].Device < entries[j].Device
		}
		return entries[i].URL < entries[j].URL
	})
	return entries, nil
}

func parseDeviceInfo(value []byte) (syncDevice, bool) {
	fields := decodeProtoFields(value)
	guid := protoStringField(fields, 1)
	if guid == "" {
		return syncDevice{}, false
	}
	return syncDevice{
		guid:  guid,
		name:  protoStringField(fields, 2),
		model: protoStringField(fields, 11),
		raw:   protoStringField(fields, 2),
	}, true
}

func parseSessionSpecifics(value []byte) (*syncSessionHeader, *syncSessionTab, bool) {
	fields := decodeProtoFields(value)
	tag := protoStringField(fields, 1)
	if tag == "" {
		return nil, nil, false
	}
	if headerBytes := protoBytesField(fields, 2); len(headerBytes) > 0 {
		hf := decodeProtoFields(headerBytes)
		name := protoStringField(hf, 3)
		return &syncSessionHeader{sessionTag: tag, clientName: name}, nil, true
	}
	if tabBytes := protoBytesField(fields, 3); len(tabBytes) > 0 {
		tf := decodeProtoFields(tabBytes)
		navs := protoRepeatedBytes(tf, 7)
		if len(navs) == 0 {
			return nil, nil, false
		}
		idx, ok := protoIntField(tf, 4)
		if !ok || idx < 0 || idx >= len(navs) {
			idx = len(navs) - 1
		}
		nf := decodeProtoFields(navs[idx])
		url := protoStringField(nf, 2)
		if url == "" || shouldSkipSyncedURL(url) {
			return nil, nil, false
		}
		windowID, _ := protoIntField(tf, 2)
		visual, _ := protoIntField(tf, 3)
		return nil, &syncSessionTab{
			sessionTag: tag,
			windowID:   windowID,
			index:      visual,
			pinned:     protoBoolField(tf, 5),
			url:        url,
			title:      protoStringField(nf, 4),
		}, true
	}
	return nil, nil, false
}

func protoBytesField(fields []protoField, num int) []byte {
	for _, f := range fields {
		if f.num == num && f.wire == 2 {
			return f.bytes
		}
	}
	return nil
}

func shouldSkipSyncedURL(url string) bool {
	switch {
	case strings.HasPrefix(url, "chrome-extension://"),
		strings.HasPrefix(url, "moz-extension://"),
		strings.HasPrefix(url, "edge-extension://"),
		strings.HasPrefix(url, "brave-extension://"),
		strings.HasPrefix(url, "chrome://vivaldi-data/"):
		return true
	default:
		return false
	}
}

func deviceNameFor(tag string, headers map[string]syncSessionHeader, devices map[string]syncDevice) string {
	if h, ok := headers[tag]; ok && strings.TrimSpace(h.clientName) != "" && !isChromeBuildName(h.clientName) {
		return h.clientName
	}
	if d, ok := devices[tag]; ok {
		if d.model != "" && !isChromeBuildName(d.model) {
			return d.model
		}
		if pretty := humanDeviceName(d.name); pretty != "" {
			return pretty
		}
		if d.name != "" {
			return d.name
		}
	}
	if h, ok := headers[tag]; ok && h.clientName != "" {
		return h.clientName
	}
	return "Unknown device"
}

func isChromeBuildName(s string) bool {
	return strings.HasPrefix(s, "Chrome ") || strings.Contains(s, "Developer Build")
}

func humanDeviceName(raw string) string {
	// "Chrome ANDROID-PHONE … Vivaldi 8.1…" → prefer model via caller.
	// "Chrome MAC …" is the local desktop.
	switch {
	case strings.Contains(raw, "ANDROID-PHONE"):
		return "Android"
	case strings.Contains(raw, "ANDROID-TABLET"):
		return "Android tablet"
	case strings.Contains(raw, "Chrome MAC"):
		return "This Mac"
	case strings.Contains(raw, "Chrome WIN"):
		return "Windows"
	case strings.Contains(raw, "Chrome LINUX"):
		return "Linux"
	default:
		return ""
	}
}

func localSessionTags(devices map[string]syncDevice, headers map[string]syncSessionHeader) map[string]bool {
	out := map[string]bool{}
	hosts := localHostnames()
	for guid, d := range devices {
		if isLocalDevice(d) {
			out[guid] = true
		}
	}
	for tag, h := range headers {
		if isLocalClientName(h.clientName, hosts) {
			out[tag] = true
		}
	}
	return out
}

func localHostnames() []string {
	var names []string
	if h, err := os.Hostname(); err == nil {
		names = append(names, h)
		if short, _, ok := strings.Cut(h, "."); ok {
			names = append(names, short)
		}
	}
	return names
}

func isLocalClientName(name string, hosts []string) bool {
	n := strings.TrimSpace(name)
	if n == "" {
		return false
	}
	for _, h := range hosts {
		if strings.EqualFold(n, h) {
			return true
		}
	}
	return false
}

func isLocalDevice(d syncDevice) bool {
	raw := d.name + " " + d.raw
	switch runtime.GOOS {
	case "darwin":
		return strings.Contains(raw, "Chrome MAC") || strings.Contains(raw, "Mac OS X") || strings.Contains(raw, "Mac OS")
	case "windows":
		return strings.Contains(raw, "Chrome WIN") || strings.Contains(raw, " Windows")
	case "linux":
		return strings.Contains(raw, "Chrome LINUX") || strings.Contains(raw, " Linux")
	default:
		return false
	}
}

func syncedTabTags(device string) []string {
	tags := []string{"synced"}
	if slug := deviceSlug(device); slug != "" && slug != "synced" {
		tags = append(tags, slug)
	}
	return tags
}

func deviceSlug(device string) string {
	s := strings.ToLower(strings.TrimSpace(device))
	s = deviceSlugRe.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")
	return s
}

func copyLevelDB(src string) (string, error) {
	dst, err := os.MkdirTemp("", "web-recap-sync-*")
	if err != nil {
		return "", fmt.Errorf("temp dir: %w", err)
	}
	entries, err := os.ReadDir(src)
	if err != nil {
		os.RemoveAll(dst)
		return "", fmt.Errorf("read sync data: %w", err)
	}
	for _, entry := range entries {
		if entry.IsDir() || entry.Name() == "LOCK" {
			continue
		}
		if err := copyFile(filepath.Join(src, entry.Name()), filepath.Join(dst, entry.Name())); err != nil {
			os.RemoveAll(dst)
			return "", fmt.Errorf("copy %s: %w", entry.Name(), err)
		}
	}
	return dst, nil
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o600)
	if err != nil {
		return err
	}
	defer out.Close()
	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	return out.Close()
}

// QueryMultipleBrowsersSyncedTabs extracts synced tabs from every Chromium browser.
func QueryMultipleBrowsersSyncedTabs(detector *browser.Detector, skipLocal bool) ([]models.TabEntry, error) {
	var all []models.TabEntry
	for _, b := range detector.Detect() {
		if !browser.IsChromiumBased(b.Type) {
			continue
		}
		syncPath, err := browser.GetSyncDataPath(b.Type)
		if err != nil {
			continue
		}
		entries, err := QuerySyncedTabs(&b, syncPath, skipLocal)
		if err != nil {
			continue
		}
		all = append(all, entries...)
	}
	return all, nil
}
