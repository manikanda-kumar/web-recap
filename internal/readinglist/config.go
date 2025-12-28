package readinglist

import (
	"fmt"
	"os"
	"time"
)

// LoadConfigFromEnv loads reading list configuration from environment variables
func LoadConfigFromEnv(platform PlatformType) (*Config, error) {
	config := &Config{
		Platform:       string(platform),
		RateLimitDelay: 2 * time.Second,
	}

	switch platform {
	case PlatformMedium:
		config.SessionToken = os.Getenv("MEDIUM_SESSION_TOKEN")
		config.Cookie = os.Getenv("MEDIUM_COOKIE")
		config.Username = os.Getenv("MEDIUM_USERNAME")
		config.FilePath = os.Getenv("MEDIUM_FILE_PATH")

	case PlatformSubstack:
		config.SessionToken = os.Getenv("SUBSTACK_SESSION_TOKEN")
		config.Cookie = os.Getenv("SUBSTACK_COOKIE")
		config.FilePath = os.Getenv("SUBSTACK_FILE_PATH")

	default:
		return nil, fmt.Errorf("%w: %s", ErrUnsupportedPlatform, platform)
	}

	return config, nil
}

// LoadConfigFromFlags loads configuration from CLI flags
func LoadConfigFromFlags(platform PlatformType, sessionToken, cookie, username, filePath, publicURL string) *Config {
	config := &Config{
		Platform:       string(platform),
		SessionToken:   sessionToken,
		Cookie:         cookie,
		Username:       username,
		FilePath:       filePath,
		PublicURL:      publicURL,
		RateLimitDelay: 2 * time.Second,
	}

	return config
}

// MergeConfigs merges flag-based config with env-based config
// Flags take precedence over environment variables
func MergeConfigs(flagConfig, envConfig *Config) *Config {
	merged := &Config{
		Platform:       flagConfig.Platform,
		RateLimitDelay: flagConfig.RateLimitDelay,
	}

	// SessionToken: flag > env
	if flagConfig.SessionToken != "" {
		merged.SessionToken = flagConfig.SessionToken
	} else {
		merged.SessionToken = envConfig.SessionToken
	}

	// Cookie: flag > env
	if flagConfig.Cookie != "" {
		merged.Cookie = flagConfig.Cookie
	} else {
		merged.Cookie = envConfig.Cookie
	}

	// Username: flag > env
	if flagConfig.Username != "" {
		merged.Username = flagConfig.Username
	} else {
		merged.Username = envConfig.Username
	}

	// FilePath: flag > env
	if flagConfig.FilePath != "" {
		merged.FilePath = flagConfig.FilePath
	} else {
		merged.FilePath = envConfig.FilePath
	}

	// PublicURL: flag only (no env var needed)
	merged.PublicURL = flagConfig.PublicURL

	return merged
}
