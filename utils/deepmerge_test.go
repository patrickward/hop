package utils_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/patrickward/hop/v2/utils"
)

func TestDeepMerge(t *testing.T) {
	tests := []struct {
		name string
		dst  map[string]any
		src  map[string]any
		want map[string]any
	}{
		{
			name: "merge into empty map",
			dst:  map[string]any{},
			src: map[string]any{
				"key1": "value1",
			},
			want: map[string]any{
				"key1": "value1",
			},
		},
		{
			name: "merge nested maps",
			dst: map[string]any{
				"nested": map[string]any{
					"a": 1,
				},
			},
			src: map[string]any{
				"nested": map[string]any{
					"b": 2,
				},
			},
			want: map[string]any{
				"nested": map[string]any{
					"a": 1,
					"b": 2,
				},
			},
		},
		{
			name: "merge with nil nested map",
			dst: map[string]any{
				"nested": nil,
			},
			src: map[string]any{
				"nested": map[string]any{
					"key": "value",
				},
			},
			want: map[string]any{
				"nested": map[string]any{
					"key": "value",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			utils.DeepMerge(&tt.dst, tt.src)
			assert.Equal(t, tt.want, tt.dst)
		})
	}
}
