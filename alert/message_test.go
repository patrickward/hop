package alert

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestMessages_ByType(t *testing.T) {
	tests := []struct {
		name     string
		messages Messages
		filter   Type
		expected Messages
	}{
		{
			name: "no messages of type",
			messages: Messages{
				{Type: "info", Content: "Information message"},
				{Type: "warning", Content: "Warning message"},
			},
			filter:   "error",
			expected: nil,
		},
		{
			name: "some messages of type",
			messages: Messages{
				{Type: "info", Content: "Information message"},
				{Type: "error", Content: "Error 1"},
				{Type: "warning", Content: "Warning message"},
				{Type: "error", Content: "Error 2"},
			},
			filter: "error",
			expected: Messages{
				{Type: "error", Content: "Error 1"},
				{Type: "error", Content: "Error 2"},
			},
		},
		{
			name: "all messages of type",
			messages: Messages{
				{Type: "info", Content: "Message 1"},
				{Type: "info", Content: "Message 2"},
			},
			filter: "info",
			expected: Messages{
				{Type: "info", Content: "Message 1"},
				{Type: "info", Content: "Message 2"},
			},
		},
		{
			name:     "empty messages",
			messages: Messages{},
			filter:   "info",
			expected: nil,
		},
		{
			name: "mixed case types",
			messages: Messages{
				{Type: "Info", Content: "Message 1"},
				{Type: "info", Content: "Message 2"},
			},
			filter: "info",
			expected: Messages{
				{Type: "info", Content: "Message 2"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.messages.ByType(tt.filter)
			if !cmp.Equal(result, tt.expected) {
				t.Errorf("ByType() = %v, expected %v", result, tt.expected)
				t.Error(cmp.Diff(result, tt.expected))
			}
		})
	}
}
