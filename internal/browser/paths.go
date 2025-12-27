package browser

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// GetDatabasePath returns the database path for a given browser type on the current platform
func GetDatabasePath(browserType Type) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	switch runtime.GOOS {
	case "linux":
		return getLinuxPath(home, browserType)
	case "darwin":
		return getDarwinPath(home, browserType)
	case "windows":
		return getWindowsPath(browserType)
	default:
		return "", ErrUnsupportedPlatform
	}
}

func getLinuxPath(home string, browserType Type) (string, error) {
	switch browserType {
	case Chrome:
		return filepath.Join(home, ".config/google-chrome/Default/History"), nil
	case Chromium:
		return filepath.Join(home, ".config/chromium/Default/History"), nil
	case Edge:
		return filepath.Join(home, ".config/microsoft-edge/Default/History"), nil
	case Brave:
		return filepath.Join(home, ".config/BraveSoftware/Brave-Browser/Default/History"), nil
	case Vivaldi:
		return filepath.Join(home, ".config/vivaldi/Default/History"), nil
	case Firefox:
		// Firefox uses profile directory, we'll handle this in detector
		return filepath.Join(home, ".mozilla/firefox"), nil
	case Safari:
		// Safari not available on Linux
		return "", ErrBrowserNotAvailable
	case Auto:
		return "", nil
	default:
		return "", ErrUnknownBrowser
	}
}

func getDarwinPath(home string, browserType Type) (string, error) {
	switch browserType {
	case Chrome:
		return filepath.Join(home, "Library/Application Support/Google/Chrome/Default/History"), nil
	case Chromium:
		return filepath.Join(home, "Library/Application Support/Chromium/Default/History"), nil
	case Edge:
		return filepath.Join(home, "Library/Application Support/Microsoft Edge/Default/History"), nil
	case Brave:
		return filepath.Join(home, "Library/Application Support/BraveSoftware/Brave-Browser/Default/History"), nil
	case Vivaldi:
		return filepath.Join(home, "Library/Application Support/Vivaldi/Default/History"), nil
	case Firefox:
		return filepath.Join(home, "Library/Application Support/Firefox"), nil
	case Safari:
		return filepath.Join(home, "Library/Safari/History.db"), nil
	case Auto:
		return "", nil
	default:
		return "", ErrUnknownBrowser
	}
}

func getWindowsPath(browserType Type) (string, error) {
	appData := os.Getenv("LOCALAPPDATA")
	if appData == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		appData = filepath.Join(home, "AppData/Local")
	}

	switch browserType {
	case Chrome:
		return filepath.Join(appData, `Google\Chrome\User Data\Default\History`), nil
	case Chromium:
		return filepath.Join(appData, `Chromium\User Data\Default\History`), nil
	case Edge:
		return filepath.Join(appData, `Microsoft\Edge\User Data\Default\History`), nil
	case Brave:
		return filepath.Join(appData, `BraveSoftware\Brave-Browser\User Data\Default\History`), nil
	case Vivaldi:
		return filepath.Join(appData, `Vivaldi\User Data\Default\History`), nil
	case Firefox:
		return filepath.Join(appData, "Mozilla/Firefox"), nil
	case Safari:
		// Safari not available on Windows
		return "", ErrBrowserNotAvailable
	case Auto:
		return "", nil
	default:
		return "", ErrUnknownBrowser
	}
}

