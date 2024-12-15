package time_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	hopTime "github.com/patrickward/hop/templates/funcmap/time"
)

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

func TestApproximateDuration(t *testing.T) {
	tests := []struct {
		input    time.Duration
		expected string
	}{
		{time.Second, "1 second"},
		{2 * time.Second, "2 seconds"},
		{time.Minute, "1 minute"},
		{2 * time.Minute, "2 minutes"},
		{time.Hour, "1 hour"},
		{2 * time.Hour, "2 hours"},
		{24 * time.Hour, "1 day"},
		{48 * time.Hour, "2 days"},
		{365 * 24 * time.Hour, "1 year"},
		{2 * 365 * 24 * time.Hour, "2 years"},
	}

	for _, tt := range tests {
		result := hopTime.FuncMap()["time_duration"].(func(time.Duration) string)(tt.input)
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
