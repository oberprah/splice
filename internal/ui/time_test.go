package ui

import (
	"testing"
	"time"
)

func TestToRelativeTimeFrom(t *testing.T) {
	// Fixed "now" time for consistent testing
	now := time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name     string
		input    time.Time
		expected string
	}{
		// Seconds
		{"just now", now.Add(-30 * time.Second), "just now"},
		{"59 seconds", now.Add(-59 * time.Second), "just now"},

		// Minutes
		{"1 minute", now.Add(-1 * time.Minute), "1 min ago"},
		{"2 minutes", now.Add(-2 * time.Minute), "2 mins ago"},
		{"30 minutes", now.Add(-30 * time.Minute), "30 mins ago"},
		{"59 minutes", now.Add(-59 * time.Minute), "59 mins ago"},

		// Hours
		{"1 hour", now.Add(-1 * time.Hour), "1 hour ago"},
		{"2 hours", now.Add(-2 * time.Hour), "2 hours ago"},
		{"12 hours", now.Add(-12 * time.Hour), "12 hours ago"},
		{"23 hours", now.Add(-23 * time.Hour), "23 hours ago"},

		// Days
		{"1 day", now.Add(-24 * time.Hour), "1 day ago"},
		{"2 days", now.Add(-48 * time.Hour), "2 days ago"},
		{"6 days", now.Add(-6 * 24 * time.Hour), "6 days ago"},

		// Weeks
		{"1 week", now.Add(-7 * 24 * time.Hour), "1 week ago"},
		{"2 weeks", now.Add(-14 * 24 * time.Hour), "2 weeks ago"},
		{"3 weeks", now.Add(-21 * 24 * time.Hour), "3 weeks ago"},
		{"4 weeks", now.Add(-28 * 24 * time.Hour), "4 weeks ago"},

		// Months
		{"1 month", now.Add(-30 * 24 * time.Hour), "1 month ago"},
		{"2 months", now.Add(-60 * 24 * time.Hour), "2 months ago"},
		{"11 months", now.Add(-330 * 24 * time.Hour), "11 months ago"},

		// Years (showing absolute date)
		{"1 year", now.Add(-365 * 24 * time.Hour), "1 year ago"},
		{"2 years", time.Date(2022, 1, 15, 12, 0, 0, 0, time.UTC), "Jan 15, 2022"},
		{"10 years", time.Date(2014, 6, 20, 8, 30, 0, 0, time.UTC), "Jun 20, 2014"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ToRelativeTimeFrom(tt.input, now)
			if result != tt.expected {
				t.Errorf("ToRelativeTimeFrom(%v, %v) = %q, want %q",
					tt.input, now, result, tt.expected)
			}
		})
	}
}
