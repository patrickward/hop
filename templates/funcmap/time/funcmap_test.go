package time_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	hopTime "github.com/patrickward/hop/templates/funcmap/time"
)

func TestTimeAgo(t *testing.T) {
	now := time.Now()
	tests := []struct {
		input    time.Time
		expected string
	}{
		{now.Add(-time.Second), "just now"},
		{now.Add(-time.Second * 3), "just now"},
		{now.Add(-time.Second * 10), "10 seconds ago"},
		{now.Add(-time.Second * 59), "59 seconds ago"},
		{now.Add(-time.Minute), "1 minute ago"},
		{now.Add(-time.Minute * 2), "2 minutes ago"},
		{now.Add(-time.Hour), "1 hour ago"},
		{now.Add(-time.Hour * 2), "2 hours ago"},
		{now.Add(-time.Hour * 24), "1 day ago"},
		{now.Add(-time.Hour * 24 * 2), "2 days ago"},
		{now.Add(-time.Hour * 24 * 7), "7 days ago"},
		{now.Add(-time.Hour * 24 * 7 * 2), "14 days ago"},
		{now.Add(-time.Hour * 24 * 30), "1 month ago"},
		{now.Add(-time.Hour * 24 * 30 * 2), "2 months ago"},
		{now.Add(-time.Hour * 24 * 365), "1 year ago"},
		{now.Add(-time.Hour * 24 * 365 * 2), "2 years ago"},
		{now.Add(time.Second), "in the future"},
		{now, "just now"},
		{time.Time{}, ""},
	}

	for _, tt := range tests {
		result := hopTime.FuncMap()["time_ago"].(func(time.Time) string)(tt.input)
		assert.Equal(t, tt.expected, result)
	}
}

func TestFormatTime(t *testing.T) {
	tests := []struct {
		format   string
		input    time.Time
		expected string
	}{
		{"2006-01-02", time.Date(2023, 10, 1, 0, 0, 0, 0, time.UTC), "2023-10-01"},
		{"15:04:05", time.Date(2023, 10, 1, 14, 30, 0, 0, time.UTC), "14:30:00"},
	}

	for _, tt := range tests {
		result := hopTime.FuncMap()["time_format"].(func(string, time.Time) string)(tt.format, tt.input)
		assert.Equal(t, tt.expected, result)
	}
}

func TestIsToday(t *testing.T) {
	now := time.Now()
	tests := []struct {
		input    time.Time
		expected bool
	}{
		{now, true},
		{now.AddDate(0, 0, -1), false},
	}

	for _, tt := range tests {
		result := hopTime.FuncMap()["time_isToday"].(func(time.Time) bool)(tt.input)
		assert.Equal(t, tt.expected, result)
	}
}

func TestTimeNow(t *testing.T) {
	result := hopTime.FuncMap()["time_now"].(func() time.Time)()
	assert.WithinDuration(t, time.Now(), result, time.Second)
}

func TestTimeSince(t *testing.T) {
	now := time.Now()
	start := now.Add(-time.Hour)
	result := hopTime.FuncMap()["time_since"].(func(time.Time) time.Duration)(start)
	// result should be a duration of 1 hour
	assert.WithinDuration(t, now, now.Add(-result), time.Hour+time.Second)
}

func TestTimeUntil(t *testing.T) {
	now := time.Now()
	future := now.Add(time.Hour)
	result := hopTime.FuncMap()["time_until"].(func(time.Time) time.Duration)(future)
	assert.WithinDuration(t, now, now.Add(result), time.Hour+time.Second)
}
