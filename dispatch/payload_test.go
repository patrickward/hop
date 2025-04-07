package dispatch_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/patrickward/hop/v2/dispatch"
)

type testUser struct {
	ID   string
	Name string
}

func TestPayloadAs(t *testing.T) {
	tests := []struct {
		name        string
		payload     any
		wantErr     bool
		errContains string
	}{
		{
			name: "valid payload",
			payload: testUser{
				ID:   "123",
				Name: "John",
			},
			wantErr: false,
		},
		{
			name:        "nil payload",
			payload:     nil,
			wantErr:     true,
			errContains: "payload is nil",
		},
		{
			name:        "wrong type",
			payload:     "not a user",
			wantErr:     true,
			errContains: "invalid payload type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			evt := dispatch.NewEvent("test", tt.payload)

			result, err := dispatch.PayloadAs[testUser](evt)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.payload, result)
		})
	}
}

func TestMustPayloadAs(t *testing.T) {
	t.Run("valid payload", func(t *testing.T) {
		user := testUser{ID: "123", Name: "John"}
		evt := dispatch.NewEvent("test", user)

		assert.NotPanics(t, func() {
			result := dispatch.MustPayloadAs[testUser](evt)
			assert.Equal(t, user, result)
		})
	})

	t.Run("invalid payload panics", func(t *testing.T) {
		evt := dispatch.NewEvent("test", "not a user")

		assert.Panics(t, func() {
			_ = dispatch.MustPayloadAs[testUser](evt)
		})
	})
}

func TestHandlePayload(t *testing.T) {
	t.Run("successful handling", func(t *testing.T) {
		var handled testUser
		handler := dispatch.HandlePayload[testUser](func(ctx context.Context, user testUser) {
			handled = user
		})

		user := testUser{ID: "123", Name: "John"}
		evt := dispatch.NewEvent("test", user)

		handler(context.Background(), evt)
		assert.Equal(t, user, handled)
	})

	t.Run("wrong type not handled", func(t *testing.T) {
		var handled bool
		handler := dispatch.HandlePayload[testUser](func(ctx context.Context, user testUser) {
			handled = true
		})

		evt := dispatch.NewEvent("test", "not a user")

		handler(context.Background(), evt)
		assert.False(t, handled)
	})
}

func TestIsPayloadType(t *testing.T) {
	t.Run("correct type returns true", func(t *testing.T) {
		evt := dispatch.NewEvent("test", testUser{ID: "123", Name: "John"})
		assert.True(t, dispatch.IsPayloadType[testUser](evt))
	})

	t.Run("wrong type returns false", func(t *testing.T) {
		evt := dispatch.NewEvent("test", "not a user")
		assert.False(t, dispatch.IsPayloadType[testUser](evt))
	})

	t.Run("nil payload returns false", func(t *testing.T) {
		evt := dispatch.NewEvent("test", nil)
		assert.False(t, dispatch.IsPayloadType[testUser](evt))
	})
}

func TestPayloadAsMap(t *testing.T) {
	t.Run("valid map payload", func(t *testing.T) {
		payload := map[string]any{"key": "value"}
		evt := dispatch.NewEvent("test", payload)

		result, err := dispatch.PayloadAsMap(evt)
		require.NoError(t, err)
		assert.Equal(t, payload, result)
	})

	t.Run("non-map payload", func(t *testing.T) {
		evt := dispatch.NewEvent("test", "not a map")

		_, err := dispatch.PayloadAsMap(evt)
		require.Error(t, err)
	})
}

func TestPayloadAsSlice(t *testing.T) {
	t.Run("valid slice payload", func(t *testing.T) {
		payload := []any{"one", "two", "three"}
		evt := dispatch.NewEvent("test", payload)

		result, err := dispatch.PayloadAsSlice(evt)
		require.NoError(t, err)
		assert.Equal(t, payload, result)
	})

	t.Run("non-slice payload", func(t *testing.T) {
		evt := dispatch.NewEvent("test", "not a slice")

		_, err := dispatch.PayloadAsSlice(evt)
		require.Error(t, err)
	})
}

type TestItem struct {
	ID   string
	Name string
}

func TestPayloadSliceAs(t *testing.T) {
	tests := []struct {
		name        string
		payload     any
		wantItems   []TestItem
		wantErr     bool
		errContains string
	}{
		{
			name: "valid slice",
			payload: []any{
				TestItem{ID: "1", Name: "Item 1"},
				TestItem{ID: "2", Name: "Item 2"},
			},
			wantItems: []TestItem{
				{ID: "1", Name: "Item 1"},
				{ID: "2", Name: "Item 2"},
			},
			wantErr: false,
		},
		{
			name: "slice with nil elements",
			payload: []any{
				TestItem{ID: "1", Name: "Item 1"},
				nil,
				TestItem{ID: "3", Name: "Item 3"},
			},
			wantItems: []TestItem{
				{ID: "1", Name: "Item 1"},
				{ID: "3", Name: "Item 3"},
			},
			wantErr: false,
		},
		{
			name:        "not a slice",
			payload:     "not a slice",
			wantErr:     true,
			errContains: "payload is not a slice",
		},
		{
			name: "slice with wrong type",
			payload: []any{
				TestItem{ID: "1", Name: "Item 1"},
				"not a test item",
			},
			wantErr:     true,
			errContains: "invalid type at index 1",
		},
		{
			name:        "nil payload",
			payload:     nil,
			wantErr:     true,
			errContains: "payload is nil",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			evt := dispatch.NewEvent("test.slice", tt.payload)

			items, err := dispatch.PayloadSliceAs[TestItem](evt)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.wantItems, items)
		})
	}
}

func TestPayloadMapAs(t *testing.T) {
	tests := []struct {
		name        string
		payload     any
		wantMap     map[string]TestItem
		wantErr     bool
		errContains string
	}{
		{
			name: "valid map",
			payload: map[string]any{
				"item1": TestItem{ID: "1", Name: "Item 1"},
				"item2": TestItem{ID: "2", Name: "Item 2"},
			},
			wantMap: map[string]TestItem{
				"item1": {ID: "1", Name: "Item 1"},
				"item2": {ID: "2", Name: "Item 2"},
			},
			wantErr: false,
		},
		{
			name: "map with nil values",
			payload: map[string]any{
				"item1": TestItem{ID: "1", Name: "Item 1"},
				"nil":   nil,
				"item3": TestItem{ID: "3", Name: "Item 3"},
			},
			wantMap: map[string]TestItem{
				"item1": {ID: "1", Name: "Item 1"},
				"item3": {ID: "3", Name: "Item 3"},
			},
			wantErr: false,
		},
		{
			name:        "not a map",
			payload:     "not a map",
			wantErr:     true,
			errContains: "payload is not a map",
		},
		{
			name: "map with wrong type",
			payload: map[string]any{
				"item1": TestItem{ID: "1", Name: "Item 1"},
				"bad":   "not a test item",
			},
			wantErr:     true,
			errContains: `invalid type for key "bad"`,
		},
		{
			name:        "nil payload",
			payload:     nil,
			wantErr:     true,
			errContains: "payload is nil",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			evt := dispatch.NewEvent("test.map", tt.payload)

			items, err := dispatch.PayloadMapAs[TestItem](evt)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.wantMap, items)
		})
	}
}
