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

func TestEventBus_On(t *testing.T) {
	bus := events.NewEventBus(newTestLogger(os.Stdout))
	done := make(chan events.Event)

	handler := func(ctx context.Context, event events.Event) {
		done <- event
	}

	// Register handler
	bus.On("test.event", handler)

	// Emit event
	expectedPayload := "test-payload"
	bus.Emit(context.Background(), "test.event", expectedPayload)

	// Wait for event with timeout
	select {
	case receivedEvent := <-done:
		assert.Equal(t, "test.event", receivedEvent.Signature)
		assert.Equal(t, expectedPayload, receivedEvent.Payload)
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for event handler")
	}
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bus := events.NewEventBus(newTestLogger(os.Stdout))
			done := make(chan struct{})

			bus.On(tt.pattern, func(ctx context.Context, event events.Event) {
				done <- struct{}{}
			})

			bus.Emit(context.Background(), tt.eventSignature, nil)

			select {
			case <-done:
				assert.True(t, tt.shouldMatch, "handler was called but shouldn't have been")
			case <-time.After(100 * time.Millisecond):
				assert.False(t, tt.shouldMatch, "handler wasn't called but should have been")
			}
		})
	}
}

func TestEventBus_MultipleHandlers(t *testing.T) {
	bus := events.NewEventBus(newTestLogger(os.Stdout))
	var wg sync.WaitGroup
	handlerCount := 3
	wg.Add(handlerCount)

	for i := 0; i < handlerCount; i++ {
		bus.On("test.event", func(ctx context.Context, event events.Event) {
			defer wg.Done()
		})
	}

	bus.Emit(context.Background(), "test.event", nil)

	// Use a channel to convert WaitGroup completion to select pattern
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// All handlers completed successfully
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for all handlers to complete")
	}
}

func TestEventBus_PanicRecovery(t *testing.T) {
	bus := events.NewEventBus(newTestLogger(os.Stdout))
	done := make(chan struct{})

	// First handler panics
	bus.On("test.event", func(ctx context.Context, event events.Event) {
		panic("test panic")
	})

	// Second handler should still run
	bus.On("test.event", func(ctx context.Context, event events.Event) {
		done <- struct{}{}
	})

	// Should not panic
	require.NotPanics(t, func() {
		bus.Emit(context.Background(), "test.event", nil)
	})

	select {
	case <-done:
		// Second handler completed successfully
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for second handler")
	}
}

func TestEventBus_ConcurrentEmit(t *testing.T) {
	bus := events.NewEventBus(newTestLogger(os.Stdout))
	eventCount := 100

	// Create a buffered channel to collect results
	results := make(chan string, eventCount)

	// Register handler that sends event IDs to results channel
	bus.On("test.event", func(ctx context.Context, event events.Event) {
		results <- event.ID
	})

	// Emit events concurrently
	var wg sync.WaitGroup
	for i := 0; i < eventCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			bus.Emit(context.Background(), "test.event", nil)
		}()
	}

	// Wait for all emits to complete
	wg.Wait()

	// Collect results with timeout
	receivedIDs := make(map[string]bool)
	timeout := time.After(time.Second)

	for i := 0; i < eventCount; i++ {
		select {
		case id := <-results:
			assert.False(t, receivedIDs[id], "received duplicate event ID: %s", id)
			receivedIDs[id] = true
		case <-timeout:
			t.Fatalf("timeout waiting for events, received only %d/%d", i, eventCount)
		}
	}
}

func TestEventBus_EmitSync(t *testing.T) {
	bus := events.NewEventBus(newTestLogger(os.Stdout))
	done := make(chan struct{})

	bus.On("test.event", func(ctx context.Context, event events.Event) {
		time.Sleep(50 * time.Millisecond) // Simulate work
		close(done)
	})

	bus.EmitSync(context.Background(), "test.event", nil)

	// Channel should already be closed since EmitSync is synchronous
	select {
	case <-done:
		// Handler completed as expected
	default:
		t.Fatal("EmitSync returned before handler completed")
	}
}

func TestEventBus_ContextCancellation(t *testing.T) {
	bus := events.NewEventBus(newTestLogger(os.Stdout))
	started := make(chan struct{})
	completed := make(chan struct{})

	bus.On("test.event", func(ctx context.Context, event events.Event) {
		close(started)
		<-ctx.Done()
		close(completed)
	})

	ctx, cancel := context.WithCancel(context.Background())
	// Use Emit instead of EmitSync so we don't block
	bus.Emit(ctx, "test.event", nil)

	// Wait for handler to start
	select {
	case <-started:
		// Handler started
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for handler to start")
	}

	// Now we can cancel the context
	cancel()

	// Wait for handler to complete
	select {
	case <-completed:
		// Handler completed
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for handler to complete after cancellation")
	}
}
