package time

import (
	"fmt"
	"html/template"
	"math"
	"time"
)

const (
	secondsInDay  = 24 * time.Hour     // Number of seconds in a day
	secondsInYear = 365 * secondsInDay // Number of seconds in a year
)

// FuncMap returns a function map with functions for working with time.Time values.
func FuncMap() template.FuncMap {
	return template.FuncMap{
		"time_duration": approximateDuration,
		"time_format":   formatTime,
		"time_isToday":  isToday,
		"time_now":      time.Now,
		"time_since":    time.Since,
		"time_until":    time.Until,
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

// approximateDuration returns a human-readable approximation of a duration
func approximateDuration(d time.Duration) string {
	if d < time.Second {
		return "less than 1 second"
	}

	ds := int(math.Round(d.Seconds()))
	if ds == 1 {
		return "1 second"
	} else if ds < 60 {
		return fmt.Sprintf("%d seconds", ds)
	}

	dm := int(math.Round(d.Minutes()))
	if dm == 1 {
		return "1 minute"
	} else if dm < 60 {
		return fmt.Sprintf("%d minutes", dm)
	}

	dh := int(math.Round(d.Hours()))
	if dh == 1 {
		return "1 hour"
	} else if dh < 24 {
		return fmt.Sprintf("%d hours", dh)
	}

	dd := int(math.Round(float64(d / secondsInDay)))
	if dd == 1 {
		return "1 day"
	} else if dd < 365 {
		return fmt.Sprintf("%d days", dd)
	}

	dy := int(math.Round(float64(d / secondsInYear)))
	if dy == 1 {
		return "1 year"
	}

	return fmt.Sprintf("%d years", dy)
}