// GetFirefoxProfilePath returns the active Firefox profile path
func GetFirefoxProfilePath(profileBaseDir string) (string, error) {
	if !fileExists(profileBaseDir) {
		return "", ErrFirefoxProfileNotFound
	}

	// Try to find the default profile or most recently modified profile
	entries, err := os.ReadDir(profileBaseDir)
	if err != nil {
		return "", err
	}

	var mostRecentPath string
	var mostRecentTime int64

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		name := entry.Name()
		// Look for .default-release or .default profiles first
		if strings.HasSuffix(name, ".default-release") || strings.HasSuffix(name, ".default") {
			placesPath := filepath.Join(profileBaseDir, name, "places.sqlite")
			if fileExists(placesPath) {
				return placesPath, nil
			}
		}

		// Otherwise, keep track of the most recently modified profile
		info, err := entry.Info()
		if err != nil {
			continue
		}

		modTime := info.ModTime().Unix()
		if modTime > mostRecentTime {
			mostRecentTime = modTime
			placesPath := filepath.Join(profileBaseDir, name, "places.sqlite")
			if fileExists(placesPath) {
				mostRecentPath = placesPath
			}
		}
	}

	if mostRecentPath != "" {
		return mostRecentPath, nil
	}

	return "", ErrFirefoxProfileNotFound
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// GetBookmarkPath returns the bookmark database path for a given browser type on the current platform
func GetBookmarkPath(browserType Type) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	switch runtime.GOOS {
	case "linux":
		return getLinuxBookmarkPath(home, browserType)
	case "darwin":
		return getDarwinBookmarkPath(home, browserType)
	case "windows":
		return getWindowsBookmarkPath(browserType)
	default:
		return "", ErrUnsupportedPlatform
	}
}

func getLinuxBookmarkPath(home string, browserType Type) (string, error) {
	switch browserType {
	case Chrome:
		return filepath.Join(home, ".config/google-chrome/Default/Bookmarks"), nil
	case Chromium:
		return filepath.Join(home, ".config/chromium/Default/Bookmarks"), nil
	case Edge:
		return filepath.Join(home, ".config/microsoft-edge/Default/Bookmarks"), nil
	case Brave:
		return filepath.Join(home, ".config/BraveSoftware/Brave-Browser/Default/Bookmarks"), nil
	case Vivaldi:
		return filepath.Join(home, ".config/vivaldi/Default/Bookmarks"), nil
	case Firefox:
		// Firefox bookmarks are in places.sqlite (same as history)
		return filepath.Join(home, ".mozilla/firefox"), nil
	case Safari:
		// Safari not available on Linux
		return "", ErrBrowserNotAvailable
	case Auto:
		return "", nil
	default:
		return "", ErrUnknownBrowser
	}
}

func getDarwinBookmarkPath(home string, browserType Type) (string, error) {
	switch browserType {
	case Chrome:
		return filepath.Join(home, "Library/Application Support/Google/Chrome/Default/Bookmarks"), nil
	case Chromium:
		return filepath.Join(home, "Library/Application Support/Chromium/Default/Bookmarks"), nil
	case Edge:
		return filepath.Join(home, "Library/Application Support/Microsoft Edge/Default/Bookmarks"), nil
	case Brave:
		return filepath.Join(home, "Library/Application Support/BraveSoftware/Brave-Browser/Default/Bookmarks"), nil
	case Vivaldi:
		return filepath.Join(home, "Library/Application Support/Vivaldi/Default/Bookmarks"), nil
	case Firefox:
		return filepath.Join(home, "Library/Application Support/Firefox"), nil
	case Safari:
		return filepath.Join(home, "Library/Safari/Bookmarks.plist"), nil
	case Auto:
		return "", nil
	default:
		return "", ErrUnknownBrowser
	}
}

func getWindowsBookmarkPath(browserType Type) (string, error) {
	appData := os.Getenv("LOCALAPPDATA")
	if appData == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		appData = filepath.Join(home, "AppData/Local")
	}

	switch browserType {
	case Chrome:
		return filepath.Join(appData, `Google\Chrome\User Data\Default\Bookmarks`), nil
	case Chromium:
		return filepath.Join(appData, `Chromium\User Data\Default\Bookmarks`), nil
	case Edge:
		return filepath.Join(appData, `Microsoft\Edge\User Data\Default\Bookmarks`), nil
	case Brave:
		return filepath.Join(appData, `BraveSoftware\Brave-Browser\User Data\Default\Bookmarks`), nil
	case Vivaldi:
		return filepath.Join(appData, `Vivaldi\User Data\Default\Bookmarks`), nil
	case Firefox:
		return filepath.Join(appData, "Mozilla/Firefox"), nil
	case Safari:
		// Safari not available on Windows
		return "", ErrBrowserNotAvailable
	case Auto:
		return "", nil
	default:
		return "", ErrUnknownBrowser
	}
}

