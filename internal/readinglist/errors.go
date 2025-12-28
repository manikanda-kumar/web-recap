package readinglist

import "errors"

var (
	// ErrUnsupportedPlatform is returned when the platform is not supported
	ErrUnsupportedPlatform = errors.New("unsupported platform")

	// ErrNoAuthProvided is returned when no authentication credentials are provided
	ErrNoAuthProvided = errors.New("no authentication provided")

	// ErrAuthenticationFailed is returned when authentication fails
	ErrAuthenticationFailed = errors.New("authentication failed")

	// ErrRateLimited is returned when the platform rate limits requests
	ErrRateLimited = errors.New("rate limited by platform")

	// ErrNoStrategiesAvailable is returned when no fetching strategies are available
	ErrNoStrategiesAvailable = errors.New("no fetching strategies available")

	// ErrInvalidFileFormat is returned when the file format is invalid
	ErrInvalidFileFormat = errors.New("invalid file format")
)
