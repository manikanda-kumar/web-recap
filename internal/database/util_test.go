package database

import (
	"testing"
	"time"
)

func TestConvertChromeTimestamp(t *testing.T) {
	tests := []struct {
		name       string
		chromeVal  int64
		expectZero bool
	}{
		{
			name:       "Zero timestamp",
			chromeVal:  0,
			expectZero: true,
		},
		{
			name:       "Valid timestamp",
			chromeVal:  13289816330000000, // Some arbitrary timestamp
			expectZero: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ConvertChromeTimestamp(tt.chromeVal)

			if tt.expectZero && !result.IsZero() {
				t.Errorf("expected zero time, got %v", result)
			}

			if !tt.expectZero && result.IsZero() {
				t.Errorf("expected non-zero time, got zero")
			}
		})
	}
}

func TestConvertChromeTimestampPreservesMicroseconds(t *testing.T) {
	// 13289816330000001 => 1 microsecond past an exact second.
	result := ConvertChromeTimestamp(13289816330000001)
	if result.Nanosecond() != 1000 {
		t.Fatalf("expected 1000ns remainder, got %dns", result.Nanosecond())
	}
}

func TestConvertFirefoxTimestamp(t *testing.T) {
	tests := []struct {
		name       string
		firefoxVal int64
		expectZero bool
	}{
		{
			name:       "Zero timestamp",
			firefoxVal: 0,
			expectZero: true,
		},
		{
			name:       "Valid timestamp",
			firefoxVal: 1702742400000000, // December 16, 2023
			expectZero: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ConvertFirefoxTimestamp(tt.firefoxVal)

			if tt.expectZero && !result.IsZero() {
				t.Errorf("expected zero time, got %v", result)
			}

			if !tt.expectZero && result.IsZero() {
				t.Errorf("expected non-zero time, got zero")
			}
		})
	}
}

func TestConvertSafariTimestamp(t *testing.T) {
	tests := []struct {
		name            string
		safariVal       int64
		expectZero      bool
		minExpectedYear int
	}{
		{
			name:            "Zero timestamp (Safari epoch)",
			safariVal:       0,
			expectZero:      false, // 0 = 2001-01-01
			minExpectedYear: 2001,
		},
		{
			name:            "Valid timestamp",
			safariVal:       730000000, // Some arbitrary seconds since 2001
			expectZero:      false,
			minExpectedYear: 2024, // 730M seconds since 2001 is ~2024
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ConvertSafariTimestamp(tt.safariVal)

			if tt.expectZero && !result.IsZero() {
				t.Errorf("expected zero time, got %v", result)
			}

			if !tt.expectZero && result.IsZero() {
				t.Errorf("expected non-zero time, got zero")
			}

			// Verify result is a valid time (year >= min expected)
			if result.Year() < tt.minExpectedYear {
				t.Errorf("expected year >= %d, got %d", tt.minExpectedYear, result.Year())
			}
		})
	}
}

func TestExtractDomain(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		expected string
	}{
		{
			name:     "Valid HTTPS URL",
			url:      "https://example.com/page",
			expected: "example.com",
		},
		{
			name:     "Valid HTTP URL",
			url:      "http://www.google.com/search",
			expected: "www.google.com",
		},
		{
			name:     "URL with port",
			url:      "https://localhost:8080/app",
			expected: "localhost:8080",
		},
		{
			name:     "Empty URL",
			url:      "",
			expected: "",
		},
		{
			name:     "Subdomain",
			url:      "https://api.github.com/repos",
			expected: "api.github.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExtractDomain(tt.url)

			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestFilterByDateRange(t *testing.T) {
	startDate := time.Date(2025, 12, 15, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2025, 12, 16, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name         string
		startDate    time.Time
		endDate      time.Time
		inputLen     int
		minOutputLen int
	}{
		{
			name:         "No date filter",
			startDate:    time.Time{},
			endDate:      time.Time{},
			inputLen:     5,
			minOutputLen: 5,
		},
		{
			name:         "With date range",
			startDate:    startDate,
			endDate:      endDate,
			inputLen:     5,
			minOutputLen: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create dummy entries
			entries := make([]interface{}, tt.inputLen)
			for i := range entries {
				entries[i] = nil
			}

			result := FilterByDateRange(entries, tt.startDate, tt.endDate)

			if len(result) < tt.minOutputLen {
				t.Errorf("expected at least %d entries, got %d", tt.minOutputLen, len(result))
			}
		})
	}
}

func TestWithinHalfOpenRange(t *testing.T) {
	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	end := start.Add(24 * time.Hour)

	tests := []struct {
		name  string
		value time.Time
		ok    bool
	}{
		{name: "inside range", value: start.Add(12 * time.Hour), ok: true},
		{name: "equal to start", value: start, ok: true},
		{name: "equal to end excluded", value: end, ok: false},
		{name: "before start", value: start.Add(-time.Second), ok: false},
		{name: "zero time", value: time.Time{}, ok: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := WithinHalfOpenRange(tt.value, start, end); got != tt.ok {
				t.Fatalf("WithinHalfOpenRange() = %v, want %v", got, tt.ok)
			}
		})
	}
}
