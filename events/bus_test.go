package events_test

import (
	"context"
	"io"
	"log/slog"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/patrickward/hop/events"
)

func newTestLogger(out io.Writer) *slog.Logger {
	return slog.New(slog.NewTextHandler(out, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
}

func TestNewEvent(t *testing.T) {
	tests := []struct {
		name      string
		signature string
		payload   any
	}{
		{
			name:      "basic event",
			signature: "test.basic",
			payload:   nil,
		},
		{
			name:      "event with string payload",
			signature: "test.string.payload",
			payload:   "hello",
		},
		{
			name:      "event with struct payload",
			signature: "test.struct.payload",
			payload:   struct{ Name string }{"test"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			event := events.NewEvent(tt.signature, tt.payload)

			assert.NotEmpty(t, event.ID)
			assert.Equal(t, tt.signature, event.Signature)
			assert.Equal(t, tt.payload, event.Payload)
			assert.False(t, event.Timestamp.IsZero())
		})
	}
}

func TestEventBus_On(t *testing.T) {
	bus := events.NewEventBus(newTestLogger(os.Stdout))
	var handlerCalled bool
	var receivedEvent events.Event

	handler := func(ctx context.Context, event events.Event) {
		handlerCalled = true
		receivedEvent = event
	}

	// Register handler
	bus.On("test.event", handler)

	// Emit event
	bus.Emit(context.Background(), "test.event", "test-payload")

	// Allow async handlers to complete
	time.Sleep(50 * time.Millisecond)

	assert.True(t, handlerCalled)
	assert.NotNil(t, receivedEvent)
	assert.Equal(t, "test.event", receivedEvent.Signature)
	assert.Equal(t, "test-payload", receivedEvent.Payload)
}

func TestEventBus_Wildcards(t *testing.T) {
	tests := []struct {
		name           string
		pattern        string
		eventSignature string
		shouldMatch    bool
	}{
		{
			name:           "exact match",
			pattern:        "test.event",
			eventSignature: "test.event",
			shouldMatch:    true,
		},
		{
			name:           "wildcard prefix",
			pattern:        "*.event",
			eventSignature: "test.event",
			shouldMatch:    true,
		},
		{
			name:           "wildcard suffix",
			pattern:        "test.*",
			eventSignature: "test.event",
			shouldMatch:    true,
		},
		{
			name:           "multiple wildcards",
			pattern:        "*.system.*",
			eventSignature: "test.system.start",
			shouldMatch:    true,
		},
		{
			name:           "no match",
			pattern:        "other.event",
			eventSignature: "test.event",
			shouldMatch:    false,
		},
		{
			name:           "different segments",
			pattern:        "test.*",
			eventSignature: "test.system.start",
			shouldMatch:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bus := events.NewEventBus(newTestLogger(os.Stdout))
			var handlerCalled bool

			bus.On(tt.pattern, func(ctx context.Context, event events.Event) {
				handlerCalled = true
			})

			bus.Emit(context.Background(), tt.eventSignature, nil)
			time.Sleep(50 * time.Millisecond)

			assert.Equal(t, tt.shouldMatch, handlerCalled)
		})
	}
}

func TestEventBus_MultipleHandlers(t *testing.T) {
	bus := events.NewEventBus(newTestLogger(os.Stdout))
	var mu sync.Mutex
	handlerCalls := make(map[string]bool)

	handlers := []string{"handler1", "handler2", "handler3"}
	for _, name := range handlers {
		handlerName := name // Capture for closure
		bus.On("test.event", func(ctx context.Context, event events.Event) {
			mu.Lock()
			handlerCalls[handlerName] = true
			mu.Unlock()
		})
	}

	bus.Emit(context.Background(), "test.event", nil)
	time.Sleep(50 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()
	for _, name := range handlers {
		assert.True(t, handlerCalls[name], "handler %s should have been called", name)
	}
}

func TestEventBus_EmitSync(t *testing.T) {
	bus := events.NewEventBus(newTestLogger(os.Stdout))
	var handlerCalled bool

	bus.On("test.event", func(ctx context.Context, event events.Event) {
		time.Sleep(10 * time.Millisecond) // Simulate work
		handlerCalled = true
	})

	// EmitSync should wait for handler to complete
	bus.EmitSync(context.Background(), "test.event", nil)

	// No need to wait since EmitSync is synchronous
	assert.True(t, handlerCalled)
}

func TestEventBus_PanicRecovery(t *testing.T) {
	bus := events.NewEventBus(newTestLogger(os.Stdout))
	var secondHandlerCalled bool

	// First handler panics
	bus.On("test.event", func(ctx context.Context, event events.Event) {
		panic("test panic")
	})

	// Second handler should still run
	bus.On("test.event", func(ctx context.Context, event events.Event) {
		secondHandlerCalled = true
	})

	// Should not panic
	require.NotPanics(t, func() {
		bus.Emit(context.Background(), "test.event", nil)
	})

	time.Sleep(50 * time.Millisecond)
	assert.True(t, secondHandlerCalled)
}

func TestEventBus_ConcurrentEmit(t *testing.T) {
	bus := events.NewEventBus(newTestLogger(os.Stdout))
	var mu sync.Mutex
	handlerCalls := make(map[string]int)

	bus.On("test.event", func(ctx context.Context, event events.Event) {
		mu.Lock()
		handlerCalls[event.ID]++
		mu.Unlock()
	})

	// Emit events concurrently
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			bus.Emit(context.Background(), "test.event", nil)
		}()
	}

	wg.Wait()
	time.Sleep(50 * time.Millisecond)

	mu.Lock()
	assert.Equal(t, 100, len(handlerCalls))
	for _, count := range handlerCalls {
		assert.Equal(t, 1, count)
	}
	mu.Unlock()
}

func TestEventBus_ContextCancellation(t *testing.T) {
	bus := events.NewEventBus(newTestLogger(os.Stdout))
	ctx, cancel := context.WithCancel(context.Background())
	var handlerStarted, handlerCompleted bool

	bus.On("test.event", func(ctx context.Context, event events.Event) {
		handlerStarted = true
		<-ctx.Done() // Wait for cancellation
		handlerCompleted = true
	})

	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()

	bus.EmitSync(ctx, "test.event", nil)

	assert.True(t, handlerStarted)
	assert.True(t, handlerCompleted)
}

func BenchmarkEventEmit(b *testing.B) {
	bus := events.NewEventBus(newTestLogger(os.Stdout))
	bus.On("bench.event", func(ctx context.Context, event events.Event) {})

	ctx := context.Background()
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			bus.Emit(ctx, "bench.event", nil)
		}
	})
}