// GetSessionPath returns the session directory path for a given browser type on the current platform
// This is used for extracting open tabs from Chromium-based browsers
func GetSessionPath(browserType Type) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	switch runtime.GOOS {
	case "linux":
		return getLinuxSessionPath(home, browserType)
	case "darwin":
		return getDarwinSessionPath(home, browserType)
	case "windows":
		return getWindowsSessionPath(browserType)
	default:
		return "", ErrUnsupportedPlatform
	}
}

func getLinuxSessionPath(home string, browserType Type) (string, error) {
	switch browserType {
	case Chrome:
		return filepath.Join(home, ".config/google-chrome/Default/Sessions"), nil
	case Chromium:
		return filepath.Join(home, ".config/chromium/Default/Sessions"), nil
	case Edge:
		return filepath.Join(home, ".config/microsoft-edge/Default/Sessions"), nil
	case Brave:
		return filepath.Join(home, ".config/BraveSoftware/Brave-Browser/Default/Sessions"), nil
	case Vivaldi:
		return filepath.Join(home, ".config/vivaldi/Default/Sessions"), nil
	case Firefox, Safari:
		return "", ErrBrowserNotAvailable
	case Auto:
		return "", nil
	default:
		return "", ErrUnknownBrowser
	}
}

func getDarwinSessionPath(home string, browserType Type) (string, error) {
	switch browserType {
	case Chrome:
		return filepath.Join(home, "Library/Application Support/Google/Chrome/Default/Sessions"), nil
	case Chromium:
		return filepath.Join(home, "Library/Application Support/Chromium/Default/Sessions"), nil
	case Edge:
		return filepath.Join(home, "Library/Application Support/Microsoft Edge/Default/Sessions"), nil
	case Brave:
		return filepath.Join(home, "Library/Application Support/BraveSoftware/Brave-Browser/Default/Sessions"), nil
	case Vivaldi:
		return filepath.Join(home, "Library/Application Support/Vivaldi/Default/Sessions"), nil
	case Firefox, Safari:
		return "", ErrBrowserNotAvailable
	case Auto:
		return "", nil
	default:
		return "", ErrUnknownBrowser
	}
}

func getWindowsSessionPath(browserType Type) (string, error) {
	appData := os.Getenv("LOCALAPPDATA")
	if appData == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		appData = filepath.Join(home, "AppData/Local")
	}

	switch browserType {
	case Chrome:
		return filepath.Join(appData, `Google\Chrome\User Data\Default\Sessions`), nil
	case Chromium:
		return filepath.Join(appData, `Chromium\User Data\Default\Sessions`), nil
	case Edge:
		return filepath.Join(appData, `Microsoft\Edge\User Data\Default\Sessions`), nil
	case Brave:
		return filepath.Join(appData, `BraveSoftware\Brave-Browser\User Data\Default\Sessions`), nil
	case Vivaldi:
		return filepath.Join(appData, `Vivaldi\User Data\Default\Sessions`), nil
	case Firefox, Safari:
		return "", ErrBrowserNotAvailable
	case Auto:
		return "", nil
	default:
		return "", ErrUnknownBrowser
	}
}

// IsChromiumBased returns true if the browser uses Chromium's SNSS session format
func IsChromiumBased(browserType Type) bool {
	switch browserType {
	case Chrome, Chromium, Edge, Brave, Vivaldi:
		return true
	default:
		return false
	}
}
