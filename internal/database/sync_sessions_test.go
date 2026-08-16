package database

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rzolkos/web-recap/internal/browser"
	"github.com/syndtr/goleveldb/leveldb"
)

func encodeNavigation(url, title string) []byte {
	var out []byte
	out = append(out, encodeProtoString(2, url)...)
	out = append(out, encodeProtoString(4, title)...)
	return out
}

func encodeSessionTab(tag, url, title string, windowID, navIndex int, pinned bool) []byte {
	nav := encodeNavigation(url, title)
	var tab []byte
	tab = append(tab, encodeProtoVarintField(1, 11)...)
	tab = append(tab, encodeProtoVarintField(2, uint64(windowID))...)
	tab = append(tab, encodeProtoVarintField(3, 0)...)
	tab = append(tab, encodeProtoVarintField(4, uint64(navIndex))...)
	if pinned {
		tab = append(tab, encodeProtoVarintField(5, 1)...)
	}
	tab = append(tab, encodeProtoBytes(7, nav)...)

	var out []byte
	out = append(out, encodeProtoString(1, tag)...)
	out = append(out, encodeProtoBytes(3, tab)...)
	return out
}

func encodeSessionHeader(tag, clientName string) []byte {
	var header []byte
	header = append(header, encodeProtoString(3, clientName)...)
	var out []byte
	out = append(out, encodeProtoString(1, tag)...)
	out = append(out, encodeProtoBytes(2, header)...)
	return out
}

func encodeDeviceInfo(guid, name, model string) []byte {
	var out []byte
	out = append(out, encodeProtoString(1, guid)...)
	out = append(out, encodeProtoString(2, name)...)
	out = append(out, encodeProtoString(11, model)...)
	return out
}

func TestParseSessionSpecificsTab(t *testing.T) {
	raw := encodeSessionTab("pixel-guid", "https://example.com/a", "Example", 7, 0, true)
	header, tab, ok := parseSessionSpecifics(raw)
	if !ok || header != nil || tab == nil {
		t.Fatalf("expected tab, got header=%v tab=%v ok=%v", header, tab, ok)
	}
	if tab.url != "https://example.com/a" || tab.title != "Example" || !tab.pinned || tab.windowID != 7 {
		t.Fatalf("unexpected tab: %+v", tab)
	}
}

func TestParseSessionSpecificsHeader(t *testing.T) {
	raw := encodeSessionHeader("pixel-guid", "Pixel 9a")
	header, tab, ok := parseSessionSpecifics(raw)
	if !ok || header == nil || tab != nil {
		t.Fatalf("expected header, got header=%v tab=%v ok=%v", header, tab, ok)
	}
	if header.clientName != "Pixel 9a" || header.sessionTag != "pixel-guid" {
		t.Fatalf("unexpected header: %+v", header)
	}
}

func TestQuerySyncedTabsFromFixture(t *testing.T) {
	dir := t.TempDir()
	db, err := leveldb.OpenFile(dir, nil)
	if err != nil {
		t.Fatal(err)
	}
	pixel := "d8XoS+n5VFxgsQWcFO2P/w=="
	mac := "17Ttg4xJGghJTXr0KWSXlw=="
	host, err := os.Hostname()
	if err != nil {
		t.Fatal(err)
	}
	if short, _, ok := strings.Cut(host, "."); ok {
		host = short
	}
	puts := map[string][]byte{
		"device_info-dt-" + pixel: encodeDeviceInfo(pixel, "Chrome ANDROID-PHONE Vivaldi", "Pixel 9a"),
		"device_info-dt-" + mac:   encodeDeviceInfo(mac, "Chrome LINUX 150 Vivaldi", "SomeBox"),
		"sessions-dt-pixel-h":     encodeSessionHeader(pixel, "Pixel 9a"),
		"sessions-dt-pixel-t":     encodeSessionTab(pixel, "https://ampcode.com/notes", "Orbs", 1, 0, false),
		"sessions-dt-mac-h":       encodeSessionHeader(mac, host),
		"sessions-dt-mac-t":       encodeSessionTab(mac, "https://grok.com/", "Grok", 1, 0, false),
		"sessions-md-ignore":      []byte("not a session"),
	}
	for k, v := range puts {
		if err := db.Put([]byte(k), v, nil); err != nil {
			t.Fatal(err)
		}
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}

	b := &browser.Browser{Type: browser.Vivaldi, Name: "Vivaldi"}
	entries, err := QuerySyncedTabs(b, dir, true)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 remote tab after skipping local, got %d: %+v", len(entries), entries)
	}
	got := entries[0]
	if got.URL != "https://ampcode.com/notes" || got.Device != "Pixel 9a" || !got.Synced {
		t.Fatalf("unexpected entry: %+v", got)
	}
	if got.Domain != "ampcode.com" || got.Browser != "Vivaldi" {
		t.Fatalf("unexpected domain/browser: %+v", got)
	}
	if len(got.Tags) < 2 || got.Tags[0] != "synced" || got.Tags[1] != "pixel-9a" {
		t.Fatalf("unexpected tags: %v", got.Tags)
	}

	all, err := QuerySyncedTabs(b, dir, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(all) != 2 {
		t.Fatalf("expected 2 tabs without skipLocal, got %d", len(all))
	}
}

func TestQuerySyncedTabsMissingDir(t *testing.T) {
	b := &browser.Browser{Type: browser.Vivaldi, Name: "Vivaldi"}
	entries, err := QuerySyncedTabs(b, filepath.Join(t.TempDir(), "nope"), true)
	if err != nil {
		t.Fatal(err)
	}
	if entries != nil {
		t.Fatalf("expected nil entries, got %v", entries)
	}
}

func TestDeviceSlug(t *testing.T) {
	if got := deviceSlug("Pixel 9a"); got != "pixel-9a" {
		t.Fatalf("got %q", got)
	}
}

func TestIsLocalClientName(t *testing.T) {
	if !isLocalClientName("Manikandas-MacBook-Pro", []string{"Manikandas-MacBook-Pro.local", "Manikandas-MacBook-Pro"}) {
		t.Fatal("expected hostname match")
	}
	if isLocalClientName("Pixel 9a", []string{"Manikandas-MacBook-Pro"}) {
		t.Fatal("pixel should not match")
	}
}

func TestGetSyncDataPathFromSessionDir(t *testing.T) {
	got := browser.SyncDataPathFromSessionDir("/tmp/Default/Sessions")
	want := filepath.Join("/tmp/Default", "Sync Data", "LevelDB")
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestCopyLevelDBSkipsLock(t *testing.T) {
	src := t.TempDir()
	if err := os.WriteFile(filepath.Join(src, "CURRENT"), []byte("MANIFEST-000001\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(src, "LOCK"), []byte("held"), 0o600); err != nil {
		t.Fatal(err)
	}
	dst, err := copyLevelDB(src)
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dst)
	if _, err := os.Stat(filepath.Join(dst, "LOCK")); !os.IsNotExist(err) {
		t.Fatalf("LOCK should not be copied")
	}
	if _, err := os.Stat(filepath.Join(dst, "CURRENT")); err != nil {
		t.Fatal(err)
	}
}
