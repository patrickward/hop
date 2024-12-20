package time

import (
	"fmt"
	"html/template"
	"time"
)

// FuncMap returns a function map with functions for working with time.Time values.
func FuncMap() template.FuncMap {
	return template.FuncMap{
		"time_ago":     timeAgo,
		"time_format":  formatTime,
		"time_isToday": isToday,
		"time_now":     time.Now,
		"time_since":   time.Since,
		"time_until":   time.Until,
	}
}

// formatTime formats a time.Time value as a string
func formatTime(format string, t time.Time) string {
	return t.Format(format)
}

// isToday checks if the given time is today
func isToday(t time.Time) bool {
	now := time.Now()
	return t.Year() == now.Year() && t.YearDay() == now.YearDay()
}

// time_ago returns a human-readable string representing the time elapsed since the given time
func timeAgo(t time.Time) string {
	if t.IsZero() {
		return ""
	}

	duration := time.Since(t)

	// Handle future times
	if duration < 0 {
		return "in the future"
	}

	// Very recent
	if duration < time.Minute {
		secs := int(duration.Seconds())
		if secs <= 3 {
			return "just now"
		}
		return fmt.Sprintf("%d seconds ago", secs)
	}

	// Minutes
	if duration < time.Hour {
		mins := int(duration.Minutes())
		if mins == 1 {
			return "1 minute ago"
		}
		return fmt.Sprintf("%d minutes ago", mins)
	}

	// Hours
	if duration < 24*time.Hour {
		hours := int(duration.Hours())
		if hours == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", hours)
	}

	// Days
	if duration < 30*24*time.Hour {
		days := int(duration.Hours() / 24)
		if days == 1 {
			return "1 day ago"
		}
		return fmt.Sprintf("%d days ago", days)
	}

	// Months
	if duration < 365*24*time.Hour {
		months := int(duration.Hours() / 24 / 30)
		if months == 1 {
			return "1 month ago"
		}
		return fmt.Sprintf("%d months ago", months)
	}

	// Years
	years := int(duration.Hours() / 24 / 365)
	if years == 1 {
		return "1 year ago"
	}
	return fmt.Sprintf("%d years ago", years)
}
