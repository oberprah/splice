package format

import (
	"fmt"
	"time"
)

// ToRelativeTimeFrom converts a timestamp into a human-readable relative time
// relative to the provided "now" time. This is useful for testing.
func ToRelativeTimeFrom(t, now time.Time) string {
	diff := now.Sub(t)

	seconds := int(diff.Seconds())
	minutes := seconds / 60
	hours := minutes / 60
	days := hours / 24
	weeks := days / 7
	months := days / 30
	years := days / 365

	switch {
	case seconds < 60:
		return "just now"
	case minutes < 60:
		if minutes == 1 {
			return "1 min ago"
		}
		return fmt.Sprintf("%d mins ago", minutes)
	case hours < 24:
		if hours == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", hours)
	case days < 7:
		if days == 1 {
			return "1 day ago"
		}
		return fmt.Sprintf("%d days ago", days)
	case days < 30:
		if weeks == 1 {
			return "1 week ago"
		}
		return fmt.Sprintf("%d weeks ago", weeks)
	case days < 365:
		if months == 1 {
			return "1 month ago"
		}
		return fmt.Sprintf("%d months ago", months)
	default:
		if years == 1 {
			return "1 year ago"
		}
		// For old commits, show absolute date
		return t.Format("Jan 2, 2006")
	}
}
